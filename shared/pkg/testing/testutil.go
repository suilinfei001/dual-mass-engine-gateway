package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/quality-gateway/shared/pkg/logger"
	"github.com/quality-gateway/shared/pkg/storage"
)

// SetupTestDB 创建测试数据库 (使用 SQLite 内存数据库)
func SetupTestDB(t *testing.T) *storage.DB {
	t.Helper()

	log := logger.New(logger.Config{Level: logger.InfoLevel, DisableTime: true})
	db, err := storage.Open(storage.Config{
		Driver: "sqlite",
		DSN:    ":memory:",
	}, log)
	if err != nil {
		t.Fatalf("setup test db: %v", err)
	}

	return db
}

// SetupTestDBWithSchema 创建测试数据库并初始化 schema
func SetupTestDBWithSchema(t *testing.T, schema string) *storage.DB {
	t.Helper()

	db := SetupTestDB(t)

	if schema != "" {
		_, err := db.Exec(context.Background(), schema)
		if err != nil {
			db.Close()
			t.Fatalf("create test schema: %v", err)
		}
	}

	return db
}

// TeardownTestDB 清理测试数据库
func TeardownTestDB(t *testing.T, db *storage.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Errorf("close test db: %v", err)
	}
}

// MockLogger 创建测试用的日志记录器
func NewMockLogger() *mockLogger {
	return &mockLogger{}
}

type mockLogger struct {
	messages []string
}

func (l *mockLogger) Debug(msg string, args ...string) {
	l.log("DEBUG", msg, args...)
}

func (l *mockLogger) Info(msg string, args ...string) {
	l.log("INFO", msg, args...)
}

func (l *mockLogger) Warn(msg string, args ...string) {
	l.log("WARN", msg, args...)
}

func (l *mockLogger) Error(msg string, args ...string) {
	l.log("ERROR", msg, args...)
}

func (l *mockLogger) log(level, msg string, args ...string) {
	l.messages = append(l.messages, level+": "+msg)
}

func (l *mockLogger) String(key, value string) string {
	return value
}

// WaitForCondition 等待条件满足
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if time.Now().After(deadline) {
			t.Fatalf("condition not met within timeout %v", timeout)
		}
		if condition() {
			return
		}
		<-ticker.C
	}
}

// WaitForConditionWithPoll 等待条件满足 (自定义轮询间隔)
func WaitForConditionWithPoll(t *testing.T, condition func() bool, timeout, pollInterval time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if time.Now().After(deadline) {
			t.Fatalf("condition not met within timeout %v", timeout)
		}
		if condition() {
			return
		}
		<-ticker.C
	}
}

// AssertNoError 断言无错误
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

// AssertError 断言有错误
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// AssertErrorContains 断言错误包含指定字符串
func AssertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), substr) {
		t.Errorf("error to contain %q, got %q", substr, err.Error())
	}
}

// AssertEQ 断言相等
func AssertEQ[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNEQ 断言不等
func AssertNEQ[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got == want {
		t.Errorf("got %v, want not %v", got, want)
	}
}

// AssertTrue 断言为真
func AssertTrue(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if !condition {
		if len(msg) > 0 {
			t.Errorf("expected true, got false: %s", strings.Join(msg, " "))
		} else {
			t.Error("expected true, got false")
		}
	}
}

// AssertFalse 断言为假
func AssertFalse(t *testing.T, condition bool, msg ...string) {
	t.Helper()
	if condition {
		if len(msg) > 0 {
			t.Errorf("expected false, got true: %s", strings.Join(msg, " "))
		} else {
			t.Error("expected false, got true")
		}
	}
}

// AssertNil 断言为 nil
func AssertNil(t *testing.T, got interface{}) {
	t.Helper()
	if !isNil(got) {
		t.Errorf("expected nil, got %v", got)
	}
}

// isNil checks if a value is nil, including typed nils.
func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

// AssertNotNil 断言不为 nil
func AssertNotNil(t *testing.T, got interface{}) {
	t.Helper()
	if got == nil {
		t.Error("expected not nil, got nil")
	}
}

// AssertEmpty 断言为空
func AssertEmpty(t *testing.T, got string) {
	t.Helper()
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

// AssertNotEmpty 断言不为空
func AssertNotEmpty(t *testing.T, got string) {
	t.Helper()
	if got == "" {
		t.Error("expected non-empty string, got empty")
	}
}

// AssertLen 断言长度
func AssertLen(t *testing.T, got interface{}, length int, msg ...string) {
	t.Helper()
	gotLen := 0
	switch v := got.(type) {
	case []any:
		gotLen = len(v)
	case []string:
		gotLen = len(v)
	case []int:
		gotLen = len(v)
	case []int64:
		gotLen = len(v)
	case map[string]any:
		gotLen = len(v)
	case string:
		gotLen = len(v)
	default:
		t.Fatalf("unsupported type for AssertLen: %T", got)
		return
	}

	if gotLen != length {
		if len(msg) > 0 {
			t.Errorf("expected length %d, got %d: %s", length, gotLen, strings.Join(msg, " "))
		} else {
			t.Errorf("expected length %d, got %d", length, gotLen)
		}
	}
}

// AssertContains 断言包含
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

// AssertNotContains 断言不包含
func AssertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected %q to not contain %q", s, substr)
	}
}

