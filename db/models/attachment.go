package models

import "time"

// Attachment 附件模型
type Attachment struct {
	ID            int        `json:"id"`
	FileName      string     `json:"file_name"`   // 原始文件名
	StoredName    string     `json:"stored_name"` // 存储文件名
	FilePath      string     `json:"file_path"`   // 文件路径
	FileType      string     `json:"file_type"`   // 文件类型: image, document, video, audio, archive
	ContentType   string     `json:"content_type"` // MIME类型
	FileSize      int64      `json:"file_size"`   // 文件大小（字节）
	PassageID     *int       `json:"passage_id"`  // 关联文章ID，NULL表示未关联
	Visibility    string     `json:"visibility"`  // 可见性: public(公开), private(私密), protected(受保护)
	ShowInPassage bool       `json:"show_in_passage"` // 是否在文章中显示
	UploadedAt    time.Time  `json:"uploaded_at"` // 上传时间
}