package metrics

import (
	"runtime"
	"sync"
	"time"
)

// Metrics 性能指标
type Metrics struct {
	mu sync.RWMutex

	// HTTP 请求指标
	TotalRequests      int64
	TotalErrors        int64
	ResponseTimes      []time.Duration
	RequestsPerMethod  map[string]int64

	// 数据库指标
	DBQueryCount       int64
	DBQueryErrors      int64
	DBQueryTimes       []time.Duration

	// 工作池指标
	WorkerPoolSize     int32
	WorkerQueueLength  int32
	WorkerActiveTasks  int32
	WorkerRejectedTasks int64

	// 限流指标
	RateLimitRejected  int64

	// 内存指标
	LastMemoryUsage    runtime.MemStats
}

var (
	instance *Metrics
	once     sync.Once
)

// GetMetrics 获取全局指标实例
func GetMetrics() *Metrics {
	once.Do(func() {
		instance = &Metrics{
			RequestsPerMethod: make(map[string]int64),
			ResponseTimes:     make([]time.Duration, 0, 1000),
			DBQueryTimes:      make([]time.Duration, 0, 1000),
		}
	})
	return instance
}

// RecordRequest 记录 HTTP 请求
func (m *Metrics) RecordRequest(method string, duration time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.RequestsPerMethod[method]++

	// 只保留最近的 1000 个响应时间
	if len(m.ResponseTimes) >= 1000 {
		m.ResponseTimes = m.ResponseTimes[1:]
	}
	m.ResponseTimes = append(m.ResponseTimes, duration)

	if isError {
		m.TotalErrors++
	}
}

// RecordDBQuery 记录数据库查询
func (m *Metrics) RecordDBQuery(duration time.Duration, isError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DBQueryCount++

	// 只保留最近的 1000 个查询时间
	if len(m.DBQueryTimes) >= 1000 {
		m.DBQueryTimes = m.DBQueryTimes[1:]
	}
	m.DBQueryTimes = append(m.DBQueryTimes, duration)

	if isError {
		m.DBQueryErrors++
	}
}

// RecordRateLimitRejected 记录限流拒绝
func (m *Metrics) RecordRateLimitRejected() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RateLimitRejected++
}

// UpdateWorkerPoolStats 更新工作池统计
func (m *Metrics) UpdateWorkerPoolStats(size, queueLength, activeTasks int32, rejectedTasks int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.WorkerPoolSize = size
	m.WorkerQueueLength = queueLength
	m.WorkerActiveTasks = activeTasks
	m.WorkerRejectedTasks = rejectedTasks
}

// UpdateMemoryStats 更新内存统计
func (m *Metrics) UpdateMemoryStats() {
	m.mu.Lock()
	defer m.mu.Unlock()
	runtime.ReadMemStats(&m.LastMemoryUsage)
}

// GetPercentile 计算百分位数
func (m *Metrics) GetPercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	// 复制切片以避免修改原始数据
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)

	// 简单排序
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := int(float64(len(sorted)-1) * percentile)
	return sorted[index]
}

// GetStats 获取统计信息
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 计算响应时间百分位数
	p50 := m.GetPercentile(m.ResponseTimes, 0.50)
	p95 := m.GetPercentile(m.ResponseTimes, 0.95)
	p99 := m.GetPercentile(m.ResponseTimes, 0.99)

	// 计算数据库查询时间百分位数
	dbP50 := m.GetPercentile(m.DBQueryTimes, 0.50)
	dbP95 := m.GetPercentile(m.DBQueryTimes, 0.95)
	dbP99 := m.GetPercentile(m.DBQueryTimes, 0.99)

	// 更新内存统计
	runtime.ReadMemStats(&m.LastMemoryUsage)

	return map[string]interface{}{
		"http": map[string]interface{}{
			"total_requests":      m.TotalRequests,
			"total_errors":        m.TotalErrors,
			"requests_per_method": m.RequestsPerMethod,
			"response_times": map[string]interface{}{
				"p50": p50.Milliseconds(),
				"p95": p95.Milliseconds(),
				"p99": p99.Milliseconds(),
			},
		},
		"database": map[string]interface{}{
			"query_count":  m.DBQueryCount,
			"query_errors": m.DBQueryErrors,
			"query_times": map[string]interface{}{
				"p50": dbP50.Milliseconds(),
				"p95": dbP95.Milliseconds(),
				"p99": dbP99.Milliseconds(),
			},
		},
		"worker_pool": map[string]interface{}{
			"size":          m.WorkerPoolSize,
			"queue_length":  m.WorkerQueueLength,
			"active_tasks":  m.WorkerActiveTasks,
			"rejected_tasks": m.WorkerRejectedTasks,
		},
		"rate_limit": map[string]interface{}{
			"rejected": m.RateLimitRejected,
		},
		"memory": map[string]interface{}{
			"alloc":       m.LastMemoryUsage.Alloc / 1024 / 1024,      // MB
			"total_alloc": m.LastMemoryUsage.TotalAlloc / 1024 / 1024, // MB
			"sys":         m.LastMemoryUsage.Sys / 1024 / 1024,        // MB
			"num_gc":      m.LastMemoryUsage.NumGC,
			"goroutines":  runtime.NumGoroutine(),
		},
	}
}

// Reset 重置所有指标
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests = 0
	m.TotalErrors = 0
	m.ResponseTimes = make([]time.Duration, 0, 1000)
	m.RequestsPerMethod = make(map[string]int64)
	m.DBQueryCount = 0
	m.DBQueryErrors = 0
	m.DBQueryTimes = make([]time.Duration, 0, 1000)
	m.RateLimitRejected = 0
}