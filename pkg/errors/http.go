package errors

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrorResponse HTTP 错误响应结构
type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Error   string `json:"error,omitempty"`
}

// SendError 发送错误响应
func SendError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var statusCode int
	var code string
	var message string
	var details string
	var errorStr string

	if appErr, ok := AsAppError(err); ok {
		statusCode = appErr.HTTPStatus()
		code = appErr.Code()
		message = appErr.Message()
		details = appErr.Details()
		errorStr = err.Error()
	} else {
		statusCode = http.StatusInternalServerError
		code = "INTERNAL_ERROR"
		message = "服务器内部错误"
		errorStr = err.Error()
	}

	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Success: false,
		Code:    code,
		Message: message,
		Details: details,
	}

	if errorStr != "" {
		response.Error = errorStr
	}

	json.NewEncoder(w).Encode(response)
}

// SendErrorWithDetails 发送带详细信息的错误响应
func SendErrorWithDetails(w http.ResponseWriter, statusCode int, code, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Success: false,
		Code:    code,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	json.NewEncoder(w).Encode(response)
}

// HandleHTTPError 处理 HTTP 错误
func HandleHTTPError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, ErrNotFound):
		SendError(w, ErrNotFound)
	case errors.Is(err, ErrBadRequest):
		SendError(w, ErrBadRequest)
	case errors.Is(err, ErrUnauthorized):
		SendError(w, ErrUnauthorized)
	case errors.Is(err, ErrForbidden):
		SendError(w, ErrForbidden)
	case errors.Is(err, ErrConflict):
		SendError(w, ErrConflict)
	default:
		SendError(w, err)
	}
}

// SendBadRequest 发送 400 错误
func SendBadRequest(w http.ResponseWriter, code, message string) {
	SendErrorWithDetails(w, http.StatusBadRequest, code, message, nil)
}

// SendUnauthorized 发送 401 错误
func SendUnauthorized(w http.ResponseWriter, code, message string) {
	SendErrorWithDetails(w, http.StatusUnauthorized, code, message, nil)
}

// SendForbidden 发送 403 错误
func SendForbidden(w http.ResponseWriter, code, message string) {
	SendErrorWithDetails(w, http.StatusForbidden, code, message, nil)
}

// SendNotFound 发送 404 错误
func SendNotFound(w http.ResponseWriter, code, message string) {
	SendErrorWithDetails(w, http.StatusNotFound, code, message, nil)
}

// SendConflict 发送 409 错误
func SendConflict(w http.ResponseWriter, code, message string) {
	SendErrorWithDetails(w, http.StatusConflict, code, message, nil)
}

// SendInternalError 发送 500 错误
func SendInternalError(w http.ResponseWriter, code, message string) {
	SendErrorWithDetails(w, http.StatusInternalServerError, code, message, nil)
}

// SendInternalErrorWithError 发送 500 错误（带错误信息）
func SendInternalErrorWithError(w http.ResponseWriter, code, message string, err error) {
	SendErrorWithDetails(w, http.StatusInternalServerError, code, message, err)
}