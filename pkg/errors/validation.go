package errors

import (
	"fmt"
	"net/http"
)

// ValidationError 验证错误
type ValidationError struct {
	BaseError
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.code, e.message, e.Field)
}

// NewValidationError 创建验证错误
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{
			code:       "VALIDATION_ERROR",
			message:    "验证失败",
			httpStatus: http.StatusBadRequest,
		},
		Field:   field,
		Message: message,
	}
}

// IsValidationError 判断是否为验证错误
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}