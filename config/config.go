package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Config 应用配置结构体
type Config struct {
	Port            string
	DBDriver        string
	DBConnStr       string
	LogLevel        string
	TLSCert         string
	TLSKey          string
	EnableTLS       bool
	KafkaBrokers    string
	KafkaGroupID    string
	JWTSecret       string
	EnableFileWatch bool
	// 工作池配置
	WorkerCount     int
	WorkerQueueSize int
	// 会话清理配置
	SessionCleanupInterval int // 会话清理间隔(分钟)
	// Kafka 配置
	KafkaProducerQueueSize int // Kafka 异步生产者队列大小
	// 数据库连接池配置
	DBMaxOpenConns    int    // 最大打开连接数
	DBMaxIdleConns    int    // 最大空闲连接数
	DBConnMaxLifetime int    // 连接最大存活时间(分钟)
	DBConnMaxIdleTime int    // 连接最大空闲时间(分钟)
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
	enableFileWatch := flag.Bool("enable-file-watch", false, "Enable file watching for markdown files")
	workerCount := flag.Int("worker-count", 2, "Number of worker pool workers")
	workerQueueSize := flag.Int("worker-queue-size", 1000, "Worker pool queue size")
	sessionCleanupInterval := flag.Int("session-cleanup-interval", 5, "Session cleanup interval in minutes")
	kafkaProducerQueueSize := flag.Int("kafka-producer-queue-size", 1000, "Kafka async producer queue size")
	dbMaxOpenConns := flag.Int("db-max-open-conns", 15, "Database max open connections")
	dbMaxIdleConns := flag.Int("db-max-idle-conns", 5, "Database max idle connections")
	dbConnMaxLifetime := flag.Int("db-conn-max-lifetime", 30, "Database connection max lifetime in minutes")
	dbConnMaxIdleTime := flag.Int("db-conn-max-idle-time", 10, "Database connection max idle time in minutes")
	flag.Parse()

	// 如果使用 SQLite 且路径是相对路径，将其转换为绝对路径
	// 优先使用当前工作目录（兼容 go run），如果失败则使用可执行文件目录
	if *dbDriver == "sqlite3" && *dbConnStr == "./db/data/blog.db" {
		// 首先尝试使用当前工作目录
		if cwd, err := os.Getwd(); err == nil {
			dbPath := filepath.Join(cwd, "db", "data", "blog.db")
			// 直接使用当前工作目录，不检查文件是否存在（允许首次创建）
			*dbConnStr = dbPath
			// 确保目录存在
			if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
				log.Printf("Warning: failed to create data directory: %v", err)
			}
		} else if execPath, err := os.Executable(); err == nil {
			// 如果获取工作目录失败，使用可执行文件目录
			execDir := filepath.Dir(execPath)
			*dbConnStr = filepath.Join(execDir, "db", "data", "blog.db")
			// 确保目录存在
			if err := os.MkdirAll(filepath.Dir(*dbConnStr), 0755); err != nil {
				log.Printf("Warning: failed to create data directory: %v", err)
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
		Port:                    *port,
		DBDriver:                *dbDriver,
		DBConnStr:               *dbConnStr,
		LogLevel:                *logLevel,
		TLSCert:                 *tlsCert,
		TLSKey:                  *tlsKey,
		EnableTLS:               *enableTLS,
		KafkaBrokers:            *kafkaBrokers,
		KafkaGroupID:            *kafkaGroupID,
		JWTSecret:               jwtSecretValue,
		EnableFileWatch:         *enableFileWatch,
		WorkerCount:             *workerCount,
		WorkerQueueSize:         *workerQueueSize,
		SessionCleanupInterval:  *sessionCleanupInterval,
		KafkaProducerQueueSize:  *kafkaProducerQueueSize,
		DBMaxOpenConns:          *dbMaxOpenConns,
		DBMaxIdleConns:          *dbMaxIdleConns,
		DBConnMaxLifetime:       *dbConnMaxLifetime,
		DBConnMaxIdleTime:       *dbConnMaxIdleTime,
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