package service

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var md goldmark.Markdown

// VideoNode 视频节点
type VideoNode struct {
	ast.BaseInline
	Src string
}

// KindVideo 视频节点类型
var KindVideo = ast.NewNodeKind("Video")

// Kind 实现 Node 接口
func (n *VideoNode) Kind() ast.NodeKind {
	return KindVideo
}

// Dump 实现 Node 接口
func (n *VideoNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

// VideoASTTransformer AST 转换器，将视频链接转换为视频节点
type VideoASTTransformer struct{}

// Transform 转换 AST
func (t *VideoASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	// 遍历所有节点
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// 检查是否是链接节点
		if link, ok := n.(*ast.Link); ok {
			// 获取链接的 URL
			url := string(link.Destination)
			// 检查是否是视频链接
			if strings.HasPrefix(url, "video:/") || strings.HasPrefix(url, "video://") {
				// 移除 video:/ 或 video:// 前缀
				src := strings.TrimPrefix(url, "video:/")
				src = strings.TrimPrefix(src, "video://")

				// 创建视频节点
				videoNode := &VideoNode{Src: src}

				// 替换链接节点为视频节点
				parent := link.Parent()
				if parent != nil {
					parent.ReplaceChild(parent, link, videoNode)
				}
			}
		}
		return ast.WalkContinue, nil
	})
}

// VideoRenderer 视频渲染器
type VideoRenderer struct{}

// RegisterFuncs 注册渲染函数
func (r *VideoRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindVideo, r.renderVideo)
}

// renderVideo 渲染视频节点
func (r *VideoRenderer) renderVideo(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*VideoNode)
	
	// 确保路径以 / 开头，避免相对路径拼接问题
	src := n.Src
	if !strings.HasPrefix(src, "/") {
		src = "/" + src
	}
	
	videoType := "video/mp4"
	if strings.HasSuffix(src, ".webm") {
		videoType = "video/webm"
	} else if strings.HasSuffix(src, ".ogg") {
		videoType = "video/ogg"
	}

	fmt.Fprintf(w, `<video controls style="max-width: 100%%; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.1);"><source src="%s" type="%s">您的浏览器不支持视频播放。</video>`, src, videoType)
	return ast.WalkContinue, nil
}

// VideoExtension 视频扩展
type VideoExtension struct{}

// Extend 扩展 Goldmark
func (e *VideoExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(&VideoASTTransformer{}, 100)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(&VideoRenderer{}, 100)))
}

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.NewHighlighting(
				highlighting.WithStyle("github"),
				highlighting.WithCSSWriter(nil),
			),
			&VideoExtension{}, // 添加视频扩展
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // 允许原始 HTML，不转义实体
		),
	)
}

// MarkdownDocument 表示一个 markdown 文档
type MarkdownDocument struct {
	Title           string
	Content         string
	OriginalContent string
	Path            string
	CreatedAt       time.Time
}

// ParseMarkdownFile 解析 markdown 文件
func ParseMarkdownFile(path string) (*MarkdownDocument, error) {
	// 读取文件内容
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 提取标题（第一个 # 开头的行）
	title := extractTitle(string(content))

	// 转换 markdown 为 HTML
	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return nil, fmt.Errorf("failed to convert markdown: %w", err)
	}

	// 获取文件创建时间
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &MarkdownDocument{
		Title:           title,
		Content:         buf.String(),
		OriginalContent: string(content),
		Path:            path,
		CreatedAt:       fileInfo.ModTime(),
	}, nil
}

// extractTitle 从 markdown 内容中提取标题
func extractTitle(content string) string {
	lines := bytes.Split([]byte(content), []byte("\n"))
	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, []byte("# ")) {
			return string(bytes.TrimPrefix(trimmed, []byte("# ")))
		}
	}
	return "未命名文档"
}

