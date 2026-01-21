package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"myblog-gogogo/auth"
)

// FileInfo 文件信息结构
type FileInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	IsDir        bool      `json:"is_dir"`
	ModifiedTime time.Time `json:"modified_time"`
	Extension    string    `json:"extension"`
}

// FileManagerHandler 文件管理API处理器
func FileManagerHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// 获取文件列表
		handleGetFiles(w, r)
	case http.MethodPost:
		// 上传文件
		handleFileUpload(w, r)
	case http.MethodPut:
		// 重命名文件
		handleRenameFile(w, r)
	case http.MethodDelete:
		// 删除文件
		handleDeleteFile(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求方法不允许",
		})
	}
}

// handleGetFiles 获取文件列表
func handleGetFiles(w http.ResponseWriter, r *http.Request) {
	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	// 验证并获取安全路径
	safePath, err := validatePath(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 读取目录内容
	entries, err := os.ReadDir(safePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "读取目录失败",
			"error":   err.Error(),
		})
		return
	}

	// 构建文件信息列表（初始化为空切片，避免返回 nil）
	files := make([]FileInfo, 0)
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		ext := ""
		if !entry.IsDir() {
			ext = strings.ToLower(filepath.Ext(entry.Name()))
		}

		files = append(files, FileInfo{
			Name:         entry.Name(),
			Path:         safePath,
			Size:         info.Size(),
			IsDir:        entry.IsDir(),
			ModifiedTime: info.ModTime(),
			Extension:    ext,
		})
	}

	// 获取父目录路径
	parentPath := getParentPath(safePath)

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"files":       files,
			"currentPath": safePath,
			"parentPath":  parentPath,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleFileUpload 处理文件上传
