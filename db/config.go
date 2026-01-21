package db

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	
	"myblog-gogogo/db/drivers"
)

type DBConfig struct {
	drivers.Config
	AutoMigrate bool   // 是否自动迁移
	LogLevel    string // 日志级别
}

// ParseFlags 解析命令行参数
func ParseFlags() (*DBConfig, error) {
	var config DBConfig
	
	flag.StringVar(&config.Driver, "db-driver", "sqlite", 
		fmt.Sprintf("Database driver (%v)", drivers.AvailableDrivers()))
	flag.StringVar(&config.Host, "db-host", "localhost", "Database host")
	flag.IntVar(&config.Port, "db-port", 0, "Database port")
	flag.StringVar(&config.User, "db-user", "", "Database user")
	flag.StringVar(&config.Password, "db-password", "", "Database password")
	flag.StringVar(&config.Database, "db-name", "app.db", "Database name")
	flag.StringVar(&config.SSLMode, "db-sslmode", "disable", "SSL mode")
	flag.StringVar(&config.FilePath, "db-file", "data/app.db", "SQLite file path")
	flag.IntVar(&config.MaxConns, "db-max-conns", 10, "Maximum connections")
	flag.IntVar(&config.MaxIdle, "db-max-idle", 5, "Maximum idle connections")
	flag.BoolVar(&config.AutoMigrate, "db-migrate", false, "Auto migrate database")
	flag.StringVar(&config.LogLevel, "db-log-level", "info", "Database log level")
	
	flag.Parse()
	
	// 根据驱动类型设置默认值
	switch config.Driver {
	case "sqlite", "sqlite3":
		if config.FilePath == "" {
			// 获取程序所在目录
			execPath, err := os.Executable()
			if err != nil {
				execPath = "."
			}
			execDir := filepath.Dir(execPath)
			config.FilePath = filepath.Join(execDir, "data", "app.db")
		} else if !filepath.IsAbs(config.FilePath) {
			// 如果是相对路径，转换为基于程序所在目录的绝对路径
			execPath, err := os.Executable()
			if err != nil {
				execPath = "."
			}
			execDir := filepath.Dir(execPath)
			config.FilePath = filepath.Join(execDir, config.FilePath)
		}
		// 确保目录存在
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}
	case "mysql", "mariadb":
		if config.Port == 0 {
			config.Port = 3306
		}
	case "postgres", "pgsql":
		if config.Port == 0 {
			config.Port = 5432
		}
	}
	
	return &config, nil
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv() (*DBConfig, error) {
	config := &DBConfig{
		Config: drivers.Config{
			Driver:   getEnv("DB_DRIVER", "sqlite"),
			Host:     getEnv("DB_HOST", "localhost"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Database: getEnv("DB_NAME", "app.db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			FilePath: getEnv("DB_FILE", "data/app.db"),
		},
	}

	if portStr := getEnv("DB_PORT", ""); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Config.Port = port
		}
	}

	if maxConnsStr := getEnv("DB_MAX_CONNS", ""); maxConnsStr != "" {
		if maxConns, err := strconv.Atoi(maxConnsStr); err == nil {
			config.Config.MaxConns = maxConns
		}
	}

	config.AutoMigrate = getEnv("DB_AUTO_MIGRATE", "false") == "true"
	config.LogLevel = getEnv("DB_LOG_LEVEL", "info")

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
