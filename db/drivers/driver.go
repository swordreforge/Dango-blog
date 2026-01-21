package drivers

import (
	"database/sql"
)

// Config 数据库配置
type Config struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
	FilePath string // SQLite专用
	MaxConns int
	MaxIdle  int
}

// Driver 驱动接口
type Driver interface {
	Connect(config Config) (*sql.DB, error)
	DSN(config Config) string
}
