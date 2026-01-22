package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	apperrors "myblog-gogogo/pkg/errors"
	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
	"myblog-gogogo/pkg/dto"
	"myblog-gogogo/service"
	"myblog-gogogo/service/kafka"
)

// PassageAPIHandler 文章列表API处理器
func PassageAPIHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 获取文章列表
		page := r.URL.Query().Get("page")
		limit := r.URL.Query().Get("limit")
		category := r.URL.Query().Get("category")
		_ = r.URL.Query().Get("tag") // TODO: 使用tag进行过滤

		// 默认值
		if page == "" {
			page = "1"
		}
		if limit == "" {
			limit = "10"
		}

		// 解析分页参数
		pageNum, _ := strconv.Atoi(page)
		limitNum, _ := strconv.Atoi(limit)
		offset := (pageNum - 1) * limitNum

		// 从数据库获取文章列表
		repo := db.GetPassageRepository()
		var passages []models.Passage
		var err error

		if category != "" && category != "all" {
			passages, err = repo.GetByCategory(category, limitNum, offset)
		} else {
			passages, err = repo.GetAll(limitNum, offset)
		}

		if err != nil {
			apperrors.SendError(w, apperrors.Wrap(err, "DB_ERROR", "获取文章列表失败"))
			return
		}

		total, err := repo.Count()
		if err != nil {
			total = 0
		}

		// 转换为API响应格式
		data := make([]map[string]interface{}, len(passages))
		for i, p := range passages {
			article := map[string]interface{}{
				"id":         p.ID,
				"title":      p.Title,
				"summary":    p.Summary,
				"tags":       parseTags(p.Tags),
				"category":   p.Category,
				"created_at": p.CreatedAt.Format("2006-01-02"),
				"status":     p.Status,
				"visibility": p.Visibility,
				"is_scheduled": p.IsScheduled,
			}
			
			// 如果是定时发布，添加发布时间
			if p.IsScheduled && !p.PublishedAt.IsZero() {
				article["published_at"] = p.PublishedAt.Format("2006-01-02 15:04:05")
			}
			
			data[i] = article
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

	case http.MethodPost:
		// 创建新文章
		var passage models.Passage
		if err := json.NewDecoder(r.Body).Decode(&passage); err != nil {
			apperrors.SendBadRequest(w, "INVALID_REQUEST_BODY", "请求格式错误")
			return
		}

		// 创建文章
		repo := db.GetPassageRepository()
		if err := repo.Create(&passage); err != nil {
			apperrors.SendError(w, apperrors.Wrap(err, "DB_ERROR", "创建文章失败"))
			return
		}

		// 异步发布文章创建事件到 Kafka（不阻塞响应）
		go func() {
			ctx := context.Background()
			if err := kafka.PublishArticleEventAsync(ctx, "article.created", passage.ID, passage.Title); err != nil {
				// 如果 Kafka 不可用，只记录日志，不影响业务
				fmt.Printf("Warning: Failed to publish article event to Kafka: %v\n", err)
			}
		}()

		response := map[string]interface{}{
			"success": true,
			"message": "文章创建成功",
			"data": map[string]interface{}{
				"id": passage.ID,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		apperrors.SendError(w, apperrors.ErrMethodNotAllowed)
	}
}

// PassageDetailHandler 文章详情API处理器
func PassageDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apperrors.SendError(w, apperrors.ErrMethodNotAllowed)
		return
	}

	// 从URL路径中提取文章ID
	path := strings.TrimPrefix(r.URL.Path, "/passages/")
	idStr := strings.TrimSuffix(path, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		apperrors.SendBadRequest(w, "INVALID_PASSAGE_ID", "无效的文章ID")
		return
	}

	// 获取用户角色
	var role string
	if userID, ok := r.Context().Value(UserIDKey).(int); ok && userID > 0 {
		role = GetRole(r.Context())
	}

	// 使用 PassageService 检查访问权限
	passageSvc := service.NewPassageService()
	accessResp, err := passageSvc.CheckAccess(&dto.PassageAccessRequest{
		PassageID: id,
		UserRole:  role,
	})
	if err != nil {
		apperrors.SendError(w, err)
		return
	}

	// 检查是否允许访问
	if !accessResp.Allowed {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusLocked)
		
		response := map[string]interface{}{
			"success": false,
			"message": accessResp.Reason,
		}

		// 添加额外信息
		if accessResp.Status != "" {
			response["status"] = accessResp.Status
		}
		if accessResp.Visibility != "" {
			response["visibility"] = accessResp.Visibility
		}
		if accessResp.IsScheduled {
			response["is_scheduled"] = true
			if accessResp.PublishedAt != nil {
				response["published_at"] = accessResp.PublishedAt.Format("2006-01-02 15:04:05")
			}
		}

		json.NewEncoder(w).Encode(response)
		return
	}

	// 记录文章阅读（异步，不阻塞响应）
	go func() {
		// 获取客户端真实IP（支持代理头）
		ip := service.GetClientIP(
			r.RemoteAddr,
			r.Header.Get("X-Forwarded-For"),
			r.Header.Get("X-Real-IP"),
		)

		// 如果是本地IP，跳过记录
		if service.IsLocalIP(ip) {
			return
		}

		// 获取 User-Agent
		userAgent := r.Header.Get("User-Agent")

		// 获取地理位置信息
		country, city, region, _ := service.GetLocationFromIP(ip)

		// 记录阅读
		RecordArticleView(id, ip, userAgent, country, city, region)
	}()

	// 从数据库获取完整的文章信息（包含 Summary、Tags、Category）
	repo := db.GetPassageRepository()
	passage, err := repo.GetByID(id)
	if err != nil {
		apperrors.SendError(w, apperrors.Wrap(err, "DB_ERROR", "获取文章详情失败"))
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"id":         accessResp.Passage.ID,
			"title":      accessResp.Passage.Title,
			"content":    accessResp.Passage.Content,
			"summary":    passage.Summary,
			"tags":       parseTags(passage.Tags),
			"category":   passage.Category,
			"show_title": accessResp.Passage.ShowTitle,
			"created_at": accessResp.Passage.CreatedAt.Format("2006-01-02"),
			"updated_at": accessResp.Passage.UpdatedAt.Format("2006-01-02"),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// TagsAPIHandler 标签API处理器
func TagsAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apperrors.SendError(w, apperrors.ErrMethodNotAllowed)
		return
	}

	// 从数据库获取所有文章的标签
	repo := db.GetPassageRepository()
	passages, err := repo.GetAll(1000, 0) // 获取所有文章
	if err != nil {
		apperrors.SendError(w, apperrors.Wrap(err, "DB_ERROR", "获取标签失败"))
		return
	}

	// 统计标签使用次数
	tagCount := make(map[string]int)
	for _, p := range passages {
		tags := parseTags(p.Tags)
		for _, tag := range tags {
			tagCount[tag]++
		}
	}

	// 转换为API响应格式
	data := make([]map[string]interface{}, 0, len(tagCount))
	i := 1
	for tag, count := range tagCount {
		data = append(data, map[string]interface{}{
			"id":    i,
			"name":  tag,
			"count": count,
		})
		i++
	}

	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ArchiveAPIHandler 归档API处理器
func ArchiveAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apperrors.SendError(w, apperrors.ErrMethodNotAllowed)
		return
	}

	// 从数据库获取文章并按月份分组
	repo := db.GetPassageRepository()
	passages, err := repo.GetAll(1000, 0)
	if err != nil {
		apperrors.SendError(w, apperrors.Wrap(err, "DB_ERROR", "获取归档失败"))
		return
	}

	// 按年月分组
	archiveMap := make(map[string]int)
	for _, p := range passages {
		year := p.CreatedAt.Format("2006")
		month := p.CreatedAt.Format("01")
		key := year + "-" + month
		archiveMap[key]++
	}

	// 转换为API响应格式
	data := make([]map[string]interface{}, 0, len(archiveMap))
	for key, count := range archiveMap {
		parts := strings.Split(key, "-")
		data = append(data, map[string]interface{}{
			"year":  parts[0],
			"month": parts[1],
			"count": count,
		})
	}

	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CategoriesAPIHandler 分类API处理器
func CategoriesAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apperrors.SendError(w, apperrors.ErrMethodNotAllowed)
		return
	}

	// 从数据库获取所有分类
	repo := db.GetPassageRepository()
	categories, err := repo.GetAllCategories()
	if err != nil {
		apperrors.SendError(w, apperrors.Wrap(err, "DB_ERROR", "获取分类失败"))
		return
	}

	// 转换为API响应格式
	data := make([]map[string]interface{}, len(categories))
	for i, category := range categories {
		data[i] = map[string]interface{}{
			"id":   i + 1,
			"name": category,
		}
	}

	response := map[string]interface{}{
		"success": true,
		"data":    data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// parseTags 解析标签JSON字符串
func parseTags(tagsStr string) []string {
	if tagsStr == "" || tagsStr == "[]" {
		return []string{}
	}
	
	var tags []string
	if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
		// 如果解析失败，尝试按逗号分割
		return strings.Split(tagsStr, ",")
	}
	
	return tags
}

// filterPublishedPassages 过滤出已发布的文章
func filterPublishedPassages(passages []models.Passage) []models.Passage {
	var published []models.Passage
	for _, p := range passages {
		if p.Status == "published" {
			published = append(published, p)
		}
	}
	return published
}
