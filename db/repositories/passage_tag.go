package repositories

import (
	"database/sql"
	"fmt"
	"strings"
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
	GetAllTagCounts() (map[int]int, error)
	// 批量查询方法
	GetTagIDsByPassageIDs(passageIDs []int) (map[int][]int, error)
	// 事务方法
	BatchCreate(passageID int, tagIDs []int) error
	ReplaceAll(passageID int, tagIDs []int) error
}

type passageTagRepository struct {
	db       *sql.DB
	tagRepo  TagRepository
}

// NewPassageTagRepository 创建文章-标签关联仓库
func NewPassageTagRepository(db *sql.DB) PassageTagRepository {
	return &passageTagRepository{db: db, tagRepo: NewSQLiteTagRepository(db)}
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

	// 增加标签的使用次数
	if err := r.tagRepo.IncrementUsageCount(passageTag.TagID); err != nil {
		return fmt.Errorf("更新标签使用次数失败: %w", err)
	}

	return nil
}

// DeleteByPassageID 删除文章的所有标签关联
func (r *passageTagRepository) DeleteByPassageID(passageID int) error {
	// 先获取要删除的标签ID列表
	tagIDs, err := r.GetTagIDsByPassageID(passageID)
	if err != nil {
		return fmt.Errorf("获取标签ID失败: %w", err)
	}

	query := `DELETE FROM passage_tags WHERE passage_id = ?`
	_, err = r.db.Exec(query, passageID)
	if err != nil {
		return fmt.Errorf("删除文章标签关联失败: %w", err)
	}

	// 减少被删除标签的使用次数
	for _, tagID := range tagIDs {
		if err := r.tagRepo.DecrementUsageCount(tagID); err != nil {
			return fmt.Errorf("更新标签使用次数失败: %w", err)
		}
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

	// 将标签的使用次数重置为0
	if err := r.tagRepo.UpdateUsageCount(tagID, 0); err != nil {
		return fmt.Errorf("更新标签使用次数失败: %w", err)
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

// GetAllTagCounts 批量获取所有标签的使用次数
func (r *passageTagRepository) GetAllTagCounts() (map[int]int, error) {
	query := `SELECT tag_id, COUNT(*) as count FROM passage_tags GROUP BY tag_id`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("批量统计标签使用次数失败: %w", err)
	}
	defer rows.Close()

	result := make(map[int]int)
	for rows.Next() {
		var tagID, count int
		if err := rows.Scan(&tagID, &count); err != nil {
			return nil, fmt.Errorf("扫描标签统计结果失败: %w", err)
		}
		result[tagID] = count
	}
	return result, nil
}

// GetTagIDsByPassageIDs 批量获取多篇文章的标签ID
func (r *passageTagRepository) GetTagIDsByPassageIDs(passageIDs []int) (map[int][]int, error) {
	if len(passageIDs) == 0 {
		return make(map[int][]int), nil
	}

	// 构建查询参数
	placeholders := make([]string, len(passageIDs))
	args := make([]interface{}, len(passageIDs))
	for i, id := range passageIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`SELECT passage_id, tag_id FROM passage_tags WHERE passage_id IN (%s) ORDER BY passage_id, tag_id`,
		strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("批量查询标签ID失败: %w", err)
	}
	defer rows.Close()

	result := make(map[int][]int)
	for rows.Next() {
		var passageID, tagID int
		if err := rows.Scan(&passageID, &tagID); err != nil {
			return nil, fmt.Errorf("扫描标签ID失败: %w", err)
		}
		result[passageID] = append(result[passageID], tagID)
	}

	return result, nil
}

// BatchCreate 批量创建文章-标签关联（使用事务）
func (r *passageTagRepository) BatchCreate(passageID int, tagIDs []int) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// 开始事务
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 批量插入关联
	query := `INSERT INTO passage_tags (passage_id, tag_id, created_at) VALUES (?, ?, ?)`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("准备语句失败: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, tagID := range tagIDs {
		_, err = stmt.Exec(passageID, tagID, now)
		if err != nil {
			return fmt.Errorf("插入关联失败 (passageID=%d, tagID=%d): %w", passageID, tagID, err)
		}
	}

	// 批量更新标签使用次数
	for _, tagID := range tagIDs {
		_, err = tx.Exec(`UPDATE tags SET usage_count = usage_count + 1, updated_at = ? WHERE id = ?`, now, tagID)
		if err != nil {
			return fmt.Errorf("更新标签使用次数失败 (tagID=%d): %w", tagID, err)
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// ReplaceAll 替换文章的所有标签关联（使用事务）
func (r *passageTagRepository) ReplaceAll(passageID int, tagIDs []int) error {
	// 开始事务
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// 获取现有的标签ID
	var oldTagIDs []int
	rows, err := tx.Query(`SELECT tag_id FROM passage_tags WHERE passage_id = ?`, passageID)
	if err != nil {
		return fmt.Errorf("查询现有标签失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tagID int
		if err := rows.Scan(&tagID); err != nil {
			return fmt.Errorf("扫描标签ID失败: %w", err)
		}
		oldTagIDs = append(oldTagIDs, tagID)
	}

	// 计算差异
	oldTagMap := make(map[int]bool)
	for _, tagID := range oldTagIDs {
		oldTagMap[tagID] = true
	}

	newTagMap := make(map[int]bool)
	for _, tagID := range tagIDs {
		newTagMap[tagID] = true
	}

	// 找出需要删除的标签
	var toDelete []int
	for _, tagID := range oldTagIDs {
		if !newTagMap[tagID] {
			toDelete = append(toDelete, tagID)
		}
	}

	// 找出需要添加的标签
	var toAdd []int
	for _, tagID := range tagIDs {
		if !oldTagMap[tagID] {
			toAdd = append(toAdd, tagID)
		}
	}

	now := time.Now()

	// 删除不需要的关联
	if len(toDelete) > 0 {
		placeholders := make([]string, len(toDelete))
		args := make([]interface{}, len(toDelete)+1)
		args[0] = passageID
		for i, tagID := range toDelete {
			placeholders[i] = "?"
			args[i+1] = tagID
		}

		deleteQuery := fmt.Sprintf(`DELETE FROM passage_tags WHERE passage_id = ? AND tag_id IN (%s)`, 
			fmt.Sprintf("%s", strings.Join(placeholders, ",")))
		_, err = tx.Exec(deleteQuery, args...)
		if err != nil {
			return fmt.Errorf("删除关联失败: %w", err)
		}

		// 减少标签使用次数
		for _, tagID := range toDelete {
			_, err = tx.Exec(`UPDATE tags SET usage_count = CASE WHEN usage_count > 0 THEN usage_count - 1 ELSE 0 END, updated_at = ? WHERE id = ?`, now, tagID)
			if err != nil {
				return fmt.Errorf("减少标签使用次数失败 (tagID=%d): %w", tagID, err)
			}
		}
	}

	// 添加新的关联
	if len(toAdd) > 0 {
		query := `INSERT INTO passage_tags (passage_id, tag_id, created_at) VALUES (?, ?, ?)`
		stmt, err := tx.Prepare(query)
		if err != nil {
			return fmt.Errorf("准备语句失败: %w", err)
		}
		defer stmt.Close()

		for _, tagID := range toAdd {
			_, err = stmt.Exec(passageID, tagID, now)
			if err != nil {
				return fmt.Errorf("插入关联失败 (passageID=%d, tagID=%d): %w", passageID, tagID, err)
			}
		}

		// 增加标签使用次数
		for _, tagID := range toAdd {
			_, err = tx.Exec(`UPDATE tags SET usage_count = usage_count + 1, updated_at = ? WHERE id = ?`, now, tagID)
			if err != nil {
				return fmt.Errorf("增加标签使用次数失败 (tagID=%d): %w", tagID, err)
			}
		}
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}