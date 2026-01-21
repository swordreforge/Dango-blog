package config

import (
	"flag"
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
	flag.Parse()

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
	}
}