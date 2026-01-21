package repositories

import (
	"database/sql"
	"time"

	"myblog-gogogo/db/models"
)

// StatsRepository 统计仓库接口
type StatsRepository interface {
	GetStats() (*models.Stats, error)
	GetVisitsTrend(period string) (map[string]interface{}, error)
	GetCategoriesDistribution() (map[string]interface{}, error)
	GetPopularArticles(limit int) ([]map[string]interface{}, error)
}

// SQLiteStatsRepository SQLite统计仓库实现
type SQLiteStatsRepository struct {
	db *sql.DB
}

func NewSQLiteStatsRepository(db *sql.DB) *SQLiteStatsRepository {
	return &SQLiteStatsRepository{db: db}
}

func (r *SQLiteStatsRepository) GetStats() (*models.Stats, error) {
	stats := &models.Stats{
		LastUpdated: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 获取文章总数
	err := r.db.QueryRow("SELECT COUNT(*) FROM passages").Scan(&stats.TotalArticles)
	if err != nil {
		return nil, err
	}

	// 获取用户总数
	err = r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, err
	}

	// 获取访问者总数（从 visitors 表获取）
	err = r.db.QueryRow("SELECT COUNT(*) FROM visitors").Scan(&stats.TotalVisitors)
	if err != nil {
		return nil, err
	}

	// 获取分类总数（模拟数据）
	stats.TotalCategories = 5

	// 平均互动率（模拟数据）
	stats.AvgEngagement = "75%"

	return stats, nil
}

func (r *SQLiteStatsRepository) GetVisitsTrend(period string) (map[string]interface{}, error) {
	// 模拟访问趋势数据
	result := map[string]interface{}{
		"period": period,
		"data": []map[string]interface{}{
			{"date": "2026-01-10", "visits": 850},
			{"date": "2026-01-11", "visits": 920},
			{"date": "2026-01-12", "visits": 1050},
			{"date": "2026-01-13", "visits": 980},
			{"date": "2026-01-14", "visits": 1150},
			{"date": "2026-01-15", "visits": 1204},
		},
	}
	return result, nil
}

func (r *SQLiteStatsRepository) GetCategoriesDistribution() (map[string]interface{}, error) {
	// 模拟分类分布数据
	result := map[string]interface{}{
		"categories": []map[string]interface{}{
			{"name": "前端开发", "count": 45},
			{"name": "后端开发", "count": 38},
			{"name": "设计", "count": 28},
			{"name": "人工智能", "count": 22},
			{"name": "工具推荐", "count": 9},
		},
	}
	return result, nil
}

func (r *SQLiteStatsRepository) GetPopularArticles(limit int) ([]map[string]interface{}, error) {
	// 从数据库获取热门文章（按创建时间倒序）
	query := `SELECT id, title, author, created_at FROM passages ORDER BY created_at DESC LIMIT ?`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []map[string]interface{}
	for rows.Next() {
		var id int
		var title, author string
		var createdAt time.Time

		err := rows.Scan(&id, &title, &author, &createdAt)
		if err != nil {
			return nil, err
		}

		articles = append(articles, map[string]interface{}{
			"id":         id,
			"title":      title,
			"author":     author,
			"created_at": createdAt.Format("2006-01-02"),
		})
	}

	return articles, nil
}