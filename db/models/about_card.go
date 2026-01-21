package models

import "time"

// AboutMainCard 关于页面主卡片模型
type AboutMainCard struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Icon       string    `json:"icon"`
	LayoutType string    `json:"layout_type"` // default, grid, flex
	CustomCSS  string    `json:"custom_css"`
	SortOrder  int       `json:"sort_order"`
	IsEnabled  bool      `json:"is_enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AboutSubCard 关于页面次卡片模型
type AboutSubCard struct {
	ID          int       `json:"id"`
	MainCardID  int       `json:"main_card_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	LinkURL     string    `json:"link_url"`
	LayoutType  string    `json:"layout_type"` // default, card, list
	CustomCSS   string    `json:"custom_css"`
	SortOrder   int       `json:"sort_order"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}