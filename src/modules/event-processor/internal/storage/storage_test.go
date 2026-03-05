package storage

import (
	"database/sql"
	"testing"
	"time"

	"github-hub/event-processor/internal/models"
)

func TestMySQLUserStorage_GetUserByUsername(t *testing.T) {
	db, err := sql.Open("mysql", "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	storage := &MySQLUserStorage{db: db}

	user, err := storage.GetUserByUsername("admin")
	if err != nil {
		t.Errorf("GetUserByUsername error: %v", err)
		return
	}

	if user == nil {
		t.Error("Expected user, got nil")
		return
	}

	if user.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", user.Username)
	}

	if user.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", user.Role)
	}
}

func TestMySQLUserStorage_SearchUsers(t *testing.T) {
	db, err := sql.Open("mysql", "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	storage := &MySQLUserStorage{db: db}

	users, err := storage.SearchUsers("admin")
	if err != nil {
		t.Errorf("SearchUsers error: %v", err)
		return
	}

	if len(users) == 0 {
		t.Error("Expected at least one user, got empty list")
		return
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

func TestMySQLUserStorage_CreateSession(t *testing.T) {
	db, err := sql.Open("mysql", "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	storage := &MySQLUserStorage{db: db}

	session, err := storage.CreateSession(1)
	if err != nil {
		t.Errorf("CreateSession error: %v", err)
		return
	}

	if session == nil {
		t.Error("Expected session, got nil")
		return
	}

	sessionID, ok := session["session_id"].(string)
	if !ok {
		t.Error("Expected session_id in session map")
		return
	}

	if sessionID == "" {
		t.Error("Expected non-empty session_id")
	}

	if err := storage.DeleteSession(sessionID); err != nil {
		t.Errorf("DeleteSession error: %v", err)
	}
}

func TestMySQLUserStorage_GetSessionWithUser(t *testing.T) {
	db, err := sql.Open("mysql", "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	storage := &MySQLUserStorage{db: db}

	session, err := storage.CreateSession(1)
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}

	sessionID := session["session_id"].(string)
	defer storage.DeleteSession(sessionID)

	result, err := storage.GetSessionWithUser(sessionID)
	if err != nil {
		t.Errorf("GetSessionWithUser error: %v", err)
		return
	}

	if result == nil {
		t.Error("Expected session result, got nil")
		return
	}

	userData, ok := result["user"].(map[string]interface{})
	if !ok {
		t.Error("Expected user data in session result")
		return
	}

	if userData["username"] != "admin" {
		t.Errorf("Expected username 'admin', got '%v'", userData["username"])
	}

	if userData["role"] != "admin" {
		t.Errorf("Expected role 'admin', got '%v'", userData["role"])
	}
}

func TestMySQLConfigStorage_GetConfig(t *testing.T) {
	db, err := sql.Open("mysql", "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	storage := &MySQLConfigStorage{db: db}

	_, err = storage.GetConfig("test_key_12345")
	if err != nil {
		if err.Error() == "config not found" {
			return
		}
		t.Errorf("GetConfig error: %v", err)
	}
}

func TestMySQLConfigStorage_GetAllConfigs(t *testing.T) {
	db, err := sql.Open("mysql", "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	storage := &MySQLConfigStorage{db: db}

	configs, err := storage.GetAllConfigs()
	if err != nil {
		t.Errorf("GetAllConfigs error: %v", err)
		return
	}

	if configs == nil {
		t.Error("Expected configs list, got nil")
	}
}

func TestMySQLResourceStorage_CreateAndDeleteResource(t *testing.T) {
	db, err := sql.Open("mysql", "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	storage := &MySQLResourceStorage{db: db}

	resource := &models.ExecutableResource{
		ResourceName: "test_resource_" + time.Now().Format("20060102150405"),
		RepoPath:     "/tmp/test",
		ResourceType: "script",
		Description:  "Test resource",
		CreatorID:    1,
		CreatorName:  "test_user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = storage.CreateResource(resource)
	if err != nil {
		t.Errorf("CreateResource error: %v", err)
		return
	}

	if resource.ID == 0 {
		t.Error("Expected resource ID to be set after creation")
	}

	err = storage.DeleteResource(resource.ID)
	if err != nil {
		t.Errorf("DeleteResource error: %v", err)
	}
}

func TestMySQLConfigStorage_AIRequestPoolSize(t *testing.T) {
	db, err := sql.Open("mysql", "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	configStorage := &MySQLConfigStorage{db: db}

	// Test GetAIRequestPoolSize with default (not set)
	poolSize, err := configStorage.GetAIRequestPoolSize()
	if err != nil {
		t.Errorf("GetAIRequestPoolSize error: %v", err)
	}
	if poolSize != 50 {
		t.Errorf("Expected default pool size 50, got %d", poolSize)
	}

	// Test SetAIRequestPoolSize
	err = configStorage.SetAIRequestPoolSize(100)
	if err != nil {
		t.Errorf("SetAIRequestPoolSize error: %v", err)
	}

	// Verify the value was saved
	poolSize, err = configStorage.GetAIRequestPoolSize()
	if err != nil {
		t.Errorf("GetAIRequestPoolSize error: %v", err)
	}
	if poolSize != 100 {
		t.Errorf("Expected pool size 100, got %d", poolSize)
	}

	// Test validation: pool size must be greater than concurrency
	// First set a low concurrency
	configStorage.SetAIConcurrency(50)

	// Try to set pool size <= concurrency (should fail)
	err = configStorage.SetAIRequestPoolSize(50)
	if err == nil {
		t.Error("Expected error when pool size <= concurrency")
	}

	err = configStorage.SetAIRequestPoolSize(40)
	if err == nil {
		t.Error("Expected error when pool size < concurrency")
	}

	// Valid pool size > concurrency should succeed
	err = configStorage.SetAIRequestPoolSize(60)
	if err != nil {
		t.Errorf("SetAIRequestPoolSize with valid size should succeed: %v", err)
	}

	// Cleanup: reset to default
	configStorage.SetAIRequestPoolSize(50)
	configStorage.SetAIConcurrency(20)
}

func TestMySQLConfigStorage_AIRequestPoolSize_Bounds(t *testing.T) {
	db, err := sql.Open("mysql", "root:root123456@tcp(localhost:3307)/event_processor?parseTime=true")
	if err != nil {
		t.Skipf("Skipping test: cannot connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: cannot ping database: %v", err)
	}

	configStorage := &MySQLConfigStorage{db: db}

	// Store original concurrency
	origConcurrency, _ := configStorage.GetAIConcurrency()
	defer func() {
		configStorage.SetAIConcurrency(origConcurrency)
		configStorage.SetAIRequestPoolSize(50)
	}()

	// Set concurrency to minimum so pool size tests won't fail the > concurrency check
	configStorage.SetAIConcurrency(1)

	// Test minimum bound (1) - should succeed since 1 > 1 is false, but min bound is 1
	// Actually, since pool size must be > concurrency, and concurrency is 1, pool size must be at least 2
	err = configStorage.SetAIRequestPoolSize(2)
	if err != nil {
		t.Errorf("SetAIRequestPoolSize(2) should succeed: %v", err)
	}

	// Test maximum bound (200)
	err = configStorage.SetAIRequestPoolSize(200)
	if err != nil {
		t.Errorf("SetAIRequestPoolSize(200) should succeed: %v", err)
	}

	// Test below minimum (should fail)
	err = configStorage.SetAIRequestPoolSize(0)
	if err == nil {
		t.Error("Expected error for pool size < 1")
	}

	// Test above maximum (should fail)
	err = configStorage.SetAIRequestPoolSize(201)
	if err == nil {
		t.Error("Expected error for pool size > 200")
	}

	// Test pool size <= concurrency (should fail)
	err = configStorage.SetAIRequestPoolSize(1)
	if err == nil {
		t.Error("Expected error for pool size <= concurrency")
	}
}
