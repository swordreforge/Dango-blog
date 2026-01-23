package settings

import (
	"encoding/json"
	"fmt"
	"strconv"

	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
)

// GetByKey 根据key获取单个设置
func GetByKey(key string) (string, error) {
	repo := db.GetSettingRepository()
	setting, err := repo.GetByKey(key)
	if err != nil {
		return "", err
	}
	if setting == nil {
		return "", fmt.Errorf("setting not found: %s", key)
	}
	return setting.Value, nil
}

// UpdateByKey 根据key更新单个设置
func UpdateByKey(key, value string) error {
	repo := db.GetSettingRepository()
	return repo.UpdateByKey(key, value)
}

// GetAll 获取所有设置
func GetAll() ([]models.Setting, error) {
	repo := db.GetSettingRepository()
	return repo.GetAll(1000, 0)
}

// GetByCategory 根据分类获取设置
func GetByCategory(category string) ([]models.Setting, error) {
	repo := db.GetSettingRepository()
	return repo.GetByCategory(category, 1000, 0)
}

// boolToString 将布尔值转换为字符串
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// stringToBool 将字符串转换为布尔值
func stringToBool(s string) bool {
	return s == "true"
}

// stringToInt 将字符串转换为整数
func stringToInt(s string) int64 {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	return 0
}

// stringToStringArray 将JSON字符串转换为字符串数组
func stringToStringArray(s string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return []string{}
	}
	return result
}

// stringArrayToString 将字符串数组转换为JSON字符串
func stringArrayToString(arr []string) string {
	if arr == nil || len(arr) == 0 {
		return "[]"
	}
	data, err := json.Marshal(arr)
	if err != nil {
		return "[]"
	}
	return string(data)
}
