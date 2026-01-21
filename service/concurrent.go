package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"myblog-gogogo/db"
	"myblog-gogogo/db/repositories"
)

// ConcurrentService 并发服务
type ConcurrentService struct {
	db        repositories.ArticleViewRepository
	pool      WorkerPool
	batchSize int
}

// NewConcurrentService 创建并发服务
func NewConcurrentService(batchSize int) *ConcurrentService {
	if batchSize <= 0 {
		batchSize = 100
	}

	return &ConcurrentService{
		db:        db.GetArticleViewRepository(),
		pool:      GetWorkerPool(),
		batchSize: batchSize,
	}
}

// BatchUpdateViewCount 批量更新文章阅读量
func (s *ConcurrentService) BatchUpdateViewCount(articleIDs []int) error {
	if len(articleIDs) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(articleIDs)/s.batchSize+1)

	// 分批处理
	for i := 0; i < len(articleIDs); i += s.batchSize {
		end := i + s.batchSize
		if end > len(articleIDs) {
			end = len(articleIDs)
		}

		batch := articleIDs[i:end]

		wg.Add(1)
		err := s.pool.Submit(func() {
			defer wg.Done()
			if err := s.updateViewCountBatch(batch); err != nil {
				errChan <- fmt.Errorf("batch %d-%d: %w", i, end, err)
			}
		})

		if err != nil {
			wg.Done()
			return fmt.Errorf("failed to submit batch: %w", err)
		}
	}

	// 等待所有批次完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集错误
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d batches failed: %v", len(errors), errors)
	}

	return nil
}

// updateViewCountBatch 批量更新阅读量
func (s *ConcurrentService) updateViewCountBatch(articleIDs []int) error {
	// 这里可以实现批量更新逻辑
	// 例如：批量查询、批量更新等
	for _, id := range articleIDs {
		_, err := s.db.GetArticleViews(id)
		if err != nil {
			return fmt.Errorf("failed to get views for article %d: %w", id, err)
		}
	}
	return nil
}

// ProcessConcurrentTasks 并发处理多个任务
func (s *ConcurrentService) ProcessConcurrentTasks(ctx context.Context, tasks []func(context.Context) error) error {
	if len(tasks) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))
	semaphore := make(chan struct{}, 10) // 限制并发数

	for i, task := range tasks {
		wg.Add(1)
		err := s.pool.SubmitWithContext(ctx, func(ctx context.Context) {
			defer wg.Done()

			// 信号量控制并发数
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := task(ctx); err != nil {
				errChan <- fmt.Errorf("task %d: %w", i, err)
			}
		})

		if err != nil {
			wg.Done()
			return fmt.Errorf("failed to submit task %d: %w", i, err)
		}
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d tasks failed: %v", len(errors), errors)
	}

	return nil
}

// ParallelFetch 并行获取多个文章的阅读统计
func (s *ConcurrentService) ParallelFetch(articleIDs []int) (map[int]int, error) {
	if len(articleIDs) == 0 {
		return make(map[int]int), nil
	}

	result := make(map[int]int)
	resultMu := sync.Mutex{}
	var wg sync.WaitGroup

	for _, id := range articleIDs {
		wg.Add(1)
		err := s.pool.Submit(func() {
			defer wg.Done()

			views, err := s.db.GetArticleViews(id)
			if err != nil {
				log.Printf("Failed to get views for article %d: %v", id, err)
				return
			}

			resultMu.Lock()
			result[id] = views
			resultMu.Unlock()
		})

		if err != nil {
			wg.Done()
			return nil, fmt.Errorf("failed to submit task for article %d: %w", id, err)
		}
	}

	wg.Wait()
	return result, nil
}

// ConcurrentBatchProcessor 并发批量处理器
type ConcurrentBatchProcessor struct {
	batchSize int
	processor func([]interface{}) error
}

// NewConcurrentBatchProcessor 创建并发批量处理器
func NewConcurrentBatchProcessor(batchSize int, processor func([]interface{}) error) *ConcurrentBatchProcessor {
	return &ConcurrentBatchProcessor{
		batchSize: batchSize,
		processor: processor,
	}
}

// Process 处理数据
func (c *ConcurrentBatchProcessor) Process(data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(data)/c.batchSize+1)

	for i := 0; i < len(data); i += c.batchSize {
		end := i + c.batchSize
		if end > len(data) {
			end = len(data)
		}

		batch := data[i:end]

		wg.Add(1)
		err := GetWorkerPool().Submit(func() {
			defer wg.Done()
			if err := c.processor(batch); err != nil {
				errChan <- fmt.Errorf("batch %d-%d: %w", i, end, err)
			}
		})

		if err != nil {
			wg.Done()
			return fmt.Errorf("failed to submit batch: %w", err)
		}
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("%d batches failed: %v", len(errors), errors)
	}

	return nil
}