package service

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerPool 工作池接口
type WorkerPool interface {
	Submit(task func()) error
	SubmitWithContext(ctx context.Context, task func(ctx context.Context)) error
	Close()
	Wait()
	GetStats() PoolStats
}

// PoolStats 工作池统计信息
type PoolStats struct {
	WorkerCount    int32
	QueueSize      int32
	ActiveTasks    int32
	CompletedTasks int64
	RejectedTasks  int64
}

// WorkerPoolImpl 工作池实现
type WorkerPoolImpl struct {
	workerCount   int
	taskQueue     chan taskWrapper
	quit          chan struct{}
	wg            sync.WaitGroup
	stats         PoolStats
	mu            sync.RWMutex
	closed        atomic.Bool
}

type taskWrapper struct {
	ctx context.Context
	fn  func(context.Context)
}

// NewWorkerPool 创建工作池
func NewWorkerPool(workerCount int, queueSize int) *WorkerPoolImpl {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}
	if queueSize <= 0 {
		queueSize = 1000
	}

	pool := &WorkerPoolImpl{
		workerCount: workerCount,
		taskQueue:   make(chan taskWrapper, queueSize),
		quit:        make(chan struct{}),
	}

	// 启动 worker
	for i := 0; i < workerCount; i++ {
		pool.wg.Add(1)
		go pool.worker(i)
	}

	atomic.StoreInt32(&pool.stats.WorkerCount, int32(workerCount))

	log.Printf("Worker pool created with %d workers, queue size: %d", workerCount, queueSize)
	return pool
}

// Submit 提交任务到工作池（阻塞模式）
func (p *WorkerPoolImpl) Submit(task func()) error {
	return p.SubmitWithContext(context.Background(), func(ctx context.Context) {
		task()
	})
}

// SubmitWithContext 提交任务到工作池（带上下文）
func (p *WorkerPoolImpl) SubmitWithContext(ctx context.Context, task func(ctx context.Context)) error {
	if p.closed.Load() {
		return ErrPoolClosed
	}

	select {
	case p.taskQueue <- taskWrapper{ctx: ctx, fn: task}:
		atomic.AddInt32(&p.stats.QueueSize, 1)
		return nil
	case <-ctx.Done():
		atomic.AddInt64(&p.stats.RejectedTasks, 1)
		return ctx.Err()
	default:
		// 队列已满，拒绝任务
		atomic.AddInt64(&p.stats.RejectedTasks, 1)
		return ErrQueueFull
	}
}

// worker 工作协程
func (p *WorkerPoolImpl) worker(id int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.quit:
			log.Printf("Worker %d shutting down", id)
			return

		case wrapper := <-p.taskQueue:
			atomic.AddInt32(&p.stats.QueueSize, -1)
			atomic.AddInt32(&p.stats.ActiveTasks, 1)

			// 执行任务
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Worker %d panic recovered: %v", id, r)
					}
					atomic.AddInt32(&p.stats.ActiveTasks, -1)
					atomic.AddInt64(&p.stats.CompletedTasks, 1)
				}()

				wrapper.fn(wrapper.ctx)
			}()
		}
	}
}

// Close 关闭工作池
func (p *WorkerPoolImpl) Close() {
	if !p.closed.CompareAndSwap(false, true) {
		return
	}

	close(p.quit)
	p.wg.Wait()
	close(p.taskQueue)

	log.Printf("Worker pool closed, stats: %+v", p.GetStats())
}

// Wait 等待所有任务完成
func (p *WorkerPoolImpl) Wait() {
	for {
		stats := p.GetStats()
		if stats.QueueSize == 0 && stats.ActiveTasks == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// GetStats 获取工作池统计信息
func (p *WorkerPoolImpl) GetStats() PoolStats {
	return PoolStats{
		WorkerCount:    atomic.LoadInt32(&p.stats.WorkerCount),
		QueueSize:      atomic.LoadInt32(&p.stats.QueueSize),
		ActiveTasks:    atomic.LoadInt32(&p.stats.ActiveTasks),
		CompletedTasks: atomic.LoadInt64(&p.stats.CompletedTasks),
		RejectedTasks:  atomic.LoadInt64(&p.stats.RejectedTasks),
	}
}

// 全局工作池实例
var (
	globalPool      WorkerPool
	poolInitialized atomic.Bool
)

// InitWorkerPool 初始化全局工作池
func InitWorkerPool(workerCount int, queueSize int) {
	if poolInitialized.CompareAndSwap(false, true) {
		globalPool = NewWorkerPool(workerCount, queueSize)
	}
}

// GetWorkerPool 获取全局工作池
func GetWorkerPool() WorkerPool {
	if !poolInitialized.Load() {
		InitWorkerPool(runtime.NumCPU(), 1000)
	}
	return globalPool
}

// CloseWorkerPool 关闭全局工作池
func CloseWorkerPool() {
	if poolInitialized.CompareAndSwap(true, false) {
		globalPool.Close()
	}
}

// 错误定义
var (
	ErrPoolClosed = fmt.Errorf("worker pool is closed")
	ErrQueueFull  = fmt.Errorf("worker pool queue is full")
)

// BatchProcessor 批量处理器
type BatchProcessor struct {
	batchSize int
	timeout   time.Duration
	processor func([]interface{}) error
	batch     []interface{}
	mu        sync.Mutex
	timer     *time.Timer
}

// NewBatchProcessor 创建批量处理器
func NewBatchProcessor(batchSize int, timeout time.Duration, processor func([]interface{}) error) *BatchProcessor {
	return &BatchProcessor{
		batchSize: batchSize,
		timeout:   timeout,
		processor: processor,
		batch:     make([]interface{}, 0, batchSize),
	}
}

// Add 添加数据到批次
func (b *BatchProcessor) Add(data interface{}) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.batch = append(b.batch, data)

	// 达到批次大小，立即处理
	if len(b.batch) >= b.batchSize {
		return b.flush()
	}

	// 重置定时器
	if b.timer != nil {
		b.timer.Stop()
	}
	b.timer = time.AfterFunc(b.timeout, func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		b.flush()
	})

	return nil
}

// flush 刷新批次数据
func (b *BatchProcessor) flush() error {
	if len(b.batch) == 0 {
		return nil
	}

	batch := make([]interface{}, len(b.batch))
	copy(batch, b.batch)
	b.batch = b.batch[:0]

	// 异步处理，避免阻塞
	GetWorkerPool().Submit(func() {
		if err := b.processor(batch); err != nil {
			log.Printf("Batch processing error: %v", err)
		}
	})

	return nil
}

// Flush 手动刷新批次
func (b *BatchProcessor) Flush() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flush()
}

// Close 关闭批量处理器
func (b *BatchProcessor) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.timer != nil {
		b.timer.Stop()
	}
	b.flush()
}