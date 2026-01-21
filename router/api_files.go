package router

import (
	"net/http"

	"myblog-gogogo/controller"
)

// SetupFilesAPIRoutes 配置文件管理API路由
func SetupFilesAPIRoutes(mux *http.ServeMux) {
	// 文件列表和下载
	mux.HandleFunc("/files", controller.FileManagerHandler)
	mux.HandleFunc("/files/download", controller.FileDownloadHandler)

	// 目录创建
	mux.HandleFunc("/files/create-dir", controller.CreateDirectoryHandler)
}