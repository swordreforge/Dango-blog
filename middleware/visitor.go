package middleware

import (
	"log"
	"net/http"

	"myblog-gogogo/db"
)

// VisitorTracking 访客追踪中间件
func VisitorTracking(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 只记录页面访问，不记录静态资源和API请求
		if r.URL.Path == "/health" || r.URL.Path == "/status/" {
			next.ServeHTTP(w, r)
			return
		}

		// 获取客户端IP
		ip := r.RemoteAddr
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			ip = realIP
		}

		// 提取纯IP地址，去掉端口号
		if idx := len(ip) - 1; idx >= 0 {
			for i := idx; i >= 0; i-- {
				if ip[i] == ':' {
					ip = ip[:i]
					break
				}
				if ip[i] == ']' {
					break // IPv6 格式 [::1]:port
				}
			}
		}

		// 获取User-Agent
		userAgent := r.Header.Get("User-Agent")

		// 异步记录访问（不阻塞请求）
		go func() {
			visitorRepo := db.GetVisitorRepository()
			if err := visitorRepo.RecordVisit(ip, userAgent); err != nil {
				log.Printf("Failed to record visit: %v", err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}