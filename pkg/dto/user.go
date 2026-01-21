package dto

import "time"

// LoginRequest 登录请求
type LoginRequest struct {
	Username          string `json:"username" binding:"required"`
	Password          string `json:"password"`
	EncryptedPassword string `json:"encrypted_password"`
	SessionID         string `json:"session_id"`
	ClientPublicKey   string `json:"client_public_key"`
	Algorithm         string `json:"algorithm"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      *UserDTO  `json:"user"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User  *UserDTO `json:"user"`
	Token string   `json:"token,omitempty"`
}

// UserDTO 用户数据传输对象
type UserDTO struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Password string `json:"password"`
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	PaginationRequest
	Role   string `json:"role" form:"role"`
	Status string `json:"status" form:"status"`
	Search string `json:"search" form:"search"`
}