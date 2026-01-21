package service

import (
	"fmt"
	"time"

	apperrors "myblog-gogogo/pkg/errors"
	"myblog-gogogo/auth"
	"myblog-gogogo/crypto"
	"myblog-gogogo/db"
	"myblog-gogogo/db/repositories"
	"myblog-gogogo/pkg/dto"
)

// AuthService 认证服务
type AuthService struct {
	userRepo      repositories.UserRepository
	sessionManager *SessionManager
}

// NewAuthService 创建认证服务
func NewAuthService() *AuthService {
	return &AuthService{
		userRepo:      db.GetUserRepository(),
		sessionManager: GetSessionManager(),
	}
}

// Login 用户登录
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	// 验证用户名
	if req.Username == "" {
		return nil, apperrors.ErrUsernameRequired
	}

	// 获取密码（支持明文和加密两种方式）
	var password string
	var err error

	if req.EncryptedPassword != "" && req.SessionID != "" && req.ClientPublicKey != "" {
		// 使用ECC加密方式，需要解密
		password, err = s.decryptPassword(req.EncryptedPassword, req.SessionID, req.ClientPublicKey)
		if err != nil {
			return nil, err
		}
	} else if req.Password != "" {
		// 使用明文密码（不推荐，仅作为降级方案）
		password = req.Password
	} else {
		return nil, apperrors.ErrPasswordRequired
	}

	// 从数据库获取用户
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, apperrors.Wrap(err, "DB_ERROR", "数据库查询失败")
	}

	if user == nil {
		return nil, apperrors.ErrPasswordIncorrect
	}

	// 验证密码
	if !s.verifyPassword(password, user.Password) {
		return nil, apperrors.ErrPasswordIncorrect
	}

	// 检查用户状态
	if user.Status != "active" {
		return nil, apperrors.ErrUserInactive
	}

	// 生成JWT token
	token, err := auth.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, apperrors.Wrap(err, "TOKEN_ERROR", "生成token失败")
	}

	// 构建响应
	return &dto.LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		User: &dto.UserDTO{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}

// decryptPassword 解密密码
func (s *AuthService) decryptPassword(encryptedPassword, sessionID, clientPublicKey string) (string, error) {
	s.sessionManager.mu.RLock()
	ecc, exists := s.sessionManager.Sessions[sessionID]
	s.sessionManager.mu.RUnlock()

	if !exists {
		return "", apperrors.ErrSessionNotFound
	}

	// 检查会话是否过期
	if ecc.IsExpired() {
		return "", apperrors.ErrSessionExpired
	}

	// 解密密码
	decrypted, err := ecc.HybridDecrypt(encryptedPassword, clientPublicKey)
	if err != nil {
		fmt.Printf("密码解密失败: %v\n", err)
		return "", apperrors.WrapWithDetails(err, "DECRYPTION_FAILED", "密码解密失败", err.Error())
	}

	return string(decrypted), nil
}

// verifyPassword 验证密码（使用 Argon2id）
func (s *AuthService) verifyPassword(password, hashedPassword string) bool {
	valid, err := auth.VerifyPassword(password, hashedPassword)
	if err != nil {
		fmt.Printf("密码验证失败: %v\n", err)
		return false
	}
	return valid
}

// ValidateToken 验证token
func (s *AuthService) ValidateToken(token string) (*auth.Claims, error) {
	claims, err := auth.ValidateToken(token)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}
	return claims, nil
}

// GenerateECCSession 生成ECC加密会话
func (s *AuthService) GenerateECCSession(sessionID string) (*crypto.ECCManager, error) {
	ecc, err := crypto.NewECCManager(sessionID)
	if err != nil {
		return nil, apperrors.Wrap(err, "ECC_ERROR", "创建ECC会话失败")
	}

	s.sessionManager.mu.Lock()
	s.sessionManager.Sessions[sessionID] = ecc
	s.sessionManager.mu.Unlock()

	return ecc, nil
}

// GetECCSession 获取ECC加密会话
func (s *AuthService) GetECCSession(sessionID string) (*crypto.ECCManager, error) {
	s.sessionManager.mu.RLock()
	ecc, exists := s.sessionManager.Sessions[sessionID]
	s.sessionManager.mu.RUnlock()

	if !exists {
		return nil, apperrors.ErrSessionNotFound
	}

	if ecc.IsExpired() {
		return nil, apperrors.ErrSessionExpired
	}

	return ecc, nil
}

// CheckPermission 检查权限
func (s *AuthService) CheckPermission(userRole, requiredRole string) bool {
	roleHierarchy := map[string]int{
		"admin": 3,
		"editor": 2,
		"user": 1,
	}

	userLevel, ok := roleHierarchy[userRole]
	if !ok {
		return false
	}

	requiredLevel, ok := roleHierarchy[requiredRole]
	if !ok {
		return false
	}

	return userLevel >= requiredLevel
}

// IsAdmin 检查是否为管理员
func (s *AuthService) IsAdmin(userRole string) bool {
	return userRole == "admin"
}

// IsOwnerOrAdmin 检查是否为所有者或管理员
func (s *AuthService) IsOwnerOrAdmin(userID, resourceOwnerID int, userRole string) bool {
	return userID == resourceOwnerID || userRole == "admin"
}