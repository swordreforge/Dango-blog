package static

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// FileServer 创建一个禁止目录列表的文件服务器
func FileServer(root http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 打开请求的文件
		f, err := root.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		// 获取文件信息
		stat, err := f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// 如果是目录，返回 404
		if stat.IsDir() {
			http.NotFound(w, r)
			return
		}

		// 获取文件扩展名
		ext := filepath.Ext(r.URL.Path)
		mimeType := GetMimeType(ext)

		// 对于 .mp3 扩展名的文件，检查是否实际上是 MP4/M4A 格式
		if strings.ToLower(ext) == ".mp3" {
			// 读取文件的前12字节用于检测 ftyp 盒（MP4 格式标识）
			buffer := make([]byte, 12)
			n, err := f.Read(buffer)
			if err == nil && n >= 12 {
				// MP4 文件以 ftyp 盒开始，格式为：[4字节大小] + [4字节类型] + [4字节品牌]
				// ftyp 盒的类型是 "ftyp"
				if string(buffer[4:8]) == "ftyp" {
					// 这是 MP4 格式，使用 audio/mp4 MIME 类型
					mimeType = "audio/mp4"
				}
			}
			
			// 尝试回到文件开头，以便 http.ServeContent 能正确读取
			if seeker, ok := f.(io.Seeker); ok {
				seeker.Seek(0, 0)
			}
		}

		// 设置 Content-Type
		w.Header().Set("Content-Type", mimeType)

		// 使用 http.ServeContent 提供文件内容
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
	})
}

// GetMimeType 根据文件扩展名返回 MIME 类型
func GetMimeType(ext string) string {
	switch strings.ToLower(ext) {
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".pdf":
		return "application/pdf"
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	default:
		// 默认使用 text/plain
		return "text/plain; charset=utf-8"
	}
}