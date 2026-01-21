package dto

import "time"

// AttachmentDTO 附件数据传输对象
type AttachmentDTO struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`
	FileName    string    `json:"file_name"`
	StoredName  string    `json:"stored_name"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UploadAttachmentRequest 上传附件请求
type UploadAttachmentRequest struct {
	File       interface{} `json:"file"`
	Visibility string      `json:"visibility"`
}

// AttachmentListRequest 附件列表请求
type AttachmentListRequest struct {
	PaginationRequest
	Type       string `json:"type" form:"type"`
	Visibility string `json:"visibility" form:"visibility"`
	Search     string `json:"search" form:"search"`
}