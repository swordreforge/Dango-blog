package errors

import "net/http"

// 通用错误常量
var (
	// ErrInternal 服务器内部错误
	ErrInternal = &BaseError{
		code:       "INTERNAL_ERROR",
		message:    "服务器内部错误",
		httpStatus: http.StatusInternalServerError,
	}

	// ErrBadRequest 请求参数错误
	ErrBadRequest = &BaseError{
		code:       "BAD_REQUEST",
		message:    "请求参数错误",
		httpStatus: http.StatusBadRequest,
	}

	// ErrUnauthorized 未授权
	ErrUnauthorized = &BaseError{
		code:       "UNAUTHORIZED",
		message:    "未授权",
		httpStatus: http.StatusUnauthorized,
	}

	// ErrForbidden 禁止访问
	ErrForbidden = &BaseError{
		code:       "FORBIDDEN",
		message:    "禁止访问",
		httpStatus: http.StatusForbidden,
	}

	// ErrNotFound 资源不存在
	ErrNotFound = &BaseError{
		code:       "NOT_FOUND",
		message:    "资源不存在",
		httpStatus: http.StatusNotFound,
	}

	// ErrConflict 资源冲突
	ErrConflict = &BaseError{
		code:       "CONFLICT",
		message:    "资源冲突",
		httpStatus: http.StatusConflict,
	}

	// ErrMethodNotAllowed 方法不允许
	ErrMethodNotAllowed = &BaseError{
		code:       "METHOD_NOT_ALLOWED",
		message:    "方法不允许",
		httpStatus: http.StatusMethodNotAllowed,
	}

	// ErrServiceUnavailable 服务不可用
	ErrServiceUnavailable = &BaseError{
		code:       "SERVICE_UNAVAILABLE",
		message:    "服务不可用",
		httpStatus: http.StatusServiceUnavailable,
	}

	// ErrRequestTimeout 请求超时
	ErrRequestTimeout = &BaseError{
		code:       "REQUEST_TIMEOUT",
		message:    "请求超时",
		httpStatus: http.StatusRequestTimeout,
	}

	// ErrTooManyRequests 请求过多
	ErrTooManyRequests = &BaseError{
		code:       "TOO_MANY_REQUESTS",
		message:    "请求过多",
		httpStatus: http.StatusTooManyRequests,
	}
)