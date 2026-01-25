package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
	"myblog-gogogo/service"
)

// AdminPassagesHandler 文章管理API处理器
func AdminPassagesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// 创建新文章
		// 使用临时结构体处理请求体，因为 Passage 模型已移除 Tags 字段
		type PassageRequest struct {
			models.Passage
			Tags string `json:"tags"` // 临时字段，用于接收请求中的标签
		}
		var req PassageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "Invalid request body",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 验证必填字段
		if req.Title == "" || req.Content == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "标题和内容不能为空",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 验证分类是否存在，如果不存在则自动创建
		if req.Category != "" {
			categoryRepo := db.GetCategoryRepository()
			categories, err := categoryRepo.GetAll()
			if err != nil {
				response := map[string]interface{}{
					"success": false,
					"message": "获取分类列表失败",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			categoryExists := false
			for _, cat := range categories {
				if cat.Name == req.Category {
					categoryExists = true
					break
				}
			}

			if !categoryExists {
				// 分类不存在，自动创建
				newCategory := &models.Category{
					Name:      req.Category,
					IsEnabled: true,
				}
				if err := categoryRepo.Create(newCategory); err != nil {
					log.Printf("创建分类失败: %v", err)
					response := map[string]interface{}{
						"success": false,
						"message": fmt.Sprintf("创建分类 '%s' 失败", req.Category),
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}
				log.Printf("自动创建分类: %s", req.Category)
			}
		}

		// 验证标签是否存在，如果不存在则自动创建
		if req.Tags != "" {
			tagRepo := db.GetTagRepository()
			tags, err := tagRepo.GetAll()
			if err != nil {
				response := map[string]interface{}{
					"success": false,
					"message": "获取标签列表失败",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			// 将标签字符串分割为数组
			tagNames := strings.Split(req.Tags, ",")
			validTags := make([]string, 0)
			createdTags := make([]string, 0)

			for _, tagName := range tagNames {
				trimmedName := strings.TrimSpace(tagName)
				if trimmedName == "" {
					continue
				}

				tagExists := false
				for _, tag := range tags {
					if tag.Name == trimmedName {
						tagExists = true
						break
					}
				}

				if tagExists {
					validTags = append(validTags, trimmedName)
				} else {
					// 标签不存在，自动创建
					newTag := &models.Tag{
						Name:      trimmedName,
						IsEnabled: true,
					}
					if err := tagRepo.Create(newTag); err != nil {
						log.Printf("创建标签失败: %v", err)
						response := map[string]interface{}{
							"success": false,
							"message": fmt.Sprintf("创建标签 '%s' 失败", trimmedName),
						}
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						json.NewEncoder(w).Encode(response)
						return
					}
					createdTags = append(createdTags, trimmedName)
					validTags = append(validTags, trimmedName)
				}
			}

			// 记录创建的标签（可选，用于日志）
			if len(createdTags) > 0 {
				log.Printf("自动创建了以下标签: %s", strings.Join(createdTags, ", "))
			}
		}

		// 设置默认值
		now := time.Now()

		// 如果请求体中提供了 created_at，则使用该值；否则使用当前时间
		if !req.CreatedAt.IsZero() {
			// 使用请求体中提供的创建时间
			now = req.CreatedAt
		}

		// 创建 passage 对象
		passage := models.Passage{
			Title:       req.Title,
			Content:     req.Content,
			Author:      req.Author,
			Category:    req.Category,
			Status:      req.Status,
			Visibility:  req.Visibility,
			IsScheduled: req.IsScheduled,
			PublishedAt: req.PublishedAt,
			ShowTitle:   req.ShowTitle,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if passage.Status == "" {
			passage.Status = "draft"
		}
		if passage.Visibility == "" {
			passage.Visibility = "public"
		}
		if passage.Author == "" {
			passage.Author = "管理员"
		}
		if passage.ShowTitle {
			passage.ShowTitle = true
		}

		// 保存原始内容
		passage.OriginalContent = passage.Content

		// 转换Markdown为HTML
		htmlContent, err := service.ConvertToHTMLWithOption([]byte(passage.OriginalContent), passage.ShowTitle)
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "Markdown转换失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
		passage.Content = htmlContent

		// 创建markdown文件
		cleanedTitle := service.SanitizeFilename(passage.Title)
		dateDir := now.Format("2006/01/02")
		dirPath := filepath.Join("markdown", dateDir)

		// 创建目录
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "创建目录失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 构建文件路径
		filePath := filepath.Join(dirPath, cleanedTitle+".md")
		relativePath := strings.TrimPrefix(filePath, "markdown/")
		relativePath = strings.TrimSuffix(relativePath, ".md")
		passage.FilePath = relativePath

		// 检查文件是否已存在
		if _, err := os.Stat(filePath); err == nil {
			// 文件已存在，添加时间戳避免冲突
			timestamp := now.Format("20060102-150405")
			cleanedTitle = fmt.Sprintf("%s-%s", cleanedTitle, timestamp)
			filePath = filepath.Join(dirPath, cleanedTitle+".md")
			relativePath = strings.TrimPrefix(filePath, "markdown/")
			relativePath = strings.TrimSuffix(relativePath, ".md")
			passage.FilePath = relativePath
		}

		// 写入文件
		if err := service.UpdateMarkdownFile(filePath, passage.Title, passage.OriginalContent); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("创建文件失败: %v", err),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 保存到数据库
		repo := db.GetPassageRepository()
		if err := repo.Create(&passage); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "保存到数据库失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 处理标签关联
		if req.Tags != "" {
			tagRepo := db.GetTagRepository()
			passageTagRepo := db.GetPassageTagRepository()

			// 解析标签（支持 JSON 数组和逗号分隔的字符串）
			var tagNames []string
			if strings.HasPrefix(req.Tags, "[") {
				// JSON 格式
				if err := json.Unmarshal([]byte(req.Tags), &tagNames); err == nil {
					// 解析成功
				} else {
					// 解析失败，尝试作为逗号分隔处理
					tagNames = strings.Split(req.Tags, ",")
				}
			} else {
				// 逗号分隔格式
				tagNames = strings.Split(req.Tags, ",")
			}

			// 为每个标签创建关联
			for _, tagName := range tagNames {
				tagName = strings.TrimSpace(tagName)
				if tagName == "" {
					continue
				}

				// 查找或创建标签
				var tagID int
				tags, err := tagRepo.GetAll()
				if err != nil {
					log.Printf("获取标签列表失败: %v", err)
					continue
				}

				tagExists := false
				for _, tag := range tags {
					if tag.Name == tagName {
						tagID = tag.ID
						tagExists = true
						break
					}
				}

				if !tagExists {
					// 创建新标签
					newTag := &models.Tag{
						Name:      tagName,
						IsEnabled: true,
					}
					if err := tagRepo.Create(newTag); err != nil {
						log.Printf("创建标签失败: %v", err)
						continue
					}
					tagID = newTag.ID
				}

				// 创建文章-标签关联
				passageTag := &models.PassageTag{
					PassageID: passage.ID,
					TagID:     tagID,
				}
				if err := passageTagRepo.Create(passageTag); err != nil {
					log.Printf("创建文章-标签关联失败: %v", err)
				}
			}
		}

		response := map[string]interface{}{
			"success": true,
			"message": "文章创建成功",
			"data": map[string]interface{}{
				"id":         passage.ID,
				"title":      passage.Title,
				"status":     passage.Status,
				"created_at": passage.CreatedAt.Format("2006-01-02"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodGet:
		// 检查是否请求单篇文章详情
		idStr := r.URL.Query().Get("id")
		if idStr != "" {
			// 获取单篇文章详情
			id := 0
			if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil && id > 0 {
				repo := db.GetPassageRepository()
				passage, err := repo.GetByID(id)
				if err != nil {
					http.Error(w, "Failed to fetch passage", http.StatusInternalServerError)
					return
				}
				if passage == nil {
					response := map[string]interface{}{
						"success": false,
						"message": "文章不存在",
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
					return
				}

				// 优先返回原始Markdown内容，如果没有则返回HTML内容
				content := passage.OriginalContent
				if content == "" {
					content = passage.Content
				}

				// 从关联表获取标签
			passageTagRepo := db.GetPassageTagRepository()
			tagRepo := db.GetTagRepository()
			tagIDs, err := passageTagRepo.GetTagIDsByPassageID(passage.ID)
			tagNames := []string{}
			if err == nil && len(tagIDs) > 0 {
				allTags, err := tagRepo.GetAll()
				if err == nil {
					tagMap := make(map[int]string)
					for _, t := range allTags {
						tagMap[t.ID] = t.Name
					}
					for _, tagID := range tagIDs {
						if name, ok := tagMap[tagID]; ok {
							tagNames = append(tagNames, name)
						}
					}
				}
			}

			data := map[string]interface{}{
					"id":             passage.ID,
					"title":          passage.Title,
					"content":        content,
					"summary":        passage.Summary,
					"author":         passage.Author,
					"tags":           tagNames,
					"category":       passage.Category,
					"status":         passage.Status,
					"show_title":     passage.ShowTitle,
					"visibility":     passage.Visibility,
					"is_scheduled":    passage.IsScheduled,
					"published_at":   passage.PublishedAt,
					"created_at":     passage.CreatedAt.Format("2006-01-02"),
					"content_type":   "markdown", // 标识内容类型
				}

				response := map[string]interface{}{
					"success": true,
					"data":    data,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
				return
			}
		}

		// 获取文章列表（管理后台）
		page := r.URL.Query().Get("page")
		limit := r.URL.Query().Get("limit")
		status := r.URL.Query().Get("status")

		if page == "" {
			page = "1"
		}
		if limit == "" {
			limit = "10"
		}

		// 解析分页参数
		pageNum := 1
		limitNum := 10
		fmt.Sscanf(page, "%d", &pageNum)
		fmt.Sscanf(limit, "%d", &limitNum)

		// 计算偏移量
		offsetNum := (pageNum - 1) * limitNum

		// 从数据库获取文章列表
		repo := db.GetPassageRepository()

		var passages []models.Passage
		var err error

		if status != "" {
			passages, err = repo.GetByStatus(status, limitNum, offsetNum)
		} else {
			passages, err = repo.GetAll(limitNum, offsetNum)
		}

		if err != nil {
			http.Error(w, "Failed to fetch passages", http.StatusInternalServerError)
			return
		}

		total, err := repo.Count()
		if err != nil {
			total = 0
		}

		// 转换为API响应格式
		data := make([]map[string]interface{}, len(passages))
		for i, p := range passages {
			data[i] = map[string]interface{}{
				"id":         p.ID,
				"title":      p.Title,
				"status":     p.Status,
				"created_at": p.CreatedAt.Format("2006-01-02"),
			}
		}

		response := map[string]interface{}{
			"success": true,
			"data":    data,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPut:
		// 更新文章
		// 使用临时结构体处理请求体，因为 Passage 模型已移除 Tags 字段
		type PassageUpdateRequest struct {
			models.Passage
			Tags string `json:"tags"` // 临时字段，用于接收请求中的标签
		}
		var req PassageUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 验证必填字段
		if req.Title == "" || req.Content == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "标题和内容不能为空",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 获取ID参数
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少文章ID参数",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		id := 0
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id <= 0 {
			response := map[string]interface{}{
				"success": false,
				"message": "无效的文章ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 将 req.Passage 复制到 passage 变量
		passage := req.Passage

		// 设置文章ID
		passage.ID = id

		// 获取现有文章信息以确定文件路径
		repo := db.GetPassageRepository()
		existingPassage, err := repo.GetByID(id)
		if err != nil || existingPassage == nil {
			response := map[string]interface{}{
				"success": false,
				"message": "文章不存在",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 保留原有的创建时间
		passage.CreatedAt = existingPassage.CreatedAt

		// 如果没有提供 show_title，保留原有值或使用默认值 true
		if !passage.ShowTitle && existingPassage.ShowTitle {
			passage.ShowTitle = existingPassage.ShowTitle
		}

		// 如果没有提供 visibility，保留原有值
		if passage.Visibility == "" {
			passage.Visibility = existingPassage.Visibility
		}

		// 如果没有提供 is_scheduled，保留原有值
		if !passage.IsScheduled && existingPassage.IsScheduled {
			passage.IsScheduled = existingPassage.IsScheduled
		}

		// 如果没有提供 published_at，保留原有值
		if passage.PublishedAt.IsZero() && !existingPassage.PublishedAt.IsZero() {
			passage.PublishedAt = existingPassage.PublishedAt
		}

		// 保留原有的 FilePath（除非标题改变）
		passage.FilePath = existingPassage.FilePath

		// 如果没有提供 category，保留原有值
		if passage.Category == "" && existingPassage.Category != "" {
			passage.Category = existingPassage.Category
		}

		// 如果没有提供原始内容，将HTML内容转换为Markdown（简化处理）
		if passage.OriginalContent == "" {
			// 这里假设前端发送的是Markdown格式的内容
			// 如果前端发送的是HTML，需要进行HTML到Markdown的转换
			passage.OriginalContent = passage.Content
		}

		// 转换Markdown为HTML存储到Content字段，根据show_title决定是否移除第一行标题
		htmlContent, err := service.ConvertToHTMLWithOption([]byte(passage.OriginalContent), passage.ShowTitle)
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "Markdown转换失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
		passage.Content = htmlContent

		// 检查标题是否改变，如果改变了需要重命名文件
		var newFilePath string
		if passage.Title != existingPassage.Title {
			// 清理新标题作为文件名
			cleanedTitle := service.SanitizeFilename(passage.Title)

			// 获取日期目录
			dateDir := passage.CreatedAt.Format("2006/01/02")
			dirPath := filepath.Join("markdown", dateDir)

			// 构建新文件路径
			newFilePath = filepath.Join(dirPath, cleanedTitle+".md")

			// 检查新文件名是否已存在
			if _, err := os.Stat(newFilePath); err == nil {
				// 文件已存在，添加时间戳避免冲突
				timestamp := time.Now().Format("20060102-150405")
				cleanedTitle = fmt.Sprintf("%s-%s", cleanedTitle, timestamp)
				newFilePath = filepath.Join(dirPath, cleanedTitle+".md")
			}

			// 获取旧文件路径
			oldFilePath, err := service.GetMarkdownFilePath(existingPassage.Title, existingPassage.CreatedAt)
			if err != nil {
				response := map[string]interface{}{
					"success": false,
					"message": fmt.Sprintf("获取旧文件路径失败: %v", err),
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			// 重命名文件
			if err := os.Rename(oldFilePath, newFilePath); err != nil {
				// 重命名失败，尝试创建新文件并删除旧文件
				log.Printf("Warning: failed to rename file %s to %s: %v", oldFilePath, newFilePath, err)

				// 先创建新文件
				if err := service.UpdateMarkdownFile(newFilePath, passage.Title, passage.OriginalContent); err != nil {
					response := map[string]interface{}{
						"success": false,
						"message": fmt.Sprintf("创建新文件失败: %v", err),
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}

				// 删除旧文件
				if err := os.Remove(oldFilePath); err != nil {
					log.Printf("Warning: failed to delete old file %s: %v", oldFilePath, err)
				}
			} else {
				log.Printf("Renamed file: %s -> %s", filepath.Base(oldFilePath), filepath.Base(newFilePath))

				// 重命名成功，更新文件内容（因为标题可能变化）
				if err := service.UpdateMarkdownFile(newFilePath, passage.Title, passage.OriginalContent); err != nil {
					response := map[string]interface{}{
						"success": false,
						"message": fmt.Sprintf("更新文件内容失败: %v", err),
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}
			}

			// 更新数据库中的 FilePath 字段
			relativePath := strings.TrimPrefix(newFilePath, "markdown/")
			relativePath = strings.TrimSuffix(relativePath, ".md")
			passage.FilePath = relativePath
		} else {
			// 标题未变化，直接更新文件内容
			markdownFilePath, err := service.GetMarkdownFilePath(passage.Title, passage.CreatedAt)
			if err != nil {
				response := map[string]interface{}{
					"success": false,
					"message": "获取文件路径失败",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			if err := service.UpdateMarkdownFile(markdownFilePath, passage.Title, passage.OriginalContent); err != nil {
				response := map[string]interface{}{
					"success": false,
					"message": fmt.Sprintf("文件更新失败: %v", err),
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
		}

		// 更新数据库
		if err := repo.Update(&passage); err != nil {
			log.Printf("Error updating passage: %v", err)
			response := map[string]interface{}{
				"success": false,
				"message": fmt.Sprintf("更新文章失败: %v", err),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 更新标签关联
		tagRepo := db.GetTagRepository()
		passageTagRepo := db.GetPassageTagRepository()

		// 删除现有的标签关联
		if err := passageTagRepo.DeleteByPassageID(passage.ID); err != nil {
			log.Printf("删除文章标签关联失败: %v", err)
		}

		// 重新创建标签关联
		if req.Tags != "" {
			// 解析标签（支持 JSON 数组和逗号分隔的字符串）
			var tagNames []string
			if strings.HasPrefix(req.Tags, "[") {
				// JSON 格式
				if err := json.Unmarshal([]byte(req.Tags), &tagNames); err == nil {
					// 解析成功
				} else {
					// 解析失败，尝试作为逗号分隔处理
					tagNames = strings.Split(req.Tags, ",")
				}
			} else {
				// 逗号分隔格式
				tagNames = strings.Split(req.Tags, ",")
			}

			// 为每个标签创建关联
			for _, tagName := range tagNames {
				tagName = strings.TrimSpace(tagName)
				if tagName == "" {
					continue
				}

				// 查找或创建标签
				var tagID int
				tags, err := tagRepo.GetAll()
				if err != nil {
					log.Printf("获取标签列表失败: %v", err)
					continue
				}

				tagExists := false
				for _, tag := range tags {
					if tag.Name == tagName {
						tagID = tag.ID
						tagExists = true
						break
					}
				}

				if !tagExists {
					// 创建新标签
					newTag := &models.Tag{
						Name:      tagName,
						IsEnabled: true,
					}
					if err := tagRepo.Create(newTag); err != nil {
						log.Printf("创建标签失败: %v", err)
						continue
					}
					tagID = newTag.ID
				}

				// 创建文章-标签关联
				passageTag := &models.PassageTag{
					PassageID: passage.ID,
					TagID:     tagID,
				}
				if err := passageTagRepo.Create(passageTag); err != nil {
					log.Printf("创建文章-标签关联失败: %v", err)
				}
			}
		}

		response := map[string]interface{}{
			"success": true,
			"message": "文章更新成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPatch:
		// 增量更新文章属性（如 visibility、is_scheduled、published_at、status 等）
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少文章ID参数",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		id := 0
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id <= 0 {
			response := map[string]interface{}{
				"success": false,
				"message": "无效的文章ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 解析更新数据
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 检查文章是否存在
		repo := db.GetPassageRepository()
		existingPassage, err := repo.GetByID(id)
		if err != nil || existingPassage == nil {
			response := map[string]interface{}{
				"success": false,
				"message": "文章不存在",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 允许更新的字段
		allowedFields := map[string]bool{
			"visibility":  true,
			"is_scheduled": true,
			"published_at": true,
			"status":       true,
			"category":     true,
			"tags":         true,
			"summary":      true,
			"show_title":   true,
		}

		// 构建更新数据
		updateData := make(map[string]interface{})
		for field, value := range updates {
			if !allowedFields[field] {
				continue // 跳过不允许的字段
			}
			updateData[field] = value
		}

		if len(updateData) == 0 {
			response := map[string]interface{}{
				"success": false,
				"message": "没有提供有效的更新字段",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 更新现有文章对象
		if visibility, ok := updateData["visibility"].(string); ok {
			existingPassage.Visibility = visibility
		}
		if isScheduled, ok := updateData["is_scheduled"].(bool); ok {
			existingPassage.IsScheduled = isScheduled
		}
		if publishedAtStr, ok := updateData["published_at"].(string); ok {
			publishedAt, err := time.Parse(time.RFC3339, publishedAtStr)
			if err == nil {
				existingPassage.PublishedAt = publishedAt
			}
		}
		if status, ok := updateData["status"].(string); ok {
			existingPassage.Status = status
		}
		if category, ok := updateData["category"].(string); ok {
			existingPassage.Category = category
		}
		// 标签处理：不再更新 existingPassage.Tags，统一使用 passage_tags 关联表
		var tagsStr string
		if tags, ok := updateData["tags"].(string); ok {
			tagsStr = tags
		}
		if summary, ok := updateData["summary"].(string); ok {
			existingPassage.Summary = summary
		}
		if showTitle, ok := updateData["show_title"].(bool); ok {
			existingPassage.ShowTitle = showTitle
		}

		// 更新数据库
		if err := repo.Update(existingPassage); err != nil {
			http.Error(w, "Failed to update passage", http.StatusInternalServerError)
			return
		}

		// 更新标签关联（统一使用 passage_tags 关联表）
		// 检查是否包含 tags 字段（即使是空字符串也要更新，用于清空标签）
		if _, hasTagsField := updateData["tags"]; hasTagsField {
			syncService := service.NewSyncService(repo)
			if err := syncService.UpdatePassageTags(id, tagsStr); err != nil {
				// 记录错误但不影响主流程
				fmt.Printf("Warning: 更新标签关联失败: %v\n", err)
			}
		}

		response := map[string]interface{}{
			"success": true,
			"message": "文章属性更新成功",
			"data": map[string]interface{}{
				"id":          existingPassage.ID,
				"visibility":  existingPassage.Visibility,
				"is_scheduled": existingPassage.IsScheduled,
				"published_at": existingPassage.PublishedAt,
				"status":      existingPassage.Status,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodDelete:
		// 删除文章
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少文章ID参数",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		id := 0
		if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil || id <= 0 {
			response := map[string]interface{}{
				"success": false,
				"message": "无效的文章ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 检查是否启用回收站（默认不启用）
		enableRecycleBin := r.URL.Query().Get("recycle") == "true"

		// 获取文章信息
		repo := db.GetPassageRepository()
		passage, err := repo.GetByID(id)
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "获取文章信息失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		if passage == nil {
			response := map[string]interface{}{
				"success": false,
				"message": "文章不存在",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 如果启用回收站，将文章状态改为deleted
		if enableRecycleBin {
			passage.Status = "deleted"
			if err := repo.Update(passage); err != nil {
				response := map[string]interface{}{
					"success": false,
					"message": "移动到回收站失败",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}

			response := map[string]interface{}{
				"success": true,
				"message": "文章已移动到回收站",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// 永久删除：从数据库删除
		if err := repo.Delete(id); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "从数据库删除文章失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 尝试删除对应的markdown文件
		// 根据创建时间构建可能的文件路径
		createdDate := passage.CreatedAt.Format("2006/01/02")
		markdownDir := "markdown"
		possibleFilePath := filepath.Join(markdownDir, createdDate, passage.Title+".md")

		// 尝试删除文件
		if err := DeleteMarkdownFile(possibleFilePath); err != nil {
			// 文件删除失败不影响整体操作，只记录日志
			// 在实际应用中应该使用日志库
		}

		response := map[string]interface{}{
			"success": true,
			"message": "文章删除成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DeleteMarkdownFile 删除markdown文件
func DeleteMarkdownFile(filePath string) error {
	// 检查文件是否存在
	if _, err := filepath.Abs(filePath); err != nil {
		return err
	}

	// 删除文件
	if err := remove(filePath); err != nil {
		return err
	}

	return nil
}

// remove 删除文件或目录的辅助函数
func remove(path string) error {
	// 检查文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // 文件不存在，视为成功
	}

	// 删除文件
	return os.Remove(path)
}