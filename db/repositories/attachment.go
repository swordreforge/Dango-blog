package repositories

import (
	"database/sql"
	"time"

	"myblog-gogogo/db/models"
)

// AttachmentRepository 附件仓库接口
type AttachmentRepository interface {
	Create(attachment *models.Attachment) error
	GetByID(id int) (*models.Attachment, error)
	GetAll(limit, offset int) ([]*models.Attachment, int, error)
	GetByPassageID(passageID int, limit, offset int) ([]*models.Attachment, int, error)
	UpdateVisibility(id int, visibility string, showInPassage bool) error
	Delete(id int) error
	Count() (int, error)
	CountByPassageID(passageID int) (int, error)
}

// SQLiteAttachmentRepository SQLite附件仓库实现
type SQLiteAttachmentRepository struct {
	db *sql.DB
}

func NewSQLiteAttachmentRepository(db *sql.DB) *SQLiteAttachmentRepository {
	return &SQLiteAttachmentRepository{db: db}
}

func (r *SQLiteAttachmentRepository) Create(attachment *models.Attachment) error {
	query := `INSERT INTO attachments (file_name, stored_name, file_path, file_type, content_type, file_size, passage_id, uploaded_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	attachment.UploadedAt = now

	result, err := r.db.Exec(query, attachment.FileName, attachment.StoredName, attachment.FilePath,
		attachment.FileType, attachment.ContentType, attachment.FileSize, attachment.PassageID, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	attachment.ID = int(id)
	return nil
}

func (r *SQLiteAttachmentRepository) GetByID(id int) (*models.Attachment, error) {
	query := `SELECT id, file_name, stored_name, file_path, file_type, content_type, file_size, passage_id, visibility, show_in_passage, uploaded_at
	          FROM attachments WHERE id = ?`

	attachment := &models.Attachment{}
	err := r.db.QueryRow(query, id).Scan(
		&attachment.ID, &attachment.FileName, &attachment.StoredName, &attachment.FilePath,
		&attachment.FileType, &attachment.ContentType, &attachment.FileSize,
		&attachment.PassageID, &attachment.Visibility, &attachment.ShowInPassage,
		&attachment.UploadedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

func (r *SQLiteAttachmentRepository) GetAll(limit, offset int) ([]*models.Attachment, int, error) {
	query := `SELECT id, file_name, stored_name, file_path, file_type, content_type, file_size, passage_id, visibility, show_in_passage, uploaded_at
	          FROM attachments ORDER BY uploaded_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var attachments []*models.Attachment
	for rows.Next() {
		var attachment models.Attachment
		err := rows.Scan(
			&attachment.ID, &attachment.FileName, &attachment.StoredName, &attachment.FilePath,
			&attachment.FileType, &attachment.ContentType, &attachment.FileSize,
			&attachment.PassageID, &attachment.Visibility, &attachment.ShowInPassage,
			&attachment.UploadedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		attachments = append(attachments, &attachment)
	}

	// 获取总数
	var total int
	err = r.db.QueryRow("SELECT COUNT(*) FROM attachments").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return attachments, total, nil
}

func (r *SQLiteAttachmentRepository) GetByPassageID(passageID int, limit, offset int) ([]*models.Attachment, int, error) {
	query := `SELECT id, file_name, stored_name, file_path, file_type, content_type, file_size, passage_id, visibility, show_in_passage, uploaded_at
	          FROM attachments WHERE passage_id = ? ORDER BY uploaded_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, passageID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var attachments []*models.Attachment
	for rows.Next() {
		var attachment models.Attachment
		err := rows.Scan(
			&attachment.ID, &attachment.FileName, &attachment.StoredName, &attachment.FilePath,
			&attachment.FileType, &attachment.ContentType, &attachment.FileSize,
			&attachment.PassageID, &attachment.Visibility, &attachment.ShowInPassage,
			&attachment.UploadedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		attachments = append(attachments, &attachment)
	}

	// 获取总数
	var total int
	err = r.db.QueryRow("SELECT COUNT(*) FROM attachments WHERE passage_id = ?", passageID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return attachments, total, nil
}

func (r *SQLiteAttachmentRepository) Delete(id int) error {
	query := `DELETE FROM attachments WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteAttachmentRepository) UpdateVisibility(id int, visibility string, showInPassage bool) error {
	showInPassageInt := 0
	if showInPassage {
		showInPassageInt = 1
	}
	query := `UPDATE attachments SET visibility = ?, show_in_passage = ? WHERE id = ?`
	_, err := r.db.Exec(query, visibility, showInPassageInt, id)
	return err
}

func (r *SQLiteAttachmentRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM attachments").Scan(&count)
	return count, err
}

func (r *SQLiteAttachmentRepository) CountByPassageID(passageID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM attachments WHERE passage_id = ?", passageID).Scan(&count)
	return count, err
}