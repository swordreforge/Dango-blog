package service

import (
	"time"

	apperrors "myblog-gogogo/pkg/errors"
	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
	"myblog-gogogo/db/repositories"
	"myblog-gogogo/pkg/dto"
)

// PassageService 文章服务
type PassageService struct {
	passageRepo repositories.PassageRepository
	authSvc     *AuthService
}

// NewPassageService 创建文章服务
func NewPassageService() *PassageService {
	return &PassageService{
		passageRepo: db.GetPassageRepository(),
		authSvc:     NewAuthService(),
	}
}

// CheckAccess 检查文章访问权限
func (s *PassageService) CheckAccess(req *dto.PassageAccessRequest) (*dto.PassageAccessResponse, error) {
	// 获取文章
	passage, err := s.passageRepo.GetByID(req.PassageID)
	if err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "数据库查询失败")
	}

	if passage == nil {
		return nil, apperrors.ErrPassageNotFound
	}

	// 检查文章状态
	if passage.Status != "published" {
		if req.UserRole != "admin" {
			var publishedAt *time.Time
			if !passage.PublishedAt.IsZero() {
				publishedAt = &passage.PublishedAt
			}
			return &dto.PassageAccessResponse{
				Allowed:     false,
				Reason:      "文章尚未发布",
				Status:      passage.Status,
				IsScheduled: passage.IsScheduled,
				PublishedAt: publishedAt,
			}, nil
		}
	}

	// 检查可见性
	if passage.Visibility == "private" {
		if req.UserRole != "admin" {
			return &dto.PassageAccessResponse{
				Allowed:    false,
				Reason:     "此文章为私密文章，仅管理员可见",
				Visibility: passage.Visibility,
			}, nil
		}
	}

	// 允许访问
	return &dto.PassageAccessResponse{
		Allowed: true,
		Passage: s.toDTO(passage),
	}, nil
}

// GetPassageByID 根据ID获取文章（带权限检查）
func (s *PassageService) GetPassageByID(id int, userRole string) (*dto.PassageDTO, error) {
	// 检查访问权限
	accessResp, err := s.CheckAccess(&dto.PassageAccessRequest{
		PassageID: id,
		UserRole:  userRole,
	})
	if err != nil {
		return nil, err
	}

	if !accessResp.Allowed {
		return nil, apperrors.New("ACCESS_DENIED", accessResp.Reason)
	}

	return accessResp.Passage, nil
}

// CreatePassage 创建文章
func (s *PassageService) CreatePassage(req *dto.CreatePassageRequest) (*dto.PassageDTO, error) {
	// 转换Markdown为HTML
	htmlContent, err := ConvertToHTMLWithOption([]byte(req.Content), req.ShowTitle)
	if err != nil {
		return nil, apperrors.Wrap(err, "MARKDOWN_ERROR", "Markdown转换失败")
	}

	// 处理 PublishedAt
	var publishedAt time.Time
	if req.PublishedAt != nil {
		publishedAt = *req.PublishedAt
	}

	// 创建文章
	passage := &models.Passage{
		Title:           req.Title,
		Content:         htmlContent,
		OriginalContent: req.Content,
		Status:          req.Status,
		Visibility:      req.Visibility,
		ShowTitle:       req.ShowTitle,
		IsScheduled:     req.IsScheduled,
		PublishedAt:     publishedAt,
	}

	// 设置默认值
	if passage.Status == "" {
		passage.Status = "draft"
	}

	if passage.Visibility == "" {
		passage.Visibility = "public"
	}

	// 保存文章
	if err := s.passageRepo.Create(passage); err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "创建文章失败")
	}

	// 重新获取完整数据
	passage, _ = s.passageRepo.GetByID(passage.ID)

	return s.toDTO(passage), nil
}

