package attachment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"myblog-gogogo/db/models"
	"myblog-gogogo/service/validation"
)

// AttachmentInfo 附件信息（用于API返回）
type AttachmentInfo struct {
	ID          int    `json:"id"`
	FileName    string `json:"fileName"`
	StoredName  string `json:"storedName"`
	URL         string `json:"url"`
	FileType    string `json:"fileType"`
	ContentType string `json:"contentType"`
	FileSize    int64  `json:"fileSize"`
	UploadedAt  string `json:"uploadedAt"`
}

// GetByDate 根据文章日期获取附件列表
func (s *Service) GetByDate(year, month, day string) ([]*AttachmentInfo, error) {
	attachmentDir := filepath.Join("attachments", year, month, day)

	if _, err := os.Stat(attachmentDir); os.IsNotExist(err) {
		return []*AttachmentInfo{}, nil
	}

	entries, err := os.ReadDir(attachmentDir)
	if err != nil {
		return nil, fmt.Errorf("读取附件目录失败: %w", err)
	}

	var attachments []*AttachmentInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		fileType := string(validation.GetFileType(ext))
		contentType := validation.GetContentType(ext)

		url := fmt.Sprintf("/attachments/%s/%s/%s/%s", year, month, day, entry.Name())

		attachments = append(attachments, &AttachmentInfo{
			FileName:    entry.Name(),
			StoredName:  entry.Name(),
			URL:         url,
			FileType:    fileType,
			ContentType: contentType,
			FileSize:    fileInfo.Size(),
			UploadedAt:  fileInfo.ModTime().Format("2006-01-02 15:04:05"),
		})
	}

	return attachments, nil
}

// SyncToDB 扫描附件目录并同步文件到数据库
func (s *Service) SyncToDB() error {
	attachmentsDir := "attachments"
	if _, err := os.Stat(attachmentsDir); os.IsNotExist(err) {
		return nil
	}

	allAttachments, _, err := s.repo.GetAll(10000, 0)
	if err != nil {
		return fmt.Errorf("获取数据库附件记录失败: %w", err)
	}

	recordedFiles := make(map[string]bool)
	for _, att := range allAttachments {
		recordedFiles[att.FilePath] = true
	}

	err = filepath.Walk(attachmentsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if recordedFiles[path] {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		fileType := validation.GetFileType(ext)

		if fileType == "" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("读取文件失败: %s, 错误: %v\n", path, err)
			return nil
		}

		if err := validation.ValidateFileSecurity(content, ext); err != nil {
			fmt.Printf("文件安全验证失败: %s, 错误: %v\n", path, err)
			return nil
		}

		pathParts := strings.Split(path, string(filepath.Separator))
		if len(pathParts) < 4 {
			return nil
		}

		fileName := info.Name()
		storedName := fileName
		originalFileName := fileName

		baseName := strings.TrimSuffix(fileName, ext)
		if idx := strings.LastIndex(baseName, "-"); idx > 0 {
			timestampPart := baseName[idx+1:]
			if len(timestampPart) == 14 {
				_, err := time.Parse("20060102-150405", timestampPart)
				if err == nil {
					originalFileName = baseName[:idx] + ext
				}
			}
		}

		uploadedAt := info.ModTime()

		attachment := &models.Attachment{
			FileName:      originalFileName,
			StoredName:    storedName,
			FilePath:      path,
			FileType:      string(fileType),
			ContentType:   validation.GetContentType(ext),
			FileSize:      info.Size(),
			PassageID:     nil,
			Visibility:    "public",
			ShowInPassage: false,
			UploadedAt:    uploadedAt,
		}

		if err := s.repo.Create(attachment); err != nil {
			fmt.Printf("保存附件到数据库失败: %s, 错误: %v\n", path, err)
			return nil
		}

		fmt.Printf("已同步附件到数据库: %s (ID: %d)\n", path, attachment.ID)

		return nil
	})

	if err != nil {
		return fmt.Errorf("扫描附件目录失败: %w", err)
	}

	return nil
}