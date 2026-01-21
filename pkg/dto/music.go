package dto

import "time"

// MusicDTO 音乐数据传输对象
type MusicDTO struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	Album       string    `json:"album,omitempty"`
	Genre       string    `json:"genre,omitempty"`
	Year        int       `json:"year,omitempty"`
	Duration    float64   `json:"duration"`
	FilePath    string    `json:"file_path"`
	CoverPath   string    `json:"cover_path,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UploadMusicRequest 上传音乐请求
type UploadMusicRequest struct {
	Title     string `json:"title"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	Genre     string `json:"genre"`
	Year      int    `json:"year"`
	File      string `json:"file"`
	CoverFile string `json:"cover_file"`
}

// MusicListRequest 音乐列表请求
type MusicListRequest struct {
	PaginationRequest
	Search  string `json:"search" form:"search"`
	Artist  string `json:"artist" form:"artist"`
	Album   string `json:"album" form:"album"`
	Genre   string `json:"genre" form:"genre"`
	SortBy  string `json:"sort_by" form:"sort_by"`
	SortOrder string `json:"sort_order" form:"sort_order"`
}