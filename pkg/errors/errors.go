package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError 应用错误接口
type AppError interface {
	error
	Code() string
	Message() string
	HTTPStatus() int
	Details() string
	Unwrap() error
}

// BaseError 基础错误结构
type BaseError struct {
	code       string
	message    string
	httpStatus int
	details    string
	cause      error
}

func (e *BaseError) Error() string {
	if e.details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.code, e.message, e.details)
	}
	return fmt.Sprintf("[%s] %s", e.code, e.message)
}

func (e *BaseError) Code() string {
	return e.code
}

func (e *BaseError) Message() string {
	return e.message
}

func (e *BaseError) HTTPStatus() int {
	return e.httpStatus
}

func (e *BaseError) Details() string {
	return e.details
}

func (e *BaseError) Unwrap() error {
	return e.cause
}

// New 创建新的应用错误
func New(code, message string) AppError {
	return &BaseError{
		code:       code,
		message:    message,
		httpStatus: http.StatusInternalServerError,
	}
}

// NewWithStatus 创建带 HTTP 状态码的应用错误
func NewWithStatus(code, message string, status int) AppError {
	return &BaseError{
		code:       code,
		message:    message,
		httpStatus: status,
	}
}

// NewWithDetails 创建带详细信息的应用错误
func NewWithDetails(code, message, details string) AppError {
	return &BaseError{
		code:       code,
		message:    message,
		httpStatus: http.StatusInternalServerError,
		details:    details,
	}
}

// Wrap 包装已知错误
func Wrap(err error, code, message string) AppError {
	if err == nil {
		return nil
	}
	return &BaseError{
		code:       code,
		message:    message,
		httpStatus: http.StatusInternalServerError,
		cause:      err,
	}
}

// WrapWithStatus 包装已知错误并指定 HTTP 状态码
func WrapWithStatus(err error, code, message string, status int) AppError {
	if err == nil {
		return nil
	}
	return &BaseError{
		code:       code,
		message:    message,
		httpStatus: status,
		cause:      err,
	}
}

// WrapWithDetails 包装已知错误并添加详细信息
func WrapWithDetails(err error, code, message, details string) AppError {
	if err == nil {
		return nil
	}
	return &BaseError{
		code:       code,
		message:    message,
		httpStatus: http.StatusInternalServerError,
		details:    details,
		cause:      err,
	}
}

// IsAppError 判断是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(AppError)
	return ok
}

// GetCode 获取错误代码
func GetCode(err error) string {
	if appErr, ok := err.(AppError); ok {
		return appErr.Code()
	}
	return "UNKNOWN"
}

// GetHTTPStatus 获取 HTTP 状态码
func GetHTTPStatus(err error) int {
	if appErr, ok := err.(AppError); ok {
		return appErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}

// AsAppError 将错误转换为应用错误
func AsAppError(err error) (AppError, bool) {
	var appErr AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}