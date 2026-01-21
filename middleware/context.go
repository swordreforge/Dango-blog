package middleware

import (
	"context"
	"net/http"

	"myblog-gogogo/controller"
)

// ContextKey 用于在context中存储用户信息的key
type ContextKey string

const (
	UserIDKey   ContextKey = "user_id"
	UsernameKey ContextKey = "username"
	RoleKey     ContextKey = "role"
)

// GetUserID 从context中获取用户ID
func GetUserID(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDKey).(int)
	return userID, ok
}

// GetUsername 从context中获取用户名
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameKey).(string)
	return username, ok
}

// GetRole 从context中获取用户角色
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}

// RequireAdmin 检查是否为管理员
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := GetRole(r.Context())
		if !ok || role != "admin" {
			controller.RenderStatusPage(w, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAuth 检查是否已认证
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, ok := GetUserID(r.Context())
		if !ok {
			controller.RenderStatusPage(w, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}