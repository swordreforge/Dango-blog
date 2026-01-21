package admin

import (
	"encoding/json"
	"net/http"

	"myblog-gogogo/db"
)

// AdminStatsHandler 统计数据API处理器
func AdminStatsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 获取统计数据
		statsRepo := db.GetStatsRepository()
		stats, err := statsRepo.GetStats()
		if err != nil {
			http.Error(w, "Failed to fetch stats", http.StatusInternalServerError)
			return
		}

		// 获取今日访问量
		visitorRepo := db.GetVisitorRepository()
		todayVisits, err := visitorRepo.GetTodayVisits()
		if err != nil {
			todayVisits = 0
		}

		// 获取昨日访问量
		yesterdayVisits, err := visitorRepo.GetYesterdayVisits()
		if err != nil {
			yesterdayVisits = 0
		}

		// 计算今日相对于昨日的变化百分比
		var visitsChangePercent float64
		var visitsTrend string
		if yesterdayVisits > 0 {
			visitsChangePercent = float64(todayVisits-yesterdayVisits) / float64(yesterdayVisits) * 100
			if visitsChangePercent > 0 {
				visitsTrend = "up"
			} else if visitsChangePercent < 0 {
				visitsTrend = "down"
			} else {
				visitsTrend = "stable"
			}
		} else {
			visitsChangePercent = 0
			visitsTrend = "stable"
		}

		response := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"total_articles":        stats.TotalArticles,
				"total_users":           stats.TotalUsers,
				"total_visitors":        stats.TotalVisitors,
				"total_categories":      stats.TotalCategories,
				"avg_engagement":        stats.AvgEngagement,
				"last_updated":          stats.LastUpdated,
				"today_visits":          todayVisits,
				"yesterday_visits":      yesterdayVisits,
				"visits_change_percent": visitsChangePercent,
				"visits_trend":          visitsTrend,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}