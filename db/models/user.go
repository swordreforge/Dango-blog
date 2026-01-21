package models

import "time"

// User 用户模型
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // 不在JSON中返回
	Email     string    `json:"email"`
	Role      string    `json:"role"` // admin, editor, user
	Status    string    `json:"status"` // active, restricted, banned
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}