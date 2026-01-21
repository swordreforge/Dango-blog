package admin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
)

// AdminCategoriesHandler 分类管理API处理器
func AdminCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 检查是否请求单个分类详情
		idStr := r.URL.Query().Get("id")
		if idStr != "" {
			// 获取单个分类详情
			id := 0
			if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil && id > 0 {
				repo := db.GetCategoryRepository()
				category, err := repo.GetByID(id)
				if err != nil {
					response := map[string]interface{}{
						"success": false,
						"message": "获取分类详情失败",
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(response)
					return
				}
				if category == nil {
					response := map[string]interface{}{
						"success": false,
						"message": "分类不存在",
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
					return
				}

				data := map[string]interface{}{
					"id":          category.ID,
					"name":        category.Name,
					"description": category.Description,
					"icon":        category.Icon,
					"sort_order":  category.SortOrder,
					"is_enabled":  category.IsEnabled,
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

		// 获取所有分类
		repo := db.GetCategoryRepository()
		categories, err := repo.GetAll()
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

		// 转换为API响应格式
		data := make([]map[string]interface{}, len(categories))
		for i, c := range categories {
			data[i] = map[string]interface{}{
				"id":          c.ID,
				"name":        c.Name,
				"description": c.Description,
				"icon":        c.Icon,
				"sort_order":  c.SortOrder,
				"is_enabled":  c.IsEnabled,
			}
		}

		response := map[string]interface{}{
			"success": true,
			"data":    data,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPost:
		// 创建分类
		var category models.Category
		if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
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
		if category.Name == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "分类名称不能为空",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 创建分类
		repo := db.GetCategoryRepository()
		if err := repo.Create(&category); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "创建分类失败：" + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "分类创建成功",
			"data": map[string]interface{}{
				"id": category.ID,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPut:
		// 更新分类
		var category models.Category
		if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
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
		if category.Name == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "分类名称不能为空",
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
				"message": "缺少分类ID参数",
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
				"message": "无效的分类ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 设置分类ID
		category.ID = id

		// 更新分类
		repo := db.GetCategoryRepository()
		if err := repo.Update(&category); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "更新分类失败：" + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "分类更新成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodDelete:
		// 删除分类
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少分类ID参数",
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
				"message": "无效的分类ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 删除分类
		repo := db.GetCategoryRepository()
		if err := repo.Delete(id); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "删除分类失败：" + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "分类删除成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPatch:
		// 更新分类排序或启用状态
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少分类ID参数",
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
				"message": "无效的分类ID",
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

		repo := db.GetCategoryRepository()

		// 检查分类是否存在
		category, err := repo.GetByID(id)
		if err != nil || category == nil {
			response := map[string]interface{}{
				"success": false,
				"message": "分类不存在",
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
			"message": "分类更新成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}