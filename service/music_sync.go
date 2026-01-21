package service

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"myblog-gogogo/db"
)

// MusicSyncService 音乐同步服务
type MusicSyncService struct {
	db *sql.DB
}

// NewMusicSyncService 创建音乐同步服务
func NewMusicSyncService() *MusicSyncService {
	return &MusicSyncService{
		db: db.GetDB(),
	}
}

// SyncMusicFilesToDB 同步music目录中的文件到数据库
func (s *MusicSyncService) SyncMusicFilesToDB() error {
	musicDir := "./music"
	coversDir := filepath.Join(musicDir, "covers")

	// 检查目录是否存在
	if _, err := os.Stat(musicDir); os.IsNotExist(err) {
		log.Println("Music directory does not exist, skipping sync")
		return nil
	}

	// 读取目录中的所有文件
	entries, err := os.ReadDir(musicDir)
	if err != nil {
		return fmt.Errorf("failed to read music directory: %w", err)
	}

	// 读取 covers 目录中的所有封面文件
	coversMap := make(map[string]string) // timestamp -> cover filename
	if coverEntries, err := os.ReadDir(coversDir); err == nil {
		for _, entry := range coverEntries {
			if entry.IsDir() {
				continue
			}
			coverName := entry.Name()
			// 提取时间戳（格式：timestamp_cover.ext）
			if parts := strings.SplitN(coverName, "_cover", 2); len(parts) == 2 {
				timestamp := parts[0]
				coversMap[timestamp] = coverName
			}
		}
	}

	// 获取数据库中已存在的文件
	existingFiles, err := s.getExistingFiles()
	if err != nil {
		return fmt.Errorf("failed to get existing files: %w", err)
	}

	// 遍历文件，同步到数据库
	syncedCount := 0
	updatedCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		filePath := filepath.Join(musicDir, fileName)

		// 检查是否是音频文件
		if !isAudioFile(fileName) {
			continue
		}

		// 提取时间戳（格式：timestamp_filename.ext）
		timestamp := ""
		if parts := strings.SplitN(fileName, "_", 2); len(parts) == 2 {
			timestamp = parts[0]
		}

		// 查找匹配的封面
		coverImage := ""
		if timestamp != "" {
			if coverName, exists := coversMap[timestamp]; exists {
				// 使用 strings.Builder 优化字符串拼接
				var coverPathBuilder strings.Builder
				coverPathBuilder.Grow(len("/music/covers/") + len(coverName))
				coverPathBuilder.WriteString("/music/covers/")
				coverPathBuilder.WriteString(coverName)
				coverImage = coverPathBuilder.String()
			}
		}

		// 如果文件已存在于数据库中，更新封面信息
		if _, exists := existingFiles[fileName]; exists {
			if coverImage != "" {
				// 更新封面
				if err := s.updateTrackCover(fileName, coverImage); err != nil {
					log.Printf("Warning: Failed to update cover for %s: %v", fileName, err)
				} else {
					updatedCount++
					log.Printf("Updated cover for: %s -> %s", fileName, coverImage)
				}
			}
			continue
		}

		// 提取元数据
		metadata, err := GetAudioMetadata(filePath)
		if err != nil {
			log.Printf("Warning: Failed to extract metadata for %s: %v", fileName, err)
		}

		// 准备标题和艺术家
		title := metadata.Title
		if title == "" {
			title = strings.TrimSuffix(fileName, filepath.Ext(fileName))
		}

		artist := metadata.Artist
		if artist == "" {
			artist = "未知艺术家"
		}

		// 插入数据库
		duration := "未知"
		if err := s.insertTrack(title, artist, filePath, fileName, duration, coverImage); err != nil {
			log.Printf("Warning: Failed to insert track %s: %v", fileName, err)
			continue
		}

		syncedCount++
		log.Printf("Synced music file: %s - %s", title, artist)
	}

	// 清理数据库中不存在的文件记录
	deletedCount, err := s.cleanupOrphanedFiles(entries)
	if err != nil {
		log.Printf("Warning: Failed to cleanup orphaned files: %v", err)
	}

	log.Printf("Music sync completed: %d files synced, %d covers updated, %d orphaned records removed", syncedCount, updatedCount, deletedCount)
	return nil
}

