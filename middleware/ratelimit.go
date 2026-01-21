package middleware

import (
	"net/http"
)

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: 实现限流逻辑
		// 可以使用令牌桶或漏桶算法
		next.ServeHTTP(w, r)
	})
}