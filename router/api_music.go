package router

import (
	"net/http"

	"myblog-gogogo/controller"
)

// SetupMusicAPIRoutes 配置音乐API路由
func SetupMusicAPIRoutes(mux *http.ServeMux) {
	// 音乐上传和播放列表
	mux.HandleFunc("/music/upload", controller.MusicUploadHandler)
	mux.HandleFunc("/music/playlist", controller.MusicPlaylistHandler)

	// 音乐管理（删除、更新）
	mux.HandleFunc("/music/", func(w http.ResponseWriter, r *http.Request) {
		// 根据请求方法和查询参数分发到不同的处理器
		if r.Method == http.MethodDelete {
			controller.MusicDeleteHandler(w, r)
		} else if r.Method == http.MethodPut || r.Method == http.MethodPatch {
			// 检查是否是更新标题的请求
			if r.URL.Query().Get("action") == "title" {
				controller.MusicUpdateTitleHandler(w, r)
			} else {
				controller.MusicUpdateCoverHandler(w, r)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}