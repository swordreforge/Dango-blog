package router

import (
	"net/http"
	"path/filepath"
	"strings"

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

	// 服务根目录下的HTML文件（测试文件等）
	mux.HandleFunc("/test_", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(baseDir, r.URL.Path)
		http.ServeFile(w, r, filePath)
	})

	// 404处理器（必须放在最后）
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 检查是否是根目录下的HTML文件
		if strings.HasSuffix(r.URL.Path, ".html") {
			filePath := filepath.Join(baseDir, r.URL.Path)
			http.ServeFile(w, r, filePath)
			return
		}
		
		if r.URL.Path != "/" {
			controller.RenderStatusPage(w, http.StatusNotFound)
			return
		}
		controller.IndexHandler(w, r)
	})

	return mux
}