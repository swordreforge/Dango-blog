package controller

import (
	"encoding/json"
	"net/http"
	"time"

	"myblog-gogogo/pkg/metrics"
)

// MetricsHandler 性能指标 API 处理器
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取指标
	metrics := metrics.GetMetrics()
	stats := metrics.GetStats()

	// 返回 JSON 响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
		"timestamp": time.Now().Unix(),
	})
}

// MetricsResetHandler 重置指标 API 处理器
func MetricsResetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 重置指标
	metrics.GetMetrics().Reset()

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Metrics reset successfully",
	})
}