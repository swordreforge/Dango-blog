package repositories

import (
	"database/sql"
	"time"

	"myblog-gogogo/db/models"
)

// ArticleViewRepository 文章阅读仓库接口
type ArticleViewRepository interface {
	RecordView(passageID int, ip, userAgent, country, city, region string) error
	GetArticleViews(passageID int) (int, error)
	GetArticleStats(passageID int, days int) (*models.ArticleViewStats, error)
	GetMostViewedArticles(limit int) ([]map[string]interface{}, error)
	GetViewSources(days int) ([]map[string]interface{}, error)
	GetViewTrend(days int) ([]map[string]interface{}, error)
	GetViewByCity(days int) ([]map[string]interface{}, error)
	GetViewByIP(days int) ([]map[string]interface{}, error)
}

// SQLiteArticleViewRepository SQLite文章阅读仓库实现
type SQLiteArticleViewRepository struct {
	db *sql.DB
}

func NewSQLiteArticleViewRepository(db *sql.DB) *SQLiteArticleViewRepository {
	return &SQLiteArticleViewRepository{db: db}
}

func (r *SQLiteArticleViewRepository) RecordView(passageID int, ip, userAgent, country, city, region string) error {
	viewDate := time.Now().Format("2006-01-02")

	query := `INSERT INTO article_views (passage_id, ip, user_agent, country, city, region, view_date, view_time, created_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query, passageID, ip, userAgent, country, city, region, viewDate, time.Now(), time.Now())
	return err
}

func (r *SQLiteArticleViewRepository) GetArticleViews(passageID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM article_views WHERE passage_id = ?", passageID).Scan(&count)
	return count, err
}

func (r *SQLiteArticleViewRepository) GetArticleStats(passageID int, days int) (*models.ArticleViewStats, error) {
	stats := &models.ArticleViewStats{}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	// 获取总阅读次数
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM article_views WHERE passage_id = ? AND view_date >= ?",
		passageID, startDate,
	).Scan(&stats.TotalViews)
	if err != nil {
		return nil, err
	}

	// 获取独立访客数
	err = r.db.QueryRow(
		"SELECT COUNT(DISTINCT ip) FROM article_views WHERE passage_id = ? AND view_date >= ?",
		passageID, startDate,
	).Scan(&stats.UniqueVisitors)
	if err != nil {
		return nil, err
	}

	// 获取平均阅读时长
	err = r.db.QueryRow(
		"SELECT AVG(duration) FROM article_views WHERE passage_id = ? AND view_date >= ?",
		passageID, startDate,
	).Scan(&stats.AvgDuration)
	if err != nil {
		stats.AvgDuration = 0
	}

	// 获取热门国家
	countryQuery := `SELECT country, COUNT(*) as count
	                 FROM article_views
	                 WHERE passage_id = ? AND view_date >= ? AND country != ''
	                 GROUP BY country
	                 ORDER BY count DESC
	                 LIMIT 5`
	rows, err := r.db.Query(countryQuery, passageID, startDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var country string
			var count int
			if rows.Scan(&country, &count) == nil {
				stats.TopCountries = append(stats.TopCountries, map[string]interface{}{
					"country": country,
					"count":   count,
				})
			}
		}
	}

	// 获取热门城市
	cityQuery := `SELECT city, region, COUNT(*) as count
	              FROM article_views
	              WHERE passage_id = ? AND view_date >= ? AND city != ''
	              GROUP BY city, region
	              ORDER BY count DESC
	              LIMIT 5`
	rows, err = r.db.Query(cityQuery, passageID, startDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var city, region string
			var count int
			if rows.Scan(&city, &region, &count) == nil {
				stats.TopCities = append(stats.TopCities, map[string]interface{}{
					"city":   city,
					"region": region,
					"count":  count,
				})
			}
		}
	}

	// 获取每日趋势
	trendQuery := `SELECT view_date, COUNT(*) as count
	               FROM article_views
	               WHERE passage_id = ? AND view_date >= ?
	               GROUP BY view_date
	               ORDER BY view_date ASC`
	rows, err = r.db.Query(trendQuery, passageID, startDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var date string
			var count int
			if rows.Scan(&date, &count) == nil {
				stats.DailyTrend = append(stats.DailyTrend, map[string]interface{}{
					"date":  date,
					"count": count,
				})
			}
		}
	}

	return stats, nil
}

func (r *SQLiteArticleViewRepository) GetMostViewedArticles(limit int) ([]map[string]interface{}, error) {
	query := `SELECT p.id, p.title, p.author, COUNT(av.id) as view_count
	          FROM passages p
	          LEFT JOIN article_views av ON p.id = av.passage_id
	          WHERE p.status = 'published'
	          GROUP BY p.id
	          ORDER BY view_count DESC
	          LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, viewCount int
		var title, author string

		err := rows.Scan(&id, &title, &author, &viewCount)
		if err != nil {
			return nil, err
		}

		articles = append(articles, map[string]interface{}{
			"id":         id,
			"title":      title,
			"author":     author,
			"view_count": viewCount,
		})
	}

	return articles, nil
}

