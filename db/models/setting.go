package models

import "time"

// Setting 设置模型
type Setting struct {
	ID          int       `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Type        string    `json:"type"` // string, number, boolean, json
	Description string    `json:"description"`
	Category    string    `json:"category"` // appearance, system, content
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}