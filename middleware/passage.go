package middleware

import (
	"encoding/json"
	"fmt"
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

			// 未发布的文章，返回特殊状态码
			log.Printf("[CheckPassageAccess] Non-admin user, returning 423")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusLocked) // 423 Locked - 表示资源被锁定/不可用

			response := map[string]interface{}{
				"success":     false,
				"message":     "文章尚未发布",
				"status":      targetPassage.Status,
				"is_scheduled": targetPassage.IsScheduled,
			}

			// 如果是定时发布，添加发布时间
			publishedTime := "待定"
			if targetPassage.IsScheduled && !targetPassage.PublishedAt.IsZero() {
				publishedTime = targetPassage.PublishedAt.Format("2006-01-02 15:04:05")
				response["published_at"] = publishedTime
			}

			// 返回一个包含处理逻辑的 HTML 页面
			htmlResponse := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>文章未发布</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0;
            padding: 20px;
        }
        .notice-container {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 20px;
            padding: 40px;
            max-width: 500px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            text-align: center;
        }
        .notice-icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
        .notice-title {
            font-size: 28px;
            color: #333;
            margin-bottom: 15px;
            font-weight: 700;
        }
        .notice-message {
            font-size: 16px;
            color: #666;
            line-height: 1.6;
            margin-bottom: 20px;
        }
        .notice-time {
            font-size: 18px;
            color: #d68910;
            font-weight: 600;
            background: rgba(255, 193, 7, 0.1);
            padding: 10px 20px;
            border-radius: 10px;
            display: inline-block;
            margin-bottom: 20px;
        }
        .back-link {
            display: inline-block;
            color: #667eea;
            text-decoration: none;
            font-weight: 600;
            transition: all 0.3s ease;
        }
        .back-link:hover {
            color: #764ba2;
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="notice-container">
        <div class="notice-icon">🔒</div>
        <div class="notice-title">文章尚未发布</div>
        <div class="notice-message">您访问的文章还未发布，暂时无法查看。</div>
        <div class="notice-time">预计发布时间：%s</div>
        <a href="/" class="back-link">返回首页</a>
    </div>
</body>
</html>`, publishedTime)

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusLocked)
			w.Write([]byte(htmlResponse))
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