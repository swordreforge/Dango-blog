package attachment

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
	"myblog-gogogo/db/repositories"
	"myblog-gogogo/service/kafka"
	"myblog-gogogo/service/validation"
)

// Service 附件服务
type Service struct {
	repo repositories.AttachmentRepository
}

// NewService 创建附件服务
func NewService() *Service {
	return &Service{
		repo: db.GetAttachmentRepository(),
	}
}

// UploadResult 上传结果
type UploadResult struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	FileName    string `json:"fileName"`
	StoredName  string `json:"storedName"`
	Path        string `json:"path"`
	URL         string `json:"url"`
	Size        int64  `json:"size"`
	PassageID   int    `json:"passageId"`
	ContentType string `json:"contentType"`
}

// Upload 上传附件
func (s *Service) Upload(file multipart.File, header *multipart.FileHeader, passageID int) (*UploadResult, error) {
	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	// 获取文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))

	// 根据文件类型进行安全验证
	fileType := validation.GetFileType(ext)
	if fileType == "" {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	// 根据文件类型进行安全扫描
	if err := validation.ValidateFileSecurity(content, ext); err != nil {
		return nil, fmt.Errorf("文件安全验证失败: %w", err)
	}

	// 获取文章的创建日期（如果有关联文章）
	var articleDate time.Time
	if passageID > 0 {
		passageRepo := db.GetPassageRepository()
		passage, err := passageRepo.GetByID(passageID)
		if err == nil && passage != nil {
			articleDate = passage.CreatedAt
		}
	}

	// 如果没有关联文章或获取失败，使用当前时间
	if articleDate.IsZero() {
		articleDate = time.Now()
	}

	// 使用文章的创建日期生成存储路径
	year := strconv.Itoa(articleDate.Year())
	month := fmt.Sprintf("%02d", articleDate.Month())
	day := fmt.Sprintf("%02d", articleDate.Day())

	// 创建附件目录
	attachmentDir := filepath.Join("attachments", year, month, day)
	if err := os.MkdirAll(attachmentDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成唯一的文件名
	now := time.Now()
	timestamp := now.Format("20060102-150405")
	baseName := strings.TrimSuffix(header.Filename, ext)

	var nameBuilder strings.Builder
	nameBuilder.Grow(len(baseName) + len(timestamp) + len(ext) + 2)
	nameBuilder.WriteString(baseName)
	nameBuilder.WriteString("-")
	nameBuilder.WriteString(timestamp)
	nameBuilder.WriteString(ext)
	storedName := nameBuilder.String()

	filePath := filepath.Join(attachmentDir, storedName)

	// 写入文件
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	// 保存到数据库
	var passageIDPtr *int
	if passageID > 0 {
		passageIDPtr = &passageID
	}

	attachment := &models.Attachment{
		FileName:      header.Filename,
		StoredName:    storedName,
		FilePath:      filePath,
		FileType:      string(fileType),
		ContentType:   validation.GetContentType(ext),
		FileSize:      int64(len(content)),
		PassageID:     passageIDPtr,
		Visibility:    "public",
		ShowInPassage: true,
		UploadedAt:    now,
	}

	if err := s.repo.Create(attachment); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("保存到数据库失败: %w", err)
	}

	// 发布上传事件
	ctx := context.Background()
	if err := kafka.PublishAttachmentUploadEvent(ctx, attachment.ID, attachment.FileName, attachment.FileSize, attachment.FileType, passageID); err != nil {
		// 事件发布失败不影响上传流程
		fmt.Printf("发布附件上传事件失败: %v\n", err)
	}

	// 构建访问URL
	var urlBuilder strings.Builder
	urlBuilder.Grow(len("/attachments///") + len(year) + len(month) + len(day) + len(storedName))
	urlBuilder.WriteString("/attachments/")
	urlBuilder.WriteString(year)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(month)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(day)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(storedName)
	url := urlBuilder.String()

	return &UploadResult{
		ID:          attachment.ID,
		Type:        string(fileType),
		FileName:    header.Filename,
		StoredName:  storedName,
		Path:        filePath,
		URL:         url,
		Size:        int64(len(content)),
		PassageID:   passageID,
		ContentType: attachment.ContentType,
	}, nil
}

// List 获取附件列表
func (s *Service) List(passageID *int, limit, offset int) ([]*models.Attachment, int, error) {
	if passageID != nil {
		return s.repo.GetByPassageID(*passageID, limit, offset)
	}
	return s.repo.GetAll(limit, offset)
}

// Delete 删除附件
func (s *Service) Delete(id int) error {
	// 获取附件信息
	attachment, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("获取附件信息失败: %w", err)
	}

	// 发布删除事件
	ctx := context.Background()
	passageID := 0
	if attachment.PassageID != nil {
		passageID = *attachment.PassageID
	}
	if err := kafka.PublishAttachmentDeleteEvent(ctx, id, attachment.FileName, passageID); err != nil {
		fmt.Printf("发布附件删除事件失败: %v\n", err)
	}

	// 删除数据库记录
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("删除数据库记录失败: %w", err)
	}

	// 删除文件
	if err := os.Remove(attachment.FilePath); err != nil {
		fmt.Printf("删除文件失败: %v\n", err)
	}

	return nil
}

// GetPath 获取附件路径
func (s *Service) GetPath(id int) (string, string, error) {
	attachment, err := s.repo.GetByID(id)
	if err != nil {
		return "", "", err
	}
	return attachment.FilePath, attachment.FileName, nil
}

// UpdateVisibility 更新附件可见性
func (s *Service) UpdateVisibility(id int, visibility string) error {
	attachment, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 发布更新事件
	ctx := context.Background()
	if err := kafka.PublishAttachmentUpdateEvent(ctx, id, visibility, attachment.ShowInPassage); err != nil {
		fmt.Printf("发布附件更新事件失败: %v\n", err)
	}

	return s.repo.UpdateVisibility(id, visibility, attachment.ShowInPassage)
}

// UpdateShowInPassage 更新附件是否在文章中显示
func (s *Service) UpdateShowInPassage(id int, show bool) error {
	attachment, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// 发布更新事件
	ctx := context.Background()
	if err := kafka.PublishAttachmentUpdateEvent(ctx, id, attachment.Visibility, show); err != nil {
		fmt.Printf("发布附件更新事件失败: %v\n", err)
	}

	return s.repo.UpdateVisibility(id, attachment.Visibility, show)
}
