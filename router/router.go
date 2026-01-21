package router

import (
	"net/http"

	"myblog-gogogo/controller"
)

// SetupRoutes 配置所有路由
func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// 使用当前工作目录（支持 go run 和编译后的可执行文件）
	baseDir := "."

	// 配置静态文件路由
	SetupStaticRoutes(mux, baseDir)

	// 配置页面路由
	SetupPageRoutes(mux, baseDir)

	// 配置API路由
	SetupAPIRoutes(mux)

	// 404处理器（必须放在最后）
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			controller.RenderStatusPage(w, http.StatusNotFound)
			return
		}
		controller.IndexHandler(w, r)
	})

	return mux
}