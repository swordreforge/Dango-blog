package repositories

import (
	"database/sql"
	"time"

	"myblog-gogogo/db/models"
)

// AboutMainCardRepository 主卡片仓库接口
type AboutMainCardRepository interface {
	Create(card *models.AboutMainCard) error
	GetByID(id int) (*models.AboutMainCard, error)
	GetAll() ([]models.AboutMainCard, error)
	GetAllEnabled() ([]models.AboutMainCard, error)
	Update(card *models.AboutMainCard) error
	Delete(id int) error
	UpdateSortOrder(id, sortOrder int) error
	UpdateEnabled(id int, enabled bool) error
}

// AboutSubCardRepository 次卡片仓库接口
type AboutSubCardRepository interface {
	Create(card *models.AboutSubCard) error
	GetByID(id int) (*models.AboutSubCard, error)
	GetByMainCardID(mainCardID int) ([]models.AboutSubCard, error)
	GetByMainCardIDEnabled(mainCardID int) ([]models.AboutSubCard, error)
	GetAll() ([]models.AboutSubCard, error)
	Update(card *models.AboutSubCard) error
	Delete(id int) error
	DeleteByMainCardID(mainCardID int) error
	UpdateSortOrder(id, sortOrder int) error
	UpdateEnabled(id int, enabled bool) error
}

// SQLiteAboutMainCardRepository SQLite主卡片仓库实现
type SQLiteAboutMainCardRepository struct {
	db *sql.DB
}

func NewSQLiteAboutMainCardRepository(db *sql.DB) *SQLiteAboutMainCardRepository {
	return &SQLiteAboutMainCardRepository{db: db}
}

