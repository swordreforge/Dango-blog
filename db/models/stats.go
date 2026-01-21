package models

// Stats 统计数据模型
type Stats struct {
	TotalVisitors   int    `json:"total_visitors"`
	TotalArticles   int    `json:"total_articles"`
	TotalUsers      int    `json:"total_users"`
	TotalCategories int    `json:"total_categories"`
	AvgEngagement   string `json:"avg_engagement"`
	LastUpdated     string `json:"last_updated"`
}