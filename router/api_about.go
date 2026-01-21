package router

import (
	"net/http"

	"myblog-gogogo/controller"
)

// SetupAboutAPIRoutes 配置关于页面API路由
func SetupAboutAPIRoutes(mux *http.ServeMux) {
	// 主卡片API
	mux.HandleFunc("/about/main-cards", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controller.AboutMainCardsHandler(w, r)
		case http.MethodPost:
			controller.AboutMainCardCreateHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/about/main-cards/admin", controller.AboutMainCardsAdminHandler)
	mux.HandleFunc("/about/main-cards/update", controller.AboutMainCardUpdateHandler)
	mux.HandleFunc("/about/main-cards/delete", controller.AboutMainCardDeleteHandler)
	mux.HandleFunc("/about/main-cards/sort", controller.AboutMainCardUpdateSortHandler)
	mux.HandleFunc("/about/main-cards/enabled", controller.AboutMainCardUpdateEnabledHandler)

	// 子卡片API
	mux.HandleFunc("/about/sub-cards", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			controller.AboutSubCardsHandler(w, r)
		case http.MethodPost:
			controller.AboutSubCardCreateHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/about/sub-cards/admin", controller.AboutSubCardsAdminHandler)
	mux.HandleFunc("/about/sub-cards/update", controller.AboutSubCardUpdateHandler)
	mux.HandleFunc("/about/sub-cards/delete", controller.AboutSubCardDeleteHandler)
	mux.HandleFunc("/about/sub-cards/sort", controller.AboutSubCardUpdateSortHandler)
	mux.HandleFunc("/about/sub-cards/enabled", controller.AboutSubCardUpdateEnabledHandler)
}