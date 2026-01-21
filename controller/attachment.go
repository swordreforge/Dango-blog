package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"myblog-gogogo/auth"
	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
	"myblog-gogogo/service/kafka"
	"myblog-gogogo/service/attachment"
)

// AttachmentHandler 附件上传处理器
func AttachmentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodPost:
		handleAttachmentUpload(w, r)
	case http.MethodGet:
		handleAttachmentList(w, r)
	case http.MethodDelete:
		handleAttachmentDelete(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求方法不允许",
			"code":    "METHOD_NOT_ALLOWED",
		})
	}
}

// handleAttachmentUpload 处理附件上传
func handleAttachmentUpload(w http.ResponseWriter, r *http.Request) {
	// 解析表单数据，限制最大文件大小为 500MB
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "解析表单失败",
			"error":   err.Error(),
			"code":    "PARSE_FORM_FAILED",
		})
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "获取文件失败，请确保选择了文件",
			"error":   err.Error(),
			"code":    "NO_FILE_PROVIDED",
		})
		return
	}
	defer file.Close()

	// 获取关联的文章ID（可选）
	passageIDStr := r.FormValue("passage_id")
	var passageID int
	if passageIDStr != "" {
		passageID, err = strconv.Atoi(passageIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "无效的文章ID",
				"error":   err.Error(),
				"code":    "INVALID_PASSAGE_ID",
			})
			return
		}
	}

	// 创建附件服务
	attachmentService := attachment.NewService()

	// 处理上传
	result, err := attachmentService.Upload(file, header, passageID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "上传失败",
			"error":   err.Error(),
			"code":    "UPLOAD_FAILED",
		})
		return
	}

	// 异步发布附件上传事件到 Kafka（不阻塞响应）
	go func() {
		ctx := context.Background()
		if err := kafka.PublishAttachmentUploadEvent(ctx, result.ID, result.FileName, result.Size, result.Type, result.PassageID); err != nil {
			// 如果 Kafka 不可用，只记录日志，不影响业务
			fmt.Printf("Warning: Failed to publish attachment upload event to Kafka: %v\n", err)
		}
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "上传成功",
		"data":    result,
		"code":    "UPLOAD_SUCCESS",
	})
}

// handleAttachmentList 处理获取附件列表
func handleAttachmentList(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	passageIDStr := r.URL.Query().Get("passage_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var passageID *int
	if passageIDStr != "" {
		id, err := strconv.Atoi(passageIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "无效的文章ID",
				"code":    "INVALID_PASSAGE_ID",
			})
			return
		}
		passageID = &id
	}

	limit := 20
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	// 创建附件服务
	attachmentService := attachment.NewService()

	// 获取附件列表
	attachments, total, err := attachmentService.List(passageID, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "获取附件列表失败",
			"error":   err.Error(),
			"code":    "LIST_FAILED",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    attachments,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// handleAttachmentDelete 处理删除附件
func handleAttachmentDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "缺少附件ID",
			"code":    "MISSING_ID",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无效的附件ID",
			"code":    "INVALID_ID",
		})
		return
	}

	// 创建附件服务
	attachmentService := attachment.NewService()

	// 删除附件
	if err := attachmentService.Delete(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "删除失败",
			"error":   err.Error(),
			"code":    "DELETE_FAILED",
		})
		return
	}

	// 异步发布附件删除事件到 Kafka（不阻塞响应）
	go func() {
		ctx := context.Background()
		// 获取附件信息用于事件
		repo := db.GetAttachmentRepository()
		if attachment, err := repo.GetByID(id); err == nil && attachment != nil {
			passageID := 0
			if attachment.PassageID != nil {
				passageID = *attachment.PassageID
			}
			if err := kafka.PublishAttachmentDeleteEvent(ctx, id, attachment.FileName, passageID); err != nil {
				fmt.Printf("Warning: Failed to publish attachment delete event to Kafka: %v\n", err)
			}
		}
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "删除成功",
		"code":    "DELETE_SUCCESS",
	})
}

// AttachmentDownloadHandler 附件下载处理器
func AttachmentDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求方法不允许",
			"code":    "METHOD_NOT_ALLOWED",
		})
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "缺少附件ID",
			"code":    "MISSING_ID",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无效的附件ID",
			"code":    "INVALID_ID",
		})
		return
	}

	// 获取附件信息
	repo := db.GetAttachmentRepository()
	dbAttachment, err := repo.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "附件不存在",
			"error":   err.Error(),
			"code":    "NOT_FOUND",
		})
		return
	}

	// 权限检查
	if dbAttachment.Visibility != "public" {
		// private 和 protected 需要登录验证
		claims, err := auth.GetTokenFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "需要登录才能下载此附件",
				"code":    "UNAUTHORIZED",
			})
			return
		}

		// private 权限需要管理员角色
		if dbAttachment.Visibility == "private" && claims.Role != "admin" {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "需要管理员权限才能下载此附件",
				"code":    "FORBIDDEN",
			})
			return
		}
	}

	// 创建附件服务
	attachmentService := attachment.NewService()

	// 下载附件
	filePath, fileName, err := attachmentService.GetPath(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "附件文件不存在",
			"error":   err.Error(),
			"code":    "FILE_NOT_FOUND",
		})
		return
	}

	// 设置下载头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	// 提供文件下载
	http.ServeFile(w, r, filePath)
}

