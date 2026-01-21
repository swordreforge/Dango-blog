package controller

import (
	"encoding/json"
	"net/http"
)

// LogoutHandler 登出API处理器
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 清除cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	// JWT token已经在AuthMiddleware中验证过了
	// 这里只需要返回成功响应即可
	// 客户端会清除本地存储的token

	response := map[string]interface{}{
		"success": true,
		"message": "退出登录成功",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}