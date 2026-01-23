package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
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
	// 获取当前工作目录
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	markdownDir := filepath.Join(workingDir, "markdown")
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
	// 获取当前工作目录
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	markdownDir := filepath.Join(workingDir, "markdown")

	// 检查 markdown 目录是否存在
	if _, err := os.Stat(markdownDir); os.IsNotExist(err) {
		return fmt.Errorf("markdown directory does not exist: %s", markdownDir)
	}

	// 创建文件系统监控器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	// 递归添加 markdown 目录到监控
	if err := s.addWatchRecursive(watcher, markdownDir); err != nil {
		return fmt.Errorf("failed to add watch: %w", err)
	}

	fmt.Printf("Watching markdown directory: %s\n", markdownDir)

	// 开始监控
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// 只处理 .md 文件
			if filepath.Ext(event.Name) != ".md" {
				continue
			}

			// 处理不同的事件类型
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				fmt.Printf("File created: %s\n", event.Name)
				// 等待一小段时间，确保文件写入完成
				time.Sleep(100 * time.Millisecond)
				if err := s.SyncFile(event.Name); err != nil {
					fmt.Printf("Failed to sync created file: %v\n", err)
				}

			case event.Op&fsnotify.Write == fsnotify.Write:
				fmt.Printf("File modified: %s\n", event.Name)
				// 等待一小段时间，确保文件写入完成
				time.Sleep(100 * time.Millisecond)
				if err := s.SyncFile(event.Name); err != nil {
					fmt.Printf("Failed to sync modified file: %v\n", err)
				}

			case event.Op&fsnotify.Remove == fsnotify.Remove,
				event.Op&fsnotify.Rename == fsnotify.Rename:
				fmt.Printf("File removed/renamed: %s\n", event.Name)
				if err := s.removePassageByFilePath(event.Name); err != nil {
					fmt.Printf("Failed to remove passage: %v\n", err)
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

// addWatchRecursive 递归添加目录到监控
func (s *SyncService) addWatchRecursive(watcher *fsnotify.Watcher, dir string) error {
	// 添加当前目录
	if err := watcher.Add(dir); err != nil {
		return err
	}

	// 遍历子目录
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 如果是目录，添加到监控
		if info.IsDir() && path != dir {
			if err := watcher.Add(path); err != nil {
				return err
			}
		}

		return nil
	})
}

// removePassageByFilePath 通过文件路径删除文章
func (s *SyncService) removePassageByFilePath(filePath string) error {
	// 获取当前工作目录
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}
	markdownDir := filepath.Join(workingDir, "markdown")
	relativePath := strings.TrimPrefix(filePath, markdownDir+string(filepath.Separator))
	relativePath = strings.TrimSuffix(relativePath, ".md")

	// 查找文章
	existingPassage, err := s.findPassageByFilePath(relativePath)
	if err != nil {
		return err
	}

	if existingPassage != nil {
		// 删除文章
		if err := s.repo.Delete(existingPassage.ID); err != nil {
			return fmt.Errorf("failed to delete passage: %w", err)
		}
		fmt.Printf("Deleted passage: %s (from %s)\n", existingPassage.Title, relativePath)
	}

	return nil
}
