package models

import "time"

// PassageTag 文章-标签关联模型
type PassageTag struct {
	ID        int       `json:"id"`
	PassageID int       `json:"passage_id"`
	TagID     int       `json:"tag_id"`
	CreatedAt time.Time `json:"created_at"`
}