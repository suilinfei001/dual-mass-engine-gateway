package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github-hub/event-processor/internal/storage"
)

const (
	testDBDSN     = "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true"
	testUsername  = "testuser"
	testPassword  = "testpass123"
	testEmail     = "test@aishu.cn"
	adminUsername = "admin"
	adminPassword = "admin123"
)

func getTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("mysql", testDBDSN)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	return db
}

func cleanupTestUser(db *sql.DB, identifier string) {
	db.Exec("DELETE FROM users WHERE username = ? OR email = ?", identifier, identifier)
}

func TestUserRegistration(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleRegister)

	cleanupTestUser(db, testEmail)
	defer cleanupTestUser(db, testEmail)

	body := map[string]interface{}{
		"username": testUsername,
		"password": testPassword,
		"email":    testEmail,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		t.Errorf("Expected status %d or %d, got %d. Body: %s", http.StatusOK, http.StatusCreated, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response["success"].(bool) {
		t.Errorf("Expected success=true, got %v", response["success"])
	}
}

func TestUserRegistration_InvalidEmail(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleRegister)

	body := map[string]interface{}{
		"username": "testuser2",
		"password": testPassword,
		"email":    "test@gmail.com",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Errorf("Expected success=false, got %v", response["success"])
	}

	expectedMsg := "Email must end with @aishu.cn"
	if response["message"] != expectedMsg {
		t.Errorf("Expected message '%s', got '%v'", expectedMsg, response["message"])
	}
}

func TestUserRegistration_MissingFields(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleRegister)

	body := map[string]interface{}{
		"username": "",
		"password": "",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Errorf("Expected success=false for missing fields, got %v", response["success"])
	}
}

func TestUserRegistration_MissingEmail(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleRegister)

	body := map[string]interface{}{
		"username": "testuser3",
		"password": testPassword,
		"email":    "",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Errorf("Expected success=false for missing email, got %v", response["success"])
	}

	expectedMsg := "Email is required"
	if response["message"] != expectedMsg {
		t.Errorf("Expected message '%s', got '%v'", expectedMsg, response["message"])
	}
}

func TestUserRegistration_EmailAlreadyRegistered(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleRegister)

	// First, create a test user with a specific email
	testEmail := "duplicate@aishu.cn"
	cleanupTestUser(db, testEmail)
	defer cleanupTestUser(db, testEmail)

	// Register the user first time
	body := map[string]interface{}{
		"username": "testuser_duplicate",
		"password": testPassword,
		"email":    testEmail,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	testHandler.ServeHTTP(w, req)

	// Verify first registration succeeded
	var firstResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &firstResponse)
	if firstResponse["success"] != true {
		t.Fatalf("First registration should succeed, got %v", firstResponse)
	}

	// Now try to register with the same email again
	body = map[string]interface{}{
		"username": "anotheruser",
		"password": testPassword,
		"email":    testEmail,
	}
	jsonBody, _ = json.Marshal(body)

	req = httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Errorf("Expected success=false for duplicate email, got %v", response["success"])
	}

	expectedMsg := "Email already registered"
	if response["message"] != expectedMsg {
		t.Errorf("Expected message '%s', got '%v'", expectedMsg, response["message"])
	}
}

func TestUserRegistration_PasswordTooShort(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleRegister)

	body := map[string]interface{}{
		"username": "testuser4",
		"password": "12345",
		"email":    "test4@aishu.cn",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Errorf("Expected success=false for short password, got %v", response["success"])
	}

	expectedMsg := "Password must be at least 6 characters"
	if response["message"] != expectedMsg {
		t.Errorf("Expected message '%s', got '%v'", expectedMsg, response["message"])
	}
}

func TestUserRegistration_ValidEmailFormats(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleRegister)

	testCases := []string{
		"test@aishu.cn",
		"TEST@AISHU.CN",
		"user.name@aishu.cn",
		"user123@aishu.cn",
	}

	for _, email := range testCases {
		cleanupTestUser(db, email)

		body := map[string]interface{}{
			"username": "testuser",
			"password": testPassword,
			"email":    email,
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		testHandler.ServeHTTP(w, req)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		if !response["success"].(bool) {
			t.Errorf("Expected success=true for email '%s', got %v. Message: %v", email, response["success"], response["message"])
		}

		cleanupTestUser(db, email)
	}
}

func TestUserLogin_Success(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleLogin)

	body := map[string]interface{}{
		"username": adminUsername,
		"password": adminPassword,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response["success"].(bool) {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	userData := response["user"].(map[string]interface{})
	if userData["username"] != adminUsername {
		t.Errorf("Expected username '%s', got '%s'", adminUsername, userData["username"])
	}

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("Expected session cookie to be set")
	}
}

func TestUserLogin_InvalidPassword(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleLogin)

	body := map[string]interface{}{
		"username": adminUsername,
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Errorf("Expected success=false for invalid password, got %v", response["success"])
	}
}

func TestUserLogin_NonExistentUser(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleLogin)

	body := map[string]interface{}{
		"username": "nonexistentuser",
		"password": "somepassword",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Errorf("Expected success=false for non-existent user, got %v", response["success"])
	}
}

func TestUserLogin_MissingFields(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleLogin)

	body := map[string]interface{}{
		"username": "",
		"password": "",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Errorf("Expected success=false for missing fields, got %v", response["success"])
	}
}

