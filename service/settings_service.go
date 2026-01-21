package service

import (
	"myblog-gogogo/service/settings"
)

// UpdateSettingByKey 更新单个设置
func UpdateSettingByKey(key, value string) error {
	return settings.UpdateByKey(key, value)
}

// MusicSettings 音乐设置类型别名
type MusicSettings = settings.MusicSettings