func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	// 获取目标路径
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}

	// 验证并获取安全路径
	safePath, err := validatePath(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 解析表单数据
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "解析表单失败",
		})
		return
	}

	// 获取上传的文件
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "获取文件失败",
		})
		return
	}
	defer file.Close()

	// 检查文件大小（限制50MB）
	if header.Size > 50*1024*1024 {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "文件过大，最大支持50MB",
		})
		return
	}

	// 创建目标文件
	filePath := filepath.Join(safePath, header.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "创建文件失败",
			"error":   err.Error(),
		})
		return
	}
	defer dst.Close()

	// 复制文件内容
	if _, err := io.Copy(dst, file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "保存文件失败",
			"error":   err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "上传成功",
		"data": map[string]interface{}{
			"filename": header.Filename,
			"path":     filePath,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleRenameFile 处理文件重命名
func handleRenameFile(w http.ResponseWriter, r *http.Request) {
	// 解析请求体
	var req struct {
		OldPath string `json:"old_path"`
		NewName string `json:"new_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求数据格式错误",
		})
		return
	}

	// 验证路径
	oldSafePath, err := validatePath(req.OldPath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 验证新文件名
	if req.NewName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "新文件名不能为空",
		})
		return
	}

	// 检查新文件名是否包含非法字符
	if strings.ContainsAny(req.NewName, "/\\:*?\"<>|") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "文件名包含非法字符",
		})
		return
	}

	// 构建新路径
	dir := filepath.Dir(oldSafePath)
	newSafePath := filepath.Join(dir, req.NewName)

	// 重命名文件
	if err := os.Rename(oldSafePath, newSafePath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "重命名失败",
			"error":   err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "重命名成功",
		"data": map[string]interface{}{
			"old_path": oldSafePath,
			"new_path": newSafePath,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// handleDeleteFile 处理文件删除
func handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "缺少路径参数",
		})
		return
	}

	// 验证路径
	safePath, err := validatePath(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 检查文件是否存在
	info, err := os.Stat(safePath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "文件不存在",
		})
		return
	}

	// 如果是目录，递归删除
	var deleteErr error
	if info.IsDir() {
		deleteErr = os.RemoveAll(safePath)
	} else {
		deleteErr = os.Remove(safePath)
	}

	if deleteErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "删除失败",
			"error":   deleteErr.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "删除成功",
	}

	json.NewEncoder(w).Encode(response)
}

// validatePath 验证路径安全，防止路径穿越
func validatePath(userPath string) (string, error) {
	// 规范化路径
	cleanPath := filepath.Clean(userPath)

	// 如果路径是相对路径，转换为绝对路径
	if !filepath.IsAbs(cleanPath) {
		cleanPath = filepath.Join(".", cleanPath)
	} else {
		// 如果是绝对路径，去掉开头的 /
		cleanPath = strings.TrimPrefix(cleanPath, "/")
	}

	// 使用当前工作目录
	wd := "."

	// 构建完整路径
	fullPath := filepath.Join(wd, cleanPath)

	// 解析绝对路径
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("路径解析失败")
	}

	// 获取工作目录的绝对路径
	absWd, err := filepath.Abs(wd)
	if err != nil {
		return "", fmt.Errorf("工作目录解析失败")
	}

	// 检查路径是否在工作目录或允许的子目录中
	// 允许的根目录: ./img, ./markdown, ./attachments 和 ./music
	allowedDirs := []string{
		filepath.Join(absWd, "img"),
		filepath.Join(absWd, "markdown"),
		filepath.Join(absWd, "attachments"),
		filepath.Join(absWd, "music"),
	}

	// 检查路径是否在允许的目录中
	isAllowed := false
	for _, allowedDir := range allowedDirs {
		// 确保路径以允许的目录开头，并且后面跟着路径分隔符或者完全匹配
		if absPath == allowedDir || strings.HasPrefix(absPath, allowedDir+string(filepath.Separator)) {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return "", fmt.Errorf("访问被拒绝：路径超出允许范围")
	}

	// 返回规范化后的相对路径（去掉开头的 ./）
	relPath := strings.TrimPrefix(fullPath, "."+string(filepath.Separator))
	if relPath == "" {
		relPath = cleanPath
	}

	return relPath, nil
}

// getParentPath 获取父目录路径
func getParentPath(path string) string {
	if path == "." || path == "/" {
		return ""
	}

	parent := filepath.Dir(path)
	if parent == "." {
		return ""
	}

	return parent
}

// FileDownloadHandler 文件下载处理器
func FileDownloadHandler(w http.ResponseWriter, r *http.Request) {
	// 获取路径参数
	path := r.URL.Query().Get("path")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "缺少路径参数",
		})
		return
	}

	// 验证路径
	safePath, err := validatePath(path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 检查文件是否存在
	info, err := os.Stat(safePath)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "文件不存在",
		})
		return
	}

	// 如果是目录，不允许下载
	if info.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "无法下载目录",
		})
		return
	}

	// 验证token（支持URL参数、Authorization头或cookie）
	token := r.URL.Query().Get("token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// 去掉Bearer前缀
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 如果 Authorization 头中的 token 为空，尝试从 cookie 获取
	if token == "" {
		cookie, err := r.Cookie("auth_token")
		if err == nil {
			token = cookie.Value
		}
	}

	if token == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "未授权访问",
		})
		return
	}

	// 验证token
	claims, err := auth.ValidateToken(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Token无效",
		})
		return
	}

	// 检查是否为管理员
	if claims.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "权限不足",
		})
		return
	}

	// 设置响应头
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", info.Name()))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))

	// 打开文件
	file, err := os.Open(safePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "打开文件失败",
		})
		return
	}
	defer file.Close()

	// 流式传输文件
	io.Copy(w, file)
}

// CreateDirectoryHandler 创建目录处理器
func CreateDirectoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求方法不允许",
		})
		return
	}

	// 解析请求体
	var req struct {
		Path     string `json:"path"`
		DirName  string `json:"dir_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "请求数据格式错误",
		})
		return
	}

	// 验证路径
	safePath, err := validatePath(req.Path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 验证目录名
	if req.DirName == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "目录名不能为空",
		})
		return
	}

	// 检查目录名是否包含非法字符
	if strings.ContainsAny(req.DirName, "/\\:*?\"<>|") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "目录名包含非法字符",
		})
		return
	}

	// 创建目录
	newDirPath := filepath.Join(safePath, req.DirName)
	if err := os.MkdirAll(newDirPath, 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "创建目录失败",
			"error":   err.Error(),
		})
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "目录创建成功",
		"data": map[string]interface{}{
			"path": newDirPath,
		},
	}

	json.NewEncoder(w).Encode(response)
}