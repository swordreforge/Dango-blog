package errors

import "net/http"

// 业务错误常量
var (
	// 用户相关错误
	ErrUsernameRequired = &BaseError{
		code:       "USERNAME_REQUIRED",
		message:    "用户名不能为空",
		httpStatus: http.StatusBadRequest,
	}
	ErrUsernameInvalid = &BaseError{
		code:       "USERNAME_INVALID",
		message:    "用户名格式不正确（3-20个字符，只允许字母、数字、下划线）",
		httpStatus: http.StatusBadRequest,
	}
	ErrUsernameExists = &BaseError{
		code:       "USERNAME_EXISTS",
		message:    "用户名已存在",
		httpStatus: http.StatusConflict,
	}
	ErrEmailRequired = &BaseError{
		code:       "EMAIL_REQUIRED",
		message:    "邮箱不能为空",
		httpStatus: http.StatusBadRequest,
	}
	ErrEmailInvalid = &BaseError{
		code:       "EMAIL_INVALID",
		message:    "邮箱格式不正确",
		httpStatus: http.StatusBadRequest,
	}
	ErrEmailExists = &BaseError{
		code:       "EMAIL_EXISTS",
		message:    "邮箱已被使用",
		httpStatus: http.StatusConflict,
	}
	ErrPasswordRequired = &BaseError{
		code:       "PASSWORD_REQUIRED",
		message:    "密码不能为空",
		httpStatus: http.StatusBadRequest,
	}
	ErrPasswordTooShort = &BaseError{
		code:       "PASSWORD_TOO_SHORT",
		message:    "密码长度至少为6个字符",
		httpStatus: http.StatusBadRequest,
	}
	ErrPasswordIncorrect = &BaseError{
		code:       "PASSWORD_INCORRECT",
		message:    "密码错误",
		httpStatus: http.StatusUnauthorized,
	}
	ErrUserNotFound = &BaseError{
		code:       "USER_NOT_FOUND",
		message:    "用户不存在",
		httpStatus: http.StatusNotFound,
	}
	ErrUserInactive = &BaseError{
		code:       "USER_INACTIVE",
		message:    "用户账户未激活",
		httpStatus: http.StatusForbidden,
	}
	ErrCannotDeleteAdmin = &BaseError{
		code:       "CANNOT_DELETE_ADMIN",
		message:    "无法删除管理员账户",
		httpStatus: http.StatusForbidden,
	}

	// 文章相关错误
	ErrPassageNotFound = &BaseError{
		code:       "PASSAGE_NOT_FOUND",
		message:    "文章不存在",
		httpStatus: http.StatusNotFound,
	}
	ErrPassageNotPublished = &BaseError{
		code:       "PASSAGE_NOT_PUBLISHED",
		message:    "文章尚未发布",
		httpStatus: http.StatusNotFound,
	}
	ErrPassagePrivate = &BaseError{
		code:       "PASSAGE_PRIVATE",
		message:    "此文章为私密文章，仅管理员可见",
		httpStatus: http.StatusForbidden,
	}

	// 文件相关错误
	ErrFileTooLarge = &BaseError{
		code:       "FILE_TOO_LARGE",
		message:    "文件过大",
		httpStatus: http.StatusBadRequest,
	}
	ErrUnsupportedFileType = &BaseError{
		code:       "UNSUPPORTED_FILE_TYPE",
		message:    "不支持的文件类型",
		httpStatus: http.StatusBadRequest,
	}

	// 会话相关错误
	ErrSessionExpired = &BaseError{
		code:       "SESSION_EXPIRED",
		message:    "会话已过期",
		httpStatus: http.StatusUnauthorized,
	}
	ErrSessionNotFound = &BaseError{
		code:       "SESSION_NOT_FOUND",
		message:    "会话不存在",
		httpStatus: http.StatusUnauthorized,
	}
)