package controller

import (
	"net/http"
	_ "net/http/pprof"
)

// RegisterPProfRoutes 注册 PProf 路由
func RegisterPProfRoutes(mux *http.ServeMux) {
	// PProf 路由
	mux.HandleFunc("/debug/pprof/", func(w http.ResponseWriter, r *http.Request) {
		// 仅在 DEBUG 模式下允许访问
		// 生产环境应该通过配置禁用或添加认证
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	mux.HandleFunc("/debug/pprof/cmdline", func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	mux.HandleFunc("/debug/pprof/profile", func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	mux.HandleFunc("/debug/pprof/symbol", func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	mux.HandleFunc("/debug/pprof/trace", func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	mux.HandleFunc("/debug/pprof/heap", func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	mux.HandleFunc("/debug/pprof/goroutine", func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	mux.HandleFunc("/debug/pprof/threadcreate", func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	mux.HandleFunc("/debug/pprof/block", func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})

	mux.HandleFunc("/debug/pprof/mutex", func(w http.ResponseWriter, r *http.Request) {
		http.DefaultServeMux.ServeHTTP(w, r)
	})
}