package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2 参数配置
const (
	iterations = 1
	memory     = 64 * 1024 // 64 MB
	threads    = 4
	keyLength  = 32
	saltLength = 16
)

// HashPassword 使用 Argon2id 哈希密码
func HashPassword(password string) (string, error) {
	// 生成随机盐
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// 使用 Argon2id 哈希密码
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, threads, keyLength)

	// 编码为 base64 格式: $argon2id$v=19$m=65536,t=1,p=4$salt$hash
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, iterations, threads, b64Salt, b64Hash)

	return encoded, nil
}

// VerifyPassword 验证密码是否匹配哈希值
func VerifyPassword(password, encodedHash string) (bool, error) {
	// 解析哈希字符串
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// 使用相同的参数计算密码的哈希
	otherHash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.threads, params.keyLength)

	// 使用恒定时间比较,防止时序攻击
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

// argon2Params 存储 Argon2 参数
type argon2Params struct {
	iterations uint32
	memory     uint32
	threads    uint8
	keyLength  uint32
}

// decodeHash 解码 Argon2 哈希字符串
func decodeHash(encodedHash string) (*argon2Params, []byte, []byte, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, fmt.Errorf("invalid hash format")
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid version: %w", err)
	}
	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible version: %d (expected %d)", version, argon2.Version)
	}

	params := &argon2Params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &params.memory, &params.iterations, &params.threads)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid salt: %w", err)
	}

	hash, err := base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid hash: %w", err)
	}

	params.keyLength = uint32(len(hash))

	return params, salt, hash, nil
}