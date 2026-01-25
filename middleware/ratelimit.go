package middleware

import (
	"container/list"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"myblog-gogogo/pkg/metrics"
)

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
}

// RateLimitStrategy 限流策略
type RateLimitStrategy struct {
	POST RateLimitConfig
	GET  RateLimitConfig
}

// RateLimiter 限流器
type RateLimiter struct {
	limiter     *rate.Limiter
	lastUpdate time.Time
	element     *list.Element // LRU 链表元素
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(next http.Handler) http.Handler {
	// 定义限流策略
	strategy := RateLimitStrategy{
		POST: RateLimitConfig{
			RequestsPerSecond: 5,  // 每秒5个请求
			BurstSize:         12, // 突发12个请求
		},
		GET: RateLimitConfig{
			RequestsPerSecond: 250, // 每秒250个请求
			BurstSize:         250, // 突发250个请求
		},
	}

	// LRU 缓存配置
	const maxLimiters = 10000 // 最大限流器数量
	const cleanupInterval = 1 * time.Minute // 清理间隔
	const maxIdleTime = 5 * time.Minute // 最大空闲时间

	// 使用 LRU 缓存存储限流器
	limiters := make(map[string]*RateLimiter)
	lruList := list.New()
	var mu sync.RWMutex

	// 清理过期限流器的 goroutine
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now()

			// 从链表尾部开始清理
			for lruList.Len() > 0 {
				// 获取链表尾部元素（最久未使用）
				elem := lruList.Back()
				if elem == nil {
					break
				}

				limiter := elem.Value.(*RateLimiter)

				// 如果限流器超过最大空闲时间，删除它
				if now.Sub(limiter.lastUpdate) > maxIdleTime {
					// 从链表和 map 中删除
					lruList.Remove(elem)
					// 找到对应的 key 并删除（需要反向查找）
					for key, l := range limiters {
						if l == limiter {
							delete(limiters, key)
							break
						}
					}
				} else {
					// 如果没有过期的限流器，停止清理
					break
				}
			}

			// 如果仍然超过最大容量，强制删除最旧的限流器
			for len(limiters) > maxLimiters && lruList.Len() > 0 {
				elem := lruList.Back()
				if elem != nil {
					limiter := elem.Value.(*RateLimiter)
					lruList.Remove(elem)
					for key, l := range limiters {
						if l == limiter {
							delete(limiters, key)
							break
						}
					}
				}
			}

			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method

		// PATCH 和 PUT 请求不做限流（主要是管理员请求）
		if method == http.MethodPatch || method == http.MethodPut {
			next.ServeHTTP(w, r)
			return
		}

		// 检查用户角色，管理员用户不受限流限制
		if role, ok := GetRole(r.Context()); ok && role == "admin" {
			next.ServeHTTP(w, r)
			return
		}

		// 获取客户端 IP
		ip := getClientIP(r)

		// 根据方法选择限流策略
		var config RateLimitConfig
		switch method {
		case http.MethodPost:
			config = strategy.POST
		case http.MethodGet:
			config = strategy.GET
		default:
			// 其他方法（DELETE 等）使用 GET 的限制策略
			config = strategy.GET
		}

		// 生成限流器 key
		key := method + ":" + ip

		// 获取或创建限流器
		mu.Lock()
		limiter, exists := limiters[key]
		if !exists {
			limiter = &RateLimiter{
				limiter:     rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.BurstSize),
				lastUpdate: time.Now(),
			}
			// 添加到 LRU 链表头部
			limiter.element = lruList.PushFront(limiter)
			limiters[key] = limiter
		} else {
			// 更新最后使用时间
			limiter.lastUpdate = time.Now()
			// 移动到 LRU 链表头部
			if limiter.element != nil {
				lruList.MoveToFront(limiter.element)
			}
		}
		mu.Unlock()

		// 检查是否超过限流
		if !limiter.limiter.Allow() {
			// 记录限流拒绝指标
			metrics.GetMetrics().RecordRateLimitRejected()

			// 超过限流，返回 429 Too Many Requests
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			// 计算重试时间（基于限流策略）
			retryAfter := int64(1) // 默认1秒
			if method == http.MethodPost {
				retryAfter = 12 // POST 请求建议等待12秒
			}

			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			
			response := map[string]interface{}{
				"success": false,
				"error":   "RATE_LIMIT_EXCEEDED",
				"message": "请求过于频繁，请稍后再试",
				"code":    429,
				"retry_after": retryAfter,
			}
			
			// 使用 json.Marshal 编码响应
			jsonBytes, err := json.Marshal(response)
			if err != nil {
				// 如果 JSON 编码失败，返回简单文本
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			
			w.Write(jsonBytes)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP 获取客户端 IP 地址
func getClientIP(r *http.Request) string {
	// 优先从 X-Forwarded-For 获取（代理情况）
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For 可能包含多个 IP，取第一个
		if idx := len(xff); idx > 0 {
			for i, c := range xff {
				if c == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}

	// 其次从 X-Real-IP 获取
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 最后从 RemoteAddr 获取
	if r.RemoteAddr != "" {
		// RemoteAddr 格式为 "ip:port"，只取 IP 部分
		if idx := len(r.RemoteAddr); idx > 0 {
			for i := idx - 1; i >= 0; i-- {
				if r.RemoteAddr[i] == ':' {
					return r.RemoteAddr[:i]
				}
			}
		}
		return r.RemoteAddr
	}

	return "unknown"
}