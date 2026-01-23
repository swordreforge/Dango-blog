package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用配置结构体
type Config struct {
	Port         string
	DBDriver     string
	DBConnStr    string
	LogLevel     string
	TLSCert      string
	TLSKey       string
	EnableTLS    bool
	KafkaBrokers string
	KafkaGroupID string
	JWTSecret    string
}

// Load 从命令行参数加载配置
func Load() *Config {
	port := flag.String("port", "8080", "Port to listen on")
	dbDriver := flag.String("db-driver", "sqlite3", "Database driver (sqlite3, mysql, postgres)")
	dbConnStr := flag.String("db-conn", "./db/data/blog.db", "Database connection string")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	tlsCert := flag.String("tls-cert", "", "Path to TLS certificate file (absolute path)")
	tlsKey := flag.String("tls-key", "", "Path to TLS private key file (absolute path)")
	enableTLS := flag.Bool("enable-tls", false, "Enable TLS (HTTPS/HTTP3)")
	kafkaBrokers := flag.String("kafka-brokers", "", "Kafka brokers (comma-separated, leave empty to disable)")
	kafkaGroupID := flag.String("kafka-group-id", "myblog-consumer-group", "Kafka consumer group ID")
	jwtSecret := flag.String("jwt-secret", "", "JWT secret key (leave empty to auto-generate or load from ./data/jwt-secret)")
	flag.Parse()

	// 如果使用 SQLite 且路径是相对路径，将其转换为绝对路径
	// 优先使用当前工作目录（兼容 go run），如果失败则使用可执行文件目录
	if *dbDriver == "sqlite3" && *dbConnStr == "./db/data/blog.db" {
		// 首先尝试使用当前工作目录
		if cwd, err := os.Getwd(); err == nil {
			dbPath := filepath.Join(cwd, "db", "data", "blog.db")
			// 检查数据库文件是否存在
			if _, err := os.Stat(dbPath); err == nil {
				*dbConnStr = dbPath
			} else {
				// 如果当前工作目录下不存在，尝试使用可执行文件目录
				if execPath, err := os.Executable(); err == nil {
					execDir := filepath.Dir(execPath)
					*dbConnStr = filepath.Join(execDir, "db", "data", "blog.db")
				}
			}
		}
	}

	// 处理 JWT secret
	var jwtSecretValue string
	if *jwtSecret != "" {
		jwtSecretValue = *jwtSecret
	} else {
		// 尝试从文件读取
		jwtSecretFile := filepath.Join(".", "data", "jwt-secret")
		if secretBytes, err := os.ReadFile(jwtSecretFile); err == nil {
			jwtSecretValue = string(secretBytes)
		} else {
			// 生成新的 secret 并保存
			jwtSecretValue = generateRandomSecret()
			if err := os.WriteFile(jwtSecretFile, []byte(jwtSecretValue), 0600); err != nil {
				panic(fmt.Sprintf("Failed to save JWT secret: %v", err))
			}
		}
	}

	return &Config{
		Port:         *port,
		DBDriver:     *dbDriver,
		DBConnStr:    *dbConnStr,
		LogLevel:     *logLevel,
		TLSCert:      *tlsCert,
		TLSKey:       *tlsKey,
		EnableTLS:    *enableTLS,
		KafkaBrokers: *kafkaBrokers,
		KafkaGroupID: *kafkaGroupID,
		JWTSecret:    jwtSecretValue,
	}
}

// generateRandomSecret 生成随机的 JWT secret
func generateRandomSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用硬编码的默认 secret
		return "your-secret-key-change-this-in-production"
	}
	return hex.EncodeToString(bytes)
}