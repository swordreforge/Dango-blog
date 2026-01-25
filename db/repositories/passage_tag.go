package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"myblog-gogogo/db/models"
)

// PassageTagRepository 文章-标签关联仓库接口
type PassageTagRepository interface {
	Create(passageTag *models.PassageTag) error
	DeleteByPassageID(passageID int) error
	DeleteByTagID(tagID int) error
	GetByPassageID(passageID int) ([]*models.PassageTag, error)
	GetByTagID(tagID int) ([]*models.PassageTag, error)
	GetTagIDsByPassageID(passageID int) ([]int, error)
	GetPassageIDsByTagID(tagID int) ([]int, error)
	CountByTagID(tagID int) (int, error)
}

type passageTagRepository struct {
	db *sql.DB
}

// NewPassageTagRepository 创建文章-标签关联仓库
func NewPassageTagRepository(db *sql.DB) PassageTagRepository {
	return &passageTagRepository{db: db}
}

// Create 创建文章-标签关联
func (r *passageTagRepository) Create(passageTag *models.PassageTag) error {
	query := `
		INSERT INTO passage_tags (passage_id, tag_id, created_at)
		VALUES (?, ?, ?)
	`
	result, err := r.db.Exec(query, passageTag.PassageID, passageTag.TagID, time.Now())
	if err != nil {
		return fmt.Errorf("创建文章-标签关联失败: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("获取插入ID失败: %w", err)
	}

	passageTag.ID = int(id)
	return nil
}

// DeleteByPassageID 删除文章的所有标签关联
func (r *passageTagRepository) DeleteByPassageID(passageID int) error {
	query := `DELETE FROM passage_tags WHERE passage_id = ?`
	_, err := r.db.Exec(query, passageID)
	if err != nil {
		return fmt.Errorf("删除文章标签关联失败: %w", err)
	}
	return nil
}

// DeleteByTagID 删除标签的所有文章关联
func (r *passageTagRepository) DeleteByTagID(tagID int) error {
	query := `DELETE FROM passage_tags WHERE tag_id = ?`
	_, err := r.db.Exec(query, tagID)
	if err != nil {
		return fmt.Errorf("删除标签文章关联失败: %w", err)
	}
	return nil
}

// GetByPassageID 获取文章的所有标签关联
func (r *passageTagRepository) GetByPassageID(passageID int) ([]*models.PassageTag, error) {
	query := `
		SELECT id, passage_id, tag_id, created_at
		FROM passage_tags
		WHERE passage_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, passageID)
	if err != nil {
		return nil, fmt.Errorf("获取文章标签关联失败: %w", err)
	}
	defer rows.Close()

	var passageTags []*models.PassageTag
	for rows.Next() {
		pt := &models.PassageTag{}
		err := rows.Scan(&pt.ID, &pt.PassageID, &pt.TagID, &pt.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描文章标签关联失败: %w", err)
		}
		passageTags = append(passageTags, pt)
	}

	return passageTags, nil
}

// GetByTagID 获取标签的所有文章关联
func (r *passageTagRepository) GetByTagID(tagID int) ([]*models.PassageTag, error) {
	query := `
		SELECT id, passage_id, tag_id, created_at
		FROM passage_tags
		WHERE tag_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, tagID)
	if err != nil {
		return nil, fmt.Errorf("获取标签文章关联失败: %w", err)
	}
	defer rows.Close()

	var passageTags []*models.PassageTag
	for rows.Next() {
		pt := &models.PassageTag{}
		err := rows.Scan(&pt.ID, &pt.PassageID, &pt.TagID, &pt.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("扫描标签文章关联失败: %w", err)
		}
		passageTags = append(passageTags, pt)
	}

	return passageTags, nil
}

// GetTagIDsByPassageID 获取文章的所有标签ID
func (r *passageTagRepository) GetTagIDsByPassageID(passageID int) ([]int, error) {
	query := `SELECT tag_id FROM passage_tags WHERE passage_id = ? ORDER BY tag_id`
	rows, err := r.db.Query(query, passageID)
	if err != nil {
		return nil, fmt.Errorf("获取文章标签ID失败: %w", err)
	}
	defer rows.Close()

	var tagIDs []int
	for rows.Next() {
		var tagID int
		err := rows.Scan(&tagID)
		if err != nil {
			return nil, fmt.Errorf("扫描标签ID失败: %w", err)
		}
		tagIDs = append(tagIDs, tagID)
	}

	return tagIDs, nil
}

// GetPassageIDsByTagID 获取标签的所有文章ID
func (r *passageTagRepository) GetPassageIDsByTagID(tagID int) ([]int, error) {
	query := `SELECT passage_id FROM passage_tags WHERE tag_id = ? ORDER BY passage_id`
	rows, err := r.db.Query(query, tagID)
	if err != nil {
		return nil, fmt.Errorf("获取标签文章ID失败: %w", err)
	}
	defer rows.Close()

	var passageIDs []int
	for rows.Next() {
		var passageID int
		err := rows.Scan(&passageID)
		if err != nil {
			return nil, fmt.Errorf("扫描文章ID失败: %w", err)
		}
		passageIDs = append(passageIDs, passageID)
	}

	return passageIDs, nil
}

// CountByTagID 统计标签被使用的次数
func (r *passageTagRepository) CountByTagID(tagID int) (int, error) {
	query := `SELECT COUNT(*) FROM passage_tags WHERE tag_id = ?`
	var count int
	err := r.db.QueryRow(query, tagID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("统计标签使用次数失败: %w", err)
	}
	return count, nil
}