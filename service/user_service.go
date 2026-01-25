package service

import (
	"fmt"
	"regexp"

	apperrors "myblog-gogogo/pkg/errors"
	"myblog-gogogo/auth"
	"myblog-gogogo/db"
	"myblog-gogogo/db/models"
	"myblog-gogogo/db/repositories"
	"myblog-gogogo/pkg/dto"
)

// UserService 用户服务
type UserService struct {
	userRepo repositories.UserRepository
	authSvc  *AuthService
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		userRepo: db.GetUserRepository(),
		authSvc:  NewAuthService(),
	}
}

// Register 用户注册
func (s *UserService) Register(req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	// 验证用户名
	if err := s.validateUsername(req.Username); err != nil {
		return nil, err
	}

	// 验证邮箱
	if err := s.validateEmail(req.Email); err != nil {
		return nil, err
	}

	// 获取密码（支持明文和加密两种方式）
	var password string
	var err error

	if req.EncryptedPassword != "" && req.SessionID != "" && req.ClientPublicKey != "" {
		// 使用ECC加密方式，需要解密
		if s.authSvc == nil {
			return nil, fmt.Errorf("auth service not initialized")
		}
		password, err = s.authSvc.decryptPassword(req.EncryptedPassword, req.SessionID, req.ClientPublicKey)
		if err != nil {
			return nil, err
		}
	} else if req.Password != "" {
		// 使用明文密码（不推荐，仅作为降级方案）
		password = req.Password
	} else {
		return nil, dto.ErrPasswordRequired
	}

	// 验证密码
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}
	if existingUser != nil {
		return nil, dto.ErrUsernameExists
	}

	// 检查邮箱是否已存在
	existingUser, err = s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}
	if existingUser != nil {
		return nil, dto.ErrEmailExists
	}

	// 哈希密码
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码哈希失败: %w", err)
	}

	// 创建用户
	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
		Status:   "active",
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	// 生成token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("生成token失败: %w", err)
	}

	return &dto.RegisterResponse{
		User: &dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Token: token,
	}, nil
}

// validateUsername 验证用户名
func (s *UserService) validateUsername(username string) error {
	if username == "" {
		return dto.ErrUsernameRequired
	}

	if len(username) < 3 || len(username) > 20 {
		return dto.ErrUsernameInvalid
	}

	matched, err := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	if err != nil {
		return apperrors.Wrap(err, "VALIDATION_ERROR", "用户名验证失败")
	}
	if !matched {
		return dto.ErrUsernameInvalid
	}

	return nil
}

// validateEmail 验证邮箱
func (s *UserService) validateEmail(email string) error {
	if email == "" {
		return dto.ErrEmailRequired
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return dto.ErrEmailInvalid
	}

	return nil
}

// validatePassword 验证密码
func (s *UserService) validatePassword(password string) error {
	if password == "" {
		return dto.ErrPasswordRequired
	}

	if len(password) < 6 {
		return dto.ErrPasswordTooShort
	}

	return nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id int) (*dto.UserDTO, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}

	if user == nil {
		return nil, dto.ErrUserNotFound
	}

	return s.toDTO(user), nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*dto.UserDTO, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}

	if user == nil {
		return nil, dto.ErrUserNotFound
	}

	return s.toDTO(user), nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id int, req *dto.UpdateUserRequest) (*dto.UserDTO, error) {
	// 获取现有用户
	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("数据库查询失败: %w", err)
	}

	if existingUser == nil {
		return nil, dto.ErrUserNotFound
	}

	// 如果修改用户名，检查是否重复
	if req.Username != "" && req.Username != existingUser.Username {
		if err := s.validateUsername(req.Username); err != nil {
			return nil, err
		}

		userWithSameName, err := s.userRepo.GetByUsername(req.Username)
		if err == nil && userWithSameName != nil && userWithSameName.ID != id {
			return nil, dto.ErrUsernameExists
		}

		existingUser.Username = req.Username
	}

	// 如果修改邮箱，检查是否重复
	if req.Email != "" && req.Email != existingUser.Email {
		if err := s.validateEmail(req.Email); err != nil {
			return nil, err
		}

		// 检查邮箱是否已被其他用户使用
		userWithSameEmail, err := s.userRepo.GetByEmail(req.Email)
		if err == nil && userWithSameEmail != nil && userWithSameEmail.ID != id {
			return nil, dto.ErrEmailExists
		}

		existingUser.Email = req.Email
	}

	// 更新其他字段
	if req.Role != "" {
		existingUser.Role = req.Role
	}

	if req.Status != "" {
		existingUser.Status = req.Status
	}

	// 如果提供了新密码，则更新密码
	if req.Password != "" {
		if err := s.validatePassword(req.Password); err != nil {
			return nil, err
		}

		hashedPassword, err := auth.HashPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("密码哈希失败: %w", err)
		}

		existingUser.Password = hashedPassword
	}

	// 保存更新
	if err := s.userRepo.Update(existingUser); err != nil {
		return nil, fmt.Errorf("更新用户失败: %w", err)
	}

	return s.toDTO(existingUser), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id int) error {
	// 获取用户
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("数据库查询失败: %w", err)
	}

	if user == nil {
		return dto.ErrUserNotFound
	}

	// 防止删除管理员账户
	if user.Role == "admin" {
		return dto.ErrCannotDeleteAdmin
	}

	// 删除用户
	if err := s.userRepo.Delete(id); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	return nil
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(req *dto.UserListRequest) (*dto.PaginationResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 计算偏移量
	offset := (req.Page - 1) * req.PageSize

	// 获取用户列表
	users, err := s.userRepo.GetAll(req.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	// 获取总数
	total, err := s.userRepo.Count()
	if err != nil {
		return nil, fmt.Errorf("获取用户总数失败: %w", err)
	}

	// 转换为 DTO
	userDTOs := make([]*dto.UserDTO, 0, len(users))
	for _, user := range users {
		userDTOs = append(userDTOs, &dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	// 计算总页数
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize != 0 {
		totalPages++
	}

	// 构建分页响应
	return &dto.PaginationResponse{
		Total:       int64(total),
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
		HasNext:     req.Page < totalPages,
		HasPrevious: req.Page > 1,
		Data:        userDTOs,
	}, nil
}

// toDTO 转换为DTO
func (s *UserService) toDTO(user *models.User) *dto.UserDTO {
	return &dto.UserDTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// UsernameExists 检查用户名是否存在
func (s *UserService) UsernameExists(username string) (bool, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

// EmailExists 检查邮箱是否存在
func (s *UserService) EmailExists(email string) (bool, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return false, fmt.Errorf("检查邮箱是否存在失败: %w", err)
	}
	return user != nil, nil
}