package models

import "time"

// ArticleView 文章阅读记录模型
type ArticleView struct {
	ID        int       `json:"id"`
	PassageID int       `json:"passage_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Country   string    `json:"country"` // 国家
	City      string    `json:"city"`    // 城市
	Region    string    `json:"region"`  // 地区/省份
	ViewDate  string    `json:"view_date"` // YYYY-MM-DD 格式
	ViewTime  time.Time `json:"view_time"` // 详细时间
	Duration  int       `json:"duration"` // 阅读时长（秒）
	CreatedAt time.Time `json:"created_at"`
}

// ArticleViewStats 文章阅读统计汇总
type ArticleViewStats struct {
	TotalViews     int                    `json:"total_views"`
	UniqueVisitors int                    `json:"unique_visitors"`
	AvgDuration    float64                `json:"avg_duration"`
	TopCountries   []map[string]interface{} `json:"top_countries"`
	TopCities      []map[string]interface{} `json:"top_cities"`
	DailyTrend     []map[string]interface{} `json:"daily_trend"`
}