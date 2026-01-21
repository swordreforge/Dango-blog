package models

import "time"

// Visitor 访客模型
type Visitor struct {
	ID        int       `json:"id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	VisitDate string    `json:"visit_date"` // YYYY-MM-DD 格式
	CreatedAt time.Time `json:"created_at"`
}