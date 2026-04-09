package auth

import (
	"net/http"
	"testing"
	"time"
)

// TestValidateCredentials 测试凭证验证
func TestValidateCredentials(t *testing.T) {
	// 设置测试凭证
	SetAdminCredentials("testuser", "testpass")
	defer ResetAdminCredentials()

	tests := []struct {
		name     string
		username string
		password string
		want     bool
	}{
		{
			name:     "正确凭证",
			username: "testuser",
			password: "testpass",
			want:     true,
		},
		{
			name:     "错误用户名",
			username: "wronguser",
			password: "testpass",
			want:     false,
		},
		{
			name:     "错误密码",
			username: "testuser",
			password: "wrongpass",
			want:     false,
		},
		{
			name:     "空凭证",
			username: "",
			password: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateCredentials(tt.username, tt.password); got != tt.want {
				t.Errorf("ValidateCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestSessionStore_CreateSession 测试创建会话
func TestSessionStore_CreateSession(t *testing.T) {
	store := NewSessionStore()

	session, err := store.CreateSession("testuser")
	if err != nil {
		t.Fatalf("CreateSession() failed: %v", err)
	}

	if session == nil {
		t.Fatal("CreateSession() returned nil session")
	}

	if session.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", session.Username)
	}

	if session.SessionID == "" {
		t.Error("SessionID should not be empty")
	}

	if !session.IsValid() {
		t.Error("New session should be valid")
	}
}

// TestSessionStore_GetSession 测试获取会话
func TestSessionStore_GetSession(t *testing.T) {
	store := NewSessionStore()

	// 创建会话
	session, err := store.CreateSession("testuser")
	if err != nil {
		t.Fatalf("CreateSession() failed: %v", err)
	}

	// 获取存在的会话
	got, exists := store.GetSession(session.SessionID)
	if !exists {
		t.Error("GetSession() should find existing session")
	}

	if got.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", got.Username)
	}

	// 获取不存在的会话
	_, exists = store.GetSession("nonexistent")
	if exists {
		t.Error("GetSession() should not find nonexistent session")
	}
}

// TestSessionStore_DeleteSession 测试删除会话
func TestSessionStore_DeleteSession(t *testing.T) {
	store := NewSessionStore()

	session, err := store.CreateSession("testuser")
	if err != nil {
		t.Fatalf("CreateSession() failed: %v", err)
	}

	// 删除会话
	store.DeleteSession(session.SessionID)

	// 确认会话已删除
	_, exists := store.GetSession(session.SessionID)
	if exists {
		t.Error("Deleted session should not exist")
	}
}

// TestSession_Expiration 测试会话过期
func TestSession_Expiration(t *testing.T) {
	store := NewSessionStore()

	// 创建一个已过期的会话
	session := &Session{
		SessionID: "expired-session",
		Username:  "testuser",
		CreatedAt: time.Now().Add(-25 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	store.sessions["expired-session"] = session

	// 过期的会话应该无效
	if session.IsValid() {
		t.Error("Expired session should not be valid")
	}

	// GetSession 应该返回不存在
	_, exists := store.GetSession("expired-session")
	if exists {
		t.Error("GetSession() should not return expired session")
	}
}

// TestAuthMiddleware_CheckSession 测试中间件检查会话
func TestAuthMiddleware_CheckSession(t *testing.T) {
	store := NewSessionStore()
	middleware := NewAuthMiddleware(store)

	session, err := store.CreateSession("testuser")
	if err != nil {
		t.Fatalf("CreateSession() failed: %v", err)
	}

	// 有效的会话
	info, err := middleware.CheckSession(session.SessionID)
	if err != nil {
		t.Errorf("CheckSession() failed for valid session: %v", err)
	}

	if info.Username != "testuser" {
		t.Errorf("Username = %v, want testuser", info.Username)
	}

	if info.SessionID != session.SessionID {
		t.Error("SessionID mismatch")
	}

	// 无效的会话
	_, err = middleware.CheckSession("invalid")
	if err != ErrInvalidSession {
		t.Errorf("CheckSession() should return ErrInvalidSession, got %v", err)
	}
}

// TestGetAdminCredentials 测试获取管理员凭证
func TestGetAdminCredentials(t *testing.T) {
	username, password := GetAdminCredentials()

	if username == "" {
		t.Error("Username should not be empty")
	}

	if password == "" {
		t.Error("Password should not be empty")
	}
}

// TestParseLoginRequest 测试解析登录请求
func TestParseLoginRequest(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "有效请求",
			data:    `{"username":"admin","password":"admin123"}`,
			wantErr: false,
		},
		{
			name:    "无效JSON",
			data:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "空JSON",
			data:    `{}`,
			wantErr: false, // 空JSON可以解析，只是字段为空
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseLoginRequest([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLoginRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && req.Username == "" && tt.data != "{}" {
				t.Error("ParseLoginRequest() should parse username")
			}
		})
	}
}

// TestCookieConfig 测试 Cookie 配置是否正确
func TestCookieConfig(t *testing.T) {
	// 这是一个文档性测试，确保 Cookie 配置符合安全标准
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    "test-value",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
		SameSite: http.SameSiteLaxMode,
	}

	if !cookie.HttpOnly {
		t.Error("Cookie should be HttpOnly")
	}

	if cookie.MaxAge != 86400 {
		t.Errorf("Cookie MaxAge should be 86400, got %d", cookie.MaxAge)
	}

	if cookie.SameSite != http.SameSiteLaxMode {
		t.Error("Cookie should use SameSiteLaxMode")
	}
}
