package settings

import (
	"encoding/json"
	"fmt"
	"log"

	"myblog-gogogo/db"
)

// AppearanceSettings 外观设置结构
type AppearanceSettings struct {
	BackgroundImage      string   `json:"background_image"`
	MobileBackgroundImage string   `json:"mobile_background_image"`
	GlobalOpacity        string   `json:"global_opacity"`
	BackgroundSize       string   `json:"background_size"`
	BackgroundPosition   string   `json:"background_position"`
	BackgroundRepeat     string   `json:"background_repeat"`
	BackgroundAttachment string   `json:"background_attachment"`
	BlurAmount           string   `json:"blur_amount"`
	SaturateAmount       string   `json:"saturate_amount"`
	DarkModeEnabled      bool     `json:"dark_mode_enabled"`
	NavbarGlassColor     string   `json:"navbar_glass_color"`
	NavbarTextColor      string   `json:"navbar_text_color"`
	CardGlassColor       string   `json:"card_glass_color"`
	FooterGlassColor     string   `json:"footer_glass_color"`
	FloatingTextEnabled  bool     `json:"floating_text_enabled"`
	FloatingTexts        []string `json:"floating_texts"`
}

// GetAppearance 获取外观设置
func GetAppearance() (*AppearanceSettings, error) {
	repo := db.GetSettingRepository()

	settings := AppearanceSettings{
		BackgroundImage:       "/img/test.webp",
		MobileBackgroundImage: "/img/mobile-test.webp",
		GlobalOpacity:         "0.15",
		BackgroundSize:        "cover",
		BackgroundPosition:    "center",
		BackgroundRepeat:      "no-repeat",
		BackgroundAttachment:  "fixed",
		BlurAmount:            "20px",
		SaturateAmount:        "180%",
		DarkModeEnabled:       false,
		NavbarGlassColor:      "rgba(220, 138, 221, 0.15)",
		NavbarTextColor:       "#333333",
		CardGlassColor:        "rgba(220, 138, 221, 0.2)",
		FooterGlassColor:      "rgba(220, 138, 221, 0.25)",
		FloatingTextEnabled:   false,
		FloatingTexts:         []string{"perfect", "good", "excellent", "extraordinary", "legend"},
	}

	keys := []string{
		"background_image", "mobile_background_image", "global_opacity", "background_size",
		"background_position", "background_repeat", "background_attachment",
		"blur_amount", "saturate_amount", "dark_mode_enabled",
		"navbar_glass_color", "navbar_text_color", "card_glass_color", "footer_glass_color",
		"floating_text_enabled", "floating_texts",
	}

	// 使用批量查询
	settingsMap, err := repo.GetByKeys(keys)
	if err != nil {
		log.Printf("Failed to get appearance settings: %v", err)
		return &settings, nil
	}

	for key, setting := range settingsMap {
		if setting != nil {
			switch key {
			case "background_image":
				settings.BackgroundImage = setting.Value
			case "mobile_background_image":
				settings.MobileBackgroundImage = setting.Value
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
			case "floating_text_enabled":
				settings.FloatingTextEnabled = stringToBool(setting.Value)
			case "floating_texts":
				settings.FloatingTexts = stringToStringArray(setting.Value)
			}
		}
	}

	return &settings, nil
}

// UpdateAppearance 更新外观设置
func UpdateAppearance(settings *AppearanceSettings) error {
	repo := db.GetSettingRepository()

	updates := map[string]string{
		"background_image":       settings.BackgroundImage,
		"mobile_background_image": settings.MobileBackgroundImage,
		"global_opacity":         settings.GlobalOpacity,
		"background_size":        settings.BackgroundSize,
		"background_position":    settings.BackgroundPosition,
		"background_repeat":      settings.BackgroundRepeat,
		"background_attachment":  settings.BackgroundAttachment,
		"blur_amount":            settings.BlurAmount,
		"saturate_amount":        settings.SaturateAmount,
		"dark_mode_enabled":      boolToString(settings.DarkModeEnabled),
		"navbar_glass_color":     settings.NavbarGlassColor,
		"navbar_text_color":      settings.NavbarTextColor,
		"card_glass_color":       settings.CardGlassColor,
		"footer_glass_color":     settings.FooterGlassColor,
		"floating_text_enabled":  boolToString(settings.FloatingTextEnabled),
		"floating_texts":         stringArrayToString(settings.FloatingTexts),
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
		"background_image":       "background_image",
		"mobile_background_image": "mobile_background_image",
		"global_opacity":         "global_opacity",
		"background_size":        "background_size",
		"background_position":    "background_position",
		"background_repeat":      "background_repeat",
		"background_attachment":  "background_attachment",
		"blur_amount":            "blur_amount",
		"saturate_amount":        "saturate_amount",
		"dark_mode_enabled":      "dark_mode_enabled",
		"navbar_glass_color":     "navbar_glass_color",
		"navbar_text_color":      "navbar_text_color",
		"card_glass_color":       "card_glass_color",
		"footer_glass_color":     "footer_glass_color",
		"floating_text_enabled":  "floating_text_enabled",
		"floating_texts":         "floating_texts",
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
		case []interface{}:
			// 特殊处理数组类型，转换为 JSON 字符串
			var strArray []string
			for _, item := range v {
				if str, ok := item.(string); ok {
					strArray = append(strArray, str)
				} else {
					strArray = append(strArray, fmt.Sprintf("%v", item))
				}
			}
			stringValue = stringArrayToString(strArray)
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