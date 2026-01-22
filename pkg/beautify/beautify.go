package beautify

import (
	"fmt"
	"strings"
	"time"
)

// LogLevel 日志级别类型
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelSuccess
)

// TreeLogger 树状结构日志记录器
type TreeLogger struct {
	indentLevel int
	indentStr   string
	showTime    bool
}

// NewTreeLogger 创建新的树状日志记录器
func NewTreeLogger() *TreeLogger {
	return &TreeLogger{
		indentLevel: 0,
		indentStr:   "│  ",
		showTime:    true,
	}
}

// SetIndent 设置缩进字符串
func (l *TreeLogger) SetIndent(indent string) {
	l.indentStr = indent
}

// SetShowTime 设置是否显示时间
func (l *TreeLogger) SetShowTime(show bool) {
	l.showTime = show
}

// getIndent 获取当前缩进
func (l *TreeLogger) getIndent() string {
	return strings.Repeat(l.indentStr, l.indentLevel)
}

// getPrefix 获取前缀（带时间戳）
func (l *TreeLogger) getPrefix(level LogLevel) string {
	var prefix string
	if l.showTime {
		timestamp := time.Now().Format("15:04:05.000")
		prefix = fmt.Sprintf("[%s] ", timestamp)
	}

	// 添加级别标识
	switch level {
	case LevelDebug:
		prefix += "🔍 "
	case LevelInfo:
		prefix += "ℹ️  "
	case LevelWarn:
		prefix += "⚠️  "
	case LevelError:
		prefix += "❌ "
	case LevelSuccess:
		prefix += "✅ "
	}

	return prefix
}

// formatMessage 格式化消息
func (l *TreeLogger) formatMessage(level LogLevel, message string) string {
	return fmt.Sprintf("%s%s%s", l.getPrefix(level), l.getIndent(), message)
}

// Print 打印普通消息
func (l *TreeLogger) Print(message string) {
	fmt.Println(l.formatMessage(LevelInfo, message))
}

// Debug 打印调试消息
func (l *TreeLogger) Debug(message string) {
	fmt.Println(l.formatMessage(LevelDebug, message))
}

// Info 打印信息消息
func (l *TreeLogger) Info(message string) {
	fmt.Println(l.formatMessage(LevelInfo, message))
}

// Warn 打印警告消息
func (l *TreeLogger) Warn(message string) {
	fmt.Println(l.formatMessage(LevelWarn, message))
}

// Error 打印错误消息
func (l *TreeLogger) Error(message string) {
	fmt.Println(l.formatMessage(LevelError, message))
}

// Success 打印成功消息
func (l *TreeLogger) Success(message string) {
	fmt.Println(l.formatMessage(LevelSuccess, message))
}

// Printf 格式化打印
func (l *TreeLogger) Printf(format string, args ...interface{}) {
	fmt.Println(l.formatMessage(LevelInfo, fmt.Sprintf(format, args...)))
}

// Debugf 格式化打印调试消息
func (l *TreeLogger) Debugf(format string, args ...interface{}) {
	fmt.Println(l.formatMessage(LevelDebug, fmt.Sprintf(format, args...)))
}

// Warnf 格式化打印警告消息
func (l *TreeLogger) Warnf(format string, args ...interface{}) {
	fmt.Println(l.formatMessage(LevelWarn, fmt.Sprintf(format, args...)))
}

// Errorf 格式化打印错误消息
func (l *TreeLogger) Errorf(format string, args ...interface{}) {
	fmt.Println(l.formatMessage(LevelError, fmt.Sprintf(format, args...)))
}

// Successf 格式化打印成功消息
func (l *TreeLogger) Successf(format string, args ...interface{}) {
	fmt.Println(l.formatMessage(LevelSuccess, fmt.Sprintf(format, args...)))
}

// Indent 增加缩进层级
func (l *TreeLogger) Indent() {
	l.indentLevel++
}

// Outdent 减少缩进层级
func (l *TreeLogger) Outdent() {
	if l.indentLevel > 0 {
		l.indentLevel--
	}
}

// WithIndent 在指定缩进层级执行函数
func (l *TreeLogger) WithIndent(fn func()) {
	l.Indent()
	defer l.Outdent()
	fn()
}

// Branch 打印分支节点
func (l *TreeLogger) Branch(message string) {
	indent := l.getIndent()
	// 替换最后一个缩进为分支符号
	if len(indent) > 0 {
		indent = indent[:len(indent)-len(l.indentStr)] + "├─ "
	} else {
		indent = "├─ "
	}
	fmt.Printf("%s%s%s\n", l.getPrefix(LevelInfo), indent, message)
}

// Leaf 打印叶子节点
func (l *TreeLogger) Leaf(message string) {
	indent := l.getIndent()
	// 替换最后一个缩进为叶子符号
	if len(indent) > 0 {
		indent = indent[:len(indent)-len(l.indentStr)] + "└─ "
	} else {
		indent = "└─ "
	}
	fmt.Printf("%s%s%s\n", l.getPrefix(LevelInfo), indent, message)
}

// SuccessLeaf 打印成功的叶子节点
func (l *TreeLogger) SuccessLeaf(message string) {
	indent := l.getIndent()
	if len(indent) > 0 {
		indent = indent[:len(indent)-len(l.indentStr)] + "└─ "
	} else {
		indent = "└─ "
	}
	fmt.Printf("%s%s%s\n", l.getPrefix(LevelSuccess), indent, message)
}

