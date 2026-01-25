package models

import "time"

// Tag 标签模型
type Tag struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"` // 标签颜色
	CategoryID  int       `json:"category_id"` // 所属分类ID，0表示无分类
	SortOrder   int       `json:"sort_order"`
	IsEnabled   bool      `json:"is_enabled"`
	UsageCount  int       `json:"usage_count"` // 使用次数统计
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}