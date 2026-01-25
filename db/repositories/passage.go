package repositories

import (
	"context"
	"database/sql"
	"time"

	"myblog-gogogo/db/models"
)

// PassageRepository 文章仓库接口
type PassageRepository interface {
	Create(passage *models.Passage) error
	GetByID(id int) (*models.Passage, error)
	GetAll(limit, offset int) ([]models.Passage, error)
	Update(passage *models.Passage) error
	Delete(id int) error
	GetByStatus(status string, limit, offset int) ([]models.Passage, error)
	GetByCategory(category string, limit, offset int) ([]models.Passage, error)
	GetAllCategories() ([]string, error)
	Count() (int, error)
	CountByStatus(status string) (int, error)
}

// SQLitePassageRepository SQLite文章仓库实现
type SQLitePassageRepository struct {
	db *sql.DB
}

func NewSQLitePassageRepository(db *sql.DB) *SQLitePassageRepository {
	return &SQLitePassageRepository{db: db}
}

// getContext 创建带超时的 context
func (r *SQLitePassageRepository) getContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (r *SQLitePassageRepository) Create(passage *models.Passage) error {
	query := `INSERT INTO passages (title, content, original_content, summary, author, category, status, file_path, visibility, is_scheduled, published_at, show_title, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()

	// 如果 CreatedAt 为零值，使用当前时间
	if passage.CreatedAt.IsZero() {
		passage.CreatedAt = now
	}

	passage.UpdatedAt = now

	// 设置默认值
	if passage.Visibility == "" {
		passage.Visibility = "public"
	}

	// 处理 is_scheduled 布尔值
	isScheduled := 0
	if passage.IsScheduled {
		isScheduled = 1
	}

	// 处理 show_title 布尔值
	showTitle := 1
	if passage.ShowTitle {
		showTitle = 1
	}

	result, err := r.db.Exec(query, passage.Title, passage.Content, passage.OriginalContent, passage.Summary,
		passage.Author, passage.Category, passage.Status, passage.FilePath, passage.Visibility,
		isScheduled, passage.PublishedAt, showTitle, passage.CreatedAt, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	passage.ID = int(id)
	return nil
}

func (r *SQLitePassageRepository) GetByID(id int) (*models.Passage, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	query := `SELECT id, title, content, original_content, summary, author, category, status, file_path, visibility, is_scheduled, published_at, show_title, created_at, updated_at
	          FROM passages WHERE id = ?`

	passage := &models.Passage{}
	var isScheduled int
	var showTitle int
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&passage.ID, &passage.Title, &passage.Content, &passage.OriginalContent, &passage.Summary,
		&passage.Author, &passage.Category, &passage.Status, &passage.FilePath,
		&passage.Visibility, &isScheduled, &passage.PublishedAt, &showTitle,
		&passage.CreatedAt, &passage.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// 转换 is_scheduled 为布尔值
	passage.IsScheduled = isScheduled == 1
	// 转换 show_title 为布尔值
	passage.ShowTitle = showTitle == 1

	return passage, nil
}

func (r *SQLitePassageRepository) GetAll(limit, offset int) ([]models.Passage, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	query := `SELECT id, title, content, original_content, summary, author, category, status, file_path, visibility, is_scheduled, published_at, show_title, created_at, updated_at
	          FROM passages ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passages []models.Passage
	for rows.Next() {
		var passage models.Passage
		var isScheduled int
		var showTitle int
		err := rows.Scan(
			&passage.ID, &passage.Title, &passage.Content, &passage.OriginalContent, &passage.Summary,
			&passage.Author, &passage.Category, &passage.Status, &passage.FilePath,
			&passage.Visibility, &isScheduled, &passage.PublishedAt, &showTitle,
			&passage.CreatedAt, &passage.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		// 转换 is_scheduled 为布尔值
		passage.IsScheduled = isScheduled == 1
		// 转换 show_title 为布尔值
		passage.ShowTitle = showTitle == 1
		passages = append(passages, passage)
	}

	return passages, nil
}

func (r *SQLitePassageRepository) Update(passage *models.Passage) error {
	ctx, cancel := r.getContext()
	defer cancel()

	query := `UPDATE passages SET title = ?, content = ?, original_content = ?, summary = ?, author = ?, category = ?,
	          status = ?, file_path = ?, visibility = ?, is_scheduled = ?, published_at = ?, show_title = ?, updated_at = ? WHERE id = ?`

	passage.UpdatedAt = time.Now()

	// 处理 is_scheduled 布尔值
	isScheduled := 0
	if passage.IsScheduled {
		isScheduled = 1
	}

	// 处理 show_title 布尔值
	showTitle := 0
	if passage.ShowTitle {
		showTitle = 1
	}

	_, err := r.db.ExecContext(ctx, query, passage.Title, passage.Content, passage.OriginalContent, passage.Summary,
		passage.Author, passage.Category, passage.Status, passage.FilePath, passage.Visibility,
		isScheduled, passage.PublishedAt, showTitle, passage.UpdatedAt, passage.ID)

	return err
}

func (r *SQLitePassageRepository) Delete(id int) error {
	ctx, cancel := r.getContext()
	defer cancel()

	query := `DELETE FROM passages WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *SQLitePassageRepository) GetByStatus(status string, limit, offset int) ([]models.Passage, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	query := `SELECT id, title, content, original_content, summary, author, category, status, file_path, visibility, is_scheduled, published_at, show_title, created_at, updated_at
	          FROM passages WHERE status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passages []models.Passage
	for rows.Next() {
		var passage models.Passage
		var isScheduled int
		var showTitle int
		err := rows.Scan(
			&passage.ID, &passage.Title, &passage.Content, &passage.OriginalContent, &passage.Summary,
			&passage.Author, &passage.Category, &passage.Status, &passage.FilePath,
			&passage.Visibility, &isScheduled, &passage.PublishedAt, &showTitle,
			&passage.CreatedAt, &passage.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		// 转换 is_scheduled 为布尔值
		passage.IsScheduled = isScheduled == 1
		// 转换 show_title 为布尔值
		passage.ShowTitle = showTitle == 1
		passages = append(passages, passage)
	}

	return passages, nil
}

func (r *SQLitePassageRepository) Count() (int, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM passages").Scan(&count)
	return count, err
}

func (r *SQLitePassageRepository) CountByStatus(status string) (int, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM passages WHERE status = ?", status).Scan(&count)
	return count, err
}

func (r *SQLitePassageRepository) GetByCategory(category string, limit, offset int) ([]models.Passage, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	query := `SELECT id, title, content, original_content, summary, author, category, status, file_path, visibility, is_scheduled, published_at, show_title, created_at, updated_at
	          FROM passages WHERE category = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, category, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var passages []models.Passage
	for rows.Next() {
		var passage models.Passage
		var isScheduled int
		var showTitle int
		err := rows.Scan(
			&passage.ID, &passage.Title, &passage.Content, &passage.OriginalContent, &passage.Summary,
			&passage.Author, &passage.Category, &passage.Status, &passage.FilePath,
			&passage.Visibility, &isScheduled, &passage.PublishedAt, &showTitle,
			&passage.CreatedAt, &passage.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		// 转换 is_scheduled 为布尔值
		passage.IsScheduled = isScheduled == 1
		// 转换 show_title 为布尔值
		passage.ShowTitle = showTitle == 1
		passages = append(passages, passage)
	}

	return passages, nil
}

func (r *SQLitePassageRepository) GetAllCategories() ([]string, error) {
	ctx, cancel := r.getContext()
	defer cancel()

	query := `SELECT DISTINCT category FROM passages WHERE category != '未分类' ORDER BY category`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}