func (r *SQLiteArticleViewRepository) GetViewSources(days int) ([]map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `SELECT country, COUNT(*) as count
	          FROM article_views
	          WHERE view_date >= ? AND country != ''
	          GROUP BY country
	          ORDER BY count DESC
	          LIMIT 10`

	rows, err := r.db.Query(query, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sources := make([]map[string]interface{}, 0)
	for rows.Next() {
		var country string
		var count int

		err := rows.Scan(&country, &count)
		if err != nil {
			return nil, err
		}

		sources = append(sources, map[string]interface{}{
			"country": country,
			"count":   count,
		})
	}

	return sources, nil
}

func (r *SQLiteArticleViewRepository) GetViewTrend(days int) ([]map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `SELECT view_date, COUNT(*) as count
	          FROM article_views
	          WHERE view_date >= ?
	          GROUP BY view_date
	          ORDER BY view_date ASC`

	rows, err := r.db.Query(query, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trend := make([]map[string]interface{}, 0)
	for rows.Next() {
		var date string
		var count int

		err := rows.Scan(&date, &count)
		if err != nil {
			return nil, err
		}

		trend = append(trend, map[string]interface{}{
			"date":  date,
			"count": count,
		})
	}

	return trend, nil
}

func (r *SQLiteArticleViewRepository) GetViewByCity(days int) ([]map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `SELECT country, city, region, COUNT(*) as count
	          FROM article_views
	          WHERE view_date >= ? AND city != ''
	          GROUP BY country, city, region
	          ORDER BY count DESC
	          LIMIT 20`

	rows, err := r.db.Query(query, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cities := make([]map[string]interface{}, 0)
	for rows.Next() {
		var country, city, region string
		var count int

		err := rows.Scan(&country, &city, &region, &count)
		if err != nil {
			return nil, err
		}

		cities = append(cities, map[string]interface{}{
			"country": country,
			"city":    city,
			"region":  region,
			"count":   count,
		})
	}

	return cities, nil
}

func (r *SQLiteArticleViewRepository) GetViewByIP(days int) ([]map[string]interface{}, error) {
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	query := `SELECT ip, country, city, region, COUNT(*) as count, MIN(view_date) as first_visit, MAX(view_date) as last_visit
	          FROM article_views
	          WHERE view_date >= ?
	          GROUP BY ip, country, city, region
	          ORDER BY count DESC
	          LIMIT 20`

	rows, err := r.db.Query(query, startDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ips := make([]map[string]interface{}, 0)
	for rows.Next() {
		var ip, country, city, region, firstVisit, lastVisit string
		var count int

		err := rows.Scan(&ip, &country, &city, &region, &count, &firstVisit, &lastVisit)
		if err != nil {
			return nil, err
		}

		ips = append(ips, map[string]interface{}{
			"ip":         ip,
			"country":    country,
			"city":       city,
			"region":     region,
			"count":      count,
			"firstVisit": firstVisit,
			"lastVisit":  lastVisit,
		})
	}

	return ips, nil
}