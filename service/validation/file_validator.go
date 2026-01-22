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
		// 图片格式
		".jpg":   TypeImage,
		".jpeg":  TypeImage,
		".png":   TypeImage,
		".gif":   TypeImage,
		".webp":  TypeImage,
		".svg":   TypeImage,
		".bmp":   TypeImage,
		// 文档格式
		".pdf":   TypeDocument,
		".docx":  TypeDocument,
		".txt":   TypeDocument,
		".md":    TypeDocument,
		// 视频格式
		".mp4":   TypeVideo,
		".webm":  TypeVideo,
		".ogg":   TypeVideo,
		// 音频格式
		".mp3":   TypeAudio,
		".flac":  TypeAudio,
		".wav":   TypeAudio,
		".aac":   TypeAudio,
		// 压缩包格式
		".zip":   TypeArchive,
		".tar":   TypeArchive,
		".7z":    TypeArchive,
		".gz":    TypeArchive,
		".tar.gz": TypeArchive,
		".rar":   TypeArchive,
	}
	return supportedTypes[ext]
}

// GetContentType 获取MIME类型
func GetContentType(ext string) string {
	contentTypes := map[string]string{
		// 图片格式
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".png":   "image/png",
		".gif":   "image/gif",
		".webp":  "image/webp",
		".svg":   "image/svg+xml",
		".bmp":   "image/bmp",
		// 文档格式
		".pdf":   "application/pdf",
		".docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".txt":   "text/plain",
		".md":    "text/markdown",
		// 视频格式
		".mp4":   "video/mp4",
		".webm":  "video/webm",
		".ogg":   "video/ogg",
		// 音频格式
		".mp3":   "audio/mpeg",
		".flac":  "audio/flac",
		".wav":   "audio/wav",
		".aac":   "audio/aac",
		// 压缩包格式
		".zip":   "application/zip",
		".tar":   "application/x-tar",
		".7z":    "application/x-7z-compressed",
		".gz":    "application/gzip",
		".tar.gz": "application/x-tar",
		".rar":   "application/x-rar-compressed",
	}
	return contentTypes[ext]
}

// ValidateFileSecurity 验证文件安全性
func ValidateFileSecurity(content []byte, ext string) error {
	ext = strings.ToLower(ext)

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return ValidateImage(content)
	case ".svg":
		return ValidateSVG(content)
	case ".bmp":
		return ValidateBMP(content)
	case ".pdf":
		return ValidatePDF(content)
	case ".docx":
		return ValidateDOCX(content)
	case ".txt", ".md":
		return ValidateDocument(content)
	case ".zip":
		return ValidateZIP(content)
	case ".tar", ".gz", ".tar.gz":
		return ValidateTarGz(content, ext)
	case ".7z":
		return Validate7z(content)
	case ".rar":
		return ValidateRAR(content)
	case ".mp4", ".webm", ".ogg":
		return ValidateVideo(content)
	case ".mp3", ".flac", ".wav", ".aac":
		return ValidateAudio(content)
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

// ValidateVideo 验证视频文件
func ValidateVideo(content []byte) error {
	// 检查文件大小限制（2GB）
	if err := CheckFileSize(content, 2*1024*1024*1024); err != nil {
		return err
	}

	// 检查视频文件头
	if len(content) < 12 {
		return errors.New("视频文件太小，无法识别")
	}

	// MP4: 通常包含 "ftyp" box
	if content[4] == 'f' && content[5] == 't' && content[6] == 'y' && content[7] == 'p' {
		return nil
	}

	// WebM: 以EBML头开始
	if content[0] == 0x1A && content[1] == 0x45 && content[2] == 0xDF && content[3] == 0xA3 {
		return nil
	}

	// OGG: 以 "OggS" 开头
	if string(content[0:4]) == "OggS" {
		return nil
	}

	return errors.New("无效的视频文件格式")
}

// ValidateAudio 验证音频文件
func ValidateAudio(content []byte) error {
	// 检查文件大小限制（500MB）
	if err := CheckFileSize(content, 500*1024*1024); err != nil {
		return err
	}

	// 检查音频文件头
	if len(content) < 10 {
		return errors.New("音频文件太小，无法识别")
	}

	// MP3: 以ID3v2头开始或包含MP3帧同步
	if string(content[0:3]) == "ID3" {
		return nil
	}
	// MP3帧同步 (0xFF开头，第二字节的高5位是1)
	if content[0] == 0xFF && (content[1]&0xE0) == 0xE0 {
		return nil
	}

	// FLAC: 以 "fLaC" 开头
	if string(content[0:4]) == "fLaC" {
		return nil
	}

	// WAV: 以 "RIFF" + 4 bytes + "WAVE" 开头
	if string(content[0:4]) == "RIFF" && string(content[8:12]) == "WAVE" {
		return nil
	}

	// AAC: 通常包含ADTS帧同步
	if content[0] == 0xFF && (content[1]&0xF0) == 0xF0 {
		return nil
	}

	return errors.New("无效的音频文件格式")
}