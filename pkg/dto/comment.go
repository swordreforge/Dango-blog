package dto

import "time"

// CommentDTO 评论数据传输对象
type CommentDTO struct {
	ID         int       `json:"id"`
	Content    string    `json:"content"`
	AuthorName string    `json:"author_name"`
	AuthorEmail string   `json:"author_email,omitempty"`
	PassageID  int       `json:"passage_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	Content     string `json:"content" binding:"required"`
	AuthorName  string `json:"author_name" binding:"required"`
	AuthorEmail string `json:"author_email" binding:"required,email"`
	PassageID   int    `json:"passage_id" binding:"required"`
}

// CommentListRequest 评论列表请求
type CommentListRequest struct {
	PaginationRequest
	PassageID int    `json:"passage_id" form:"passage_id"`
	Status    string `json:"status" form:"status"`
	Search    string `json:"search" form:"search"`
}