// getExistingFiles 获取数据库中已存在的文件
func (s *MusicSyncService) getExistingFiles() (map[string]bool, error) {
	query := `SELECT file_name FROM music_tracks`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make(map[string]bool)
	for rows.Next() {
		var fileName string
		if err := rows.Scan(&fileName); err != nil {
			continue
		}
		files[fileName] = true
	}

	return files, nil
}

// 插入音乐曲目到数据库
func (s *MusicSyncService) insertTrack(title, artist, filePath, fileName, duration, coverImage string) error {
	// 净化标题：移除时间戳和下划线
	cleanTitle := s.cleanTitle(title)
	
	query := `INSERT INTO music_tracks (title, artist, file_path, file_name, duration, cover_image, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query, cleanTitle, artist, filePath, fileName, duration, coverImage, time.Now())
	return err
}

// cleanTitle 净化标题，移除时间戳和下划线前缀
func (s *MusicSyncService) cleanTitle(title string) string {
	// 检查是否以数字开头，后面跟着下划线
	var timestamp int64
	matched, _ := fmt.Sscanf(title, "%d_", &timestamp)
	if matched == 1 {
		// 移除时间戳和下划线
		parts := strings.SplitN(title, "_", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	return title
}

// updateTrackCover 更新音乐曲目的封面
func (s *MusicSyncService) updateTrackCover(fileName, coverImage string) error {
	query := `UPDATE music_tracks SET cover_image = ? WHERE file_name = ?`
	_, err := s.db.Exec(query, coverImage, fileName)
	return err
}

// CleanAllTitles 清理数据库中所有音乐标题，移除时间戳和下划线
func (s *MusicSyncService) CleanAllTitles() error {
	query := `SELECT id, title FROM music_tracks`
	rows, err := s.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query music tracks: %w", err)
	}
	defer rows.Close()

	updatedCount := 0
	for rows.Next() {
		var id int
		var title string
		if err := rows.Scan(&id, &title); err != nil {
			continue
		}

		// 检查标题是否需要清理
		cleanTitle := s.cleanTitle(title)
		if cleanTitle != title {
			// 更新数据库
			updateQuery := `UPDATE music_tracks SET title = ? WHERE id = ?`
			if _, err := s.db.Exec(updateQuery, cleanTitle, id); err != nil {
				log.Printf("Warning: Failed to update title for track %d: %v", id, err)
			} else {
				updatedCount++
				log.Printf("Cleaned title for track %d: %s -> %s", id, title, cleanTitle)
			}
		}
	}

	if updatedCount > 0 {
		log.Printf("Cleaned %d music titles", updatedCount)
	}
	return nil
}

// cleanupOrphanedFiles 清理数据库中不存在的文件记录
func (s *MusicSyncService) cleanupOrphanedFiles(entries []os.DirEntry) (int, error) {
	// 构建当前目录中的文件集合
	filesInDir := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			filesInDir[entry.Name()] = true
		}
	}

	// 获取数据库中的所有文件
	query := `SELECT id, file_name FROM music_tracks`
	rows, err := s.db.Query(query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	deletedCount := 0
	for rows.Next() {
		var id int
		var fileName string
		if err := rows.Scan(&id, &fileName); err != nil {
			continue
		}

		// 如果文件不在目录中，删除记录
		if !filesInDir[fileName] {
			if _, err := s.db.Exec("DELETE FROM music_tracks WHERE id = ?", id); err != nil {
				log.Printf("Warning: Failed to delete orphaned record %s: %v", fileName, err)
				continue
			}
			deletedCount++
			log.Printf("Removed orphaned music record: %s", fileName)
		}
	}

	return deletedCount, nil
}

// isAudioFile 检查文件是否是音频文件
func isAudioFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	allowedExts := map[string]bool{
		".mp3":  true,
		".wav":  true,
		".ogg":  true,
		".m4a":  true,
		".flac": true,
		".aac":  true,
		".wma":  true,
	}
	return allowedExts[ext]
}