// ArticleAttachmentsHandler 根据文章日期返回附件列表（无需鉴权）
func ArticleAttachmentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求方法不允许",
			"code":    "METHOD_NOT_ALLOWED",
		})
		return
	}

	// 获取查询参数
	year := r.URL.Query().Get("year")
	month := r.URL.Query().Get("month")
	day := r.URL.Query().Get("day")

	// 验证参数
	if year == "" || month == "" || day == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "缺少必需参数：year、month、day",
			"code":    "MISSING_PARAMS",
		})
		return
	}

	// 创建附件服务
	attachmentService := attachment.NewService()

	// 获取指定日期的附件列表
	attachments, err := attachmentService.GetByDate(year, month, day)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "获取附件列表失败",
			"error":   err.Error(),
			"code":    "FETCH_FAILED",
		})
		return
	}

	// 过滤：只返回公开且在文章中显示的附件
	publicAttachments := make([]*attachment.AttachmentInfo, 0)
	for _, att := range attachments {
		// 获取附件的详细信息
		repo := db.GetAttachmentRepository()
		dbAttachment, err := repo.GetByID(att.ID)
		if err == nil && dbAttachment != nil {
			if dbAttachment.Visibility == "public" && dbAttachment.ShowInPassage {
				publicAttachments = append(publicAttachments, att)
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    publicAttachments,
		"total":   len(publicAttachments),
	})
}

// AttachmentManagementHandler 附件管理处理器（管理员专用）
func AttachmentManagementHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		handleAttachmentManagementList(w, r)
	case http.MethodPatch:
		handleAttachmentUpdate(w, r)
	case http.MethodDelete:
		handleAttachmentDelete(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求方法不允许",
			"code":    "METHOD_NOT_ALLOWED",
		})
	}
}

// handleAttachmentManagementList 获取所有附件列表（管理员）
func handleAttachmentManagementList(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	idStr := r.URL.Query().Get("id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	visibility := r.URL.Query().Get("visibility")
	passageIDStr := r.URL.Query().Get("passage_id")

	// 如果提供了 id 参数,返回单个附件
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "无效的附件ID",
				"code":    "INVALID_ID",
			})
			return
		}

		repo := db.GetAttachmentRepository()
		attachment, err := repo.GetByID(id)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "附件不存在",
				"code":    "NOT_FOUND",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"data":    []*models.Attachment{attachment},
			"total":   1,
			"limit":   1,
			"offset":  0,
		})
		return
	}

	limit := 50
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		o, err := strconv.Atoi(offsetStr)
		if err == nil && o >= 0 {
			offset = o
		}
	}

	// 创建附件服务
	attachmentService := attachment.NewService()

	var attachments []*models.Attachment
	var total int
	var err error

	// 根据查询参数获取附件
	if passageIDStr != "" {
		passageID, err := strconv.Atoi(passageIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "无效的文章ID",
				"code":    "INVALID_PASSAGE_ID",
			})
			return
		}
		attachments, total, err = attachmentService.List(&passageID, limit, offset)
	} else {
		attachments, total, err = attachmentService.List(nil, limit, offset)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "获取附件列表失败",
			"error":   err.Error(),
			"code":    "LIST_FAILED",
		})
		return
	}

	// 过滤可见性
	if visibility != "" {
		filtered := make([]*models.Attachment, 0)
		for _, att := range attachments {
			if att.Visibility == visibility {
				filtered = append(filtered, att)
			}
		}
		attachments = filtered
		total = len(filtered)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    attachments,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// handleAttachmentUpdate 更新附件权限设置
func handleAttachmentUpdate(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "缺少附件ID",
			"code":    "MISSING_ID",
		})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无效的附件ID",
			"code":    "INVALID_ID",
		})
		return
	}

	// 解析请求体
	var updateData struct {
		Visibility    string `json:"visibility"`
		ShowInPassage *bool  `json:"show_in_passage"` // 使用指针来区分零值和未提供
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无效的请求体",
			"error":   err.Error(),
			"code":    "INVALID_REQUEST_BODY",
		})
		return
	}

	// 验证可见性值
	if updateData.Visibility != "" && updateData.Visibility != "public" && updateData.Visibility != "private" && updateData.Visibility != "protected" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无效的可见性值，必须是 public、private 或 protected",
			"code":    "INVALID_VISIBILITY",
		})
		return
	}

	// 获取当前附件信息
	repo := db.GetAttachmentRepository()
	attachment, err := repo.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "附件不存在",
			"code":    "NOT_FOUND",
		})
		return
	}

	// 确定更新值
	visibility := updateData.Visibility
	if visibility == "" {
		visibility = attachment.Visibility
	}

	showInPassage := attachment.ShowInPassage
	if updateData.ShowInPassage != nil {
		showInPassage = *updateData.ShowInPassage
	}

	// 更新附件
	if err := repo.UpdateVisibility(id, visibility, showInPassage); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "更新失败",
			"error":   err.Error(),
			"code":    "UPDATE_FAILED",
		})
		return
	}

	// 异步发布附件更新事件到 Kafka（不阻塞响应）
	go func() {
		ctx := context.Background()
		if err := kafka.PublishAttachmentUpdateEvent(ctx, id, visibility, showInPassage); err != nil {
			fmt.Printf("Warning: Failed to publish attachment update event to Kafka: %v\n", err)
		}
	}()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "更新成功",
		"code":    "UPDATE_SUCCESS",
	})
}
