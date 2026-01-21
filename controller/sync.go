package controller

import (
	"encoding/json"
	"net/http"

	"myblog-gogogo/db"
	"myblog-gogogo/service"
)

// SyncHandler 同步处理器
func SyncHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取文章仓库
	repo := db.GetPassageRepository()
	
	// 创建同步服务
	syncService := service.NewSyncService(repo)
	
	// 执行同步
	if err := syncService.SyncAll(); err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": "同步失败",
			"error":   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "同步成功",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
