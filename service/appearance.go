package service

import (
	"myblog-gogogo/service/settings"
)

// GetAppearanceSettingsSafe 获取外观设置，如果失败则返回默认设置
func GetAppearanceSettingsSafe() *settings.AppearanceSettings {
	appearanceSettings, err := settings.GetAppearance()
	if err != nil {
		return &settings.AppearanceSettings{
			BackgroundImage:      "/img/test.png",
			GlobalOpacity:        "0.15",
			BackgroundSize:       "cover",
			BackgroundPosition:   "center",
			BackgroundRepeat:     "no-repeat",
			BackgroundAttachment: "fixed",
			BlurAmount:           "20px",
			SaturateAmount:       "180%",
			DarkModeEnabled:      false,
			NavbarGlassColor:     "rgba(220, 138, 221, 0.15)",
			CardGlassColor:       "rgba(220, 138, 221, 0.2)",
			FooterGlassColor:     "rgba(220, 138, 221, 0.25)",
		}
	}
	return appearanceSettings
}

// GetAppearanceSettingsMap 获取外观设置并转换为 map 格式
func GetAppearanceSettingsMap() map[string]interface{} {
	appearanceSettings := GetAppearanceSettingsSafe()
	return map[string]interface{}{
		"background_image":       appearanceSettings.BackgroundImage,
		"global_opacity":         appearanceSettings.GlobalOpacity,
		"background_size":        appearanceSettings.BackgroundSize,
		"background_position":    appearanceSettings.BackgroundPosition,
		"background_repeat":      appearanceSettings.BackgroundRepeat,
		"background_attachment":  appearanceSettings.BackgroundAttachment,
		"blur_amount":            appearanceSettings.BlurAmount,
		"saturate_amount":        appearanceSettings.SaturateAmount,
		"dark_mode_enabled":      appearanceSettings.DarkModeEnabled,
		"navbar_glass_color":     appearanceSettings.NavbarGlassColor,
		"card_glass_color":       appearanceSettings.CardGlassColor,
		"footer_glass_color":     appearanceSettings.FooterGlassColor,
	}
}