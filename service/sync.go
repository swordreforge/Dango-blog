package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"myblog-gogogo/db/models"
	"myblog-gogogo/db/repositories"
)

// SyncService 同步服务
type SyncService struct {
	repo repositories.PassageRepository
}

// NewSyncService 创建同步服务
func NewSyncService(repo repositories.PassageRepository) *SyncService {
	return &SyncService{
		repo: repo,
	}
}

// SyncAll 同步所有 markdown 文件到数据库
func (s *SyncService) SyncAll() error {
	files, err := ListMarkdownFiles()
	if err != nil {
		return fmt.Errorf("failed to list markdown files: %w", err)
	}

	for _, file := range files {
		if err := s.SyncFile(file); err != nil {
			// 记录错误但继续处理其他文件
			fmt.Printf("Failed to sync %s: %v\n", file, err)
		}
	}

	return nil
}

// SyncFile 同步单个 markdown 文件到数据库
func (s *SyncService) SyncFile(filePath string) error {
	// 解析 markdown 文件
	doc, err := ParseMarkdownFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse markdown file: %w", err)
	}

	// 从文件路径提取信息
	// 获取程序所在目录
	execPath, err := os.Executable()
	if err != nil {
		execPath = "."
	}
	execDir := filepath.Dir(execPath)
	markdownDir := filepath.Join(execDir, "markdown")
	relativePath := strings.TrimPrefix(filePath, markdownDir+string(filepath.Separator))
	relativePath = strings.TrimSuffix(relativePath, ".md")

	// 提取日期
	parts := strings.Split(relativePath, "/")
	var year, month, day string

	if len(parts) >= 3 {
		year = parts[0]
		month = parts[1]
		day = parts[2]
	}

	// 构建日期
	var createdAt time.Time
	if year != "" && month != "" && day != "" {
		createdAt, err = time.Parse("2006-01-02", fmt.Sprintf("%s-%s-%s", year, month, day))
		if err != nil {
			createdAt = time.Now()
		}
	} else {
		// 如果路径中没有日期信息，使用当前时间
		createdAt = time.Now()
	}

	// 生成摘要
	summary := s.extractSummary(doc.Content)

	// 生成标签
	tags := s.extractTags(relativePath)

	// 检查是否已存在（通过文件路径）
	existingPassage, err := s.findPassageByFilePath(relativePath)
	if err != nil {
		return fmt.Errorf("failed to find existing passage: %w", err)
	}

	if existingPassage != nil {
		// 更新现有文章
		existingPassage.Title = doc.Title
		existingPassage.Content = doc.Content
		existingPassage.OriginalContent = doc.OriginalContent
		existingPassage.Summary = summary
		existingPassage.Tags = tags
		existingPassage.Status = "published"
		existingPassage.FilePath = relativePath
		existingPassage.UpdatedAt = time.Now()

		if err := s.repo.Update(existingPassage); err != nil {
			return fmt.Errorf("failed to update passage: %w", err)
		}
		fmt.Printf("Updated passage: %s (from %s)\n", doc.Title, relativePath)
	} else {
		// 创建新文章
		passage := &models.Passage{
			Title:           doc.Title,
			Content:         doc.Content,
			OriginalContent: doc.OriginalContent,
			Summary:         summary,
			Author:          "Admin",
			Tags:            tags,
			Status:          "published",
			FilePath:        relativePath,
			CreatedAt:       createdAt,
			UpdatedAt:       time.Now(),
		}

		if err := s.repo.Create(passage); err != nil {
			return fmt.Errorf("failed to create passage: %w", err)
		}
		fmt.Printf("Created passage: %s (from %s)\n", doc.Title, relativePath)
	}

	return nil
}

// findPassageByFilePath 通过文件路径查找文章
func (s *SyncService) findPassageByFilePath(filePath string) (*models.Passage, error) {
	passages, err := s.repo.GetAll(1000, 0)
	if err != nil {
		return nil, err
	}

	for _, p := range passages {
		if p.FilePath == filePath {
			return &p, nil
		}
	}

	return nil, nil
}

// findPassageByTitleAndDate 通过标题和日期查找文章（保留用于向后兼容）
func (s *SyncService) findPassageByTitleAndDate(title string, date time.Time) (*models.Passage, error) {
	passages, err := s.repo.GetAll(1000, 0)
	if err != nil {
		return nil, err
	}

	for _, p := range passages {
		if p.Title == title && p.CreatedAt.Format("2006-01-02") == date.Format("2006-01-02") {
			return &p, nil
		}
	}

	return nil, nil
}

// extractSummary 从 HTML 内容中提取摘要
func (s *SyncService) extractSummary(htmlContent string) string {
	// 移除 HTML 标签
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(htmlContent, "")

	// 移除多余的空白
	text = strings.Join(strings.Fields(text), " ")

	// 转换为 rune 切片以正确处理中文字符
	runes := []rune(text)

	// 截取前 200 个字符（按字符数，不是字节数）
	if len(runes) > 200 {
		text = string(runes[:200]) + "..."
	}

	return text
}

// extractTags 从路径中提取标签
func (s *SyncService) extractTags(path string) string {
	parts := strings.Split(path, "/")
	
	// 使用年份和月份作为标签
	var tags []string
	if len(parts) >= 2 {
		tags = append(tags, parts[0])  // 年份
		tags = append(tags, parts[1])  // 月份
	}

	// 转换为 JSON 格式
	if len(tags) == 0 {
		return "[]"
	}

	// 使用 strings.Builder 优化字符串拼接
	var builder strings.Builder
	// 预分配容量：每个标签平均 10 字符 + 3 个引号/逗号
	builder.Grow(len(tags) * 13)
	builder.WriteString("[")

	for i, tag := range tags {
		if i > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(`"`)
		builder.WriteString(tag)
		builder.WriteString(`"`)
	}

	builder.WriteString("]")
	return builder.String()
}

// GetPassageByPath 通过路径获取文章
func (s *SyncService) GetPassageByPath(path string) (*models.Passage, error) {
	// 移除 .md 后缀
	path = strings.TrimSuffix(path, ".md")
	
	// 提取标题
	parts := strings.Split(path, "/")
	var title string
	if len(parts) > 0 {
		title = parts[len(parts)-1]
	}

	// 查找文章
	passages, err := s.repo.GetAll(1000, 0)
	if err != nil {
		return nil, err
	}

	for _, p := range passages {
		// 将标题转换为小写并替换空格和特殊字符
		sanitizedTitle := sanitizeTitle(p.Title)
		if sanitizedTitle == title {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("passage not found: %s", path)
}

// sanitizeTitle 清理标题用于路径匹配
func sanitizeTitle(title string) string {
	// 转换为小写
	title = strings.ToLower(title)
	
	// 替换空格为连字符
	title = strings.ReplaceAll(title, " ", "-")
	
	// 移除特殊字符
	re := regexp.MustCompile(`[^\w\-]`)
	title = re.ReplaceAllString(title, "")
	
	return title
}

// WatchAndSync 监控文件变化并自动同步
func (s *SyncService) WatchAndSync() error {
	// TODO: 实现文件监控功能
	// 可以使用 fsnotify 库来监控文件系统变化
	return nil
}
