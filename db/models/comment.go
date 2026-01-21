package models

import "time"

// Comment 评论模型
type Comment struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	PassageID int       `json:"passage_id"`
	CreatedAt time.Time `json:"created_at"`
}