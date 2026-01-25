package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"myblog-gogogo/db"
)

// UploadService 上传服务
type UploadService struct {
	syncService *SyncService
}

// NewUploadService 创建上传服务
func NewUploadService() *UploadService {
	repo := db.GetPassageRepository()
	return &UploadService{
		syncService: NewSyncService(repo),
	}
}

// UploadResult 上传结果
type UploadResult struct {
	Type     string `json:"type"`     // image 或 markdown
	Path     string `json:"path"`     // 文件路径
	URL      string `json:"url"`      // 访问URL
	Size     int64  `json:"size"`     // 文件大小
	FileName string `json:"fileName"` // 原始文件名
}

// HandleUpload 处理文件上传
func (s *UploadService) HandleUpload(file multipart.File, header *multipart.FileHeader, year, month, day, tags string) (*UploadResult, error) {
	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 根据文件类型处理
	ext := strings.ToLower(filepath.Ext(header.Filename))
	
	switch ext {
	case ".md":
		return s.handleMarkdownUpload(content, header, year, month, day, tags)
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg":
		return s.handleImageUpload(content, header)
	default:
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}
}

// handleMarkdownUpload 处理 markdown 文件上传
func (s *UploadService) handleMarkdownUpload(content []byte, header *multipart.FileHeader, year, month, day, tags string) (*UploadResult, error) {
	// 如果没有提供日期，使用当前日期
	if year == "" || month == "" || day == "" {
		now := time.Now()
		year = strconv.Itoa(now.Year())
		month = fmt.Sprintf("%02d", now.Month())
		day = fmt.Sprintf("%02d", now.Day())
	}

	// 格式化月份和日（确保是两位数）
	if len(month) == 1 {
		month = "0" + month
	}
	if len(day) == 1 {
		day = "0" + day
	}

	// 创建目录路径
	dateDir := filepath.Join("markdown", year, month, day)
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成文件名（使用原始文件名）
	fileName := header.Filename
	filePath := filepath.Join(dateDir, fileName)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件已存在，添加时间戳
		timestamp := time.Now().Format("20060102-150405")
		baseName := strings.TrimSuffix(fileName, ".md")

		// 使用 strings.Builder 优化文件名生成
		var nameBuilder strings.Builder
		nameBuilder.Grow(len(baseName) + len(timestamp) + 5) // +5 for "-.md"
		nameBuilder.WriteString(baseName)
		nameBuilder.WriteString("-")
		nameBuilder.WriteString(timestamp)
		nameBuilder.WriteString(".md")
		fileName = nameBuilder.String()

		filePath = filepath.Join(dateDir, fileName)
	}

	// 写入文件
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	// 同步到数据库，传递标签参数
	if err := s.syncService.SyncFile(filePath, tags); err != nil {
		// 记录错误但不返回，因为文件已经成功上传
		fmt.Printf("同步到数据库失败: %v\n", err)
	}

	// 构建访问URL
	var urlBuilder strings.Builder
	urlBuilder.Grow(len("/passage///") + len(year) + len(month) + len(day) + len(fileName))
	urlBuilder.WriteString("/passage/")
	urlBuilder.WriteString(year)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(month)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(day)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(strings.TrimSuffix(fileName, ".md"))
	url := urlBuilder.String()

	return &UploadResult{
		Type:     "markdown",
		Path:     filePath,
		URL:      url,
		Size:     int64(len(content)),
		FileName: header.Filename,
	}, nil
}

// handleImageUpload 处理图片文件上传
func (s *UploadService) handleImageUpload(content []byte, header *multipart.FileHeader) (*UploadResult, error) {
	// 创建 img 目录（如果不存在）
	if err := os.MkdirAll("img", 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成文件名（使用原始文件名）
	fileName := header.Filename
	filePath := filepath.Join("img", fileName)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件已存在，添加时间戳
		timestamp := time.Now().Format("20060102-150405")
		ext := filepath.Ext(fileName)
		baseName := strings.TrimSuffix(fileName, ext)

		// 使用 strings.Builder 优化文件名生成
		var nameBuilder strings.Builder
		nameBuilder.Grow(len(baseName) + len(timestamp) + len(ext) + 1) // +1 for hyphen
		nameBuilder.WriteString(baseName)
		nameBuilder.WriteString("-")
		nameBuilder.WriteString(timestamp)
		nameBuilder.WriteString(ext)
		fileName = nameBuilder.String()

		filePath = filepath.Join("img", fileName)
	}

	// 写入文件
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	// 构建访问URL
	var urlBuilder strings.Builder
	urlBuilder.Grow(len("/img/") + len(fileName))
	urlBuilder.WriteString("/img/")
	urlBuilder.WriteString(fileName)
	url := urlBuilder.String()

	return &UploadResult{
		Type:     "image",
		Path:     filePath,
		URL:      url,
		Size:     int64(len(content)),
		FileName: header.Filename,
	}, nil
}