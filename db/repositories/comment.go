package repositories

import (
	"database/sql"
	"time"

	"myblog-gogogo/db/models"
)

// CommentRepository 评论仓库接口
type CommentRepository interface {
	Create(comment *models.Comment) error
	GetByID(id int) (*models.Comment, error)
	GetByPassageID(passageID int, limit, offset int) ([]models.Comment, error)
	GetAll(limit, offset int) ([]models.Comment, error)
	Delete(id int) error
	Count() (int, error)
	CountByPassageID(passageID int) (int, error)
}

// SQLiteCommentRepository SQLite评论仓库实现
type SQLiteCommentRepository struct {
	db *sql.DB
}

func NewSQLiteCommentRepository(db *sql.DB) *SQLiteCommentRepository {
	return &SQLiteCommentRepository{db: db}
}

func (r *SQLiteCommentRepository) Create(comment *models.Comment) error {
	query := `INSERT INTO comments (username, content, passage_id, created_at)
	          VALUES (?, ?, ?, ?)`

	now := time.Now()
	comment.CreatedAt = now

	result, err := r.db.Exec(query, comment.Username, comment.Content, comment.PassageID, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	comment.ID = int(id)
	return nil
}

func (r *SQLiteCommentRepository) GetByID(id int) (*models.Comment, error) {
	query := `SELECT id, username, content, passage_id, created_at
	          FROM comments WHERE id = ?`

	comment := &models.Comment{}
	err := r.db.QueryRow(query, id).Scan(
		&comment.ID, &comment.Username, &comment.Content,
		&comment.PassageID, &comment.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *SQLiteCommentRepository) GetByPassageID(passageID int, limit, offset int) ([]models.Comment, error) {
	query := `SELECT id, username, content, passage_id, created_at
	          FROM comments WHERE passage_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, passageID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID, &comment.Username, &comment.Content,
			&comment.PassageID, &comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *SQLiteCommentRepository) GetAll(limit, offset int) ([]models.Comment, error) {
	query := `SELECT id, username, content, passage_id, created_at
	          FROM comments ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		err := rows.Scan(
			&comment.ID, &comment.Username, &comment.Content,
			&comment.PassageID, &comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (r *SQLiteCommentRepository) Delete(id int) error {
	query := `DELETE FROM comments WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteCommentRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM comments").Scan(&count)
	return count, err
}

func (r *SQLiteCommentRepository) CountByPassageID(passageID int) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM comments WHERE passage_id = ?", passageID).Scan(&count)
	return count, err
}