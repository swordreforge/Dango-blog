package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"myblog-gogogo/auth"
	"myblog-gogogo/controller"
)

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[AuthMiddleware] Processing request: %s", r.URL.Path)

		// 检查管理后台页面访问
		if r.URL.Path == "/admin" || strings.HasPrefix(r.URL.Path, "/admin/") {
			// 从cookie中获取token
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				// 重定向到首页
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			// 验证token
			claims, err := auth.ValidateToken(cookie.Value)
			if err != nil {
				log.Printf("Admin access - Token validation failed: %v", err)
				// 重定向到首页
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			// 检查是否为管理员
			if claims.Role != "admin" {
				log.Printf("Admin access denied for user %s (role: %s)", claims.Username, claims.Role)
				// 重定向到首页
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			// 将用户信息存入context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			// 使用新的context继续处理请求
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 对API路由进行认证检查
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			// 公开API列表（不需要认证）
			publicAPIs := map[string]bool{
				"/api/login":                 true,
				"/api/register":              true,
				"/api/passages":              true,
				"/api/tags":                  true,
				"/api/categories":            true,
				"/api/archive":               true,
				"/api/stats":                 true,
				"/api/comments":              true, // 评论API公开，允许未登录用户发表评论
				"/api/about/main-cards":      true, // 关于页面主卡片API公开
				"/api/about/sub-cards":       true, // 关于页面次卡片API公开
				"/api/settings/appearance":   true, // 外观设置API公开，允许所有用户查看
				"/api/settings/music":       true, // 音乐设置API公开，允许所有用户查看
				"/api/music/playlist":       true, // 音乐播放列表API公开
				"/music/":                   true, // 音乐文件访问公开
				"/api/attachments":          true, // 附件列表API公开，允许普通用户查看附件
				"/api/attachments/download":  true, // 附件下载API公开，允许普通用户下载附件
				"/api/attachments/by-date":   true, // 根据文章日期获取附件列表API公开，无需鉴权
				"/api/crypto/public-key":     true, // ECC公钥获取API公开
				"/api/user/info":             true, // 用户信息API公开，用于检查登录状态
				//"/api/crypto/decrypt":        true, // ECC解密API公开
			}

			// 检查是否是公开API
			if publicAPIs[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// 检查是否是公开API路径（如 /api/passages/123）
			for apiPath := range publicAPIs {
				if len(r.URL.Path) > len(apiPath) && r.URL.Path[:len(apiPath)] == apiPath {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 需要认证的API，验证JWT token
			// 支持从 Authorization header 或 cookie 中获取 token
			var tokenString string

			// 首先尝试从 Authorization header 获取
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				// 支持两种格式: "Bearer <token>" 或直接 "<token>"
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				if tokenString == authHeader {
					// 如果没有Bearer前缀，直接使用整个header值
					tokenString = authHeader
				}
			} else {
				// 如果 Authorization header 为空，尝试从 cookie 获取
				// 对于管理后台的 API 路由（/api/admin 和 /api/files），支持 cookie 认证
				if strings.HasPrefix(r.URL.Path, "/api/admin") || strings.HasPrefix(r.URL.Path, "/api/files") {
					cookie, err := r.Cookie("auth_token")
					if err == nil {
						tokenString = cookie.Value
					}
				}
			}

			if tokenString == "" {
				controller.RenderStatusPage(w, http.StatusUnauthorized)
				return
			}

			// 验证token
			claims, err := auth.ValidateToken(tokenString)
			if err != nil {
				log.Printf("Token validation failed: %v", err)
				controller.RenderStatusPage(w, http.StatusUnauthorized)
				return
			}

			// 将用户信息存入context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			// 检查管理员权限
			if strings.HasPrefix(r.URL.Path, "/api/admin") && claims.Role != "admin" {
				controller.RenderStatusPage(w, http.StatusForbidden)
				return
			}

			// 使用新的context继续处理请求
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 对于其他路径（如 /passage/），尝试从 cookie 中获取用户信息并设置到 context
		// 这样其他中间件（如 CheckPassageAccess）可以使用这些信息
		log.Printf("[AuthMiddleware] Processing non-admin/non-api path: %s", r.URL.Path)
		cookie, err := r.Cookie("auth_token")
		if err == nil {
			log.Printf("[AuthMiddleware] Found auth_token cookie")
			// 尝试验证 token
			claims, validateErr := auth.ValidateToken(cookie.Value)
			if validateErr == nil {
				log.Printf("[AuthMiddleware] Token validated for user %s (role: %s) on path %s", claims.Username, claims.Role, r.URL.Path)
				// 将用户信息存入context
				ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
				ctx = context.WithValue(ctx, UsernameKey, claims.Username)
				ctx = context.WithValue(ctx, RoleKey, claims.Role)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				log.Printf("[AuthMiddleware] Token validation failed on path %s: %v", r.URL.Path, validateErr)
			}
		} else {
			log.Printf("[AuthMiddleware] No auth_token cookie found on path %s: %v", r.URL.Path, err)
		}

		next.ServeHTTP(w, r)
	})
}