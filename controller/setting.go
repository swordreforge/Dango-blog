package controller

import (
	"encoding/json"
	"net/http"

	"myblog-gogogo/service"
	"myblog-gogogo/service/settings"
)

// SettingAPIHandler 设置API处理器
func SettingAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 检查是否是外观设置路径
	if r.URL.Path == "/settings/appearance" {
		switch r.Method {
		case http.MethodGet:
			appearanceSettings, err := settings.GetAppearance()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			json.NewEncoder(w).Encode(appearanceSettings)
		case http.MethodPut:
			var appearanceSettings settings.AppearanceSettings
			if err := json.NewDecoder(r.Body).Decode(&appearanceSettings); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
				return
			}

			// 验证透明度值
			if appearanceSettings.GlobalOpacity == "" {
				appearanceSettings.GlobalOpacity = "0.15"
			}

			if err := settings.UpdateAppearance(&appearanceSettings); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			json.NewEncoder(w).Encode(map[string]string{"message": "Appearance settings updated successfully"})
		case http.MethodPatch:
			var updates map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
				return
			}

			if err := settings.UpdateAppearancePartial(updates); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			json.NewEncoder(w).Encode(map[string]string{"message": "Appearance settings updated successfully"})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		}
		return
	}

	// 通用设置处理
	switch r.Method {
	case http.MethodGet:
		getSettings(w, r)
	case http.MethodPut:
		updateSettings(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

// getSettings 获取设置
func getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := settings.GetAppearance()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(settings)
}

// updateSettings 更新设置
func updateSettings(w http.ResponseWriter, r *http.Request) {
	var appearanceSettings settings.AppearanceSettings
	if err := json.NewDecoder(r.Body).Decode(&appearanceSettings); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// 验证透明度值
	if appearanceSettings.GlobalOpacity == "" {
		appearanceSettings.GlobalOpacity = "0.15"
	}

	if err := settings.UpdateAppearance(&appearanceSettings); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Settings updated successfully"})
}

// GetAllSettingsHandler 获取所有设置
func GetAllSettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	settings, err := settings.GetAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(settings)
}

// GetAppearanceSettingsHandler 获取外观设置（用于前端渲染）
func GetAppearanceSettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	settings, err := settings.GetAppearance()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(settings)
}

// UpdateSingleSettingHandler 更新单个设置
func UpdateSingleSettingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Key == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Key is required"})
		return
	}

	if err := service.UpdateSettingByKey(req.Key, req.Value); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Setting updated successfully"})
}

// GetMusicSettingsHandler 获取音乐设置
func GetMusicSettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	settings, err := settings.GetMusic()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(settings)
}

// UpdateMusicSettingsHandler 更新音乐设置
func UpdateMusicSettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var musicSettings service.MusicSettings
	if err := json.NewDecoder(r.Body).Decode(&musicSettings); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if err := settings.UpdateMusic(&musicSettings); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Music settings updated successfully"})
}

// UpdateMusicSettingsPartialHandler 增量更新音乐设置
func UpdateMusicSettingsPartialHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if err := settings.UpdateMusicPartial(updates); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Music settings updated successfully"})
}