func TestHandleAuthStatus_WithSession(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	loginHandler := http.HandlerFunc(handler.handleLogin)
	authHandler := http.HandlerFunc(handler.handleAuthStatus)

	loginBody := map[string]interface{}{
		"username": adminUsername,
		"password": adminPassword,
	}
	jsonLoginBody, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonLoginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()

	loginHandler.ServeHTTP(loginW, loginReq)

	sessionCookie := loginW.Result().Cookies()[0]

	authReq := httptest.NewRequest("GET", "/api/auth/status", nil)
	authReq.AddCookie(sessionCookie)
	authW := httptest.NewRecorder()

	authHandler.ServeHTTP(authW, authReq)

	if authW.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, authW.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(authW.Body.Bytes(), &response)

	if !response["loggedIn"].(bool) {
		t.Errorf("Expected loggedIn=true, got %v", response["loggedIn"])
	}
}

func TestHandleAuthStatus_WithoutSession(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleAuthStatus)

	req := httptest.NewRequest("GET", "/api/auth/status", nil)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["loggedIn"] != false {
		t.Errorf("Expected loggedIn=false without session, got %v", response["loggedIn"])
	}
}

func TestHandleSearchUsers_WithPagination(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleSearchUsers)

	adminSession, _ := userStorage.GetUserByUsername(adminUsername)
	session, err := userStorage.CreateSession(adminSession.ID)
	if err != nil {
		t.Skipf("Skipping test: cannot create session: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/admin/users?page=1&pageSize=20", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: session["session_id"].(string)})
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Body: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response["success"].(bool) {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	data := response["data"].([]interface{})
	if len(data) == 0 {
		t.Error("Expected at least one user, got empty list")
	}

	if response["total"] == nil {
		t.Error("Expected total count in response")
	}
}

func TestHandleSearchUsers_WithKeyword(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleSearchUsers)

	adminSession, _ := userStorage.GetUserByUsername(adminUsername)
	session, err := userStorage.CreateSession(adminSession.ID)
	if err != nil {
		t.Skipf("Skipping test: cannot create session: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/admin/users?keyword=admin&page=1&pageSize=20", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: session["session_id"].(string)})
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response["success"].(bool) {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	data := response["data"].([]interface{})
	if len(data) == 0 {
		t.Error("Expected to find admin user with keyword search")
	}
}

func TestHandleSearchUsers_WithoutAuth(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testHandler := http.HandlerFunc(handler.handleSearchUsers)

	req := httptest.NewRequest("GET", "/api/admin/users?page=1&pageSize=20", nil)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Errorf("Expected success=true (handler doesn't check auth), got %v", response["success"])
	}
}

func TestHandleUpdateUserPassword(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	testPassword := "newpassword123"

	adminSession, _ := userStorage.GetUserByUsername(adminUsername)
	session, err := userStorage.CreateSession(adminSession.ID)
	if err != nil {
		t.Skipf("Skipping test: cannot create session: %v", err)
	}

	body := map[string]interface{}{
		"new_password": testPassword,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("PUT", "/api/admin/users/1/password", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session_id", Value: session["session_id"].(string)})
	w := httptest.NewRecorder()

	handler.handleUpdateUserPassword(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response["success"].(bool) {
		t.Errorf("Expected success=true, got %v. Message: %v", response["success"], response["message"])
	}
}

func TestHandleLogout(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)
	handler := NewUserHandler(userStorage)

	adminSession, _ := userStorage.GetUserByUsername(adminUsername)
	session, _ := userStorage.CreateSession(adminSession.ID)

	testHandler := http.HandlerFunc(handler.handleLogout)

	req := httptest.NewRequest("POST", "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: session["session_id"].(string)})
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if !response["success"].(bool) {
		t.Errorf("Expected success=true, got %v", response["success"])
	}

	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Error("Expected cookie to be cleared")
	}

	for _, cookie := range cookies {
		if cookie.Name == "session_id" && cookie.MaxAge != -1 {
			t.Error("Expected session cookie to be expired")
		}
	}
}

func TestStorageGetUsersWithPagination(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)

	users, total, err := userStorage.GetUsersWithPagination(1, 20)
	if err != nil {
		t.Errorf("GetUsersWithPagination error: %v", err)
		return
	}

	if total == 0 {
		t.Error("Expected at least one user in database")
	}

	if len(users) == 0 {
		t.Error("Expected users list to not be empty")
	}
}

func TestStorageGetUsersWithPaginationAndKeyword(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)

	users, total, err := userStorage.GetUsersWithPaginationAndKeyword(1, 20, "admin")
	if err != nil {
		t.Errorf("GetUsersWithPaginationAndKeyword error: %v", err)
		return
	}

	if total == 0 {
		t.Error("Expected to find admin user with keyword search")
	}

	found := false
	for _, u := range users {
		if u.Username == "admin" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find 'admin' user in search results")
	}
}

func TestStorageGetUsersWithPaginationAndKeyword_Empty(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)

	users, total, err := userStorage.GetUsersWithPaginationAndKeyword(1, 20, "nonexistentkeyword123")
	if err != nil {
		t.Errorf("GetUsersWithPaginationAndKeyword error: %v", err)
		return
	}

	if total != 0 {
		t.Errorf("Expected 0 users for non-existent keyword, got %d", total)
	}

	if len(users) != 0 {
		t.Error("Expected empty users list for non-existent keyword")
	}
}

func TestStorageCreateAndDeleteSession(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	userStorage := storage.NewMySQLUserStorage(db)

	adminUser, err := userStorage.GetUserByUsername(adminUsername)
	if err != nil {
		t.Skipf("Skipping test: cannot get admin user: %v", err)
	}

	session, err := userStorage.CreateSession(adminUser.ID)
	if err != nil {
		t.Errorf("CreateSession error: %v", err)
		return
	}

	sessionID := session["session_id"].(string)

	err = userStorage.DeleteSession(sessionID)
	if err != nil {
		t.Errorf("DeleteSession error: %v", err)
	}
}