// GetMarkdownPath 根据 URL 路径构建 markdown 文件路径
// 路径格式: /passage/:year/:month/:day/:name
// 文件路径: markdown/:year/:month/:day/:name.md
// 如果直接匹配失败，会尝试根据日期目录下的文件标题进行模糊匹配
func GetMarkdownPath(urlPath string) (string, error) {
	// 移除 /passage 前缀
	path := filepath.Clean(urlPath)
	
	// 检查是否是有效的 passage 路径
	if path == "/passage" || path == "/passage/" {
		return "", fmt.Errorf("invalid passage path")
	}
	
	// 移除 /passage 前缀
	if len(path) > len("/passage") && path[:len("/passage")] == "/passage" {
		path = path[len("/passage"):]
	}
	
	// 清理路径，移除前导和尾随的斜杠
	path = strings.Trim(path, "/")
	if path == "" {
		return "", fmt.Errorf("invalid passage path")
	}
	
	// 分离日期部分和标题部分
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid passage path format")
	}
	
	year, month, day := parts[0], parts[1], parts[2]
	titleFromURL := strings.Join(parts[3:], "/")
	
	// 对 URL 编码的标题进行解码
	decodedTitle, err := url.QueryUnescape(titleFromURL)
	if err == nil {
		titleFromURL = decodedTitle
	}
	
	// 构建 markdown 文件路径
	dateDir := filepath.Join("markdown", year, month, day)
	markdownPath := filepath.Join(dateDir, titleFromURL) + ".md"

	// 检查文件是否存在
	if _, err := os.Stat(markdownPath); err == nil {
		return markdownPath, nil
	}
	
	// 如果直接匹配失败，尝试根据标题模糊匹配
	// 读取该日期目录下的所有 .md 文件
	files, err := os.ReadDir(dateDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}
	
	// 尝试匹配文件
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".md" {
			continue
		}
		
		filePath := filepath.Join(dateDir, file.Name())
		
		// 读取文件内容，提取标题
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		
		fileTitle := extractTitle(string(content))
		
		// 比较标题（去除空格和特殊字符后比较）
		if normalizeTitle(fileTitle) == normalizeTitle(titleFromURL) {
			return filePath, nil
		}
	}
	
	return "", fmt.Errorf("markdown file not found: %s", markdownPath)
}

// normalizeTitle 标准化标题用于比较
func normalizeTitle(title string) string {
	var builder strings.Builder
	builder.Grow(len(title)) // 预分配容量

	for _, r := range strings.ToLower(title) {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
		// 跳过空格和特殊字符
	}

	return builder.String()
}

// ListMarkdownFiles 列出 markdown 目录下的所有文件
func ListMarkdownFiles() ([]string, error) {
	var files []string

	err := filepath.Walk("markdown", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录和非 markdown 文件
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// ConvertToHTML 将 markdown 内容转换为 HTML
func ConvertToHTML(markdownContent []byte) (string, error) {
	var buf bytes.Buffer
	if err := md.Convert(markdownContent, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ConvertToHTMLWithOption 将 markdown 内容转换为 HTML，可选择是否移除第一行标题
func ConvertToHTMLWithOption(markdownContent []byte, showTitle bool) (string, error) {
	// 如果 showTitle 为 false，移除第一行标题
	if !showTitle {
		lines := bytes.Split(markdownContent, []byte("\n"))
		if len(lines) > 0 {
			// 检查第一行是否是标题（以 # 开头）
			firstLine := bytes.TrimSpace(lines[0])
			if bytes.HasPrefix(firstLine, []byte("#")) {
				// 移除第一行标题
				markdownContent = bytes.Join(lines[1:], []byte("\n"))
			}
		}
	}

	var buf bytes.Buffer
	if err := md.Convert(markdownContent, &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// UpdateMarkdownFile 更新现有的 markdown 文件，如果文件不存在则创建
func UpdateMarkdownFile(filePath, title, content string) error {
	// 确保文件路径是绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 构建完整的markdown内容
	fullContent := fmt.Sprintf("# %s\n\n%s", title, content)

	// 写入文件
	if err := os.WriteFile(absPath, []byte(fullContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetMarkdownFilePath 根据文章ID和创建时间获取markdown文件路径
func GetMarkdownFilePath(title string, createdAt time.Time) (string, error) {
	// 清理标题作为文件名
	cleanedTitle := SanitizeFilename(title)
	// 根据创建时间构建文件路径
	dateDir := createdAt.Format("2006/01/02")
	markdownPath := filepath.Join("markdown", dateDir, cleanedTitle+".md")
	return markdownPath, nil
}

// CalculateReadTime 计算文章阅读时长（按200字/分钟估算）
func CalculateReadTime(content string) int {
	// 移除HTML标签，只保留纯文本
	text := removeHTMLTags(content)
	
	// 移除空白字符
	text = strings.TrimSpace(text)
	
	// 计算字数
	wordCount := len([]rune(text))
	
	// 按200字/分钟计算阅读时长
	readTime := wordCount / 200
	
	// 至少显示1分钟
	if readTime == 0 {
		readTime = 1
	}
	
	return readTime
}

// removeHTMLTags 移除HTML标签
func removeHTMLTags(html string) string {
	var result strings.Builder
	result.Grow(len(html)) // 预分配容量，避免扩容
	inTag := false

	for _, r := range html {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// SanitizeFilename 清理文件名，移除或替换不安全的字符
func SanitizeFilename(name string) string {
	// 定义不允许的字符
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	
	// 替换不允许的字符
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}
	
	// 移除首尾空格
	result = strings.TrimSpace(result)
	
	// 如果结果为空，使用默认名称
	if result == "" {
		result = "未命名文档"
	}
	
	return result
}