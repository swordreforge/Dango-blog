package repositories

import (
	"database/sql"
	"strings"
	"time"

	"myblog-gogogo/db/models"
)

// UserRepository 用户仓库接口
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id int) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetAll(limit, offset int) ([]models.User, error)
	Update(user *models.User) error
	UpdatePartial(id int, updates map[string]interface{}) error
	Delete(id int) error
	Count() (int, error)
}

// SQLiteUserRepository SQLite用户仓库实现
type SQLiteUserRepository struct {
	db *sql.DB
}

func NewSQLiteUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) Create(user *models.User) error {
	query := `INSERT INTO users (username, password, email, role, status, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()

	// 如果 CreatedAt 为零值，使用当前时间
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}

	user.UpdatedAt = now

	result, err := r.db.Exec(query, user.Username, user.Password, user.Email,
		user.Role, user.Status, user.CreatedAt, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

func (r *SQLiteUserRepository) GetByID(id int) (*models.User, error) {
	query := `SELECT id, username, password, email, role, status, created_at, updated_at
	          FROM users WHERE id = ?`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *SQLiteUserRepository) GetByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, password, email, role, status, created_at, updated_at
	          FROM users WHERE username = ?`

	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *SQLiteUserRepository) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, password, email, role, status, created_at, updated_at
	          FROM users WHERE email = ?`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *SQLiteUserRepository) GetAll(limit, offset int) ([]models.User, error) {
	query := `SELECT id, username, password, email, role, status, created_at, updated_at
	          FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Password, &user.Email,
			&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *SQLiteUserRepository) Update(user *models.User) error {
	query := `UPDATE users SET username = ?, password = ?, email = ?, role = ?,
	          status = ?, updated_at = ? WHERE id = ?`

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(query, user.Username, user.Password, user.Email,
		user.Role, user.Status, user.UpdatedAt, user.ID)

	return err
}

func (r *SQLiteUserRepository) UpdatePartial(id int, updates map[string]interface{}) error {
	// 构建动态 UPDATE 语句
	if len(updates) == 0 {
		return nil
	}

	// 允许更新的字段
	allowedFields := map[string]bool{
		"username": true,
		"password": true,
		"email":    true,
		"role":     true,
		"status":   true,
	}

	// 构建SET子句和参数
	var setClauses []string
	var args []interface{}

	for field, value := range updates {
		if !allowedFields[field] {
			continue // 跳过不允许的字段
		}
		setClauses = append(setClauses, field+" = ?")
		args = append(args, value)
	}

	if len(setClauses) == 0 {
		return nil // 没有可更新的字段
	}

	// 添加 updated_at 字段
	setClauses = append(setClauses, "updated_at = ?")
	args = append(args, time.Now())

	// 添加 WHERE 条件的参数
	args = append(args, id)

	// 构建完整查询
	query := "UPDATE users SET " + strings.Join(setClauses, ", ") + " WHERE id = ?"

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *SQLiteUserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteUserRepository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}