// ErrorLeaf 打印错误的叶子节点
func (l *TreeLogger) ErrorLeaf(message string) {
	indent := l.getIndent()
	if len(indent) > 0 {
		indent = indent[:len(indent)-len(l.indentStr)] + "└─ "
	} else {
		indent = "└─ "
	}
	fmt.Printf("%s%s%s\n", l.getPrefix(LevelError), indent, message)
}

// Separator 打印分隔线
func (l *TreeLogger) Separator(char string, length int) {
	if length <= 0 {
		length = 50
	}
	if char == "" {
		char = "─"
	}
	fmt.Println(strings.Repeat(char, length))
}

// Header 打印标题
func (l *TreeLogger) Header(title string) {
	width := len(title) + 4
	border := strings.Repeat("═", width)
	fmt.Printf("\n%s\n║ %s ║\n%s\n\n", border, title, border)
}

// Section 打印章节标题
func (l *TreeLogger) Section(title string) {
	// 计算显示宽度（中文字符占2个宽度）
	displayWidth := 0
	for _, r := range title {
		if r < 128 {
			displayWidth++ // ASCII 字符占1个宽度
		} else {
			displayWidth += 2 // 中文字符占2个宽度
		}
	}
	borderWidth := displayWidth + 4
	border := strings.Repeat("─", borderWidth)
	fmt.Printf("\n%s\n│  %s\n%s\n\n", border, title, border)
}

// Table 打印表格
func (l *TreeLogger) Table(headers []string, rows [][]string) {
	if len(headers) == 0 || len(rows) == 0 {
		return
	}

	// 计算每列宽度
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// 打印分隔线
	printSeparator := func() {
		for _, w := range colWidths {
			fmt.Print("┼─" + strings.Repeat("─", w))
		}
		fmt.Println("┼")
	}

	// 打印表头
	fmt.Print("├─")
	for i, w := range colWidths {
		fmt.Printf(" %-*s ", w, headers[i])
		if i < len(colWidths)-1 {
			fmt.Print("│")
		}
	}
	fmt.Println("┤")
	printSeparator()

	// 打印数据行
	for _, row := range rows {
		fmt.Print("├─")
		for i, cell := range row {
			if i < len(colWidths) {
				fmt.Printf(" %-*s ", colWidths[i], cell)
			}
			if i < len(colWidths)-1 {
				fmt.Print("│")
			}
		}
		fmt.Println("┤")
	}
	printSeparator()
}

// KeyValue 打印键值对
func (l *TreeLogger) KeyValue(key string, value interface{}) {
	fmt.Printf("%s%s%s: %v\n", l.getPrefix(LevelInfo), l.getIndent(), key, value)
}

// Progress 进度条
type Progress struct {
	total      int
	current    int
	barWidth   int
	logger     *TreeLogger
	startTime  time.Time
	lastUpdate time.Time
}

// NewProgress 创建进度条
func (l *TreeLogger) NewProgress(total int) *Progress {
	return &Progress{
		total:      total,
		current:    0,
		barWidth:   40,
		logger:     l,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

// Update 更新进度
func (p *Progress) Update(increment int, message string) {
	p.current += increment
	if p.current > p.total {
		p.current = p.total
	}

	// 限制更新频率（每100ms更新一次）
	now := time.Now()
	if now.Sub(p.lastUpdate) < 100*time.Millisecond && p.current < p.total {
		return
	}
	p.lastUpdate = now

	percentage := float64(p.current) / float64(p.total) * 100
	filled := int(percentage / 100 * float64(p.barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", p.barWidth-filled)

	p.logger.Printf("[%s] %d/%d (%.1f%%) %s", bar, p.current, p.total, percentage, message)
}

// Done 完成进度
func (p *Progress) Done() {
	p.current = p.total
	bar := strings.Repeat("█", p.barWidth)
	elapsed := time.Since(p.startTime).Round(time.Second)
	p.logger.Successf("[%s] %d/%d (100.0%%) 完成! 耗时: %s", bar, p.total, p.total, elapsed)
}

// 全局默认日志记录器
var DefaultLogger = NewTreeLogger()

// 便捷函数（使用默认日志记录器）
func Print(message string)                { DefaultLogger.Print(message) }
func Debug(message string)               { DefaultLogger.Debug(message) }
func Info(message string)                { DefaultLogger.Info(message) }
func Warn(message string)                { DefaultLogger.Warn(message) }
func Error(message string)               { DefaultLogger.Error(message) }
func Success(message string)             { DefaultLogger.Success(message) }
func Printf(format string, args ...interface{})    { DefaultLogger.Printf(format, args...) }
func Debugf(format string, args ...interface{})   { DefaultLogger.Debugf(format, args...) }
func Warnf(format string, args ...interface{})    { DefaultLogger.Warnf(format, args...) }
func Errorf(format string, args ...interface{})   { DefaultLogger.Errorf(format, args...) }
func Successf(format string, args ...interface{}) { DefaultLogger.Successf(format, args...) }
func Indent()                           { DefaultLogger.Indent() }
func Outdent()                          { DefaultLogger.Outdent() }
func Branch(message string)             { DefaultLogger.Branch(message) }
func Leaf(message string)               { DefaultLogger.Leaf(message) }
func SuccessLeaf(message string)        { DefaultLogger.SuccessLeaf(message) }
func ErrorLeaf(message string)          { DefaultLogger.ErrorLeaf(message) }
func Separator(char string, length int) { DefaultLogger.Separator(char, length) }
func Header(title string)               { DefaultLogger.Header(title) }
func Section(title string)              { DefaultLogger.Section(title) }
func Table(headers []string, rows [][]string) { DefaultLogger.Table(headers, rows) }
func KeyValue(key string, value interface{})  { DefaultLogger.KeyValue(key, value) }
