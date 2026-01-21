package controller

import (
	"encoding/json"
	"net/http"

	apperrors "myblog-gogogo/pkg/errors"
	"myblog-gogogo/pkg/dto"
	"myblog-gogogo/service"
)

var authService = service.NewAuthService()

// LoginHandler 登录API处理器
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apperrors.SendError(w, apperrors.ErrMethodNotAllowed)
		return
	}

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.SendBadRequest(w, "INVALID_REQUEST_BODY", "请求格式错误")
		return
	}

	// 调用认证服务
	resp, err := authService.Login(&req)
	if err != nil {
		apperrors.SendError(w, err)
		return
	}

	// 设置cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    resp.Token,
		Path:     "/",
		MaxAge:   3600 * 24 * 7,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "登录成功",
		"token":   resp.Token,
		"user":    resp.User,
	})
}