// TruncateTable 清空表
func TruncateTable(t *testing.T, db *storage.DB, table string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), fmt.Sprintf("DELETE FROM %s", table))
	if err != nil {
		t.Fatalf("truncate table %s: %v", table, err)
	}
}

// ExecSQL 执行 SQL 语句
func ExecSQL(t *testing.T, db *storage.DB, query string, args ...any) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), query, args...)
	if err != nil {
		t.Fatalf("exec SQL %q: %v", query, err)
	}
}

// QuerySQL 查询 SQL
func QuerySQL(t *testing.T, db *storage.DB, dest any, query string, args ...any) {
	t.Helper()
	err := db.QueryRow(context.Background(), query, args...).Scan(dest)
	if err != nil {
		t.Fatalf("query SQL %q: %v", query, err)
	}
}

// NewTestRequest 创建测试 HTTP 请求
func NewTestRequest(method, url string, body any) *http.Request {
	var reader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			reader = strings.NewReader(v)
		case []byte:
			reader = io.NopCloser(strings.NewReader(string(v)))
		default:
			// assume JSON for other types
			jsonData, err := json.Marshal(body)
			if err != nil {
				panic(fmt.Sprintf("marshal request body: %v", err))
			}
			reader = strings.NewReader(string(jsonData))
		}
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		panic(fmt.Sprintf("create request: %v", err))
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

// DoTestRequest 执行测试 HTTP 请求
func DoTestRequest(t *testing.T, handler http.Handler, method, url string, body any) *httptest.ResponseRecorder {
	t.Helper()

	req := NewTestRequest(method, url, body)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

// AssertStatusCode 断言 HTTP 状态码
func AssertStatusCode(t *testing.T, w *httptest.ResponseRecorder, status int) {
	t.Helper()
	if w.Code != status {
		t.Errorf("expected status %d, got %d", status, w.Code)
		t.Logf("response body: %s", w.Body.String())
	}
}

// AssertBodyContains 断言响应体包含
func AssertBodyContains(t *testing.T, w *httptest.ResponseRecorder, substr string) {
	t.Helper()
	body := w.Body.String()
	if !strings.Contains(body, substr) {
		t.Errorf("expected body to contain %q, got %q", substr, body)
	}
}

// AssertJSONEq 断言 JSON 相等
func AssertJSONEq(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("expected JSON:\n%s\n\ngot:\n%s", want, got)
	}
}

// GetEnv 获取环境变量，如果不存在返回默认值
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetEnv 设置环境变量 (测试用)
func SetEnv(t *testing.T, key, value string) {
	t.Helper()

	oldValue := os.Getenv(key)
	os.Setenv(key, value)
	t.Cleanup(func() {
		if oldValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, oldValue)
		}
	})
}

// TempDir 创建临时目录
func TempDir(t *testing.T) string {
	t.Helper()

	dir, err := os.MkdirTemp("", "test-")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// TempFile 创建临时文件
func TempFile(t *testing.T, content string) string {
	t.Helper()

	file, err := os.CreateTemp("", "test-")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer file.Close()

	if content != "" {
		if _, err := file.WriteString(content); err != nil {
			t.Fatalf("write temp file: %v", err)
		}
	}

	t.Cleanup(func() {
		os.Remove(file.Name())
	})
	return file.Name()
}

// CaptureOutput 捕获输出
func CaptureOutput(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		var buf strings.Builder
		io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = old

	output := <-done
	return output
}

// Retry 重试函数直到成功或超时
func Retry(t *testing.T, fn func() error, maxAttempts int) error {
	t.Helper()

	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}

	return fmt.Errorf("retry failed after %d attempts, last error: %w", maxAttempts, lastErr)
}

// Parallel 并行运行测试
func Parallel(t *testing.T, fns ...func(t *testing.T)) {
	t.Helper()

	done := make(chan struct{})
	for _, fn := range fns {
		go func(f func(t *testing.T)) {
			defer func() { done <- struct{}{} }()
			fn(t)
		}(fn)
	}

	for range fns {
		<-done
	}
}
