package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"myblog-gogogo/crypto"
)

// SessionManager 会话管理器
type SessionManager struct {
	Sessions map[string]*crypto.ECCManager
	mu       sync.RWMutex
}

// 全局会话管理器实例
var sessionManager = &SessionManager{
	Sessions: make(map[string]*crypto.ECCManager),
}

// GetSessionManager 获取会话管理器实例
func GetSessionManager() *SessionManager {
	return sessionManager
}

// Lock 获取写锁
func (sm *SessionManager) Lock() {
	sm.mu.Lock()
}

// Unlock 释放写锁
func (sm *SessionManager) Unlock() {
	sm.mu.Unlock()
}

// RLock 获取读锁
func (sm *SessionManager) RLock() {
	sm.mu.RLock()
}

// RUnlock 释放读锁
func (sm *SessionManager) RUnlock() {
	sm.mu.RUnlock()
}

// generateSessionID 生成会话ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "session_" + hex.EncodeToString(bytes)
}

// CleanupExpiredSessions 清理过期会话（定期调用）
func CleanupExpiredSessions() {
	sessionManager.mu.Lock()
	defer sessionManager.mu.Unlock()

	for sessionID, ecc := range sessionManager.Sessions {
		if ecc.IsExpired() {
			delete(sessionManager.Sessions, sessionID)
		}
	}
}

// GetSessionCount 获取当前活跃会话数
func GetSessionCount() int {
	sessionManager.mu.RLock()
	defer sessionManager.mu.RUnlock()
	return len(sessionManager.Sessions)
}