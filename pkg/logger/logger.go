package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// LogLevel 日志级别
type LogLevel int

const (
	// DEBUG 调试级别
	DEBUG LogLevel = iota
	// INFO 信息级别
	INFO
	// WARN 警告级别
	WARN
	// ERROR 错误级别
	ERROR
)

var (
	currentLevel = INFO
	mu           sync.RWMutex
)

// SetLevel 设置日志级别
func SetLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	currentLevel = level
}

// GetLevel 获取当前日志级别
func GetLevel() LogLevel {
	mu.RLock()
	defer mu.RUnlock()
	return currentLevel
}

// SetLevelFromString 从字符串设置日志级别
func SetLevelFromString(level string) {
	switch level {
	case "debug":
		SetLevel(DEBUG)
	case "info":
		SetLevel(INFO)
	case "warn":
		SetLevel(WARN)
	case "error":
		SetLevel(ERROR)
	default:
		SetLevel(INFO)
	}
}

// levelToString 将日志级别转换为字符串
func levelToString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// shouldLog 判断是否应该记录日志
func shouldLog(level LogLevel) bool {
	return level >= GetLevel()
}

// logMessage 记录日志消息
func logMessage(level LogLevel, format string, v ...interface{}) {
	if !shouldLog(level) {
		return
	}

	message := fmt.Sprintf("[%s] %s", levelToString(level), fmt.Sprintf(format, v...))

	switch level {
	case DEBUG, INFO:
		log.Println(message)
	case WARN:
		log.SetOutput(os.Stderr)
		log.Println(message)
		log.SetOutput(os.Stdout)
	case ERROR:
		log.SetOutput(os.Stderr)
		log.Println(message)
		log.SetOutput(os.Stdout)
	}
}

// Debug 记录调试日志
func Debug(format string, v ...interface{}) {
	logMessage(DEBUG, format, v...)
}

// Info 记录信息日志
func Info(format string, v ...interface{}) {
	logMessage(INFO, format, v...)
}

// Warn 记录警告日志
func Warn(format string, v ...interface{}) {
	logMessage(WARN, format, v...)
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
	logMessage(ERROR, format, v...)
}