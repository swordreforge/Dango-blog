package router

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"myblog-gogogo/controller"
	"myblog-gogogo/static"
)

// SetupStaticRoutes 配置静态文件路由
func SetupStaticRoutes(mux *http.ServeMux, baseDir string) {
	// ========== 模板相关资源：优先使用嵌入的文件系统，回退到本地文件系统 ==========
	
	// 尝试从嵌入的文件系统加载静态资源
	if templateFS := controller.GetTemplateFS(); templateFS != nil {
		// 使用嵌入的文件系统,需要使用 fs.Sub 获取 template 子目录
		subFS, err := fs.Sub(templateFS, "template")
		if err == nil {
			mux.Handle("/static/", http.StripPrefix("/static/", static.FileServer(http.FS(subFS))))
		}
	} else {
		// 回退到本地文件系统
		fs := static.FileServer(http.Dir(filepath.Join(baseDir, "template")))
		mux.Handle("/static/", http.StripPrefix("/static/", fs))
	}

	// CSS文件服务（优先使用嵌入的文件系统）
	if templateFS := controller.GetTemplateFS(); templateFS != nil {
		subFS, err := fs.Sub(templateFS, "template/css")
		if err == nil {
			mux.Handle("/css/", http.StripPrefix("/css/", static.FileServer(http.FS(subFS))))
		} else {
			// 回退到本地文件系统
			cssFs := static.FileServer(http.Dir(filepath.Join(baseDir, "template", "css")))
			mux.Handle("/css/", http.StripPrefix("/css/", cssFs))
		}
	} else {
		// 回退到本地文件系统
		cssFs := static.FileServer(http.Dir(filepath.Join(baseDir, "template", "css")))
		mux.Handle("/css/", http.StripPrefix("/css/", cssFs))
	}

	// JavaScript文件服务（优先使用嵌入的文件系统）
	if templateFS := controller.GetTemplateFS(); templateFS != nil {
		subFS, err := fs.Sub(templateFS, "template/js")
		if err == nil {
			mux.Handle("/js/", http.StripPrefix("/js/", static.FileServer(http.FS(subFS))))
		} else {
			// 回退到本地文件系统
			jsFs := static.FileServer(http.Dir(filepath.Join(baseDir, "template", "js")))
			mux.Handle("/js/", http.StripPrefix("/js/", jsFs))
		}
	} else {
		// 回退到本地文件系统
		jsFs := static.FileServer(http.Dir(filepath.Join(baseDir, "template", "js")))
		mux.Handle("/js/", http.StripPrefix("/js/", jsFs))
	}

	// ========== 用户资源：优先使用本地文件系统（释放的文件夹），回退到嵌入文件系统 ==========

	// 图片文件服务（优先使用本地文件系统）
	imgLocalPath := filepath.Join(baseDir, "img")
	if _, err := os.Stat(imgLocalPath); err == nil {
		// 本地文件系统存在，优先使用
		imgFs := static.FileServer(http.Dir(imgLocalPath))
		mux.Handle("/img/", http.StripPrefix("/img/", imgFs))
	} else if templateFS := controller.GetTemplateFS(); templateFS != nil {
		// 回退到嵌入的文件系统
		subFS, err := fs.Sub(templateFS, "img")
		if err == nil {
			mux.Handle("/img/", http.StripPrefix("/img/", static.FileServer(http.FS(subFS))))
		}
	}

	// 音乐文件服务（优先使用本地文件系统）
	musicLocalPath := filepath.Join(baseDir, "music")
	if _, err := os.Stat(musicLocalPath); err == nil {
		// 本地文件系统存在，优先使用
		musicFs := static.FileServer(http.Dir(musicLocalPath))
		mux.Handle("/music/", http.StripPrefix("/music/", musicFs))
	} else if templateFS := controller.GetTemplateFS(); templateFS != nil {
		// 回退到嵌入的文件系统
		subFS, err := fs.Sub(templateFS, "music")
		if err == nil {
			mux.Handle("/music/", http.StripPrefix("/music/", static.FileServer(http.FS(subFS))))
		}
	}

	// 附件文件服务（仅使用本地文件系统）
	attachmentFs := static.FileServer(http.Dir(filepath.Join(baseDir, "attachments")))
	mux.Handle("/attachments/", http.StripPrefix("/attachments/", attachmentFs))

	// Markdown文件服务（仅使用本地文件系统）
	markdownFs := static.FileServer(http.Dir(filepath.Join(baseDir, "markdown")))
	mux.Handle("/markdown/", http.StripPrefix("/markdown/", markdownFs))
}