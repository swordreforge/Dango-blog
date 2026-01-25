package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
	"myblog-gogogo/service"
	"myblog-gogogo/service/settings"
)

// MarkdownEditorHandler Markdown编辑器页面处理器
func MarkdownEditorHandler(w http.ResponseWriter, r *http.Request) {
	// 获取外观设置
	appearanceSettings := getAppearanceSettings()

	// 获取模板设置
	templateSettings, err := settings.GetTemplate()
	if err != nil {
		// 如果获取失败，使用默认值
		templateSettings = &settings.TemplateSettings{
			Name:               "欢迎来到我的博客",
			Greting:            "这是一个使用 Go 语言构建的个人博客系统",
			Year:               "2026",
			Foodes:             "我的博客",
			ArticleTitle:       true,
			ArticleTitlePrefix: "文章",
			SwitchNotice:       true,
			SwitchNoticeText:   "回来继续阅读",
		}
	}

	data := map[string]interface{}{
		"title":                    "在线Markdown编辑器",
		"Settings":                 appearanceSettings,
		"SwitchNotice":             templateSettings.SwitchNotice,
		"SwitchNoticeText":         templateSettings.SwitchNoticeText,
		"ExternalLinkWarning":      templateSettings.ExternalLinkWarning,
		"ExternalLinkWhitelist":    templateSettings.ExternalLinkWhitelist,
		"ExternalLinkWarningText":  templateSettings.ExternalLinkWarningText,
		"Live2DEnabled":            templateSettings.Live2DEnabled,
	}
	renderTemplate(w, "markdown-editor.html", data)
}

// MarkdownEditorSaveHandler 保存Markdown文章到数据库
func MarkdownEditorSaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查用户权限（需要管理员权限）
	role, ok := r.Context().Value(RoleKey).(string)
	if !ok || role != "admin" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "需要管理员权限才能保存文章",
		})
		return
	}

	// 解析请求体
	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Tags     string `json:"tags"`
		Summary  string `json:"summary"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无效的请求数据",
		})
		return
	}

	// 验证必填字段
	if req.Title == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "文章标题不能为空",
		})
		return
	}

	if req.Content == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "文章内容不能为空",
		})
		return
	}

	// 转换 Markdown 为 HTML
	htmlContent, err := service.ConvertToHTML([]byte(req.Content))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Markdown转换失败",
		})
		return
	}

	// 构建文件路径（按日期组织）
	now := time.Now()
	dateDir := now.Format("2006/01/02")
	filePath := strings.Join([]string{"markdown", dateDir, req.Title + ".md"}, "/")

	// 保存 Markdown 文件到磁盘
	if err := service.UpdateMarkdownFile(filePath, req.Title, req.Content); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "保存Markdown文件失败: " + err.Error(),
		})
		return
	}

	// 设置默认分类
	if req.Category == "" {
		req.Category = "未分类"
	}

	// 设置默认摘要
	if req.Summary == "" {
		req.Summary = "暂无摘要"
	}

	// 获取用户信息
	username, _ := r.Context().Value(UsernameKey).(string)

	// 创建文章记录
	passage := &models.Passage{
		Title:           req.Title,
		Content:         htmlContent,
		OriginalContent: req.Content,
		Summary:         req.Summary,
		Author:          username,
		Category:        req.Category,
		Status:          "published",
		FilePath:        filePath,
		ShowTitle:       true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// 保存到数据库
	repo := db.GetPassageRepository()
	if err := repo.Create(passage); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "保存到数据库失败: " + err.Error(),
		})
		return
	}

	// 处理标签关联（使用 passage_tags 关联表）
	if req.Tags != "" {
		syncService := service.NewSyncService(repo)
		if err := syncService.UpdatePassageTags(passage.ID, req.Tags); err != nil {
			// 记录错误但不影响主流程
			log.Printf("Warning: 创建标签关联失败: %v", err)
		}
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "文章保存成功",
		"data": map[string]interface{}{
			"id":         passage.ID,
			"title":      passage.Title,
			"file_path":  passage.FilePath,
			"created_at": passage.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}