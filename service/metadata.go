package service

import (
	"fmt"
	"os"
	"strings"

	"github.com/dhowden/tag"
)

// AudioMetadata 音频元数据
type AudioMetadata struct {
	Duration    float64 `json:"duration"`    // 时长（秒）
	Format      string  `json:"format"`      // 格式
	Title       string  `json:"title"`       // 标题
	Artist      string  `json:"artist"`      // 艺术家
	Album       string  `json:"album"`       // 专辑
	Year        int     `json:"year"`        // 年份
	Genre       string  `json:"genre"`       // 流派
	Track       int     `json:"track"`       // 曲目号
	TotalTracks int     `json:"total_tracks"` // 总曲目数
}

// GetAudioMetadata 使用tag库提取音频元数据
func GetAudioMetadata(filePath string) (*AudioMetadata, error) {
	// 打开音频文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 解析音频元数据
	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	// 获取曲目号和总曲目数
	track, totalTracks := metadata.Track()

	// 构建返回的元数据结构
	result := &AudioMetadata{
		Duration:    0, // tag库不提供时长信息
		Format:      strings.ToLower(filepathExt(filePath)),
		Title:       metadata.Title(),
		Artist:      metadata.Artist(),
		Album:       metadata.Album(),
		Year:        metadata.Year(),
		Genre:       metadata.Genre(),
		Track:       track,
		TotalTracks: totalTracks,
	}

	return result, nil
}

// FormatDuration 格式化时长（秒 -> MM:SS 或 HH:MM:SS）
func FormatDuration(seconds float64) string {
	if seconds <= 0 {
		return "00:00"
	}

	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	secs := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

// filepathExt 获取文件扩展名（小写）
func filepathExt(path string) string {
	ext := ""
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			break
		}
		if path[i] == '.' {
			ext = path[i:]
			break
		}
	}
	return strings.ToLower(ext)
}