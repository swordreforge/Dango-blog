package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ConcurrentOperations 并发操作示例
type ConcurrentOperations struct {
	pool WorkerPool
}

// NewConcurrentOperations 创建并发操作实例
func NewConcurrentOperations() *ConcurrentOperations {
	return &ConcurrentOperations{
		pool: GetWorkerPool(),
	}
}

// Example1_BatchUpdateArticles 批量更新文章（示例1）
func (c *ConcurrentOperations) Example1_BatchUpdateArticles(articleIDs []int) error {
	if len(articleIDs) == 0 {
		return nil
	}

	batchSize := 50
	concurrentService := NewConcurrentService(batchSize)

	return concurrentService.BatchUpdateViewCount(articleIDs)
}

// Example2_ParallelFetchStats 并行获取统计信息（示例2）
func (c *ConcurrentOperations) Example2_ParallelFetchStats(articleIDs []int) (map[int]int, error) {
	if len(articleIDs) == 0 {
		return make(map[int]int), nil
	}

	concurrentService := NewConcurrentService(100)
	return concurrentService.ParallelFetch(articleIDs)
}

// Example3_ProcessWithTimeout 带超时的并发处理（示例3）
func (c *ConcurrentOperations) Example3_ProcessWithTimeout(articleIDs []int, timeout time.Duration) error {
	if len(articleIDs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tasks := make([]func(context.Context) error, len(articleIDs))
	for i, id := range articleIDs {
		captureID := id // 避免闭包问题
		tasks[i] = func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// 模拟处理
				time.Sleep(100 * time.Millisecond)
				log.Printf("Processed article %d", captureID)
				return nil
			}
		}
	}

	concurrentService := NewConcurrentService(100)
	return concurrentService.ProcessConcurrentTasks(ctx, tasks)
}

// Example4_BatchProcessor 使用批量处理器（示例4）
func (c *ConcurrentOperations) Example4_BatchProcessor() {
	// 创建批量处理器
	processor := NewBatchProcessor(100, 5*time.Second, func(batch []interface{}) error {
		// 处理批次数据
		log.Printf("Processing batch of %d items", len(batch))
		for _, item := range batch {
			log.Printf("Item: %v", item)
		}
		return nil
	})

	// 添加数据
	for i := 0; i < 250; i++ {
		processor.Add(fmt.Sprintf("item-%d", i))
	}

	// 手动刷新剩余数据
	processor.Flush()
	processor.Close()
}

// Example5_ConcurrentBatchProcessor 使用并发批量处理器（示例5）
func (c *ConcurrentOperations) Example5_ConcurrentBatchProcessor(data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

	processor := NewConcurrentBatchProcessor(50, func(batch []interface{}) error {
		log.Printf("Processing batch of %d items", len(batch))
		// 批量处理逻辑
		return nil
	})

	return processor.Process(data)
}

// Example6_GetPoolStats 获取工作池统计信息（示例6）
func (c *ConcurrentOperations) Example6_GetPoolStats() PoolStats {
	return c.pool.GetStats()
}

// Example7_WaitForAllTasks 等待所有任务完成（示例7）
func (c *ConcurrentOperations) Example7_WaitForAllTasks(articleIDs []int) {
	if len(articleIDs) == 0 {
		return
	}

	// 提交所有任务
	for _, id := range articleIDs {
		captureID := id
		c.pool.Submit(func() {
			// 模拟处理
			time.Sleep(100 * time.Millisecond)
			log.Printf("Processed article %d", captureID)
		})
	}

	// 等待所有任务完成
	c.pool.Wait()
	log.Println("All tasks completed")
}

// Example8_CombinedUsage 组合使用示例（示例8）
func (c *ConcurrentOperations) Example8_CombinedUsage(articleIDs []int) error {
	if len(articleIDs) == 0 {
		return nil
	}

	// 1. 并行获取统计信息
	stats, err := c.Example2_ParallelFetchStats(articleIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch stats: %w", err)
	}

	log.Printf("Fetched stats for %d articles", len(stats))

	// 2. 批量更新文章
	if err := c.Example1_BatchUpdateArticles(articleIDs); err != nil {
		return fmt.Errorf("failed to batch update: %w", err)
	}

	// 3. 检查工作池状态
	poolStats := c.Example6_GetPoolStats()
	log.Printf("Pool stats: %+v", poolStats)

	return nil
}

// Example9_RateLimitedProcessing 限流处理（示例9）
func (c *ConcurrentOperations) Example9_RateLimitedProcessing(articleIDs []int, maxConcurrent int) error {
	if len(articleIDs) == 0 {
		return nil
	}

	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}

	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, id := range articleIDs {
		wg.Add(1)
		err := c.pool.Submit(func() {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 处理文章
			time.Sleep(100 * time.Millisecond)
			log.Printf("Processed article %d with rate limiting", id)
		})

		if err != nil {
			wg.Done()
			return fmt.Errorf("failed to submit task: %w", err)
		}
	}

	wg.Wait()
	return nil
}

// Example10_ErrorHandling 错误处理示例（示例10）
func (c *ConcurrentOperations) Example10_ErrorHandling(articleIDs []int) ([]error, error) {
	if len(articleIDs) == 0 {
		return nil, nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(articleIDs))
	var mu sync.Mutex
	var errors []error

	for _, id := range articleIDs {
		wg.Add(1)
		captureID := id
		err := c.pool.Submit(func() {
			defer wg.Done()

			// 模拟可能失败的操作
			if captureID%10 == 0 {
				errChan <- fmt.Errorf("article %d processing failed", captureID)
				return
			}

			time.Sleep(50 * time.Millisecond)
			log.Printf("Successfully processed article %d", captureID)
		})

		if err != nil {
			wg.Done()
			return nil, fmt.Errorf("failed to submit task: %w", err)
		}
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		mu.Lock()
		errors = append(errors, err)
		mu.Unlock()
	}

	return errors, nil
}

// GetArticleViewStats 获取文章阅读统计（实际应用示例）
func GetArticleViewStats(articleIDs []int) (map[int]int, error) {
	if len(articleIDs) == 0 {
		return make(map[int]int), nil
	}

	ops := NewConcurrentOperations()
	return ops.Example2_ParallelFetchStats(articleIDs)
}

// BatchUpdateArticleViews 批量更新文章阅读量（实际应用示例）
func BatchUpdateArticleViews(articleIDs []int) error {
	if len(articleIDs) == 0 {
		return nil
	}

	ops := NewConcurrentOperations()
	return ops.Example1_BatchUpdateArticles(articleIDs)
}