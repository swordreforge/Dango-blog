package middleware

import (
	"net/http"
	"time"

	"myblog-gogogo/pkg/logger"
	"myblog-gogogo/pkg/metrics"
)

// responseWriter 包装 http.ResponseWriter 以捕获状态码
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// Logging 日志中间件
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		logger.Info("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		statusColor := getStatusColor(rw.statusCode)

		// 记录性能指标
		isError := rw.statusCode >= 400
		metrics.GetMetrics().RecordRequest(r.Method, duration, isError)

		// 根据状态码选择日志级别
		switch {
		case rw.statusCode >= 500:
			logger.Error("Completed %s %s %s%d%s in %v", r.Method, r.URL.Path, statusColor, rw.statusCode, "\x1b[0m", duration)
		case rw.statusCode >= 400:
			logger.Warn("Completed %s %s %s%d%s in %v", r.Method, r.URL.Path, statusColor, rw.statusCode, "\x1b[0m", duration)
		default:
			logger.Info("Completed %s %s %s%d%s in %v", r.Method, r.URL.Path, statusColor, rw.statusCode, "\x1b[0m", duration)
		}
	})
}

// getStatusColor 根据状态码返回颜色 ANSI 码
func getStatusColor(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "\x1b[32m" // 绿色
	case status >= 300 && status < 400:
		return "\x1b[33m" // 黄色
	case status >= 400 && status < 500:
		return "\x1b[31m" // 红色
	case status >= 500:
		return "\x1b[35m" // 紫色
	default:
		return "\x1b[0m" // 默认
	}
}