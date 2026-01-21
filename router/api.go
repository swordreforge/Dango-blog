package router

import (
	"net/http"

	"myblog-gogogo/controller"
	"myblog-gogogo/controller/admin"
)

// SetupAPIRoutes 配置API路由
func SetupAPIRoutes(mux *http.ServeMux) {
	apiMux := http.NewServeMux()

	// 认证API
	SetupAuthAPIRoutes(apiMux)

	// 文章相关API
	apiMux.HandleFunc("/passages", controller.PassageAPIHandler)
	apiMux.HandleFunc("/passages/", controller.PassageDetailHandler)
	apiMux.HandleFunc("/tags", controller.TagsAPIHandler)
	apiMux.HandleFunc("/categories", controller.CategoriesAPIHandler)
	apiMux.HandleFunc("/archive", controller.ArchiveAPIHandler)

	// 评论API
	apiMux.HandleFunc("/comments", controller.CommentHandler)

	// 同步和上传API
	apiMux.HandleFunc("/sync", controller.SyncHandler)
	apiMux.HandleFunc("/upload", controller.UploadHandler)

	// Markdown编辑器API
	apiMux.HandleFunc("/markdown-editor/save", controller.MarkdownEditorSaveHandler)

	// 用户信息API
	apiMux.HandleFunc("/user/info", controller.UserInfoHandler)

	// ECC加密API
	apiMux.HandleFunc("/crypto/public-key", controller.GetPublicKey)
	apiMux.HandleFunc("/crypto/decrypt", controller.DecryptData)

	// 设置API
	SetupSettingsAPIRoutes(apiMux)

	// 音乐API
	SetupMusicAPIRoutes(apiMux)

	// 关于页面API
	SetupAboutAPIRoutes(apiMux)

	// 文件管理API
	SetupFilesAPIRoutes(apiMux)

	// 附件管理API
	SetupAttachmentsAPIRoutes(apiMux)

	// 管理后台API
	apiMux.HandleFunc("/admin/users", admin.AdminUsersHandler)
	apiMux.HandleFunc("/admin/passages", admin.AdminPassagesHandler)
	apiMux.HandleFunc("/admin/categories", admin.AdminCategoriesHandler)
	apiMux.HandleFunc("/admin/tags", admin.AdminTagsHandler)
	apiMux.HandleFunc("/admin/stats", admin.AdminStatsHandler)
	apiMux.HandleFunc("/admin/comments", admin.AdminCommentsHandler)
	apiMux.HandleFunc("/admin/analytics", controller.AdminAnalyticsHandler)

	// 将所有API路由挂载到 /api/
	mux.Handle("/api/", http.StripPrefix("/api", apiMux))
}