package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
)

// CheckPassageAccess 检查文章访问权限
// 如果文章未发布，返回特殊状态码，让前端显示提示
func CheckPassageAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 只检查文章详情页面
		if !strings.HasPrefix(r.URL.Path, "/passage/") {
			next.ServeHTTP(w, r)
			return
		}

		// 提取文章ID
		path := strings.TrimPrefix(r.URL.Path, "/passage/")
		path = strings.TrimSuffix(path, "/")

		// 路径格式: /passage/:year/:month/:day/:name
		parts := strings.Split(path, "/")
		if len(parts) < 4 {
			next.ServeHTTP(w, r)
			return
		}

		// 获取文章路径
		dateDir := strings.Join(parts[:3], "/")
		title := strings.Join(parts[3:], "/")
		targetPath := dateDir + "/" + title

		// 从数据库获取文章信息
		repo := db.GetPassageRepository()
		passages, err := repo.GetAll(1000, 0)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// 查找匹配的文章
		var targetPassage *models.Passage
		for _, p := range passages {
			// 比较文件路径
			if p.FilePath == targetPath {
				targetPassage = &p
				break
			}
		}

		// 如果没有找到匹配的文章，放行
		if targetPassage == nil {
			next.ServeHTTP(w, r)
			return
		}

		// 检查文章状态
		if targetPassage.Status != "published" {
			// 检查是否是管理员
			role, _ := GetRole(r.Context())
			log.Printf("[CheckPassageAccess] Passage not published, user role: %s", role)
			if role == "admin" {
				// 管理员可以访问未发布的文章
				log.Printf("[CheckPassageAccess] Admin user, allowing access to unpublished article")
				next.ServeHTTP(w, r)
				return
			}

			// 未发布的文章，将信息添加到上下文，让前端在右侧显示提示
			log.Printf("[CheckPassageAccess] Non-admin user, adding unpublished info to context")
			
			// 将未发布信息添加到上下文中
			ctx := context.WithValue(r.Context(), "passage_unpublished", true)
			ctx = context.WithValue(ctx, "passage_status", targetPassage.Status)
			ctx = context.WithValue(ctx, "passage_is_scheduled", targetPassage.IsScheduled)
			
			// 如果是定时发布，添加发布时间
			if targetPassage.IsScheduled && !targetPassage.PublishedAt.IsZero() {
				ctx = context.WithValue(ctx, "passage_published_at", targetPassage.PublishedAt.Format("2006-01-02 15:04:05"))
			}
			
			// 继续处理请求，让前端在右侧显示未发布提示
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 检查可见性
		if targetPassage.Visibility == "private" {
			// 检查是否是管理员
			role, _ := GetRole(r.Context())
			if role != "admin" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusLocked)

				response := map[string]interface{}{
					"success":    false,
					"message":    "此文章为私密文章，仅管理员可见",
					"visibility": targetPassage.Visibility,
				}

				json.NewEncoder(w).Encode(response)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}