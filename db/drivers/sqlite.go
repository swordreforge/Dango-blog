package drivers

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	_ "modernc.org/sqlite"
	// 或者使用 github.com/mattn/go-sqlite3 (需要CGO)
)

type SQLiteDriver struct{}

func (d *SQLiteDriver) Connect(config Config) (*sql.DB, error) {
	// 确保数据库文件的父目录存在
	dbDir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
	}

	dsn := d.DSN(config)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// 博客应用优化的 SQLite 连接池配置
	// SQLite 写操作是串行的，但 WAL 模式允许并发读
	db.SetMaxOpenConns(15)     // 最大连接数：WAL 模式下支持并发读
	db.SetMaxIdleConns(5)      // 最大空闲连接数：保持少量活跃连接
	db.SetConnMaxLifetime(30 * time.Minute)  // 连接最大存活时间
	db.SetConnMaxIdleTime(10 * time.Minute)  // 连接最大空闲时间

	// 执行 PRAGMA 优化设置
	pragmas := []string{
		"PRAGMA foreign_keys = ON",        // 启用外键约束
		"PRAGMA journal_mode = WAL",       // WAL 模式：提高并发读性能
		"PRAGMA synchronous = NORMAL",     // NORMAL 模式：平衡性能和安全性
		"PRAGMA cache_size = -10000",      // 10MB 缓存：提升查询性能
		"PRAGMA busy_timeout = 5000",      // 忙等待 5 秒：处理并发冲突
		"PRAGMA secure_delete = FAST",     // 快速删除：提升删除性能
		"PRAGMA temp_store = MEMORY",      // 临时表使用内存：提升性能
		"PRAGMA mmap_size = 268435456",    // 256MB 内存映射：提升大文件性能
		"PRAGMA page_size = 4096",         // 4KB 页大小：优化 I/O
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to execute pragma %q: %w", pragma, err)
		}
	}

	return db, nil
}

func (d *SQLiteDriver) DSN(config Config) string {
	// 构建 DSN，包含优化参数
	params := []string{
		"_journal=WAL",           // WAL 日志模式
		"_timeout=5000",          // 超时 5 秒
		"_sync=NORMAL",           // NORMAL 同步模式
		"_cache=shared",          // 共享缓存模式
		"_mutex=no",              // 禁用互斥锁（在单进程应用中）
	}

	return fmt.Sprintf("file:%s?%s", config.FilePath, strings.Join(params, "&"))
}

func init() {
	RegisterDriver("sqlite", &SQLiteDriver{})
	RegisterDriver("sqlite3", &SQLiteDriver{})
}
