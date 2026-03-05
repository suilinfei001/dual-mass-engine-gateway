package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github-hub/internal/quality/models"
	"github-hub/internal/quality/storage"
)

func setupTestServer(t *testing.T) (*Server, storage.Storage) {
	store := storage.NewMockStorage()
	server, err := NewServerWithStorage(store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	return server, store
}

func TestHandleQualityCheckUpdate(t *testing.T) {
	server, store := setupTestServer(t)

	event := &models.GitHubEvent{
		EventID:     "test-event-quality-check",
		EventType:   models.EventTypePush,
		EventStatus: models.EventStatusPending,
		Repository:  "test/repo",
		Branch:      "main",
		Payload:     []byte(`{}`),
		CreatedAt:   models.Now(),
		UpdatedAt:   models.Now(),
	}
	store.CreateEvent(event)

	check := &models.PRQualityCheck{
		GitHubEventID: event.EventID,
		CheckType:     models.QualityCheckTypeCompilation,
		CheckStatus:   models.QualityCheckStatusPending,
		Stage:         models.StageTypeBasicCI,
		StageOrder:    1,
		CheckOrder:    1,
		RetryCount:    0,
		CreatedAt:     models.Now(),
		UpdatedAt:     models.Now(),
	}
	store.CreateQualityCheck(check)

	tests := []struct {
		name           string
		checkID        int
		payload        map[string]interface{}
		expectedStatus int
		wantStatus     models.QualityCheckStatus
		wantOutput     *string
		wantDuration   *float64
	}{
		{
			name:    "update check_status to passed",
			checkID: check.ID,
			payload: map[string]interface{}{
				"check_status": "passed",
			},
			expectedStatus: http.StatusOK,
			wantStatus:     models.QualityCheckStatusPassed,
		},
		{
			name:    "update with output and duration",
			checkID: check.ID,
			payload: map[string]interface{}{
				"check_status":     "passed",
				"output":           "Compilation successful",
				"duration_seconds": 5.5,
			},
			expectedStatus: http.StatusOK,
			wantStatus:     models.QualityCheckStatusPassed,
			wantOutput:     strPtr("Compilation successful"),
			wantDuration:   floatPtr(5.5),
		},
		{
			name:    "update with error_message",
			checkID: check.ID,
			payload: map[string]interface{}{
				"check_status":  "failed",
				"error_message": "Build failed: undefined variable",
			},
			expectedStatus: http.StatusOK,
			wantStatus:     models.QualityCheckStatusFailed,
		},
		{
			name:    "invalid check_status",
			checkID: check.ID,
			payload: map[string]interface{}{
				"check_status": "invalid_status",
			},
			expectedStatus: http.StatusBadRequest,
			wantStatus:     models.QualityCheckStatusFailed,
		},
		{
			name:    "non-existent check",
			checkID: 9999,
			payload: map[string]interface{}{
				"check_status": "passed",
			},
			expectedStatus: http.StatusNotFound,
			wantStatus:     models.QualityCheckStatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest(http.MethodPut, "/api/quality-checks/"+strconv.Itoa(tt.checkID), bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			server.handleQualityCheckUpdate(rec, req, tt.checkID)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
				t.Errorf("response body: %s", rec.Body.String())
				return
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(rec.Body.Bytes(), &response)

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Fatal("response data is not a map")
				}

				if data["check_status"] != string(tt.wantStatus) {
					t.Errorf("expected check_status '%s', got '%s'", tt.wantStatus, data["check_status"])
				}

				if tt.wantOutput != nil {
					if data["output"] != *tt.wantOutput {
						t.Errorf("expected output '%s', got '%s'", *tt.wantOutput, data["output"])
					}
				}

				if tt.wantDuration != nil {
					if data["duration_seconds"] != *tt.wantDuration {
						t.Errorf("expected duration_seconds %v, got %v", *tt.wantDuration, data["duration_seconds"])
					}
				}
			}
		})
	}
}

