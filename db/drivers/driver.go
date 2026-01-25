package drivers

import (
	"database/sql"
	"time"
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
	// 连接池配置
	MaxOpenConns    int           // 最大打开连接数
	MaxIdleConns    int           // 最大空闲连接数
	ConnMaxLifetime time.Duration // 连接最大存活时间
	ConnMaxIdleTime time.Duration // 连接最大空闲时间
}

// Driver 驱动接口
type Driver interface {
	Connect(config Config) (*sql.DB, error)
	DSN(config Config) string
}
