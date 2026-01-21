package router

import (
	"net/http"

	"myblog-gogogo/controller"
)

// SetupSettingsAPIRoutes 配置设置API路由
func SetupSettingsAPIRoutes(mux *http.ServeMux) {
	// 通用设置API
	mux.HandleFunc("/settings", controller.SettingAPIHandler)
	mux.HandleFunc("/settings/appearance", controller.SettingAPIHandler)
	mux.HandleFunc("/settings/all", controller.GetAllSettingsHandler)

	// 音乐设置API
	mux.HandleFunc("/settings/music", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controller.GetMusicSettingsHandler(w, r)
		case http.MethodPut:
			controller.UpdateMusicSettingsHandler(w, r)
		case http.MethodPatch:
			controller.UpdateMusicSettingsPartialHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 模板设置API
	mux.HandleFunc("/settings/template", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controller.GetTemplateSettingsHandler(w, r)
		case http.MethodPut, http.MethodPatch:
			controller.UpdateTemplateSettingsHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// 单个设置更新API
	mux.HandleFunc("/settings/single", controller.UpdateSingleSettingHandler)
}