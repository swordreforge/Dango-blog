package repositories

import (
	"database/sql"
	"time"
)

// VisitorRepository 访客仓库接口
type VisitorRepository interface {
	RecordVisit(ip, userAgent string) error
	GetTodayVisits() (int, error)
	GetYesterdayVisits() (int, error)
	GetTotalVisitors() (int, error)
	GetVisitsByDate(date string) (int, error)
}

// SQLiteVisitorRepository SQLite访客仓库实现
type SQLiteVisitorRepository struct {
	db *sql.DB
}

func NewSQLiteVisitorRepository(db *sql.DB) *SQLiteVisitorRepository {
	return &SQLiteVisitorRepository{db: db}
}

func (r *SQLiteVisitorRepository) RecordVisit(ip, userAgent string) error {
	visitDate := time.Now().Format("2006-01-02")

	// 检查今天是否已经记录过该IP
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM visitors WHERE ip = ? AND visit_date = ?",
		ip, visitDate,
	).Scan(&count)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// 如果今天没有记录，则插入新记录
	if count == 0 {
		query := `INSERT INTO visitors (ip, user_agent, visit_date, created_at)
		          VALUES (?, ?, ?, ?)`

		_, err = r.db.Exec(query, ip, userAgent, visitDate, time.Now())
		return err
	}

	return nil
}

func (r *SQLiteVisitorRepository) GetTodayVisits() (int, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM visitors WHERE visit_date = ?",
		today,
	).Scan(&count)
	return count, err
}

func (r *SQLiteVisitorRepository) GetYesterdayVisits() (int, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM visitors WHERE visit_date = ?",
		yesterday,
	).Scan(&count)
	return count, err
}

func (r *SQLiteVisitorRepository) GetTotalVisitors() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM visitors").Scan(&count)
	return count, err
}

func (r *SQLiteVisitorRepository) GetVisitsByDate(date string) (int, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM visitors WHERE visit_date = ?",
		date,
	).Scan(&count)
	return count, err
}