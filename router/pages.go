package router

import (
	"net/http"
	"path/filepath"

	"myblog-gogogo/controller"
	"myblog-gogogo/controller/admin"
)

// SetupPageRoutes 配置页面路由
func SetupPageRoutes(mux *http.ServeMux, baseDir string) {
	// HTML页面路由
	mux.HandleFunc("/keyboard-test", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(baseDir, "template", "keyboard-test.html"))
	})
	mux.HandleFunc("/index", controller.IndexHandler)
	mux.HandleFunc("/passage", controller.PassageHandler)
	mux.HandleFunc("/passage/", controller.PassageHandler)
	mux.HandleFunc("/collect", controller.CollectHandler)
	mux.HandleFunc("/about", controller.AboutHandler)
	mux.HandleFunc("/markdown-editor", controller.MarkdownEditorHandler)

	// 管理后台
	mux.HandleFunc("/admin", admin.AdminHandler)

	// 状态页面
	mux.HandleFunc("/status/", controller.StatusHandler)

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	})
}