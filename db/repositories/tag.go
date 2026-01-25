package repositories

import (
	"database/sql"
	"time"

	"myblog-gogogo/db/models"
)

// TagRepository 标签仓库接口
type TagRepository interface {
	Create(tag *models.Tag) error
	GetByID(id int) (*models.Tag, error)
	GetAll() ([]models.Tag, error)
	GetAllEnabled() ([]models.Tag, error)
	GetByCategoryID(categoryID int) ([]models.Tag, error)
	Update(tag *models.Tag) error
	Delete(id int) error
	UpdateSortOrder(id, sortOrder int) error
	UpdateEnabled(id int, enabled bool) error
	UpdateUsageCount(id, count int) error
	IncrementUsageCount(id int) error
	DecrementUsageCount(id int) error
	Count() (int, error)
	// 事务方法
	UpdateUsageCountInTx(tx *sql.Tx, id, count int) error
	IncrementUsageCountInTx(tx *sql.Tx, id int) error
	DecrementUsageCountInTx(tx *sql.Tx, id int) error
}

// SQLiteTagRepository SQLite标签仓库实现
type SQLiteTagRepository struct {
	db *sql.DB
}

func NewSQLiteTagRepository(db *sql.DB) *SQLiteTagRepository {
	return &SQLiteTagRepository{db: db}
}

func (r *SQLiteTagRepository) Create(tag *models.Tag) error {
	query := `INSERT INTO tags (name, description, color, category_id, sort_order, is_enabled, usage_count, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if tag.CreatedAt.IsZero() {
		tag.CreatedAt = now
	}
	tag.UpdatedAt = now

	result, err := r.db.Exec(query, tag.Name, tag.Description, tag.Color,
		tag.CategoryID, tag.SortOrder, tag.IsEnabled, tag.UsageCount, tag.CreatedAt, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	tag.ID = int(id)
	return nil
}

func (r *SQLiteTagRepository) GetByID(id int) (*models.Tag, error) {
	query := `SELECT id, name, description, color, category_id, sort_order, is_enabled, usage_count, created_at, updated_at
	          FROM tags WHERE id = ?`

	tag := &models.Tag{}
	err := r.db.QueryRow(query, id).Scan(
		&tag.ID, &tag.Name, &tag.Description, &tag.Color,
		&tag.CategoryID, &tag.SortOrder, &tag.IsEnabled, &tag.UsageCount, &tag.CreatedAt, &tag.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func (r *SQLiteTagRepository) GetAll() ([]models.Tag, error) {
	query := `SELECT id, name, description, color, category_id, sort_order, is_enabled, usage_count, created_at, updated_at
	          FROM tags ORDER BY sort_order ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Description, &tag.Color,
			&tag.CategoryID, &tag.SortOrder, &tag.IsEnabled, &tag.UsageCount, &tag.CreatedAt, &tag.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *SQLiteTagRepository) GetAllEnabled() ([]models.Tag, error) {
	query := `SELECT id, name, description, color, category_id, sort_order, is_enabled, usage_count, created_at, updated_at
	          FROM tags WHERE is_enabled = 1 ORDER BY sort_order ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Description, &tag.Color,
			&tag.CategoryID, &tag.SortOrder, &tag.IsEnabled, &tag.UsageCount, &tag.CreatedAt, &tag.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *SQLiteTagRepository) GetByCategoryID(categoryID int) ([]models.Tag, error) {
	query := `SELECT id, name, description, color, category_id, sort_order, is_enabled, usage_count, created_at, updated_at
	          FROM tags WHERE category_id = ? ORDER BY sort_order ASC`

	rows, err := r.db.Query(query, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(
			&tag.ID, &tag.Name, &tag.Description, &tag.Color,
			&tag.CategoryID, &tag.SortOrder, &tag.IsEnabled, &tag.UsageCount, &tag.CreatedAt, &tag.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *SQLiteTagRepository) Update(tag *models.Tag) error {
	query := `UPDATE tags SET name = ?, description = ?, color = ?, category_id = ?, sort_order = ?, is_enabled = ?, usage_count = ?, updated_at = ?
	          WHERE id = ?`

	tag.UpdatedAt = time.Now()

	_, err := r.db.Exec(query, tag.Name, tag.Description, tag.Color,
		tag.CategoryID, tag.SortOrder, tag.IsEnabled, tag.UsageCount, tag.UpdatedAt, tag.ID)

	return err
}

func (r *SQLiteTagRepository) Delete(id int) error {
	query := `DELETE FROM tags WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteTagRepository) UpdateSortOrder(id, sortOrder int) error {
	query := `UPDATE tags SET sort_order = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, sortOrder, time.Now(), id)
	return err
}

func (r *SQLiteTagRepository) UpdateEnabled(id int, enabled bool) error {
	query := `UPDATE tags SET is_enabled = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, enabled, time.Now(), id)
	return err
}

func (r *SQLiteTagRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM tags").Scan(&count)
	return count, err
}

func (r *SQLiteTagRepository) UpdateUsageCount(id, count int) error {
	query := `UPDATE tags SET usage_count = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, count, time.Now(), id)
	return err
}

func (r *SQLiteTagRepository) IncrementUsageCount(id int) error {
	query := `UPDATE tags SET usage_count = usage_count + 1, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

func (r *SQLiteTagRepository) DecrementUsageCount(id int) error {
	query := `UPDATE tags SET usage_count = CASE WHEN usage_count > 0 THEN usage_count - 1 ELSE 0 END, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), id)
	return err
}

// UpdateUsageCountInTx 在事务中更新标签使用次数
func (r *SQLiteTagRepository) UpdateUsageCountInTx(tx *sql.Tx, id, count int) error {
	query := `UPDATE tags SET usage_count = ?, updated_at = ? WHERE id = ?`
	_, err := tx.Exec(query, count, time.Now(), id)
	return err
}

// IncrementUsageCountInTx 在事务中增加标签使用次数
func (r *SQLiteTagRepository) IncrementUsageCountInTx(tx *sql.Tx, id int) error {
	query := `UPDATE tags SET usage_count = usage_count + 1, updated_at = ? WHERE id = ?`
	_, err := tx.Exec(query, time.Now(), id)
	return err
}

// DecrementUsageCountInTx 在事务中减少标签使用次数
func (r *SQLiteTagRepository) DecrementUsageCountInTx(tx *sql.Tx, id int) error {
	query := `UPDATE tags SET usage_count = CASE WHEN usage_count > 0 THEN usage_count - 1 ELSE 0 END, updated_at = ? WHERE id = ?`
	_, err := tx.Exec(query, time.Now(), id)
	return err
}