func (r *SQLiteAboutMainCardRepository) Create(card *models.AboutMainCard) error {
	query := `INSERT INTO about_main_cards (title, icon, layout_type, custom_css, sort_order, is_enabled, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if card.CreatedAt.IsZero() {
		card.CreatedAt = now
	}
	card.UpdatedAt = now

	result, err := r.db.Exec(query, card.Title, card.Icon, card.LayoutType, card.CustomCSS,
		card.SortOrder, card.IsEnabled, card.CreatedAt, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	card.ID = int(id)
	return nil
}

func (r *SQLiteAboutMainCardRepository) GetByID(id int) (*models.AboutMainCard, error) {
	query := `SELECT id, title, icon, layout_type, custom_css, sort_order, is_enabled, created_at, updated_at
	          FROM about_main_cards WHERE id = ?`

	card := &models.AboutMainCard{}
	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.Title, &card.Icon, &card.LayoutType, &card.CustomCSS,
		&card.SortOrder, &card.IsEnabled, &card.CreatedAt, &card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (r *SQLiteAboutMainCardRepository) GetAll() ([]models.AboutMainCard, error) {
	query := `SELECT id, title, icon, layout_type, custom_css, sort_order, is_enabled, created_at, updated_at
	          FROM about_main_cards ORDER BY sort_order ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.AboutMainCard
	for rows.Next() {
		var card models.AboutMainCard
		err := rows.Scan(
			&card.ID, &card.Title, &card.Icon, &card.LayoutType, &card.CustomCSS,
			&card.SortOrder, &card.IsEnabled, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (r *SQLiteAboutMainCardRepository) GetAllEnabled() ([]models.AboutMainCard, error) {
	query := `SELECT id, title, icon, layout_type, custom_css, sort_order, is_enabled, created_at, updated_at
	          FROM about_main_cards WHERE is_enabled = 1 ORDER BY sort_order ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.AboutMainCard
	for rows.Next() {
		var card models.AboutMainCard
		err := rows.Scan(
			&card.ID, &card.Title, &card.Icon, &card.LayoutType, &card.CustomCSS,
			&card.SortOrder, &card.IsEnabled, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (r *SQLiteAboutMainCardRepository) Update(card *models.AboutMainCard) error {
	query := `UPDATE about_main_cards SET title = ?, icon = ?, layout_type = ?, custom_css = ?,
	          sort_order = ?, is_enabled = ?, updated_at = ? WHERE id = ?`

	card.UpdatedAt = time.Now()

	_, err := r.db.Exec(query, card.Title, card.Icon, card.LayoutType, card.CustomCSS,
		card.SortOrder, card.IsEnabled, card.UpdatedAt, card.ID)

	return err
}

func (r *SQLiteAboutMainCardRepository) Delete(id int) error {
	query := `DELETE FROM about_main_cards WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteAboutMainCardRepository) UpdateSortOrder(id, sortOrder int) error {
	query := `UPDATE about_main_cards SET sort_order = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, sortOrder, time.Now(), id)
	return err
}

func (r *SQLiteAboutMainCardRepository) UpdateEnabled(id int, enabled bool) error {
	query := `UPDATE about_main_cards SET is_enabled = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, enabled, time.Now(), id)
	return err
}

// SQLiteAboutSubCardRepository SQLite次卡片仓库实现
type SQLiteAboutSubCardRepository struct {
	db *sql.DB
}

func NewSQLiteAboutSubCardRepository(db *sql.DB) *SQLiteAboutSubCardRepository {
	return &SQLiteAboutSubCardRepository{db: db}
}

func (r *SQLiteAboutSubCardRepository) Create(card *models.AboutSubCard) error {
	query := `INSERT INTO about_sub_cards (main_card_id, title, description, icon, link_url, layout_type, custom_css, sort_order, is_enabled, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if card.CreatedAt.IsZero() {
		card.CreatedAt = now
	}
	card.UpdatedAt = now

	result, err := r.db.Exec(query, card.MainCardID, card.Title, card.Description, card.Icon,
		card.LinkURL, card.LayoutType, card.CustomCSS, card.SortOrder, card.IsEnabled, card.CreatedAt, now)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	card.ID = int(id)
	return nil
}

func (r *SQLiteAboutSubCardRepository) GetByID(id int) (*models.AboutSubCard, error) {
	query := `SELECT id, main_card_id, title, description, icon, link_url, layout_type, custom_css, sort_order, is_enabled, created_at, updated_at
	          FROM about_sub_cards WHERE id = ?`

	card := &models.AboutSubCard{}
	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.MainCardID, &card.Title, &card.Description, &card.Icon,
		&card.LinkURL, &card.LayoutType, &card.CustomCSS, &card.SortOrder,
		&card.IsEnabled, &card.CreatedAt, &card.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (r *SQLiteAboutSubCardRepository) GetByMainCardID(mainCardID int) ([]models.AboutSubCard, error) {
	query := `SELECT id, main_card_id, title, description, icon, link_url, layout_type, custom_css, sort_order, is_enabled, created_at, updated_at
	          FROM about_sub_cards WHERE main_card_id = ? ORDER BY sort_order ASC`

	rows, err := r.db.Query(query, mainCardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.AboutSubCard
	for rows.Next() {
		var card models.AboutSubCard
		err := rows.Scan(
			&card.ID, &card.MainCardID, &card.Title, &card.Description, &card.Icon,
			&card.LinkURL, &card.LayoutType, &card.CustomCSS, &card.SortOrder,
			&card.IsEnabled, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (r *SQLiteAboutSubCardRepository) GetByMainCardIDEnabled(mainCardID int) ([]models.AboutSubCard, error) {
	query := `SELECT id, main_card_id, title, description, icon, link_url, layout_type, custom_css, sort_order, is_enabled, created_at, updated_at
	          FROM about_sub_cards WHERE main_card_id = ? AND is_enabled = 1 ORDER BY sort_order ASC`

	rows, err := r.db.Query(query, mainCardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.AboutSubCard
	for rows.Next() {
		var card models.AboutSubCard
		err := rows.Scan(
			&card.ID, &card.MainCardID, &card.Title, &card.Description, &card.Icon,
			&card.LinkURL, &card.LayoutType, &card.CustomCSS, &card.SortOrder,
			&card.IsEnabled, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (r *SQLiteAboutSubCardRepository) GetAll() ([]models.AboutSubCard, error) {
	query := `SELECT id, main_card_id, title, description, icon, link_url, layout_type, custom_css, sort_order, is_enabled, created_at, updated_at
	          FROM about_sub_cards ORDER BY main_card_id, sort_order ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.AboutSubCard
	for rows.Next() {
		var card models.AboutSubCard
		err := rows.Scan(
			&card.ID, &card.MainCardID, &card.Title, &card.Description, &card.Icon,
			&card.LinkURL, &card.LayoutType, &card.CustomCSS, &card.SortOrder,
			&card.IsEnabled, &card.CreatedAt, &card.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}

	return cards, nil
}

func (r *SQLiteAboutSubCardRepository) Update(card *models.AboutSubCard) error {
	query := `UPDATE about_sub_cards SET main_card_id = ?, title = ?, description = ?, icon = ?, link_url = ?,
	          layout_type = ?, custom_css = ?, sort_order = ?, is_enabled = ?, updated_at = ? WHERE id = ?`

	card.UpdatedAt = time.Now()

	_, err := r.db.Exec(query, card.MainCardID, card.Title, card.Description, card.Icon,
		card.LinkURL, card.LayoutType, card.CustomCSS, card.SortOrder, card.IsEnabled, card.UpdatedAt, card.ID)

	return err
}

func (r *SQLiteAboutSubCardRepository) Delete(id int) error {
	query := `DELETE FROM about_sub_cards WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *SQLiteAboutSubCardRepository) DeleteByMainCardID(mainCardID int) error {
	query := `DELETE FROM about_sub_cards WHERE main_card_id = ?`
	_, err := r.db.Exec(query, mainCardID)
	return err
}

func (r *SQLiteAboutSubCardRepository) UpdateSortOrder(id, sortOrder int) error {
	query := `UPDATE about_sub_cards SET sort_order = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, sortOrder, time.Now(), id)
	return err
}

func (r *SQLiteAboutSubCardRepository) UpdateEnabled(id int, enabled bool) error {
	query := `UPDATE about_sub_cards SET is_enabled = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, enabled, time.Now(), id)
	return err
}