func TestHandleQualityCheckUpdate_AllFields(t *testing.T) {
	server, store := setupTestServer(t)

	event := &models.GitHubEvent{
		EventID:     "test-event-all-fields",
		EventType:   models.EventTypePush,
		EventStatus: models.EventStatusPending,
		Repository:  "test/repo",
		Branch:      "main",
		Payload:     []byte(`{}`),
		CreatedAt:   models.Now(),
		UpdatedAt:   models.Now(),
	}
	store.CreateEvent(event)

	check := &models.PRQualityCheck{
		GitHubEventID: event.EventID,
		CheckType:     models.QualityCheckTypeCompilation,
		CheckStatus:   models.QualityCheckStatusPending,
		Stage:         models.StageTypeBasicCI,
		StageOrder:    1,
		CheckOrder:    1,
		RetryCount:    0,
		CreatedAt:     models.Now(),
		UpdatedAt:     models.Now(),
	}
	store.CreateQualityCheck(check)

	payload := map[string]interface{}{
		"check_status":     "passed",
		"output":           "All tests passed",
		"error_message":    nil,
		"started_at":       "2026-02-12T10:00:00Z",
		"completed_at":     "2026-02-12T10:05:30Z",
		"duration_seconds": 330.0,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/quality-checks/"+strconv.Itoa(check.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.handleQualityCheckUpdate(rec, req, check.ID)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})

	if data["check_status"] != "passed" {
		t.Errorf("expected check_status 'passed', got '%s'", data["check_status"])
	}

	if data["output"] != "All tests passed" {
		t.Errorf("expected output 'All tests passed', got '%s'", data["output"])
	}

	if data["duration_seconds"] != 330.0 {
		t.Errorf("expected duration_seconds 330.0, got %v", data["duration_seconds"])
	}
}

func TestHandleBatchUpdateQualityChecks(t *testing.T) {
	server, store := setupTestServer(t)

	event := &models.GitHubEvent{
		EventID:       "test-event-batch-update",
		EventType:     models.EventTypePush,
		EventStatus:   models.EventStatusPending,
		Repository:    "test/repo",
		Branch:        "main",
		QualityChecks: models.CreateChecksForEvent("test-event-batch-update"),
		Payload:       []byte(`{}`),
		CreatedAt:     models.Now(),
		UpdatedAt:     models.Now(),
	}
	store.CreateEvent(event)

	checkIDs := make([]int, len(event.QualityChecks))
	for i, qc := range event.QualityChecks {
		checkIDs[i] = qc.ID
	}

	updates := []map[string]interface{}{
		{
			"id":               checkIDs[0],
			"check_status":     "passed",
			"output":           "Compilation successful",
			"duration_seconds": 10.5,
		},
		{
			"id":               checkIDs[1],
			"check_status":     "passed",
			"output":           "All unit tests passed",
			"duration_seconds": 25.0,
		},
		{
			"id":               checkIDs[2],
			"check_status":     "failed",
			"error_message":    "Integration test failed",
			"duration_seconds": 60.0,
		},
	}

	payload := map[string]interface{}{
		"quality_checks": updates,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/events/"+strconv.Itoa(event.ID)+"/quality-checks/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.handleBatchUpdateQualityChecks(rec, req, event.ID)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("expected success to be true")
	}

	check1, _ := store.GetQualityCheck(checkIDs[0])
	if check1.CheckStatus != models.QualityCheckStatusPassed {
		t.Errorf("expected check %d status 'passed', got '%s'", checkIDs[0], check1.CheckStatus)
	}
	if check1.Output == nil || *check1.Output != "Compilation successful" {
		t.Errorf("expected check %d output 'Compilation successful', got %v", checkIDs[0], check1.Output)
	}

	check2, _ := store.GetQualityCheck(checkIDs[1])
	if check2.CheckStatus != models.QualityCheckStatusPassed {
		t.Errorf("expected check %d status 'passed', got '%s'", checkIDs[1], check2.CheckStatus)
	}

	check3, _ := store.GetQualityCheck(checkIDs[2])
	if check3.CheckStatus != models.QualityCheckStatusFailed {
		t.Errorf("expected check %d status 'failed', got '%s'", checkIDs[2], check3.CheckStatus)
	}
	if check3.ErrorMessage == nil || *check3.ErrorMessage != "Integration test failed" {
		t.Errorf("expected check %d error_message 'Integration test failed', got %v", checkIDs[2], check3.ErrorMessage)
	}
}

func TestHandleEventStatusUpdate(t *testing.T) {
	server, store := setupTestServer(t)

	event := &models.GitHubEvent{
		EventID:     "test-event-status-update",
		EventType:   models.EventTypePush,
		EventStatus: models.EventStatusPending,
		Repository:  "test/repo",
		Branch:      "main",
		Payload:     []byte(`{}`),
		CreatedAt:   models.Now(),
		UpdatedAt:   models.Now(),
	}
	store.CreateEvent(event)

	payload := map[string]interface{}{
		"event_status": "completed",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPut, "/api/events/"+strconv.Itoa(event.ID)+"/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.handleUpdateEventStatus(rec, req, event.ID)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d. Body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	updatedEvent, _ := store.GetEvent(event.ID)
	if updatedEvent.EventStatus != models.EventStatusCompleted {
		t.Errorf("expected event status 'completed', got '%s'", updatedEvent.EventStatus)
	}
}

func strPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

func TestHandleLogin(t *testing.T) {
	server, _ := setupTestServer(t)

	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		expectSuccess  bool
		expectCookie   bool
	}{
		{
			name:           "valid login",
			username:       "admin",
			password:       "admin",
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			expectCookie:   true,
		},
		{
			name:           "invalid username",
			username:       "wrong",
			password:       "admin",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
			expectCookie:   false,
		},
		{
			name:           "invalid password",
			username:       "admin",
			password:       "wrong",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
			expectCookie:   false,
		},
		{
			name:           "empty credentials",
			username:       "",
			password:       "",
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
			expectCookie:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"username": tt.username,
				"password": tt.password,
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			server.handleLogin(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(rec.Body.Bytes(), &response)

			success, _ := response["success"].(bool)
			if success != tt.expectSuccess {
				t.Errorf("expected success %v, got %v", tt.expectSuccess, success)
			}

			cookies := rec.Result().Cookies()
			hasSessionCookie := false
			for _, cookie := range cookies {
				if cookie.Name == "session_id" {
					hasSessionCookie = true
					if tt.expectCookie && cookie.Value == "" {
						t.Errorf("expected non-empty session cookie")
					}
					if !tt.expectCookie && cookie.MaxAge > 0 {
						t.Errorf("expected no session cookie, got one with MaxAge=%d", cookie.MaxAge)
					}
				}
			}

			if hasSessionCookie != tt.expectCookie {
				t.Errorf("expected session cookie %v, got %v", tt.expectCookie, hasSessionCookie)
			}
		})
	}
}

func TestHandleCheckLogin(t *testing.T) {
	server, _ := setupTestServer(t)

	t.Run("not logged in", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
		rec := httptest.NewRecorder()

		server.handleCheckLogin(rec, req)

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response["is_logged_in"] != false {
			t.Errorf("expected is_logged_in false, got %v", response["is_logged_in"])
		}
	})

	t.Run("logged in with valid session", func(t *testing.T) {
		// First login
		loginPayload := map[string]interface{}{
			"username": "admin",
			"password": "admin",
		}
		body, _ := json.Marshal(loginPayload)
		loginReq := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
		loginReq.Header.Set("Content-Type", "application/json")
		loginRec := httptest.NewRecorder()

		server.handleLogin(loginRec, loginReq)

		// Get session cookie
		cookies := loginRec.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session_id" {
				sessionCookie = cookie
				break
			}
		}

		if sessionCookie == nil {
			t.Fatal("expected session cookie to be set")
		}

		// Now check login status with the session cookie
		checkReq := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
		checkReq.AddCookie(sessionCookie)
		checkRec := httptest.NewRecorder()

		server.handleCheckLogin(checkRec, checkReq)

		var response map[string]interface{}
		json.Unmarshal(checkRec.Body.Bytes(), &response)

		if response["is_logged_in"] != true {
			t.Errorf("expected is_logged_in true, got %v", response["is_logged_in"])
		}

		if response["username"] != "admin" {
			t.Errorf("expected username 'admin', got %v", response["username"])
		}
	})

	t.Run("logged in with invalid session", func(t *testing.T) {
		invalidCookie := &http.Cookie{
			Name:  "session_id",
			Value: "invalid_session_id",
		}

		req := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
		req.AddCookie(invalidCookie)
		rec := httptest.NewRecorder()

		server.handleCheckLogin(rec, req)

		var response map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &response)

		if response["is_logged_in"] != false {
			t.Errorf("expected is_logged_in false for invalid session, got %v", response["is_logged_in"])
		}
	})
}

