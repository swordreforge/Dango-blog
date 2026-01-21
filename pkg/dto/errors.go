package dto

import "fmt"

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// BusinessError 业务错误
type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *BusinessError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 预定义业务错误
var (
	ErrUsernameRequired    = &BusinessError{Code: "USERNAME_REQUIRED", Message: "用户名不能为空"}
	ErrUsernameInvalid     = &BusinessError{Code: "USERNAME_INVALID", Message: "用户名格式不正确（3-20个字符，只允许字母、数字、下划线）"}
	ErrUsernameExists      = &BusinessError{Code: "USERNAME_EXISTS", Message: "用户名已存在"}
	ErrEmailRequired       = &BusinessError{Code: "EMAIL_REQUIRED", Message: "邮箱不能为空"}
	ErrEmailInvalid        = &BusinessError{Code: "EMAIL_INVALID", Message: "邮箱格式不正确"}
	ErrEmailExists         = &BusinessError{Code: "EMAIL_EXISTS", Message: "邮箱已被使用"}
	ErrPasswordRequired    = &BusinessError{Code: "PASSWORD_REQUIRED", Message: "密码不能为空"}
	ErrPasswordTooShort    = &BusinessError{Code: "PASSWORD_TOO_SHORT", Message: "密码长度至少为6个字符"}
	ErrPasswordIncorrect   = &BusinessError{Code: "PASSWORD_INCORRECT", Message: "密码错误"}
	ErrUserNotFound        = &BusinessError{Code: "USER_NOT_FOUND", Message: "用户不存在"}
	ErrUserInactive        = &BusinessError{Code: "USER_INACTIVE", Message: "用户账户未激活"}
	ErrUnauthorized        = &BusinessError{Code: "UNAUTHORIZED", Message: "未授权访问"}
	ErrForbidden           = &BusinessError{Code: "FORBIDDEN", Message: "权限不足"}
	ErrPassageNotFound     = &BusinessError{Code: "PASSAGE_NOT_FOUND", Message: "文章不存在"}
	ErrPassageNotPublished = &BusinessError{Code: "PASSAGE_NOT_PUBLISHED", Message: "文章尚未发布"}
	ErrPassagePrivate      = &BusinessError{Code: "PASSAGE_PRIVATE", Message: "此文章为私密文章，仅管理员可见"}
	ErrFileTooLarge        = &BusinessError{Code: "FILE_TOO_LARGE", Message: "文件过大"}
	ErrUnsupportedFileType = &BusinessError{Code: "UNSUPPORTED_FILE_TYPE", Message: "不支持的文件类型"}
	ErrSessionExpired      = &BusinessError{Code: "SESSION_EXPIRED", Message: "会话已过期"}
	ErrSessionNotFound     = &BusinessError{Code: "SESSION_NOT_FOUND", Message: "会话不存在"}
	ErrCannotDeleteAdmin   = &BusinessError{Code: "CANNOT_DELETE_ADMIN", Message: "无法删除管理员账户"}
)

// NewValidationError 创建验证错误
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewBusinessError 创建业务错误
func NewBusinessError(code, message, details string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
		Details: details,
	}
}