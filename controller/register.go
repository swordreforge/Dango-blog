package controller

import (
	"encoding/json"
	"net/http"

	apperrors "myblog-gogogo/pkg/errors"
	"myblog-gogogo/pkg/dto"
	"myblog-gogogo/service"
)

var userService = service.NewUserService()

// RegisterHandler 注册API处理器
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		apperrors.SendError(w, apperrors.ErrMethodNotAllowed)
		return
	}

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperrors.SendBadRequest(w, "INVALID_REQUEST_BODY", "请求格式错误")
		return
	}

	// 调用用户服务
	resp, err := userService.Register(&req)
	if err != nil {
		apperrors.SendError(w, err)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "注册成功",
		"user":    resp.User,
		"token":   resp.Token,
	})
}