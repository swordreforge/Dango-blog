package controller

import (
	"encoding/json"
	"net/http"

	apperrors "myblog-gogogo/pkg/errors"
)

// UserInfoHandler 获取当前用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		apperrors.SendError(w, apperrors.ErrMethodNotAllowed)
		return
	}

	// 尝试从context中获取用户信息
	userID, hasUserID := r.Context().Value(UserIDKey).(int)
	username, hasUsername := r.Context().Value(UsernameKey).(string)
	role, hasRole := r.Context().Value(RoleKey).(string)

	// 如果没有用户信息，返回未登录状态
	if !hasUserID || !hasUsername || !hasRole {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未登录",
			"data": map[string]interface{}{
				"logged_in": false,
			},
		})
		return
	}

	// 返回用户信息
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "获取用户信息成功",
		"data": map[string]interface{}{
			"logged_in": true,
			"user_id":   userID,
			"username":  username,
			"role":      role,
		},
	})
}