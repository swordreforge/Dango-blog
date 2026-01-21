package drivers

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/lib/pq"
)

type PostgreSQLDriver struct{}

func (d *PostgreSQLDriver) Connect(config Config) (*sql.DB, error) {
	dsn := d.DSN(config)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// 博客应用优化的 PostgreSQL 连接池配置
	if config.MaxConns == 0 {
		config.MaxConns = 25  // 默认最大连接数
	}
	if config.MaxIdle == 0 {
		config.MaxIdle = 10   // 默认空闲连接数
	}

	db.SetMaxOpenConns(config.MaxConns)
	db.SetMaxIdleConns(config.MaxIdle)
	db.SetConnMaxLifetime(30 * time.Minute)  // 连接最大存活时间
	db.SetConnMaxIdleTime(10 * time.Minute)  // 连接最大空闲时间

	return db, nil
}

func (d *PostgreSQLDriver) DSN(config Config) string {
	sslMode := config.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	// 构建 DSN，包含优化参数
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=5&statement_timeout=30000",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		sslMode,
	)
}

func init() {
	RegisterDriver("postgres", &PostgreSQLDriver{})
	RegisterDriver("pgsql", &PostgreSQLDriver{})
}
