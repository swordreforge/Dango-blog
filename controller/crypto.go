package controller

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"myblog-gogogo/crypto"
	"myblog-gogogo/service"
)

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

	// 获取全局会话管理器
	sessionManager := service.GetSessionManager()

	// 生成或获取会话ID
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = generateSessionID()
	}

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

	// 存储会话（使用全局会话管理器）
	sessionManager.Lock()
	sessionManager.Sessions[sessionID] = newECC
	sessionManager.Unlock()

	// 获取公钥（JWK格式）
	publicKeyJWK := newECC.GetPublicKeyJWK()

	// 返回公钥信息
	response := map[string]interface{}{
		"success":     true,
		"session_id":  sessionID,
		"public_key":  publicKeyJWK,
		"key_format":  "jwk",
		"algorithm":   "ECDH-ES",
		"curve":       "P-256",
		"expires_at":  newECC.GetExpiry().Unix(),
		"expires_in":  int(time.Until(newECC.GetExpiry()).Seconds()),
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

	// 获取全局会话管理器
	sessionManager := service.GetSessionManager()

	// 获取会话
	sessionManager.RLock()
	ecc, exists := sessionManager.Sessions[req.SessionID]
	sessionManager.RUnlock()

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
	sessionManager := service.GetSessionManager()
	sessionManager.Lock()
	defer sessionManager.Unlock()

	for sessionID, ecc := range sessionManager.Sessions {
		if ecc.IsExpired() {
			delete(sessionManager.Sessions, sessionID)
		}
	}
}

// GetSessionCount 获取当前活跃会话数
func GetSessionCount() int {
	sessionManager := service.GetSessionManager()
	sessionManager.RLock()
	defer sessionManager.RUnlock()
	return len(sessionManager.Sessions)
}