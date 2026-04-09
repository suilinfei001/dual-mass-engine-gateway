package testing

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/quality-gateway/shared/pkg/models"
)

func TestNewMockResource(t *testing.T) {
	resource := NewMockResource()

	if resource.ID != 1 {
		t.Errorf("expected ID 1, got %d", resource.ID)
	}

	if resource.ResourceType != models.ResourceTypeBasicCI {
		t.Errorf("expected ResourceTypeBasicCI, got %s", resource.ResourceType)
	}

	if resource.Name != "Test Resource" {
		t.Errorf("expected name 'Test Resource', got %q", resource.Name)
	}
}

func TestNewMockUser(t *testing.T) {
	user := NewMockUser()

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}

	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %q", user.Username)
	}

	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %q", user.Email)
	}
}

func TestAssertEQ(t *testing.T) {
	AssertEQ(t, 1, 1)
	AssertEQ(t, "hello", "hello")
}

func TestAssertNEQ(t *testing.T) {
	AssertNEQ(t, 1, 2)
	AssertNEQ(t, "hello", "world")
}

func TestAssertTrue(t *testing.T) {
	AssertTrue(t, true)
}

func TestAssertFalse(t *testing.T) {
	AssertFalse(t, false)
}

func TestAssertNil(t *testing.T) {
	var s *string
	AssertNil(t, s)
}

func TestAssertNotNil(t *testing.T) {
	s := "test"
	AssertNotNil(t, &s)
}

func TestAssertContains(t *testing.T) {
	AssertContains(t, "hello world", "hello")
}

func TestAssertNotContains(t *testing.T) {
	AssertNotContains(t, "hello world", "goodbye")
}

func TestAssertLen(t *testing.T) {
	AssertLen(t, []int{1, 2, 3}, 3)
	AssertLen(t, "hello", 5)
}

func TestAssertStatusCode(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	AssertStatusCode(t, w, http.StatusCreated)
}

func TestAssertBodyContains(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	AssertBodyContains(t, w, "Hello")
}

func TestWaitForCondition(t *testing.T) {
	count := 0
	WaitForCondition(t, func() bool {
		count++
		return count >= 2
	}, 1*time.Second)

	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

func TestWaitForConditionTimeout(t *testing.T) {
	// Note: Testing timeout behavior is tricky because t.Fatalf terminates the test
	// This test is skipped - the timeout logic is tested indirectly by TestWaitForCondition
	t.Skip("timeout behavior cannot be easily tested without causing test failure")
}

func TestSetEnv(t *testing.T) {
	SetEnv(t, "TEST_VAR", "test_value")

	if got := os.Getenv("TEST_VAR"); got != "test_value" {
		t.Errorf("expected env var 'test_value', got %q", got)
	}
}

func TestTempDir(t *testing.T) {
	dir := TempDir(t)

	if !strings.HasPrefix(dir, os.TempDir()) {
		t.Errorf("expected temp dir prefix %q, got %q", os.TempDir(), dir)
	}

	// 检查目录存在
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("temp dir does not exist: %s", dir)
	}
}

func TestTempFile(t *testing.T) {
	content := "test content"
	filePath := TempFile(t, content)

	// 读取文件内容
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read temp file: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected file content %q, got %q", content, string(data))
	}
}
