package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// JWT secret key - 每次启动应用时随机生成 32 位 key
	jwtSecret = generateRandomKey()
	// Token过期时间
	tokenExpiration = 24 * time.Hour
)

// generateRandomKey 生成随机的 32 位 JWT key
func generateRandomKey() []byte {
	bytes := make([]byte, 16) // 16 字节 = 32 个十六进制字符
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败,使用硬编码的默认 key
		return []byte("your-secret-key-change-this-in-production")
	}
	return []byte(hex.EncodeToString(bytes))
}

// Claims 自定义JWT claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT token
func GenerateToken(userID int, username, role string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "myblog-gogogo",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken 验证JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 检查token是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// RefreshToken 刷新token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 生成新的token
	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}

// SetSecret 设置JWT密钥（用于测试或从环境变量加载）
func SetSecret(secret string) {
	jwtSecret = []byte(secret)
}

// SetTokenExpiration 设置token过期时间
func SetTokenExpiration(expiration time.Duration) {
	tokenExpiration = expiration
}

// GetTokenFromRequest 从请求中获取并验证 token
func GetTokenFromRequest(r *http.Request) (*Claims, error) {
	var tokenString string

	// 首先尝试从 Authorization header 获取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// 支持两种格式: "Bearer <token>" 或直接 "<token>"
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			// 如果没有Bearer前缀，直接使用整个header值
			tokenString = authHeader
		}
	}

	// 如果 header 中没有 token,尝试从 cookie 获取
	if tokenString == "" {
		cookie, err := r.Cookie("auth_token")
		if err == nil {
			tokenString = cookie.Value
		}
	}

	// 如果还是没有 token,返回错误
	if tokenString == "" {
		return nil, errors.New("missing authorization token")
	}

	// 验证 token
	return ValidateToken(tokenString)
}

// IsAdmin 检查用户是否是管理员
func IsAdmin(r *http.Request) bool {
	// 从请求头获取 token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// 解析 Bearer token
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return false
	}

	tokenString := authHeader[7:]

	// 验证 token
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return false
	}

	// 检查角色是否为 admin
	return claims.Role == "admin"
}