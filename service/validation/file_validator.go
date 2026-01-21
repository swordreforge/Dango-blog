package validation

import (
	"errors"
	"fmt"
	"strings"
)

// FileValidator 文件验证器接口
type FileValidator interface {
	Validate(content []byte) error
	GetSupportedExtensions() []string
}

// FileType 文件类型
type FileType string

const (
	TypeImage    FileType = "image"
	TypeDocument FileType = "document"
	TypeVideo    FileType = "video"
	TypeAudio    FileType = "audio"
	TypeArchive  FileType = "archive"
)

// FileTypeInfo 文件类型信息
type FileTypeInfo struct {
	Type        FileType
	Extensions  []string
	ContentType string
}

// GetFileType 获取文件类型
func GetFileType(ext string) FileType {
	supportedTypes := map[string]FileType{
		".svg":   TypeImage,
		".bmp":   TypeImage,
		".pdf":   TypeDocument,
		".docx":  TypeDocument,
		".mp4":   TypeVideo,
		".mp3":   TypeAudio,
		".flac":  TypeAudio,
		".zip":   TypeArchive,
		".tar":   TypeArchive,
		".7z":    TypeArchive,
		".gz":    TypeArchive,
		".tar.gz": TypeArchive,
	}
	return supportedTypes[ext]
}

// GetContentType 获取MIME类型
func GetContentType(ext string) string {
	contentTypes := map[string]string{
		".svg":   "image/svg+xml",
		".bmp":   "image/bmp",
		".pdf":   "application/pdf",
		".docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".mp4":   "video/mp4",
		".mp3":   "audio/mpeg",
		".flac":  "audio/flac",
		".zip":   "application/zip",
		".tar":   "application/x-tar",
		".7z":    "application/x-7z-compressed",
		".gz":    "application/gzip",
		".tar.gz": "application/x-tar",
	}
	return contentTypes[ext]
}

// ValidateFileSecurity 验证文件安全性
func ValidateFileSecurity(content []byte, ext string) error {
	ext = strings.ToLower(ext)
	
	switch ext {
	case ".svg":
		return ValidateSVG(content)
	case ".bmp":
		return ValidateBMP(content)
	case ".pdf":
		return ValidatePDF(content)
	case ".docx":
		return ValidateDOCX(content)
	case ".zip":
		return ValidateZIP(content)
	case ".tar", ".gz", ".tar.gz":
		return ValidateTarGz(content, ext)
	case ".7z":
		return Validate7z(content)
	default:
		return nil
	}
}

// CheckDangerousPatterns 检查危险模式
func CheckDangerousPatterns(content string, patterns []string) error {
	for _, pattern := range patterns {
		if strings.Contains(content, pattern) {
			return fmt.Errorf("文件包含危险内容: %s", pattern)
		}
	}
	return nil
}

// CheckFileSize 检查文件大小
func CheckFileSize(content []byte, maxSize int64) error {
	if int64(len(content)) > maxSize {
		return errors.New("文件过大")
	}
	return nil
}