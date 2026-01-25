package models

import "time"

// Passage 文章模型
type Passage struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`         // HTML格式内容
	OriginalContent string    `json:"original_content"` // 原始Markdown内容
	Summary         string    `json:"summary"`
	Author          string    `json:"author"`
	Category        string    `json:"category"` // 文章分类
	Status          string    `json:"status"` // published, draft, pending
	FilePath        string    `json:"file_path"` // Markdown 文件的相对路径
	ShowTitle       bool      `json:"show_title"` // 是否显示标题，默认为true
	Visibility      string    `json:"visibility"` // public, private - 文章可见性
	IsScheduled     bool      `json:"is_scheduled"` // 是否定时发布
	PublishedAt     time.Time `json:"published_at"` // 定时发布时间
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}