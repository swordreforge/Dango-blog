package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
	"myblog-gogogo/service/kafka"
)

// CommentHandler 评论API处理器
func CommentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 获取评论列表
		passageIDStr := r.URL.Query().Get("passage_id")
		page := r.URL.Query().Get("page")
		limit := r.URL.Query().Get("limit")

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

		repo := db.GetCommentRepository()

		var comments []models.Comment
		var err error
		var total int

		if passageIDStr != "" {
			// 获取特定文章的评论
			passageID := 0
			fmt.Sscanf(passageIDStr, "%d", &passageID)
			comments, err = repo.GetByPassageID(passageID, limitNum, offsetNum)
			if err == nil {
				total, err = repo.CountByPassageID(passageID)
				if err != nil {
					total = 0
				}
			}
		} else {
			// 获取所有评论
			comments, err = repo.GetAll(limitNum, offsetNum)
			if err == nil {
				total, err = repo.Count()
				if err != nil {
					total = 0
				}
			}
		}

		if err != nil {
			http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
			return
		}

		// 转换为API响应格式
		data := make([]map[string]interface{}, len(comments))
		for i, c := range comments {
			data[i] = map[string]interface{}{
				"id":         c.ID,
				"username":   c.Username,
				"content":    c.Content,
				"passage_id": c.PassageID,
				"created_at": c.CreatedAt.Format("2006-01-02 15:04:05"),
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

	case http.MethodPost:
		// 创建评论
		var comment models.Comment
		if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
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
		if comment.Username == "" || comment.Content == "" || comment.PassageID == 0 {
			response := map[string]interface{}{
				"success": false,
				"message": "Username, content and passage_id are required",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 创建评论
		repo := db.GetCommentRepository()
		if err := repo.Create(&comment); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "Failed to create comment",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 异步发布评论创建事件到 Kafka（不阻塞响应）
		go func() {
			ctx := context.Background()
			if err := kafka.PublishCommentEventAsync(ctx, "comment.created", comment.ID, comment.PassageID, comment.Content); err != nil {
				// 如果 Kafka 不可用，只记录日志，不影响业务
				fmt.Printf("Warning: Failed to publish comment event to Kafka: %v\n", err)
			}
		}()

		response := map[string]interface{}{
			"success": true,
			"message": "评论创建成功",
			"data": map[string]interface{}{
				"id":         comment.ID,
				"username":   comment.Username,
				"content":    comment.Content,
				"passage_id": comment.PassageID,
				"created_at": comment.CreatedAt.Format("2006-01-02 15:04:05"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodDelete:
		// 删除评论
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少评论ID参数",
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
				"message": "无效的评论ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		repo := db.GetCommentRepository()
		if err := repo.Delete(id); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "删除评论失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "评论删除成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