// UpdatePassage 更新文章
func (s *PassageService) UpdatePassage(id int, req *dto.UpdatePassageRequest) (*dto.PassageDTO, error) {
	// 获取现有文章
	existingPassage, err := s.passageRepo.GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "数据库查询失败")
	}

	if existingPassage == nil {
		return nil, apperrors.ErrPassageNotFound
	}

	// 更新字段
	if req.Title != nil {
		existingPassage.Title = *req.Title
	}

	if req.Content != nil {
		// 转换Markdown为HTML
		htmlContent, err := ConvertToHTMLWithOption([]byte(*req.Content), existingPassage.ShowTitle)
		if err != nil {
			return nil, apperrors.Wrap(err, "MARKDOWN_ERROR", "Markdown转换失败")
		}

		existingPassage.Content = htmlContent
		existingPassage.OriginalContent = *req.Content
	}

	if req.Status != nil {
		existingPassage.Status = *req.Status
	}

	if req.Visibility != nil {
		existingPassage.Visibility = *req.Visibility
	}

	if req.ShowTitle != nil {
		existingPassage.ShowTitle = *req.ShowTitle
	}

	if req.IsScheduled != nil {
		existingPassage.IsScheduled = *req.IsScheduled
	}

	if req.PublishedAt != nil {
		existingPassage.PublishedAt = *req.PublishedAt
	}

	// 保存更新
	if err := s.passageRepo.Update(existingPassage); err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "更新文章失败")
	}

	// 重新获取完整数据
	passage, err := s.passageRepo.GetByID(id)
	if err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "获取更新后的文章失败")
	}

	return s.toDTO(passage), nil
}

// PartialUpdatePassage 增量更新文章
func (s *PassageService) PartialUpdatePassage(req *dto.PassageUpdateRequest) error {
	// 获取现有文章
	passage, err := s.passageRepo.GetByID(req.PassageID)
	if err != nil {
		return apperrors.Wrap(err, "DB_ERROR", "获取文章失败")
	}

	if passage == nil {
		return apperrors.ErrPassageNotFound
	}

	// 允许更新的字段
	allowedFields := map[string]bool{
		"visibility":   true,
		"is_scheduled": true,
		"published_at": true,
		"status":       true,
	}

	// 更新字段
	for field, value := range req.Updates {
		if !allowedFields[field] {
			continue
		}

		switch field {
		case "visibility":
			if v, ok := value.(string); ok {
				passage.Visibility = v
			}
		case "is_scheduled":
			if v, ok := value.(bool); ok {
				passage.IsScheduled = v
			}
		case "published_at":
			if v, ok := value.(time.Time); ok {
				passage.PublishedAt = v
			}
		case "status":
			if v, ok := value.(string); ok {
				passage.Status = v
			}
		}
	}

	// 保存更新
	if err := s.passageRepo.Update(passage); err != nil {
		return apperrors.Wrap(err, "DB_ERROR", "更新文章失败")
	}

	return nil
}

// DeletePassage 删除文章
func (s *PassageService) DeletePassage(id int) error {
	// 获取文章
	passage, err := s.passageRepo.GetByID(id)
	if err != nil {
		return apperrors.Wrap(err, "DB_ERROR", "数据库查询失败")
	}

	if passage == nil {
		return apperrors.ErrPassageNotFound
	}

	// 删除文章
	if err := s.passageRepo.Delete(id); err != nil {
		return apperrors.Wrap(err, "DB_ERROR", "删除文章失败")
	}

	return nil
}

// ListPassages 获取文章列表
func (s *PassageService) ListPassages(req *dto.PassageListRequest) (*dto.PaginationResponse, error) {
	// 计算分页参数
	page := req.Page
	if page <= 0 {
		page = 1
	}

	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// 获取文章列表（使用基本的 GetAll 方法）
	passages, err := s.passageRepo.GetAll(pageSize, offset)
	if err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "数据库查询失败")
	}

	// 获取总数
	total, err := s.passageRepo.Count()
	if err != nil {
		total = len(passages)
	}

	// 转换为DTO
	passageDTOs := make([]*dto.PassageDTO, len(passages))
	for i, passage := range passages {
		passageDTOs[i] = s.toDTO(&passage)
	}

	// 计算总页数
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}

	return &dto.PaginationResponse{
		Total:       int64(total),
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
		Data:        passageDTOs,
	}, nil
}

// toDTO 转换为DTO
func (s *PassageService) toDTO(passage *models.Passage) *dto.PassageDTO {
	var publishedAt *time.Time
	if !passage.PublishedAt.IsZero() {
		publishedAt = &passage.PublishedAt
	}

	dto := &dto.PassageDTO{
		ID:             passage.ID,
		Title:          passage.Title,
		Content:        passage.Content,
		OriginalContent: passage.OriginalContent,
		Status:         passage.Status,
		Visibility:     passage.Visibility,
		ShowTitle:      passage.ShowTitle,
		IsScheduled:    passage.IsScheduled,
		PublishedAt:    publishedAt,
		CreatedAt:      passage.CreatedAt,
		UpdatedAt:      passage.UpdatedAt,
	}

	return dto
}