func TestHandleLogout(t *testing.T) {
	server, _ := setupTestServer(t)

	t.Run("logout with valid session", func(t *testing.T) {
		// First login
		loginPayload := map[string]interface{}{
			"username": "admin",
			"password": "admin",
		}
		body, _ := json.Marshal(loginPayload)
		loginReq := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(body))
		loginReq.Header.Set("Content-Type", "application/json")
		loginRec := httptest.NewRecorder()

		server.handleLogin(loginRec, loginReq)

		// Get session cookie
		cookies := loginRec.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "session_id" {
				sessionCookie = cookie
				break
			}
		}

		// Logout
		logoutReq := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
		logoutReq.AddCookie(sessionCookie)
		logoutRec := httptest.NewRecorder()

		server.handleLogout(logoutRec, logoutReq)

		var response map[string]interface{}
		json.Unmarshal(logoutRec.Body.Bytes(), &response)

		if response["success"] != true {
			t.Errorf("expected success true, got %v", response["success"])
		}

		// Check that cookie is cleared
		logoutCookies := logoutRec.Result().Cookies()
		hasClearedCookie := false
		for _, cookie := range logoutCookies {
			if cookie.Name == "session_id" && cookie.MaxAge < 0 {
				hasClearedCookie = true
			}
		}

		if !hasClearedCookie {
			t.Errorf("expected session cookie to be cleared")
		}

		// Verify session is no longer valid
		checkReq := httptest.NewRequest(http.MethodGet, "/api/check-login", nil)
		checkReq.AddCookie(sessionCookie)
		checkRec := httptest.NewRecorder()

		server.handleCheckLogin(checkRec, checkReq)

		var checkResponse map[string]interface{}
		json.Unmarshal(checkRec.Body.Bytes(), &checkResponse)

		if checkResponse["is_logged_in"] != false {
			t.Errorf("expected is_logged_in false after logout, got %v", checkResponse["is_logged_in"])
		}
	})
}

