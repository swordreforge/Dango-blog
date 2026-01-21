package repositories

import (
	"database/sql"
	"time"

	"myblog-gogogo/db/models"
)

// CategoryRepository 分类仓库接口
type CategoryRepository interface {
	Create(category *models.Category) error
	GetByID(id int) (*models.Category, error)
	GetAll() ([]models.Category, error)
	GetAllEnabled() ([]models.Category, error)
	Update(category *models.Category) error
	Delete(id int) error
	UpdateSortOrder(id, sortOrder int) error
	UpdateEnabled(id int, enabled bool) error
	Count() (int, error)
}

// SQLiteCategoryRepository SQLite分类仓库实现
type SQLiteCategoryRepository struct {
	db *sql.DB
}

func NewSQLiteCategoryRepository(db *sql.DB) *SQLiteCategoryRepository {
	return &SQLiteCategoryRepository{db: db}
}

func (r *SQLiteCategoryRepository) Create(category *models.Category) error {
	query := `INSERT INTO categories (name, description, icon, sort_order, is_enabled, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if category.CreatedAt.IsZero() {
		category.CreatedAt = now
	}
	category.UpdatedAt = now

	result, err := r.db.Exec(query, category.Name, category.Description, category.Icon,
		category.SortOrder, category.IsEnabled, category.CreatedAt, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	category.ID = int(id)
	return nil
}

func (r *SQLiteCategoryRepository) GetByID(id int) (*models.Category, error) {
	query := `SELECT id, name, description, icon, sort_order, is_enabled, created_at, updated_at
	          FROM categories WHERE id = ?`

	category := &models.Category{}
	err := r.db.QueryRow(query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.Icon,
		&category.SortOrder, &category.IsEnabled, &category.CreatedAt, &category.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (r *SQLiteCategoryRepository) GetAll() ([]models.Category, error) {
	query := `SELECT id, name, description, icon, sort_order, is_enabled, created_at, updated_at
	          FROM categories ORDER BY sort_order ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Icon,
			&category.SortOrder, &category.IsEnabled, &category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *SQLiteCategoryRepository) GetAllEnabled() ([]models.Category, error) {
	query := `SELECT id, name, description, icon, sort_order, is_enabled, created_at, updated_at
	          FROM categories WHERE is_enabled = 1 ORDER BY sort_order ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.Icon,
			&category.SortOrder, &category.IsEnabled, &category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func (r *SQLiteCategoryRepository) Update(category *models.Category) error {
	query := `UPDATE categories SET name = ?, description = ?, icon = ?, sort_order = ?, is_enabled = ?, updated_at = ?
	          WHERE id = ?`

	category.UpdatedAt = time.Now()

	_, err := r.db.Exec(query, category.Name, category.Description, category.Icon,
		category.SortOrder, category.IsEnabled, category.UpdatedAt, category.ID)

	return err
}

func (r *SQLiteCategoryRepository) Delete(id int) error {
	query := `DELETE FROM categories WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteCategoryRepository) UpdateSortOrder(id, sortOrder int) error {
	query := `UPDATE categories SET sort_order = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, sortOrder, time.Now(), id)
	return err
}

func (r *SQLiteCategoryRepository) UpdateEnabled(id int, enabled bool) error {
	query := `UPDATE categories SET is_enabled = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, enabled, time.Now(), id)
	return err
}

func (r *SQLiteCategoryRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	return count, err
}