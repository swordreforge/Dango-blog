package repositories

import (
	"database/sql"
	"strings"
	"time"

	"myblog-gogogo/db/models"
)

// SettingRepository 设置仓库接口
type SettingRepository interface {
	Create(setting *models.Setting) error
	GetByID(id int) (*models.Setting, error)
	GetByKey(key string) (*models.Setting, error)
	GetByKeys(keys []string) (map[string]*models.Setting, error)
	GetAll(limit, offset int) ([]models.Setting, error)
	GetByCategory(category string, limit, offset int) ([]models.Setting, error)
	Update(setting *models.Setting) error
	UpdateByKey(key, value string) error
	Delete(id int) error
	Count() (int, error)
}

// SQLiteSettingRepository SQLite设置仓库实现
type SQLiteSettingRepository struct {
	db *sql.DB
}

func NewSQLiteSettingRepository(db *sql.DB) *SQLiteSettingRepository {
	return &SQLiteSettingRepository{db: db}
}

func (r *SQLiteSettingRepository) Create(setting *models.Setting) error {
	query := `INSERT INTO settings (key, value, type, description, category, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if setting.CreatedAt.IsZero() {
		setting.CreatedAt = now
	}
	setting.UpdatedAt = now

	result, err := r.db.Exec(query, setting.Key, setting.Value, setting.Type,
		setting.Description, setting.Category, setting.CreatedAt, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	setting.ID = int(id)
	return nil
}

func (r *SQLiteSettingRepository) GetByID(id int) (*models.Setting, error) {
	query := `SELECT id, key, value, type, description, category, created_at, updated_at
	          FROM settings WHERE id = ?`

	setting := &models.Setting{}
	err := r.db.QueryRow(query, id).Scan(
		&setting.ID, &setting.Key, &setting.Value, &setting.Type,
		&setting.Description, &setting.Category, &setting.CreatedAt, &setting.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return setting, nil
}

func (r *SQLiteSettingRepository) GetByKey(key string) (*models.Setting, error) {
	query := `SELECT id, key, value, type, description, category, created_at, updated_at
	          FROM settings WHERE key = ?`

	setting := &models.Setting{}
	err := r.db.QueryRow(query, key).Scan(
		&setting.ID, &setting.Key, &setting.Value, &setting.Type,
		&setting.Description, &setting.Category, &setting.CreatedAt, &setting.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return setting, nil
}

func (r *SQLiteSettingRepository) GetByKeys(keys []string) (map[string]*models.Setting, error) {
	if len(keys) == 0 {
		return make(map[string]*models.Setting), nil
	}

	// 构建查询参数
	placeholders := make([]string, len(keys))
	args := make([]interface{}, len(keys))
	for i, key := range keys {
		placeholders[i] = "?"
		args[i] = key
	}

	query := `SELECT id, key, value, type, description, category, created_at, updated_at
	          FROM settings WHERE key IN (` + strings.Join(placeholders, ",") + `)`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]*models.Setting)
	for rows.Next() {
		setting := &models.Setting{}
		err := rows.Scan(
			&setting.ID, &setting.Key, &setting.Value, &setting.Type,
			&setting.Description, &setting.Category, &setting.CreatedAt, &setting.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result[setting.Key] = setting
	}

	return result, nil
}

func (r *SQLiteSettingRepository) GetAll(limit, offset int) ([]models.Setting, error) {
	query := `SELECT id, key, value, type, description, category, created_at, updated_at
	          FROM settings ORDER BY category, key LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.Setting
	for rows.Next() {
		var setting models.Setting
		err := rows.Scan(
			&setting.ID, &setting.Key, &setting.Value, &setting.Type,
			&setting.Description, &setting.Category, &setting.CreatedAt, &setting.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

func (r *SQLiteSettingRepository) GetByCategory(category string, limit, offset int) ([]models.Setting, error) {
	query := `SELECT id, key, value, type, description, category, created_at, updated_at
	          FROM settings WHERE category = ? ORDER BY key LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, category, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []models.Setting
	for rows.Next() {
		var setting models.Setting
		err := rows.Scan(
			&setting.ID, &setting.Key, &setting.Value, &setting.Type,
			&setting.Description, &setting.Category, &setting.CreatedAt, &setting.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

func (r *SQLiteSettingRepository) Update(setting *models.Setting) error {
	query := `UPDATE settings SET key = ?, value = ?, type = ?, description = ?, category = ?, updated_at = ?
	          WHERE id = ?`

	setting.UpdatedAt = time.Now()

	_, err := r.db.Exec(query, setting.Key, setting.Value, setting.Type,
		setting.Description, setting.Category, setting.UpdatedAt, setting.ID)

	return err
}

func (r *SQLiteSettingRepository) UpdateByKey(key, value string) error {
	query := `UPDATE settings SET value = ?, updated_at = ? WHERE key = ?`

	_, err := r.db.Exec(query, value, time.Now(), key)
	return err
}

func (r *SQLiteSettingRepository) Delete(id int) error {
	query := `DELETE FROM settings WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteSettingRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM settings").Scan(&count)
	return count, err
}