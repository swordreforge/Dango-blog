package admin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
)

// AdminTagsHandler 标签管理API处理器
func AdminTagsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 检查是否请求单个标签详情
		idStr := r.URL.Query().Get("id")
		if idStr != "" {
			// 获取单个标签详情
			id := 0
			if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil && id > 0 {
				repo := db.GetTagRepository()
				tag, err := repo.GetByID(id)
				if err != nil {
					response := map[string]interface{}{
						"success": false,
						"message": "获取标签详情失败",
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}
				if tag == nil {
					response := map[string]interface{}{
						"success": false,
						"message": "标签不存在",
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
					return
				}

				data := map[string]interface{}{
					"id":          tag.ID,
					"name":        tag.Name,
					"description": tag.Description,
					"color":       tag.Color,
					"category_id": tag.CategoryID,
					"sort_order":  tag.SortOrder,
					"is_enabled":  tag.IsEnabled,
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

		// 获取所有标签
		repo := db.GetTagRepository()
		tags, err := repo.GetAll()
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

		// 转换为API响应格式
		data := make([]map[string]interface{}, len(tags))
		for i, t := range tags {
			data[i] = map[string]interface{}{
				"id":          t.ID,
				"name":        t.Name,
				"description": t.Description,
				"color":       t.Color,
				"category_id": t.CategoryID,
				"sort_order":  t.SortOrder,
				"is_enabled":  t.IsEnabled,
			}
		}

		response := map[string]interface{}{
			"success": true,
			"data":    data,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPost:
		// 创建标签
		var tag models.Tag
		if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "无效的请求数据",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 验证必填字段
		if tag.Name == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "标签名称不能为空",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 创建标签
		repo := db.GetTagRepository()
		if err := repo.Create(&tag); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "创建标签失败：" + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "标签创建成功",
			"data": map[string]interface{}{
				"id": tag.ID,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPut:
		// 更新标签
		var tag models.Tag
		if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "无效的请求数据",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 验证必填字段
		if tag.Name == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "标签名称不能为空",
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
				"message": "缺少标签ID参数",
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
				"message": "无效的标签ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 设置标签ID
		tag.ID = id

		// 更新标签
		repo := db.GetTagRepository()
		if err := repo.Update(&tag); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "更新标签失败：" + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "标签更新成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodDelete:
		// 删除标签
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少标签ID参数",
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
				"message": "无效的标签ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 删除标签
		repo := db.GetTagRepository()
		if err := repo.Delete(id); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "删除标签失败：" + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "标签删除成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPatch:
		// 更新标签排序或启用状态
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少标签ID参数",
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
				"message": "无效的标签ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 解析更新数据
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "无效的请求数据",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		repo := db.GetTagRepository()

		// 检查标签是否存在
		tag, err := repo.GetByID(id)
		if err != nil || tag == nil {
			response := map[string]interface{}{
				"success": false,
				"message": "标签不存在",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 更新排序
		if sortOrder, ok := updates["sort_order"].(float64); ok {
			if err := repo.UpdateSortOrder(id, int(sortOrder)); err != nil {
				response := map[string]interface{}{
					"success": false,
					"message": "更新排序失败",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
		}

		// 更新启用状态
		if isEnabled, ok := updates["is_enabled"].(bool); ok {
			if err := repo.UpdateEnabled(id, isEnabled); err != nil {
				response := map[string]interface{}{
					"success": false,
					"message": "更新启用状态失败",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
		}

		response := map[string]interface{}{
			"success": true,
			"message": "标签更新成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}