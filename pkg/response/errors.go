package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Error 错误响应
func Error(w http.ResponseWriter, statusCode int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Message: message,
		Code:    code,
	})
}

// ErrorWithDetails 错误响应（带详细信息）
func ErrorWithDetails(w http.ResponseWriter, statusCode int, code string, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := Response{
		Success: false,
		Message: message,
		Code:    code,
	}
	
	if err != nil {
		response.Error = err.Error()
	}
	
	json.NewEncoder(w).Encode(response)
}

// BadRequest 400 错误响应
func BadRequest(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusBadRequest, code, message)
}

// BadRequestWithError 400 错误响应（带错误信息）
func BadRequestWithError(w http.ResponseWriter, code string, message string, err error) {
	ErrorWithDetails(w, http.StatusBadRequest, code, message, err)
}

// Unauthorized 401 错误响应
func Unauthorized(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusUnauthorized, code, message)
}

// Forbidden 403 错误响应
func Forbidden(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusForbidden, code, message)
}

// NotFound 404 错误响应
func NotFound(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusNotFound, code, message)
}

// Conflict 409 错误响应
func Conflict(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusConflict, code, message)
}

// InternalError 500 错误响应
func InternalError(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusInternalServerError, code, message)
}

// InternalErrorWithError 500 错误响应（带错误信息）
func InternalErrorWithError(w http.ResponseWriter, code string, message string, err error) {
	ErrorWithDetails(w, http.StatusInternalServerError, code, message, err)
}

// ServiceUnavailable 503 错误响应
func ServiceUnavailable(w http.ResponseWriter, code string, message string) {
	Error(w, http.StatusServiceUnavailable, code, message)
}

// HandleError 处理错误并返回相应的响应
func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	
	// 根据错误类型返回不同的状态码
	var statusCode int
	var code string
	
	switch {
	case errors.Is(err, ErrNotFound):
		statusCode = http.StatusNotFound
		code = "NOT_FOUND"
	case errors.Is(err, ErrBadRequest):
		statusCode = http.StatusBadRequest
		code = "BAD_REQUEST"
	case errors.Is(err, ErrUnauthorized):
		statusCode = http.StatusUnauthorized
		code = "UNAUTHORIZED"
	case errors.Is(err, ErrForbidden):
		statusCode = http.StatusForbidden
		code = "FORBIDDEN"
	case errors.Is(err, ErrConflict):
		statusCode = http.StatusConflict
		code = "CONFLICT"
	default:
		statusCode = http.StatusInternalServerError
		code = "INTERNAL_ERROR"
	}
	
	ErrorWithDetails(w, statusCode, code, err.Error(), err)
}

// 错误类型定义
var (
	ErrNotFound    = errors.New("资源不存在")
	ErrBadRequest  = errors.New("请求参数错误")
	ErrUnauthorized = errors.New("未授权")
	ErrForbidden   = errors.New("禁止访问")
	ErrConflict    = errors.New("资源冲突")
)

// NewError 创建新的错误
func NewError(code string, message string) error {
	return fmt.Errorf("[%s] %s", code, message)
}