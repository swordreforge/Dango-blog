package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"myblog-gogogo/db"
	"myblog-gogogo/service"
)

// MusicTrack 音乐曲目模型
type MusicTrack struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Artist     string    `json:"artist"`
	FilePath   string    `json:"file_path"`
	FileName   string    `json:"file_name"`
	Duration   string    `json:"duration"`
	CoverImage string    `json:"cover_image"`
	CreatedAt  time.Time `json:"created_at"`
}

// MusicUploadHandler 处理音乐文件上传
func MusicUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 解析 multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// 获取文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 验证文件类型
	allowedTypes := map[string]bool{
		"audio/mpeg":     true,
		"audio/mp3":      true,
		"audio/wav":      true,
		"audio/wave":     true,
		"audio/ogg":      true,
		"audio/x-m4a":    true,
		"audio/mp4":      true,
		"application/octet-stream": true, // 允许一些浏览器不识别的音频格式
	}

	contentType := handler.Header.Get("Content-Type")
	if !allowedTypes[contentType] && !strings.HasSuffix(strings.ToLower(handler.Filename), ".mp3") &&
		!strings.HasSuffix(strings.ToLower(handler.Filename), ".wav") &&
		!strings.HasSuffix(strings.ToLower(handler.Filename), ".ogg") &&
		!strings.HasSuffix(strings.ToLower(handler.Filename), ".m4a") {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// 创建音乐目录
	musicDir := "./music"
	if err := os.MkdirAll(musicDir, 0755); err != nil {
		http.Error(w, "Failed to create music directory", http.StatusInternalServerError)
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(handler.Filename)
	uniqueName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), strings.TrimSuffix(handler.Filename, ext), ext)
	filePath := filepath.Join(musicDir, uniqueName)

	// 保存文件
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// 获取标题和艺术家
	title := r.FormValue("title")
	if title == "" {
		title = strings.TrimSuffix(handler.Filename, ext)
	}
	artist := r.FormValue("artist")
	if artist == "" {
		artist = "未知艺术家"
	}

	// 提取音频元数据
	duration := "未知"
	metadata, err := service.GetAudioMetadata(filePath)
	if err != nil {
		// 如果提取失败，记录警告但继续处理
		fmt.Printf("Warning: Failed to extract metadata for %s: %v\n", filePath, err)
	} else {
		// 如果用户没有提供标题或艺术家，使用元数据中的值
		if title == strings.TrimSuffix(handler.Filename, ext) && metadata.Title != "" {
			title = metadata.Title
		}
		if artist == "未知艺术家" && metadata.Artist != "" {
			artist = metadata.Artist
		}
		// 注意：tag库不提供时长信息，duration保持为"未知"
	}

	// 处理封面上传
	coverImage := ""
	coverFile, coverHandler, err := r.FormFile("cover")
	if err == nil && coverFile != nil {
		defer coverFile.Close()

		// 验证封面文件类型
		coverAllowedTypes := map[string]bool{
			"image/jpeg": true,
			"image/jpg":  true,
			"image/png":  true,
			"image/gif":  true,
			"image/webp": true,
		}

		coverContentType := coverHandler.Header.Get("Content-Type")
		if coverAllowedTypes[coverContentType] {
			// 创建封面目录
			coverDir := filepath.Join("music", "covers")
			if err := os.MkdirAll(coverDir, 0755); err != nil {
				fmt.Printf("Warning: Failed to create cover directory: %v\n", err)
			} else {
				// 生成唯一封面文件名
				coverExt := filepath.Ext(coverHandler.Filename)
				coverUniqueName := fmt.Sprintf("%d_cover%s", time.Now().Unix(), coverExt)
				coverPath := filepath.Join(coverDir, coverUniqueName)

				// 保存封面文件
				coverDst, err := os.Create(coverPath)
				if err == nil {
					defer coverDst.Close()
					if _, err := io.Copy(coverDst, coverFile); err == nil {
						coverImage = "/music/covers/" + coverUniqueName
					}
				}
			}
		}
		coverFile.Close()
	}

	// 保存到数据库
	database := db.GetDB()
	query := `INSERT INTO music_tracks (title, artist, file_path, file_name, duration, cover_image, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := database.Exec(query, title, artist, filePath, uniqueName, duration, coverImage, time.Now())
	if err != nil {
		// 删除已上传的文件
		os.Remove(filePath)
		http.Error(w, "Failed to save to database", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Warning: failed to get last insert id: %v", err)
		id = 0
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Music uploaded successfully",
		"id":      id,
	})
}

// MusicPlaylistHandler 获取音乐播放列表
func MusicPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	database := db.GetDB()
	query := `SELECT id, title, artist, file_path, file_name, duration, cover_image, created_at
	          FROM music_tracks ORDER BY created_at DESC`

	rows, err := database.Query(query)
	if err != nil {
		http.Error(w, "Failed to fetch playlist", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	playlist := make([]MusicTrack, 0)
	for rows.Next() {
		var track MusicTrack
		err := rows.Scan(&track.ID, &track.Title, &track.Artist, &track.FilePath, &track.FileName, &track.Duration, &track.CoverImage, &track.CreatedAt)
		if err != nil {
			continue
		}
		playlist = append(playlist, track)
	}

	json.NewEncoder(w).Encode(playlist)
}

// MusicDeleteHandler 删除音乐文件
func MusicDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 中提取 ID
	// 注意：路由已经通过 http.StripPrefix("/api", apiMux) 处理，所以 r.URL.Path 是 /music/3
	path := strings.TrimPrefix(r.URL.Path, "/music/")
	var id int
	_, err := fmt.Sscanf(path, "%d", &id)
	if err != nil {
		http.Error(w, "Invalid music ID", http.StatusBadRequest)
		return
	}

	database := db.GetDB()

	// 获取音乐文件路径
	var filePath string
	err = database.QueryRow("SELECT file_path FROM music_tracks WHERE id = ?", id).Scan(&filePath)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Music not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to fetch music", http.StatusInternalServerError)
		}
		return
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		// 记录错误但继续删除数据库记录
		fmt.Printf("Warning: Failed to delete file %s: %v\n", filePath, err)
	}

	// 从数据库删除
	_, err = database.Exec("DELETE FROM music_tracks WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete from database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Music deleted successfully",
	})
}

// MusicFileHandler 处理音乐文件访问
func MusicFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 中提取文件名
	fileName := strings.TrimPrefix(r.URL.Path, "/music/")
	if fileName == "" {
		http.Error(w, "File not specified", http.StatusBadRequest)
		return
	}

	// 使用当前工作目录
	filePath := filepath.Join("music", fileName)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 设置响应头
	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	// 发送文件
	http.ServeFile(w, r, filePath)
}

// MusicUpdateCoverHandler 更新音乐封面
func MusicUpdateCoverHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 中提取 ID
	path := strings.TrimPrefix(r.URL.Path, "/music/")
	var id int
	_, err := fmt.Sscanf(path, "%d", &id)
	if err != nil {
		http.Error(w, "Invalid music ID", http.StatusBadRequest)
		return
	}

	// 获取音乐文件信息
	database := db.GetDB()
	var oldFileName, oldFilePath, oldCoverImage string
	err = database.QueryRow("SELECT file_name, file_path, cover_image FROM music_tracks WHERE id = ?", id).Scan(&oldFileName, &oldFilePath, &oldCoverImage)
	if err != nil {
		http.Error(w, "Music not found", http.StatusNotFound)
		return
	}

	// 解析 multipart form
	err = r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// 获取封面文件
	coverFile, coverHandler, err := r.FormFile("cover")
	if err != nil {
		http.Error(w, "Failed to get cover file", http.StatusBadRequest)
		return
	}
	defer coverFile.Close()

	// 验证封面文件类型
	coverAllowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	coverContentType := coverHandler.Header.Get("Content-Type")
	if !coverAllowedTypes[coverContentType] {
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	// 检查文件名是否已包含时间戳
	timestamp := time.Now().Unix()
	var newFileName string
	var newFilePath string
	var needsRename bool

	// 检查文件名格式：数字_文件名.扩展名
	var existingTimestamp int64
	matched, _ := fmt.Sscanf(oldFileName, "%d_", &existingTimestamp)
	if matched != 1 {
		// 文件名不包含时间戳，需要重命名
		ext := filepath.Ext(oldFileName)
		baseName := strings.TrimSuffix(oldFileName, ext)

		// 使用 strings.Builder 优化字符串拼接
		var fileNameBuilder strings.Builder
		fileNameBuilder.Grow(len(baseName) + len(ext) + 20) // 预分配足够容量
		fileNameBuilder.WriteString(strconv.FormatInt(timestamp, 10))
		fileNameBuilder.WriteString("_")
		fileNameBuilder.WriteString(baseName)
		fileNameBuilder.WriteString(ext)
		newFileName = fileNameBuilder.String()

		newFilePath = filepath.Join("music", newFileName)
		needsRename = true
	} else {
		// 文件名已包含时间戳，使用现有时间戳
		timestamp = existingTimestamp
		newFileName = oldFileName
		newFilePath = oldFilePath
		needsRename = false
	}

	// 创建封面目录
	coverDir := filepath.Join("music", "covers")
	if err := os.MkdirAll(coverDir, 0755); err != nil {
		http.Error(w, "Failed to create cover directory", http.StatusInternalServerError)
		return
	}

	// 生成封面文件名（使用相同的时间戳）
	coverExt := filepath.Ext(coverHandler.Filename)

	var coverNameBuilder strings.Builder
	coverNameBuilder.Grow(len(coverExt) + 20) // 预分配足够容量
	coverNameBuilder.WriteString(strconv.FormatInt(timestamp, 10))
	coverNameBuilder.WriteString("_cover")
	coverNameBuilder.WriteString(coverExt)
	coverUniqueName := coverNameBuilder.String()

	coverPath := filepath.Join(coverDir, coverUniqueName)

	// 保存封面文件
	coverDst, err := os.Create(coverPath)
	if err != nil {
		http.Error(w, "Failed to create cover file", http.StatusInternalServerError)
		return
	}
	defer coverDst.Close()

	if _, err := io.Copy(coverDst, coverFile); err != nil {
		http.Error(w, "Failed to save cover file", http.StatusInternalServerError)
		return
	}

	// 如果需要重命名音乐文件
	if needsRename {
		// 重命名文件
		if err := os.Rename(oldFilePath, newFilePath); err != nil {
			// 删除已上传的封面文件
			os.Remove(coverPath)
			http.Error(w, "Failed to rename music file", http.StatusInternalServerError)
			return
		}
		fmt.Printf("Renamed music file: %s -> %s\n", oldFileName, newFileName)
	}

	// 删除旧封面文件（如果存在）
	if oldCoverImage != "" {
		oldCoverPath := filepath.Join(".", oldCoverImage)
		if _, err := os.Stat(oldCoverPath); err == nil {
			os.Remove(oldCoverPath)
			fmt.Printf("Removed old cover: %s\n", oldCoverImage)
		}
	}

	// 更新数据库
	coverImage := "/music/covers/" + coverUniqueName
	_, err = database.Exec("UPDATE music_tracks SET file_name = ?, file_path = ?, cover_image = ? WHERE id = ?", 
		newFileName, newFilePath, coverImage, id)
	if err != nil {
		// 回滚：删除新上传的封面文件
		os.Remove(coverPath)
		// 如果重命名了文件，尝试恢复
		if needsRename {
			os.Rename(newFilePath, oldFilePath)
		}
		http.Error(w, "Failed to update database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Cover updated successfully",
		"cover":   coverImage,
		"renamed": needsRename,
	})
}

// MusicUpdateTitleHandler 更新音乐标题
func MusicUpdateTitleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 中提取 ID
	path := strings.TrimPrefix(r.URL.Path, "/music/")
	var id int
	_, err := fmt.Sscanf(path, "%d", &id)
	if err != nil {
		http.Error(w, "Invalid music ID", http.StatusBadRequest)
		return
	}

	// 解析请求体
	var request struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证标题
	if strings.TrimSpace(request.Title) == "" {
		http.Error(w, "Title cannot be empty", http.StatusBadRequest)
		return
	}

	// 更新数据库
	database := db.GetDB()
	_, err = database.Exec("UPDATE music_tracks SET title = ? WHERE id = ?", request.Title, id)
	if err != nil {
		http.Error(w, "Failed to update database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Title updated successfully",
		"title":   request.Title,
	})
}
