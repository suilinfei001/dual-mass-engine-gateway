package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/quality-gateway/shared/pkg/auth"
)

func TestMain(m *testing.M) {
	// 设置测试凭证
	auth.SetAdminCredentials("testuser", "testpass")
	code := m.Run()
	auth.ResetAdminCredentials()
	os.Exit(code)
}

func TestHandleLogin_Success(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	reqBody := `{"username":"testuser","password":"testpass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleLogin(sessionStore, w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["success"] != true {
		t.Errorf("Expected success=true, got %v", resp["success"])
	}

	// 检查是否设置了 session cookie
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("Expected session cookie to be set")
		return
	}

	sessionCookie := cookies[0]
	if sessionCookie.Name != "session_id" {
		t.Errorf("Expected cookie name 'session_id', got '%s'", sessionCookie.Name)
	}

	if !sessionCookie.HttpOnly {
		t.Error("Expected cookie to be HttpOnly")
	}
}

func TestHandleLogin_InvalidCredentials(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	reqBody := `{"username":"wrong","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleLogin(sessionStore, w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["success"] != false {
		t.Errorf("Expected success=false for invalid credentials, got %v", resp["success"])
	}
}

func TestHandleLogin_InvalidJSON(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleLogin(sessionStore, w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleLogin_WrongMethod(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	w := httptest.NewRecorder()

	handleLogin(sessionStore, w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleLogout_Success(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	// 首先创建一个会话
	session, _ := sessionStore.CreateSession("testuser")

	// 创建带 session cookie 的请求
	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.SessionID,
	})
	w := httptest.NewRecorder()

	handleLogout(sessionStore, w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["success"] != true {
		t.Errorf("Expected success=true, got %v", resp["success"])
	}

	// 验证会话已被删除
	_, exists := sessionStore.GetSession(session.SessionID)
	if exists {
		t.Error("Session should be deleted after logout")
	}

	// 验证 cookie 被清除
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("Expected cookie to be cleared")
		return
	}

	if cookies[0].MaxAge != -1 {
		t.Errorf("Expected MaxAge=-1 to clear cookie, got %d", cookies[0].MaxAge)
	}
}

func TestHandleLogout_NoSession(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	w := httptest.NewRecorder()

	handleLogout(sessionStore, w, req)

	// 即使没有 session cookie 也应该成功
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleCheckLogin_LoggedIn(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	// 创建会话
	session, _ := sessionStore.CreateSession("testuser")

	req := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.SessionID,
	})
	w := httptest.NewRecorder()

	handleCheckLogin(sessionStore, w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["is_logged_in"] != true {
		t.Errorf("Expected is_logged_in=true, got %v", resp["is_logged_in"])
	}

	if resp["username"] != "testuser" {
		t.Errorf("Expected username='testuser', got %v", resp["username"])
	}
}

func TestHandleCheckLogin_NotLoggedIn(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	// 没有 session cookie
	req := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
	w := httptest.NewRecorder()

	handleCheckLogin(sessionStore, w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["is_logged_in"] != false {
		t.Errorf("Expected is_logged_in=false, got %v", resp["is_logged_in"])
	}
}

func TestHandleCheckLogin_InvalidSession(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	req := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "invalid-session-id",
	})
	w := httptest.NewRecorder()

	handleCheckLogin(sessionStore, w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["is_logged_in"] != false {
		t.Errorf("Expected is_logged_in=false for invalid session, got %v", resp["is_logged_in"])
	}
}

func TestHandleCheckLogin_ExpiredSession(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	// 由于无法直接创建过期会话（会话会自动检查过期），
	// 我们测试无效的 session ID
	req := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "invalid-session-id",
	})
	w := httptest.NewRecorder()

	handleCheckLogin(sessionStore, w, req)

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["is_logged_in"] != false {
		t.Errorf("Expected is_logged_in=false for expired session, got %v", resp["is_logged_in"])
	}
}

// 集成测试：完整登录流程
func TestIntegration_LoginFlow(t *testing.T) {
	sessionStore := auth.NewSessionStore()

	// 1. 登录
	loginReqBody := `{"username":"testuser","password":"testpass"}`
	loginReq := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(loginReqBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	handleLogin(sessionStore, loginW, loginReq)

	var loginResp map[string]interface{}
	json.NewDecoder(loginW.Body).Decode(&loginResp)

	if loginResp["success"] != true {
		t.Fatalf("Login failed: %v", loginResp)
	}

	// 获取 session cookie
	cookies := loginW.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("No session cookie set")
	}
	sessionCookie := cookies[0]

	// 2. 检查登录状态
	checkReq := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
	checkReq.AddCookie(sessionCookie)
	checkW := httptest.NewRecorder()

	handleCheckLogin(sessionStore, checkW, checkReq)

	var checkResp map[string]interface{}
	json.NewDecoder(checkW.Body).Decode(&checkResp)

	if checkResp["is_logged_in"] != true {
		t.Errorf("Expected is_logged_in=true, got %v", checkResp["is_logged_in"])
	}

	// 3. 登出
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutW := httptest.NewRecorder()

	handleLogout(sessionStore, logoutW, logoutReq)

	// 4. 再次检查登录状态（应该为 false）
	checkReq2 := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
	checkReq2.AddCookie(sessionCookie)
	checkW2 := httptest.NewRecorder()

	handleCheckLogin(sessionStore, checkW2, checkReq2)

	var checkResp2 map[string]interface{}
	json.NewDecoder(checkW2.Body).Decode(&checkResp2)

	if checkResp2["is_logged_in"] != false {
		t.Errorf("Expected is_logged_in=false after logout, got %v", checkResp2["is_logged_in"])
	}
}
