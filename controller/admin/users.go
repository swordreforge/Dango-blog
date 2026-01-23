package admin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"myblog-gogogo/auth"
	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
)

// AdminUsersHandler 用户管理API处理器
func AdminUsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 检查是否请求单个用户详情
		idStr := r.URL.Query().Get("id")
		if idStr != "" {
			// 获取单个用户详情
			id := 0
			if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil && id > 0 {
				repo := db.GetUserRepository()
				user, err := repo.GetByID(id)
				if err != nil {
					http.Error(w, "Failed to fetch user", http.StatusInternalServerError)
					return
				}
				if user == nil {
					response := map[string]interface{}{
						"success": false,
						"message": "用户不存在",
					}
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(response)
					return
				}

				data := map[string]interface{}{
					"id":         user.ID,
					"username":   user.Username,
					"email":      user.Email,
					"role":       user.Role,
					"status":     user.Status,
					"created_at": user.CreatedAt.Format("2006-01-02"),
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

		// 获取用户列表
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

		// 从数据库获取用户列表
		repo := db.GetUserRepository()
		users, err := repo.GetAll(limitNum, offsetNum)
		if err != nil {
			http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
			return
		}

		total, err := repo.Count()
		if err != nil {
			total = 0
		}

		// 转换为API响应格式
		data := make([]map[string]interface{}, len(users))
		for i, u := range users {
			data[i] = map[string]interface{}{
				"id":         u.ID,
				"username":   u.Username,
				"email":      u.Email,
				"role":       u.Role,
				"status":     u.Status,
				"created_at": u.CreatedAt.Format("2006-01-02"),
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
		// 创建用户
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 创建用户
		repo := db.GetUserRepository()
		if err := repo.Create(&user); err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "用户创建成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodDelete:
		// 删除用户
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少用户ID参数",
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
				"message": "无效的用户ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 获取用户信息
		repo := db.GetUserRepository()
		user, err := repo.GetByID(id)
		if err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "获取用户信息失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		if user == nil {
			response := map[string]interface{}{
				"success": false,
				"message": "用户不存在",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 防止删除管理员账户
		if user.Role == "admin" {
			response := map[string]interface{}{
				"success": false,
				"message": "无法删除管理员账户",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 删除用户
		if err := repo.Delete(id); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "删除用户失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "用户删除成功",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPut:
		// 全量更新用户
		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// 验证用户ID
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少用户ID参数",
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
				"message": "无效的用户ID",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 设置用户ID
		user.ID = id

		// 验证必填字段
		if user.Username == "" || user.Email == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "用户名和邮箱不能为空",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 检查用户是否存在
		repo := db.GetUserRepository()
		existingUser, err := repo.GetByID(id)
		if err != nil || existingUser == nil {
			response := map[string]interface{}{
				"success": false,
				"message": "用户不存在",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 检查用户名是否已被其他用户使用
		userWithSameName, err := repo.GetByUsername(user.Username)
		if err == nil && userWithSameName != nil && userWithSameName.ID != id {
			response := map[string]interface{}{
				"success": false,
				"message": "用户名已被使用",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 如果密码为空，使用原密码
		if user.Password == "" {
			user.Password = existingUser.Password
		} else {
			// 对新密码进行 Argon2id 哈希处理
			hashedPassword, err := auth.HashPassword(user.Password)
			if err != nil {
				response := map[string]interface{}{
					"success": false,
					"message": "密码哈希失败",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(response)
				return
			}
			user.Password = hashedPassword
		}

		// 更新用户
		if err := repo.Update(&user); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "更新用户失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "用户更新成功",
			"data": map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
				"status":   user.Status,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	case http.MethodPatch:
		// 增量更新用户
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			response := map[string]interface{}{
				"success": false,
				"message": "缺少用户ID参数",
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
				"message": "无效的用户ID",
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

		// 检查用户是否存在
		repo := db.GetUserRepository()
		existingUser, err := repo.GetByID(id)
		if err != nil || existingUser == nil {
			response := map[string]interface{}{
				"success": false,
				"message": "用户不存在",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 如果更新用户名，检查是否已被其他用户使用
		if username, ok := updates["username"].(string); ok && username != "" {
			userWithSameName, err := repo.GetByUsername(username)
			if err == nil && userWithSameName != nil && userWithSameName.ID != id {
				response := map[string]interface{}{
					"success": false,
					"message": "用户名已被使用",
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(response)
				return
			}
		}

		// 增量更新用户
		if err := repo.UpdatePartial(id, updates); err != nil {
			response := map[string]interface{}{
				"success": false,
				"message": "更新用户失败",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 获取更新后的用户信息
		updatedUser, err := repo.GetByID(id)
		if err != nil {
			response := map[string]interface{}{
				"success": true,
				"message": "用户更新成功",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"success": true,
			"message": "用户更新成功",
			"data": map[string]interface{}{
				"id":         updatedUser.ID,
				"username":   updatedUser.Username,
				"email":      updatedUser.Email,
				"role":       updatedUser.Role,
				"status":     updatedUser.Status,
				"created_at": updatedUser.CreatedAt.Format("2006-01-02"),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}