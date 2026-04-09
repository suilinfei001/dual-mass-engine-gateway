// Package auth 提供简单的认证和授权功能
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

var (
	// 默认管理员凭证
	defaultAdminUser = "admin"
	defaultAdminPass = "admin123"

	// 可通过环境变量覆盖
	adminUser string
	adminPass string

	once sync.Once
)

// 初始化管理员凭证
func init() {
	adminUser = os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = defaultAdminUser
	}
	adminPass = os.Getenv("ADMIN_PASS")
	if adminPass == "" {
		adminPass = defaultAdminPass
	}
}

// GetAdminCredentials 返回管理员凭证
func GetAdminCredentials() (username, password string) {
	return adminUser, adminPass
}

// Session 会话信息
type Session struct {
	SessionID string
	Username  string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// IsValid 检查会话是否有效
func (s *Session) IsValid() bool {
	return time.Now().Before(s.ExpiresAt)
}

// SessionStore 会话存储
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

// NewSessionStore 创建新的会话存储
func NewSessionStore() *SessionStore {
	store := &SessionStore{
		sessions: make(map[string]*Session),
	}
	// 启动清理过期会话的 goroutine
	go store.cleanupExpiredSessions()
	return store
}

// generateSessionID 生成随机会话 ID
func generateSessionID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// CreateSession 创建新会话
func (s *SessionStore) CreateSession(username string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionID := generateSessionID()
	session := &Session{
		SessionID: sessionID,
		Username:  username,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	s.sessions[sessionID] = session
	return session, nil
}

// GetSession 获取会话
func (s *SessionStore) GetSession(sessionID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, false
	}
	if !session.IsValid() {
		return nil, false
	}
	return session, true
}

// DeleteSession 删除会话
func (s *SessionStore) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// cleanupExpiredSessions 定期清理过期会话
func (s *SessionStore) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, session := range s.sessions {
			if now.After(session.ExpiresAt) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Username string `json:"username,omitempty"`
}

// ValidateCredentials 验证用户凭证
func ValidateCredentials(username, password string) bool {
	return username == adminUser && password == adminPass
}

// Errors
var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidSession     = errors.New("invalid session")
)

// AuthMiddleware 认证中间件配置
type AuthMiddleware struct {
	sessionStore *SessionStore
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(store *SessionStore) *AuthMiddleware {
	return &AuthMiddleware{
		sessionStore: store,
	}
}

// GetSessionStore 获取会话存储
func (a *AuthMiddleware) GetSessionStore() *SessionStore {
	return a.sessionStore
}

// SessionInfo 从请求中提取的会话信息
type SessionInfo struct {
	SessionID string
	Username  string
}

// CheckSession 检查会话是否有效
func (a *AuthMiddleware) CheckSession(sessionID string) (*SessionInfo, error) {
	session, exists := a.sessionStore.GetSession(sessionID)
	if !exists {
		return nil, ErrInvalidSession
	}
	return &SessionInfo{
		SessionID: session.SessionID,
		Username:  session.Username,
	}, nil
}

// Helper function for JSON responses
func WriteJSONResponse(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// ParseLoginRequest 解析登录请求
func ParseLoginRequest(data []byte) (*LoginRequest, error) {
	var req LoginRequest
	err := json.Unmarshal(data, &req)
	return &req, err
}

// SetAdminCredentials 设置管理员凭证（用于测试）
func SetAdminCredentials(username, password string) {
	adminUser = username
	adminPass = password
}

// ResetAdminCredentials 重置为默认凭证
func ResetAdminCredentials() {
	adminUser = defaultAdminUser
	adminPass = defaultAdminPass
}
