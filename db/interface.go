//switch db between mariadb,sqlite,pgsql   
package db

import (
	"context"
	"database/sql"
)

// Database 数据库统一接口
type Database interface {
	// 连接操作
	Connect() error
	Close() error
	Ping(ctx context.Context) error
	
	// 查询操作
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	
	// 事务操作
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	
	// 获取原生驱动
	GetDriver() interface{}
	
	// 数据库信息
	Name() string
	Version() (string, error)
}
