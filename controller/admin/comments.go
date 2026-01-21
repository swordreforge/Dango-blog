package admin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myblog-gogogo/db"
)

// AdminCommentsHandler 评论管理API处理器
func AdminCommentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 获取评论列表
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
		comments, err := repo.GetAll(limitNum, offsetNum)
		if err != nil {
			http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
			return
		}

		total, err := repo.Count()
		if err != nil {
			total = 0
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