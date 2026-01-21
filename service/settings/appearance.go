package settings

import (
	"encoding/json"
	"fmt"
	"log"

	"myblog-gogogo/db"
)

// AppearanceSettings 外观设置结构
type AppearanceSettings struct {
	BackgroundImage      string `json:"background_image"`
	GlobalOpacity        string `json:"global_opacity"`
	BackgroundSize       string `json:"background_size"`
	BackgroundPosition   string `json:"background_position"`
	BackgroundRepeat     string `json:"background_repeat"`
	BackgroundAttachment string `json:"background_attachment"`
	BlurAmount           string `json:"blur_amount"`
	SaturateAmount       string `json:"saturate_amount"`
	DarkModeEnabled      bool   `json:"dark_mode_enabled"`
	NavbarGlassColor     string `json:"navbar_glass_color"`
	NavbarTextColor      string `json:"navbar_text_color"`
	CardGlassColor       string `json:"card_glass_color"`
	FooterGlassColor     string `json:"footer_glass_color"`
}

// GetAppearance 获取外观设置
func GetAppearance() (*AppearanceSettings, error) {
	repo := db.GetSettingRepository()

	settings := AppearanceSettings{
		BackgroundImage:      "/img/test.png",
		GlobalOpacity:        "0.15",
		BackgroundSize:       "cover",
		BackgroundPosition:   "center",
		BackgroundRepeat:     "no-repeat",
		BackgroundAttachment: "fixed",
		BlurAmount:           "20px",
		SaturateAmount:       "180%",
		DarkModeEnabled:      false,
		NavbarGlassColor:     "rgba(255, 255, 255, 0.85)",
		NavbarTextColor:      "#333333",
		CardGlassColor:       "rgba(255, 255, 255, 0.75)",
		FooterGlassColor:     "rgba(255, 255, 255, 0.9)",
	}

	keys := []string{
		"background_image", "global_opacity", "background_size",
		"background_position", "background_repeat", "background_attachment",
		"blur_amount", "saturate_amount", "dark_mode_enabled",
		"navbar_glass_color", "navbar_text_color", "card_glass_color", "footer_glass_color",
	}

	for _, key := range keys {
		setting, err := repo.GetByKey(key)
		if err != nil {
			log.Printf("Failed to get setting %s: %v", key, err)
			continue
		}
		if setting != nil {
			switch key {
			case "background_image":
				settings.BackgroundImage = setting.Value
			case "global_opacity":
				settings.GlobalOpacity = setting.Value
			case "background_size":
				settings.BackgroundSize = setting.Value
			case "background_position":
				settings.BackgroundPosition = setting.Value
			case "background_repeat":
				settings.BackgroundRepeat = setting.Value
			case "background_attachment":
				settings.BackgroundAttachment = setting.Value
			case "blur_amount":
				settings.BlurAmount = setting.Value
			case "saturate_amount":
				settings.SaturateAmount = setting.Value
			case "dark_mode_enabled":
				settings.DarkModeEnabled = stringToBool(setting.Value)
			case "navbar_glass_color":
				settings.NavbarGlassColor = setting.Value
			case "navbar_text_color":
				settings.NavbarTextColor = setting.Value
			case "card_glass_color":
				settings.CardGlassColor = setting.Value
			case "footer_glass_color":
				settings.FooterGlassColor = setting.Value
			}
		}
	}

	return &settings, nil
}

// UpdateAppearance 更新外观设置
func UpdateAppearance(settings *AppearanceSettings) error {
	repo := db.GetSettingRepository()

	updates := map[string]string{
		"background_image":      settings.BackgroundImage,
		"global_opacity":        settings.GlobalOpacity,
		"background_size":       settings.BackgroundSize,
		"background_position":   settings.BackgroundPosition,
		"background_repeat":     settings.BackgroundRepeat,
		"background_attachment": settings.BackgroundAttachment,
		"blur_amount":           settings.BlurAmount,
		"saturate_amount":       settings.SaturateAmount,
		"dark_mode_enabled":     boolToString(settings.DarkModeEnabled),
		"navbar_glass_color":    settings.NavbarGlassColor,
		"navbar_text_color":     settings.NavbarTextColor,
		"card_glass_color":      settings.CardGlassColor,
		"footer_glass_color":    settings.FooterGlassColor,
	}

	for key, value := range updates {
		if err := repo.UpdateByKey(key, value); err != nil {
			return fmt.Errorf("failed to update setting %s: %w", key, err)
		}
	}

	return nil
}

// UpdateAppearancePartial 增量更新外观设置
func UpdateAppearancePartial(updates map[string]interface{}) error {
	repo := db.GetSettingRepository()

	fieldMapping := map[string]string{
		"background_image":      "background_image",
		"global_opacity":        "global_opacity",
		"background_size":       "background_size",
		"background_position":   "background_position",
		"background_repeat":     "background_repeat",
		"background_attachment": "background_attachment",
		"blur_amount":           "blur_amount",
		"saturate_amount":       "saturate_amount",
		"dark_mode_enabled":     "dark_mode_enabled",
		"navbar_glass_color":    "navbar_glass_color",
		"navbar_text_color":     "navbar_text_color",
		"card_glass_color":      "card_glass_color",
		"footer_glass_color":    "footer_glass_color",
	}

	for jsonField, value := range updates {
		dbKey, exists := fieldMapping[jsonField]
		if !exists {
			log.Printf("Unknown appearance field: %s", jsonField)
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

// GetAppearanceJSON 获取外观设置的JSON格式
func GetAppearanceJSON() (string, error) {
	settings, err := GetAppearance()
	if err != nil {
		return "", err
	}

	data, err := json.Marshal(settings)
	if err != nil {
		return "", err
	}

	return string(data), nil
}