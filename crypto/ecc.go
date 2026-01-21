package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"time"

	"golang.org/x/crypto/hkdf"
)

// ECCManager ECC加密管理器
type ECCManager struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
	keyExpiry  time.Time
	sessionID  string
}

// JWK JWK格式的公钥
type JWK struct {
	Kty string `json:"kty"` // 密钥类型
	Crv string `json:"crv"` // 曲线名称
	X   string `json:"x"`   // X坐标
	Y   string `json:"y"`   // Y坐标
	Use string `json:"use"` // 用途
	Alg string `json:"alg"` // 算法
}

// EncryptedData 加密数据结构
type EncryptedData struct {
	SessionID      string `json:"session_id"`
	EncryptedData  string `json:"encrypted_data"`
	ClientPubKey   string `json:"client_public_key"`
	Algorithm      string `json:"algorithm"`
}

// NewECCManager 创建新的ECC管理器
func NewECCManager(sessionID string) (*ECCManager, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ECC key pair: %w", err)
	}

	return &ECCManager{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
		keyExpiry:  time.Now().Add(1 * time.Hour), // 密钥1小时过期
		sessionID:  sessionID,
	}, nil
}

// GetPublicKeyJWK 获取JWK格式的公钥
func (m *ECCManager) GetPublicKeyJWK() JWK {
	pub := m.publicKey

	return JWK{
		Kty: "EC",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(pub.X.Bytes()),
		Y:   base64.RawURLEncoding.EncodeToString(pub.Y.Bytes()),
		Use: "enc",
		Alg: "ECDH-ES+A256KW",
	}
}

// GetPublicKeyRaw 获取Raw格式的公钥（用于Web Crypto API）
func (m *ECCManager) GetPublicKeyRaw() (map[string]interface{}, error) {
	pub := m.publicKey

	// 生成未压缩格式公钥（65字节：04 + X + Y）
	publicKeyRaw := make([]byte, 65)
	publicKeyRaw[0] = 0x04 // 未压缩格式标识
	pub.X.FillBytes(publicKeyRaw[1:33])
	pub.Y.FillBytes(publicKeyRaw[33:])

	return map[string]interface{}{
		"format":      "raw",
		"keyData":     base64.StdEncoding.EncodeToString(publicKeyRaw),
		"algorithm":   map[string]string{"name": "ECDH", "namedCurve": "P-256"},
		"extractable": true,
		"usages":      []string{"deriveKey", "deriveBits"},
	}, nil
}

// GetPublicKeyPEM 获取PEM格式的公钥
func (m *ECCManager) GetPublicKeyPEM() (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(m.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPEM), nil
}

// ParsePublicKeyPEM 解析PEM格式的公钥
func ParsePublicKeyPEM(pemData string) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("not an ECDSA public key")
	}

	return ecdsaPub, nil
}

// ParsePublicKeyJWK 解析JWK格式的公钥
func ParsePublicKeyJWK(jwkData string) (*ecdsa.PublicKey, error) {
	var jwk JWK
	if err := json.Unmarshal([]byte(jwkData), &jwk); err != nil {
		return nil, fmt.Errorf("failed to parse JWK: %w", err)
	}

	if jwk.Kty != "EC" || jwk.Crv != "P-256" {
		return nil, errors.New("unsupported key type or curve")
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X coordinate: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Y coordinate: %w", err)
	}

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	return pubKey, nil
}

// DeriveSharedSecret 派生共享密钥
func (m *ECCManager) DeriveSharedSecret(clientPublicKey *ecdsa.PublicKey) ([]byte, error) {
	// 验证曲线是否匹配
	if m.privateKey.Curve != clientPublicKey.Curve {
		return nil, errors.New("curve mismatch")
	}

	// 计算共享密钥
	sharedX, _ := m.privateKey.Curve.ScalarMult(
		clientPublicKey.X,
		clientPublicKey.Y,
		m.privateKey.D.Bytes(),
	)

	if sharedX == nil {
		return nil, errors.New("failed to compute shared secret")
	}

	// 使用HKDF提取和扩展
	return m.deriveHKDF(sharedX.Bytes()), nil
}

// deriveHKDF 使用HKDF派生密钥
func (m *ECCManager) deriveHKDF(sharedSecret []byte) []byte {
	// 使用SHA-256的HKDF
	salt := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		// 如果随机数生成失败，使用固定salt（不推荐生产环境）
		salt = []byte("*\x078\x099gbcmgz7+````''")
	}

	hkdf := hkdf.New(sha256.New, sharedSecret, salt, nil)
	derivedKey := make([]byte, 32) // 256位AES密钥
	if _, err := io.ReadFull(hkdf, derivedKey); err != nil {
		// 如果HKDF失败，直接使用共享密钥（不推荐生产环境）
		return sharedSecret[:32]
	}

	return derivedKey
}

// HybridEncrypt 混合加密（ECDH + AES-GCM）
func (m *ECCManager) HybridEncrypt(plaintext []byte, clientPubKeyPEM string) (string, error) {
	// 1. 解析客户端公钥
	clientPubKey, err := ParsePublicKeyPEM(clientPubKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to parse client public key: %w", err)
	}

	// 2. 派生共享密钥
	sharedKey, err := m.DeriveSharedSecret(clientPubKey)
	if err != nil {
		return "", fmt.Errorf("failed to derive shared key: %w", err)
	}

	// 3. 使用AES-GCM加密数据
	block, err := aes.NewCipher(sharedKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 4. 返回base64编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// HybridDecrypt 混合解密（ECDH + AES-GCM）
func (m *ECCManager) HybridDecrypt(encryptedData string, clientPubKeyPEM string) ([]byte, error) {
	// 1. 解析客户端公钥
	clientPubKey, err := ParsePublicKeyPEM(clientPubKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client public key: %w", err)
	}

	// 2. 计算共享密钥（不使用HKDF，直接使用原始共享密钥）
	if m.privateKey.Curve != clientPubKey.Curve {
		return nil, errors.New("curve mismatch")
	}

	sharedX, _ := m.privateKey.Curve.ScalarMult(
		clientPubKey.X,
		clientPubKey.Y,
		m.privateKey.D.Bytes(),
	)

	if sharedX == nil {
		return nil, errors.New("failed to compute shared secret")
	}

	// 直接使用共享密钥的前32字节作为AES密钥（与Web Crypto API一致）
	sharedSecret := sharedX.Bytes()
	if len(sharedSecret) > 32 {
		sharedSecret = sharedSecret[:32]
	}
	if len(sharedSecret) < 32 {
		// 如果不足32字节，填充到32字节
		padded := make([]byte, 32)
		copy(padded, sharedSecret)
		sharedSecret = padded
	}

	// 3. 解码base64数据
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	// 4. 使用AES-GCM解密数据
	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}

// IsExpired 检查密钥是否过期
func (m *ECCManager) IsExpired() bool {
	return time.Now().After(m.keyExpiry)
}

// GetSessionID 获取会话ID
func (m *ECCManager) GetSessionID() string {
	return m.sessionID
}

// GetExpiry 获取过期时间
func (m *ECCManager) GetExpiry() time.Time {
	return m.keyExpiry
}
