package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"myblog-gogogo/db"
)

// AdminAnalyticsHandler 管理员统计分析API处理器
func AdminAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 获取分析数据
		action := r.URL.Query().Get("action")

		switch action {
		case "most-viewed":
			getMostViewedArticles(w, r)
		case "view-sources":
			getViewSources(w, r)
		case "view-trend":
			getViewTrend(w, r)
		case "article-stats":
			getArticleStats(w, r)
		case "view-by-city":
			getViewByCity(w, r)
		case "view-by-ip":
			getViewByIP(w, r)
		default:
			response := map[string]interface{}{
				"success": false,
				"message": "未知的操作类型",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// getMostViewedArticles 获取最多阅读的文章
func getMostViewedArticles(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	repo := db.GetArticleViewRepository()
	articles, err := repo.GetMostViewedArticles(limit)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "获取热门文章失败",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    articles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getViewSources 获取阅读来源（按国家统计）
func getViewSources(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	repo := db.GetArticleViewRepository()
	sources, err := repo.GetViewSources(days)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "获取阅读来源失败",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    sources,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getViewTrend 获取阅读趋势
func getViewTrend(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	repo := db.GetArticleViewRepository()
	trend, err := repo.GetViewTrend(days)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "获取阅读趋势失败",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    trend,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getArticleStats 获取单篇文章的统计信息
func getArticleStats(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		response := map[string]interface{}{
			"success": false,
			"message": "缺少文章ID参数",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		response := map[string]interface{}{
			"success": false,
			"message": "无效的文章ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	repo := db.GetArticleViewRepository()
	stats, err := repo.GetArticleStats(id, days)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "获取文章统计失败",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RecordArticleView 记录文章阅读（供其他控制器调用）
func RecordArticleView(passageID int, ip, userAgent, country, city, region string) error {
	repo := db.GetArticleViewRepository()
	return repo.RecordView(passageID, ip, userAgent, country, city, region)
}

// getViewByCity 获取按城市统计的阅读数据
func getViewByCity(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	repo := db.GetArticleViewRepository()
	cities, err := repo.GetViewByCity(days)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "获取城市统计失败",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    cities,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getViewByIP 获取按IP统计的访问数据
func getViewByIP(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	repo := db.GetArticleViewRepository()
	ips, err := repo.GetViewByIP(days)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "获取IP统计失败",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    ips,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}