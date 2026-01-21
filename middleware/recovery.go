package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"myblog-gogogo/controller"
)

// Recovery 恢复中间件，捕获panic
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				controller.RenderStatusPage(w, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}