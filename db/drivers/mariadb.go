package drivers

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/go-sql-driver/mysql"
)

type MariaDBDriver struct{}

func (d *MariaDBDriver) Connect(config Config) (*sql.DB, error) {
	dsn := d.DSN(config)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MariaDB: %w", err)
	}

	// 博客应用优化的 MySQL/MariaDB 连接池配置
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

func (d *MariaDBDriver) DSN(config Config) string {
	// 构建 DSN，包含优化参数
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local&timeout=5s&readTimeout=30s&writeTimeout=30s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)
}

func init() {
	RegisterDriver("mariadb", &MariaDBDriver{})
	RegisterDriver("mysql", &MariaDBDriver{})
}
