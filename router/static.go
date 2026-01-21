package router

import (
	"net/http"
	"path/filepath"

	"myblog-gogogo/static"
)

// SetupStaticRoutes 配置静态文件路由
func SetupStaticRoutes(mux *http.ServeMux, baseDir string) {
	// 静态文件服务（禁止目录列表）
	fs := static.FileServer(http.Dir(filepath.Join(baseDir, "template")))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// CSS文件服务（禁止目录列表）
	cssFs := static.FileServer(http.Dir(filepath.Join(baseDir, "template", "css")))
	mux.Handle("/css/", http.StripPrefix("/css/", cssFs))

	// JavaScript文件服务（禁止目录列表）
	jsFs := static.FileServer(http.Dir(filepath.Join(baseDir, "template", "js")))
	mux.Handle("/js/", http.StripPrefix("/js/", jsFs))

	// 图片文件服务（禁止目录列表）
	imgFs := static.FileServer(http.Dir(filepath.Join(baseDir, "img")))
	mux.Handle("/img/", http.StripPrefix("/img/", imgFs))

	// 附件文件服务（禁止目录列表）
	attachmentFs := static.FileServer(http.Dir(filepath.Join(baseDir, "attachments")))
	mux.Handle("/attachments/", http.StripPrefix("/attachments/", attachmentFs))

	// 音乐文件服务（禁止目录列表）
	musicFs := static.FileServer(http.Dir(filepath.Join(baseDir, "music")))
	mux.Handle("/music/", http.StripPrefix("/music/", musicFs))
}