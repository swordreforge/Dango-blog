package settings

import (
	"fmt"
	"log"

	"myblog-gogogo/db"
)

// MusicSettings 音乐设置结构
type MusicSettings struct {
	Enabled     bool   `json:"enabled"`
	AutoPlay    bool   `json:"auto_play"`
	ControlSize string `json:"control_size"`
	CustomCSS   string `json:"custom_css"`
	PlayerColor string `json:"player_color"`
	Position    string `json:"position"`
}

// GetMusic 获取音乐设置
func GetMusic() (*MusicSettings, error) {
	repo := db.GetSettingRepository()

	settings := MusicSettings{
		Enabled:     false,
		AutoPlay:    false,
		ControlSize: "medium",
		CustomCSS:   "",
		PlayerColor: "rgba(66, 133, 244, 0.9)",
		Position:    "bottom-right",
	}

	keys := []string{
		"music_enabled", "music_auto_play", "music_control_size",
		"music_custom_css", "music_player_color", "music_position",
	}

	for _, key := range keys {
		setting, err := repo.GetByKey(key)
		if err != nil {
			log.Printf("Failed to get setting %s: %v", key, err)
			continue
		}
		if setting != nil {
			switch key {
			case "music_enabled":
				settings.Enabled = stringToBool(setting.Value)
			case "music_auto_play":
				settings.AutoPlay = stringToBool(setting.Value)
			case "music_control_size":
				settings.ControlSize = setting.Value
			case "music_custom_css":
				settings.CustomCSS = setting.Value
			case "music_player_color":
				settings.PlayerColor = setting.Value
			case "music_position":
				settings.Position = setting.Value
			}
		}
	}

	return &settings, nil
}

// UpdateMusic 更新音乐设置
func UpdateMusic(settings *MusicSettings) error {
	repo := db.GetSettingRepository()

	updates := map[string]string{
		"music_enabled":     boolToString(settings.Enabled),
		"music_auto_play":   boolToString(settings.AutoPlay),
		"music_control_size": settings.ControlSize,
		"music_custom_css":  settings.CustomCSS,
		"music_player_color": settings.PlayerColor,
		"music_position":    settings.Position,
	}

	for key, value := range updates {
		if err := repo.UpdateByKey(key, value); err != nil {
			return fmt.Errorf("failed to update setting %s: %w", key, err)
		}
	}

	return nil
}

// UpdateMusicPartial 增量更新音乐设置
func UpdateMusicPartial(updates map[string]interface{}) error {
	repo := db.GetSettingRepository()

	fieldMapping := map[string]string{
		"enabled":     "music_enabled",
		"auto_play":   "music_auto_play",
		"control_size": "music_control_size",
		"custom_css":  "music_custom_css",
		"player_color": "music_player_color",
		"position":    "music_position",
	}

	for jsonField, value := range updates {
		dbKey, exists := fieldMapping[jsonField]
		if !exists {
			log.Printf("Unknown music field: %s", jsonField)
			continue
		}

		var stringValue string
		switch v := value.(type) {
		case string:
			stringValue = v
		case bool:
			stringValue = boolToString(v)
		case float64:
			stringValue = fmt.Sprintf("%.0f", v)
		default:
			stringValue = fmt.Sprintf("%v", v)
		}

		if err := repo.UpdateByKey(dbKey, stringValue); err != nil {
			return fmt.Errorf("failed to update setting %s: %w", dbKey, err)
		}
	}

	return nil
}