package router

import (
	"net/http"

	"myblog-gogogo/controller"
)

// SetupAuthAPIRoutes 配置认证相关API路由
func SetupAuthAPIRoutes(mux *http.ServeMux) {
	// 用户认证
	mux.HandleFunc("/login", controller.LoginHandler)
	mux.HandleFunc("/logout", controller.LogoutHandler)
	mux.HandleFunc("/register", controller.RegisterHandler)
}