func TestSessionStore(t *testing.T) {
	store := NewSessionStore()

	t.Run("create and get session", func(t *testing.T) {
		session, err := store.CreateSession("testuser")
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		if session.SessionID == "" {
			t.Error("expected non-empty session ID")
		}

		if session.Username != "testuser" {
			t.Errorf("expected username 'testuser', got '%s'", session.Username)
		}

		// Get the session
		retrieved := store.GetSession(session.SessionID)
		if retrieved == nil {
			t.Error("expected to retrieve session")
		}

		if retrieved.Username != "testuser" {
			t.Errorf("expected username 'testuser', got '%s'", retrieved.Username)
		}
	})

	t.Run("get non-existent session", func(t *testing.T) {
		retrieved := store.GetSession("nonexistent")
		if retrieved != nil {
			t.Error("expected nil for non-existent session")
		}
	})

	t.Run("delete session", func(t *testing.T) {
		session, _ := store.CreateSession("testuser2")

		// Verify session exists
		if store.GetSession(session.SessionID) == nil {
			t.Error("expected session to exist before deletion")
		}

		// Delete session
		store.DeleteSession(session.SessionID)

		// Verify session is deleted
		if store.GetSession(session.SessionID) != nil {
			t.Error("expected session to be deleted")
		}
	})
}
