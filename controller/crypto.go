package controller

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"myblog-gogogo/crypto"
)

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*crypto.ECCManager
	mu       sync.RWMutex
}

// 全局会话管理器实例
var sessionManager = &SessionManager{
	sessions: make(map[string]*crypto.ECCManager),
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "session_" + hex.EncodeToString(bytes)
}

// GetPublicKey 获取ECC公钥
func GetPublicKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 生成或获取会话ID
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = generateSessionID()
	}

	// 检查会话是否已存在且未过期
	sessionManager.mu.RLock()
	ecc, exists := sessionManager.sessions[sessionID]
	sessionManager.mu.RUnlock()

	if !exists || ecc.IsExpired() {
		// 创建新的ECC管理器
		newECC, err := crypto.NewECCManager(sessionID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "failed to generate ECC keys",
			})
			return
		}

		// 存储会话
		sessionManager.mu.Lock()
		sessionManager.sessions[sessionID] = newECC
		sessionManager.mu.Unlock()

		ecc = newECC
	}

	// 获取公钥（JWK格式）
	publicKeyJWK := ecc.GetPublicKeyJWK()

	// 返回公钥信息
	response := map[string]interface{}{
		"success":     true,
		"session_id":  sessionID,
		"public_key":  publicKeyJWK,
		"key_format":  "jwk",
		"algorithm":   "ECDH-ES",
		"curve":       "P-256",
		"expires_at":  ecc.GetExpiry().Unix(),
		"expires_in":  int(time.Until(ecc.GetExpiry()).Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DecryptData 解密数据
func DecryptData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		SessionID      string `json:"session_id"`
		ClientPubKey   string `json:"client_public_key"`
		EncryptedData  string `json:"encrypted_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "invalid request format",
		})
		return
	}

	// 验证请求参数
	if req.SessionID == "" || req.ClientPubKey == "" || req.EncryptedData == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "missing required fields",
		})
		return
	}

	// 获取会话
	sessionManager.mu.RLock()
	ecc, exists := sessionManager.sessions[req.SessionID]
	sessionManager.mu.RUnlock()

	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "session not found",
		})
		return
	}

	// 检查会话是否过期
	if ecc.IsExpired() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGone)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "session expired",
		})
		return
	}

	// 解密数据
	decrypted, err := ecc.HybridDecrypt(req.EncryptedData, req.ClientPubKey)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "decryption failed: " + err.Error(),
		})
		return
	}

	// 返回解密后的数据
	response := map[string]interface{}{
		"success":  true,
		"decrypted": string(decrypted),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CleanupExpiredSessions 清理过期会话（定期调用）
func CleanupExpiredSessions() {
	sessionManager.mu.Lock()
	defer sessionManager.mu.Unlock()

	for sessionID, ecc := range sessionManager.sessions {
		if ecc.IsExpired() {
			delete(sessionManager.sessions, sessionID)
		}
	}
}

// GetSessionCount 获取当前活跃会话数
func GetSessionCount() int {
	sessionManager.mu.RLock()
	defer sessionManager.mu.RUnlock()
	return len(sessionManager.sessions)
}