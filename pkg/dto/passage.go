package dto

import "time"

// PassageDTO 文章数据传输对象
type PassageDTO struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	OriginalContent string   `json:"original_content,omitempty"`
	Status         string    `json:"status"`
	Visibility     string    `json:"visibility"`
	ShowTitle      bool      `json:"show_title"`
	IsScheduled    bool      `json:"is_scheduled"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	
	// 关联数据
	Categories []string `json:"categories,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	ViewCount  int      `json:"view_count,omitempty"`
}

// PassageAccessRequest 文章访问请求
type PassageAccessRequest struct {
	PassageID int    `json:"passage_id"`
	UserRole  string `json:"user_role"`
}

// PassageAccessResponse 文章访问响应
type PassageAccessResponse struct {
	Allowed     bool        `json:"allowed"`
	Passage     *PassageDTO `json:"passage,omitempty"`
	Reason      string      `json:"reason,omitempty"`
	Status      string      `json:"status,omitempty"`
	Visibility  string      `json:"visibility,omitempty"`
	IsScheduled bool        `json:"is_scheduled,omitempty"`
	PublishedAt *time.Time  `json:"published_at,omitempty"`
}

// CreatePassageRequest 创建文章请求
type CreatePassageRequest struct {
	Title          string    `json:"title" binding:"required"`
	Content        string    `json:"content" binding:"required"`
	Status         string    `json:"status"`
	Visibility     string    `json:"visibility"`
	ShowTitle      bool      `json:"show_title"`
	IsScheduled    bool      `json:"is_scheduled"`
	PublishedAt    *time.Time `json:"published_at"`
	Categories     []string  `json:"categories"`
	Tags           []string  `json:"tags"`
}

// UpdatePassageRequest 更新文章请求
type UpdatePassageRequest struct {
	Title          *string    `json:"title"`
	Content        *string    `json:"content"`
	Status         *string    `json:"status"`
	Visibility     *string    `json:"visibility"`
	ShowTitle      *bool      `json:"show_title"`
	IsScheduled    *bool      `json:"is_scheduled"`
	PublishedAt    *time.Time `json:"published_at"`
	Categories     []string   `json:"categories"`
	Tags           []string   `json:"tags"`
}

// PassageListRequest 文章列表请求
type PassageListRequest struct {
	PaginationRequest
	Status     string `json:"status" form:"status"`
	Visibility string `json:"visibility" form:"visibility"`
	Search     string `json:"search" form:"search"`
	Category   string `json:"category" form:"category"`
	Tag        string `json:"tag" form:"tag"`
	SortBy     string `json:"sort_by" form:"sort_by"`
	SortOrder  string `json:"sort_order" form:"sort_order"`
}

// PassageUpdateRequest 增量更新文章请求
type PassageUpdateRequest struct {
	PassageID int                    `json:"passage_id"`
	Updates   map[string]interface{} `json:"updates"`
}