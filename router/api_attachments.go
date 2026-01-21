package router

import (
	"net/http"

	"myblog-gogogo/controller"
)

// SetupAttachmentsAPIRoutes 配置附件管理API路由
func SetupAttachmentsAPIRoutes(mux *http.ServeMux) {
	// 附件列表和下载
	mux.HandleFunc("/attachments", controller.AttachmentHandler)
	mux.HandleFunc("/attachments/download", controller.AttachmentDownloadHandler)

	// 按日期获取附件
	mux.HandleFunc("/attachments/by-date", controller.ArticleAttachmentsHandler)

	// 管理后台附件管理
	mux.HandleFunc("/admin/attachments", controller.AttachmentManagementHandler)
}