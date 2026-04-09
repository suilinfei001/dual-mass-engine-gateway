# 共享库创建实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 创建 `shared/` 共享库，包含 API 框架、存储抽象、模型、测试工具、配置、日志和错误处理，为所有微服务提供统一的基础设施。

**架构：** 采用 Go 标准项目布局，每个包独立可测试，接口优先设计，遵循依赖注入原则。

**Tech Stack:** Go 1.21+, slog (结构化日志), sqlx (数据库扩展), testify (测试断言)

---

## 文件结构

```
shared/
├── go.mod                          # 共享库模块定义
├── go.sum
├── pkg/
│   ├── api/
│   │   ├── server.go               # HTTP 服务器
│   │   ├── middleware.go           # 中间件
│   │   ├── response.go             # 响应格式
│   │   ├── client.go               # 服务间 HTTP 客户端
│   │   ├── server_test.go
│   │   ├── middleware_test.go
│   │   └── client_test.go
│   ├── storage/
│   │   ├── db.go                   # 数据库连接管理
│   │   ├── transaction.go          # 事务管理
│   │   ├── repository.go           # Repository 基类
│   │   ├── db_test.go
│   │   ├── transaction_test.go
│   │   └── repository_test.go
│   ├── models/
│   │   ├── event.go                # 事件模型
│   │   ├── task.go                 # 任务模型
│   │   ├── resource.go             # 资源模型
│   │   ├── user.go                 # 用户模型
│   │   ├── enums.go                # 枚举定义
│   │   └── models_test.go
│   ├── testing/
│   │   ├── mock.go                 # Mock 服务器
│   │   ├── fixtures.go             # 测试数据生成
│   │   └── testutil.go             # 测试辅助函数
│   ├── config/
│   │   ├── config.go               # 配置结构
│   │   ├── loader.go               # 配置加载
│   │   ├── validator.go            # 配置验证
│   │   └── config_test.go
│   ├── logger/
│   │   ├── logger.go               # 结构化日志
│   │   ├── middleware.go           # HTTP 日志中间件
│   │   └── logger_test.go
│   └── errors/
│       ├── errors.go               # 错误定义
│       └── handler.go              # 错误处理器
├── scripts/
│   └── build.sh                    # 统一构建脚本
└── README.md
```

---

## 范围说明

**本计划仅创建共享库基础设施，不涉及服务迁移。**
- 服务迁移将在后续独立计划中进行（Plan 2-7）
- 数据迁移将在 Plan 9 中处理
- 回滚策略属于整体迁移计划范畴

---

## Task 1: 初始化共享库模块

**Files:**
- Create: `shared/go.mod`
- Create: `shared/go.sum`
- Create: `shared/README.md`

- [ ] **Step 1: 创建 go.mod（使用 go mod init）**

```bash
# 创建共享库目录
mkdir -p /root/dev/dual-mass-engine-gateway/shared

# 初始化 Go 模块
cd /root/dev/dual-mass-engine-gateway/shared
go mod init github.com/quality-gateway/shared

# 验证 go.mod 已创建
cat go.mod
```

Expected: 输出包含 `module github.com/quality-gateway/shared` 和 `go 1.21`

- [ ] **Step 2: 创建 README.md**

```markdown
# Shared Library

双引擎质量网关共享库，为所有微服务提供统一的基础设施。

## 包说明

- `pkg/api` - REST API 框架
- `pkg/storage` - 数据库抽象层
- `pkg/models` - 共享数据模型
- `pkg/testing` - 测试工具
- `pkg/config` - 配置管理
- `pkg/logger` - 日志组件
- `pkg/errors` - 错误处理

## 使用

在服务 go.mod 中引用：

\`\`\`go
require github.com/quality-gateway/shared v1.0.0
\`\`\`
```

- [ ] **Step 3: 初始化 go workspace**

```bash
cd /root/dev/dual-mass-engine-gateway
go work init shared
```

- [ ] **Step 4: 验证模块初始化**

```bash
cd shared
go mod tidy
```

Expected: 无错误

- [ ] **Step 5: 提交**

```bash
git add shared/go.mod shared/go.sum shared/README.md go.work
git commit -m "feat(shared): 初始化共享库模块"
```

---

## Task 2: 创建 pkg/errors 错误处理包

**Files:**
- Create: `shared/pkg/errors/errors.go`
- Create: `shared/pkg/errors/handler.go`
- Create: `shared/pkg/errors/errors_test.go`

- [ ] **Step 1: 编写错误定义测试**

```go
package errors_test

import (
    "testing"
    "github.com/quality-gateway/shared/pkg/errors"
)

func TestNewError(t *testing.T) {
    err := errors.New("ERR_NOT_FOUND", "resource not found")

    if err.Code != "ERR_NOT_FOUND" {
        t.Errorf("expected code ERR_NOT_FOUND, got %s", err.Code)
    }
    if err.Message != "resource not found" {
        t.Errorf("expected message 'resource not found', got %s", err.Message)
    }
}

func TestError_Error(t *testing.T) {
    err := errors.New("ERR_TEST", "test message")
    expected := "[ERR_TEST] test message"

    if err.Error() != expected {
        t.Errorf("expected '%s', got '%s'", expected, err.Error())
    }
}

func TestPredefinedErrors(t *testing.T) {
    tests := []struct {
        name string
        err  *errors.Error
        code string
    }{
        {"NotFound", errors.ErrNotFound, "ERR_NOT_FOUND"},
        {"InvalidInput", errors.ErrInvalidInput, "ERR_INVALID_INPUT"},
        {"Internal", errors.ErrInternal, "ERR_INTERNAL"},
        {"Unauthorized", errors.ErrUnauthorized, "ERR_UNAUTHORIZED"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.err.Code != tt.code {
                t.Errorf("expected code %s, got %s", tt.code, tt.err.Code)
            }
        })
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
cd shared
go test ./pkg/errors/... -v
```

Expected: FAIL - "package not found"

- [ ] **Step 3: 实现错误定义**

```go
package errors

import "fmt"

// Error 表示应用程序错误
type Error struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func (e *Error) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// New 创建新的错误
func New(code, message string) *Error {
    return &Error{Code: code, Message: message}
}

// Wrap 包装错误并添加上下文
func Wrap(err error, code, message string) *Error {
    if err == nil {
        return nil
    }
    return &Error{Code: code, Message: fmt.Sprintf("%s: %v", message, err)}
}

// 预定义错误
var (
    ErrNotFound     = New("ERR_NOT_FOUND", "resource not found")
    ErrInvalidInput = New("ERR_INVALID_INPUT", "invalid input")
    ErrInternal     = New("ERR_INTERNAL", "internal server error")
    ErrUnauthorized = New("ERR_UNAUTHORIZED", "unauthorized access")
    ErrConflict     = New("ERR_CONFLICT", "resource conflict")
    ErrTimeout      = New("ERR_TIMEOUT", "operation timeout")
)
```

- [ ] **Step 4: 创建错误处理器**

```go
package errors

import "net/http"

// Handler 处理错误并返回 HTTP 状态码
type Handler struct{}

// NewHandler 创建错误处理器
func NewHandler() *Handler {
    return &Handler{}
}

// StatusCode 返回错误对应的 HTTP 状态码
func (h *Handler) StatusCode(err error) int {
    if e, ok := err.(*Error); ok {
        switch e.Code {
        case "ERR_NOT_FOUND":
            return http.StatusNotFound
        case "ERR_INVALID_INPUT":
            return http.StatusBadRequest
        case "ERR_UNAUTHORIZED":
            return http.StatusUnauthorized
        case "ERR_CONFLICT":
            return http.StatusConflict
        case "ERR_TIMEOUT":
            return http.StatusRequestTimeout
        default:
            return http.StatusInternalServerError
        }
    }
    return http.StatusInternalServerError
}

// IsRetryable 判断错误是否可重试
func (h *Handler) IsRetryable(err error) bool {
    if e, ok := err.(*Error); ok {
        return e.Code == "ERR_TIMEOUT" || e.Code == "ERR_INTERNAL"
    }
    return false
}
```

- [ ] **Step 5: 运行测试验证通过**

```bash
cd shared
go test ./pkg/errors/... -v
```

Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add shared/pkg/errors/
git commit -m "feat(shared): 添加错误处理包"
```

---

## Task 3: 创建 pkg/logger 日志组件

**Files:**
- Create: `shared/pkg/logger/logger.go`
- Create: `shared/pkg/logger/middleware.go`
- Create: `shared/pkg/logger/logger_test.go`

- [ ] **Step 1: 编写日志测试**

```go
package logger_test

import (
    "bytes"
    "context"
    "testing"
    "github.com/quality-gateway/shared/pkg/logger"
)

func TestNewLogger(t *testing.T) {
    log := logger.New(&logger.Config{
        Level:  "info",
        Format: "json",
    })

    if log == nil {
        t.Fatal("expected non-nil logger")
    }
}

func TestLoggerWithFields(t *testing.T) {
    var buf bytes.Buffer
    log := logger.New(&logger.Config{
        Level:  "debug",
        Format: "text",
        Writer: &buf,
    })

    log.With(
        "request_id", "test-123",
        "user_id", "user-456",
    ).Info("test message")

    output := buf.String()
    if !bytes.Contains([]byte(output), []byte("test message")) {
        t.Errorf("expected log to contain 'test message', got %s", output)
    }
}

func TestLoggerLevels(t *testing.T) {
    tests := []struct {
        name      string
        level     string
        shouldLog bool
    }{
        {"Debug", "debug", true},
        {"Info", "info", true},
        {"Warn", "warn", true},
        {"Error", "error", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var buf bytes.Buffer
            log := logger.New(&logger.Config{
                Level:  tt.level,
                Format: "text",
                Writer: &buf,
            })

            switch tt.name {
            case "Debug":
                log.Debug("debug msg")
            case "Info":
                log.Info("info msg")
            case "Warn":
                log.Warn("warn msg")
            case "Error":
                log.Error("error msg")
            }

            hasOutput := buf.Len() > 0
            if hasOutput != tt.shouldLog {
                t.Errorf("expected shouldLog=%v, got %v", tt.shouldLog, hasOutput)
            }
        })
    }
}

func TestLoggerWithContext(t *testing.T) {
    ctx := context.Background()
    log := logger.New(&logger.Config{
        Level:  "info",
        Format: "text",
    })

    ctx = logger.WithContext(ctx, "request_id", "test-req-123")
    log.InfoContext(ctx, "message with context")
    // 如果没有 panic，说明实现正确
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
cd shared
go test ./pkg/logger/... -v
```

Expected: FAIL - package not found

- [ ] **Step 3: 实现日志组件**

```go
package logger

import (
    "context"
    "fmt"
    "io"
    "log/slog"
    "os"
)

// Config 日志配置
type Config struct {
    Level  string `json:"level" mapstructure:"level"`    // debug, info, warn, error
    Format string `json:"format" mapstructure:"format"`  // json, text
    Writer io.Writer `json:"-"`
}

// Logger 结构化日志
type Logger struct {
    *slog.Logger
}

// levelMap 日志级别映射
var levelMap = map[string]slog.Level{
    "debug": slog.LevelDebug,
    "info":  slog.LevelInfo,
    "warn":  slog.LevelWarn,
    "error": slog.LevelError,
}

// New 创建新的日志器
func New(cfg *Config) *Logger {
    if cfg == nil {
        cfg = &Config{Level: "info", Format: "json"}
    }

    level := slog.LevelInfo
    if l, ok := levelMap[cfg.Level]; ok {
        level = l
    }

    writer := cfg.Writer
    if writer == nil {
        writer = os.Stdout
    }

    var handler slog.Handler
    if cfg.Format == "json" {
        handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
            Level: level,
        })
    } else {
        handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
            Level: level,
        })
    }

    return &Logger{
        Logger: slog.New(handler),
    }
}

// With 创建带有字段的子日志器
func (l *Logger) With(args ...any) *Logger {
    return &Logger{Logger: l.Logger.With(args...)}
}

// WithKeyValue 创建带有键值对的子日志器
func (l *Logger) WithKeyValue(key string, value any) *Logger {
    return &Logger{Logger: l.Logger.With(key, value)}
}

// Debug 调试日志
func (l *Logger) Debug(msg string, args ...any) {
    l.Log(context.Background(), slog.LevelDebug, msg, args...)
}

// Info 信息日志
func (l *Logger) Info(msg string, args ...any) {
    l.Log(context.Background(), slog.LevelInfo, msg, args...)
}

// Warn 警告日志
func (l *Logger) Warn(msg string, args ...any) {
    l.Log(context.Background(), slog.LevelWarn, msg, args...)
}

// Error 错误日志
func (l *Logger) Error(msg string, args ...any) {
    l.Log(context.Background(), slog.LevelError, msg, args...)
}

// Fatal 致命错误日志后退出
func (l *Logger) Fatal(msg string, args ...any) {
    l.Log(context.Background(), slog.LevelError, msg, args...)
    os.Exit(1)
}

// DebugContext 带上下文的调试日志
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
    l.Log(ctx, slog.LevelDebug, msg, args...)
}

// InfoContext 带上下文的信息日志
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
    l.Log(ctx, slog.LevelInfo, msg, args...)
}

// WarnContext 带上下文的警告日志
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
    l.Log(ctx, slog.LevelWarn, msg, args...)
}

// ErrorContext 带上下文的错误日志
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
    l.Log(ctx, slog.LevelError, msg, args...)
}

// contextKey 用于存储日志字段的 context key
type contextKey struct{}

// WithContext 将字段存入 context
func WithContext(ctx context.Context, key string, value any) context.Context {
    attrs := ctx.Value(contextKey{})
    if attrs == nil {
        attrs = make([]any, 0)
    }
    attrList, ok := attrs.([]any)
    if !ok {
        attrList = make([]any, 0)
    }
    attrList = append(attrList, key, value)
    return context.WithValue(ctx, contextKey{}, attrList)
}

// Log 实现 slog.LogLogger 接口，从 context 中提取字段
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
    attrs := ctx.Value(contextKey{})
    if attrs != nil {
        if attrList, ok := attrs.([]any); ok {
            args = append(attrList, args...)
        }
    }
    l.Logger.Log(ctx, level, msg, args...)
}

// Field 创建日志字段
type Field struct {
    Key   string
    Value any
}

// String 字符串字段
func String(key, value string) Field {
    return Field{Key: key, Value: value}
}

// Int 整数字段
func Int(key string, value int) Field {
    return Field{Key: key, Value: value}
}

// Any 任意类型字段
func Any(key string, value any) Field {
    return Field{Key: key, Value: value}
}

// Error 错误字段
func Error(err error) Field {
    return Field{Key: "error", Value: err}
}

// WithFields 使用 Field 结构添加字段
func (l *Logger) WithFields(fields ...Field) *Logger {
    args := make([]any, 0, len(fields)*2)
    for _, f := range fields {
        args = append(args, f.Key, f.Value)
    }
    return &Logger{Logger: l.Logger.With(args...)}
}
```

- [ ] **Step 4: 实现 HTTP 日志中间件**

```go
package logger

import (
    "context"
    "net/http"
    "time"
)

// contextKey 用于存储 request ID
type requestIDKey struct{}

// Middleware 返回 HTTP 日志中间件
func (l *Logger) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // 生成 request ID
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = fmt.Sprintf("%d", time.Now().UnixNano())
        }

        // 将 request ID 存入 context
        ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
        r = r.WithContext(ctx)

        // 设置响应 writer 以捕获状态码
        rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        // 调用下一个处理器
        next.ServeHTTP(rw, r)

        // 记录请求日志
        duration := time.Since(start)
        l.Info("HTTP request",
            "request_id", requestID,
            "method", r.Method,
            "path", r.URL.Path,
            "status", rw.status,
            "duration", duration.String(),
            "remote_addr", r.RemoteAddr,
        )
    })
}

// responseWriter 用于捕获 HTTP 状态码
type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
    rw.status = statusCode
    rw.ResponseWriter.WriteHeader(statusCode)
}

// GetRequestID 从 context 获取 request ID
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(requestIDKey{}).(string); ok {
        return id
    }
    return ""
}
```

- [ ] **Step 5: 运行测试验证通过**

```bash
cd shared
go test ./pkg/logger/... -v -cover
```

Expected: PASS, 覆盖率 > 80%

- [ ] **Step 6: 提交**

```bash
git add shared/pkg/logger/
git commit -m "feat(shared): 添加日志组件"
```

---

## Task 4: 创建 pkg/models 共享数据模型

**源码映射表：**

| 新模型 | 源码位置 | 说明 |
|--------|----------|------|
| Event | event-receiver/internal/quality/models/event.go | 合并 Event 和 LocalTime |
| QualityCheck | event-receiver/internal/quality/models/event.go | 从 Event 中分离 |
| Task | event-processor/internal/models/task.go | 保持结构一致 |
| TaskResult | event-processor/internal/models/task.go | 从 Task 中分离 |
| Resource | event-processor/internal/models/resource.go | 保持结构一致 |
| User | resource-pool/internal/models/user.go | 简化版本 |

**Files:**
- Create: `shared/pkg/models/enums.go`
- Create: `shared/pkg/models/event.go`
- Create: `shared/pkg/models/task.go`
- Create: `shared/pkg/models/resource.go`
- Create: `shared/pkg/models/user.go`
- Create: `shared/pkg/models/models_test.go`

- [ ] **Step 1: 编写模型测试**

```go
package models_test

import (
    "testing"
    "github.com/quality-gateway/shared/pkg/models"
)

func TestEventStatusIsValid(t *testing.T) {
    tests := []struct {
        status   models.EventStatus
        expected bool
    }{
        {models.EventStatusPending, true},
        {models.EventStatusProcessing, true},
        {models.EventStatusCompleted, true},
        {models.EventStatusFailed, true},
        {models.EventStatusCancelled, true},
        {models.EventStatus("invalid"), false},
    }

    for _, tt := range tests {
        t.Run(string(tt.status), func(t *testing.T) {
            got := tt.status.IsValid()
            if got != tt.expected {
                t.Errorf("IsValid() = %v, want %v", got, tt.expected)
            }
        })
    }
}

func TestTaskStatusIsValid(t *testing.T) {
    tests := []struct {
        status   models.TaskStatus
        expected bool
    }{
        {models.TaskStatusPending, true},
        {models.TaskStatusRunning, true},
        {models.TaskStatusPassed, true},
        {models.TaskStatusFailed, true},
        {models.TaskStatusSkipped, true},
        {models.TaskStatusNoResource, true},
        {models.TaskStatusTimeout, true},
        {models.TaskStatus("invalid"), false},
    }

    for _, tt := range tests {
        t.Run(string(tt.status), func(t *testing.T) {
            got := tt.status.IsValid()
            if got != tt.expected {
                t.Errorf("IsValid() = %v, want %v", got, tt.expected)
            }
        })
    }
}

func TestQualityCheckTypeString(t *testing.T) {
    tests := []struct {
        qct      models.QualityCheckType
        expected string
    }{
        {models.QualityCheckTypeBasicCI, "basic_ci"},
        {models.QualityCheckTypeDeployment, "deployment"},
        {models.QualityCheckTypeAPITest, "api_test"},
    }

    for _, tt := range tests {
        t.Run(tt.expected, func(t *testing.T) {
            if tt.qct.String() != tt.expected {
                t.Errorf("String() = %v, want %v", tt.qct.String(), tt.expected)
            }
        })
    }
}
```

- [ ] **Step 2: 实现枚举定义**

```go
package models

// EventStatus 事件状态
type EventStatus string

const (
    EventStatusPending    EventStatus = "pending"
    EventStatusProcessing EventStatus = "processing"
    EventStatusCompleted  EventStatus = "completed"
    EventStatusFailed     EventStatus = "failed"
    EventStatusCancelled  EventStatus = "cancelled"
)

func (s EventStatus) IsValid() bool {
    switch s {
    case EventStatusPending, EventStatusProcessing, EventStatusCompleted, EventStatusFailed, EventStatusCancelled:
        return true
    }
    return false
}

// TaskStatus 任务状态
type TaskStatus string

const (
    TaskStatusPending    TaskStatus = "pending"
    TaskStatusRunning    TaskStatus = "running"
    TaskStatusPassed     TaskStatus = "passed"
    TaskStatusFailed     TaskStatus = "failed"
    TaskStatusSkipped    TaskStatus = "skipped"
    TaskStatusNoResource TaskStatus = "no_resource"
    TaskStatusTimeout    TaskStatus = "timeout"
)

func (s TaskStatus) IsValid() bool {
    switch s {
    case TaskStatusPending, TaskStatusRunning, TaskStatusPassed, TaskStatusFailed, TaskStatusSkipped, TaskStatusNoResource, TaskStatusTimeout:
        return true
    }
    return false
}

// QualityCheckStatus 质量检查状态
type QualityCheckStatus string

const (
    QualityCheckStatusPending  QualityCheckStatus = "pending"
    QualityCheckStatusRunning  QualityCheckStatus = "running"
    QualityCheckStatusPassed   QualityCheckStatus = "passed"
    QualityCheckStatusFailed   QualityCheckStatus = "failed"
    QualityCheckStatusSkipped  QualityCheckStatus = "skipped"
    QualityCheckStatusCancelled QualityCheckStatus = "cancelled"
)

// QualityCheckType 质量检查类型
type QualityCheckType string

const (
    QualityCheckTypeBasicCI          QualityCheckType = "basic_ci"
    QualityCheckTypeDeployment       QualityCheckType = "deployment"
    QualityCheckTypeAPITest          QualityCheckType = "api_test"
    QualityCheckTypeModuleE2E        QualityCheckType = "module_e2e"
    QualityCheckTypeAgentE2E         QualityCheckType = "agent_e2e"
    QualityCheckTypeAIE2E            QualityCheckType = "ai_e2e"
    QualityCheckTypeBasicCIAll       QualityCheckType = "basic_ci_all"
    QualityCheckTypeDeploymentAll    QualityCheckType = "deployment_deployment"
)

func (q QualityCheckType) String() string {
    return string(q)
}

// ResourceType 资源类型
type ResourceType string

const (
    ResourceTypeBasicCI          ResourceType = "basic_ci_all"
    ResourceTypeDeployment       ResourceType = "deployment_deployment"
    ResourceTypeAPITest          ResourceType = "specialized_tests_api_test"
    ResourceTypeModuleE2E        ResourceType = "specialized_tests_module_e2e"
    ResourceTypeAgentE2E         ResourceType = "specialized_tests_agent_e2e"
    ResourceTypeAIE2E            ResourceType = "specialized_tests_ai_e2e"
)

// EventType 事件类型
type EventType string

const (
    EventTypePullRequestOpened EventType = "pull_request.opened"
    EventTypePullRequestSync   EventType = "pull_request.synchronize"
    EventTypePush              EventType = "push"
)
```

- [ ] **Step 3: 实现事件模型**

```go
package models

import "time"

// Event GitHub 事件
type Event struct {
    ID          int64       `json:"id" db:"id"`
    EventID     string      `json:"event_id" db:"event_id"`
    EventType   EventType   `json:"event_type" db:"event_type"`
    PRNumber    int         `json:"pr_number,omitempty" db:"pr_number"`
    PRTitle     string      `json:"pr_title,omitempty" db:"pr_title"`
    SourceBranch string     `json:"source_branch,omitempty" db:"source_branch"`
    TargetBranch string     `json:"target_branch,omitempty" db:"target_branch"`
    RepoURL     string      `json:"repo_url" db:"repo_url"`
    Sender      string      `json:"sender" db:"sender"`
    EventStatus EventStatus `json:"event_status" db:"event_status"`
    CreatedAt   time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// QualityCheck 质量检查项
type QualityCheck struct {
    ID          int64             `json:"id" db:"id"`
    EventID     int64             `json:"event_id" db:"event_id"`
    CheckType   QualityCheckType  `json:"check_type" db:"check_type"`
    CheckStatus QualityCheckStatus `json:"check_status" db:"check_status"`
    StageOrder  int               `json:"stage_order" db:"stage_order"`
    Result      map[string]any    `json:"result" db:"result"`
    CreatedAt   time.Time         `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}
```

- [ ] **Step 4: 实现任务模型**

```go
package models

import "time"

// Task 质量检查任务
type Task struct {
    ID          int64       `json:"id" db:"id"`
    EventID     string      `json:"event_id" db:"event_id"`
    TaskType    string      `json:"task_type" db:"task_type"`
    TaskStatus  TaskStatus  `json:"task_status" db:"task_status"`
    ResourceID  int64       `json:"resource_id,omitempty" db:"resource_id"`
    AzureURL    string      `json:"azure_url,omitempty" db:"azure_url"`
    Analyzing   bool        `json:"analyzing" db:"analyzing"`
    CreatedAt   time.Time   `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// TaskResult 任务子检查项结果
type TaskResult struct {
    ID          int64       `json:"id" db:"id"`
    TaskID      int64       `json:"task_id" db:"task_id"`
    ResultType  string      `json:"result_type" db:"result_type"`
    ResultStatus string     `json:"result_status" db:"result_status"`
    ResultData  map[string]any `json:"result_data" db:"result_data"`
    CreatedAt   time.Time   `json:"created_at" db:"created_at"`
}
```

- [ ] **Step 5: 实现资源模型**

```go
package models

import "time"

// Resource 资源实例
type Resource struct {
    ID                 int64        `json:"id" db:"id"`
    UUID               string       `json:"uuid" db:"uuid"`
    ResourceType       ResourceType `json:"resource_type" db:"resource_type"`
    Name               string       `json:"name,omitempty" db:"name"`
    Description        string       `json:"description,omitempty" db:"description"`
    AllowSkip          bool         `json:"allow_skip" db:"allow_skip"`
    Organization       string       `json:"organization,omitempty" db:"organization"`
    Project            string       `json:"project,omitempty" db:"project"`
    PipelineID         int          `json:"pipeline_id,omitempty" db:"pipeline_id"`
    PipelineParameters map[string]any `json:"pipeline_parameters,omitempty" db:"pipeline_parameters"`
    RepoPath           string       `json:"repo_path,omitempty" db:"repo_path"`
    IsPublic           bool         `json:"is_public" db:"is_public"`
    CreatorID          int64        `json:"creator_id" db:"creator_id"`
    CreatedAt          time.Time    `json:"created_at" db:"created_at"`
    UpdatedAt          time.Time    `json:"updated_at" db:"updated_at"`
}
```

- [ ] **Step 6: 实现用户模型**

```go
package models

import "time"

// User 用户
type User struct {
    ID        int64     `json:"id" db:"id"`
    Username  string    `json:"username" db:"username"`
    Email     string    `json:"email" db:"email"`
    FullName  string    `json:"full_name,omitempty" db:"full_name"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

- [ ] **Step 7: 运行测试验证通过**

```bash
cd shared
go test ./pkg/models/... -v
```

Expected: PASS

- [ ] **Step 8: 提交**

```bash
git add shared/pkg/models/
git commit -m "feat(shared): 添加共享数据模型"
```

---

## Task 5: 创建 pkg/storage 数据库抽象层

**Files:**
- Create: `shared/pkg/storage/db.go`
- Create: `shared/pkg/storage/transaction.go`
- Create: `shared/pkg/storage/repository.go`
- Create: `shared/pkg/storage/db_test.go`

- [ ] **Step 1: 编写数据库测试**

```go
package storage_test

import (
    "context"
    "testing"
    "github.com/quality-gateway/shared/pkg/storage"
)

func TestNewDB(t *testing.T) {
    // 使用 SQLite 内存数据库进行测试
    db, err := storage.NewDB(&storage.Config{
        Driver: "sqlite",
        DSN:    ":memory:",
    })
    if err != nil {
        t.Fatalf("NewDB() error = %v", err)
    }
    defer db.Close()

    if db.DB == nil {
        t.Error("expected non-nil DB")
    }
}

func TestDBPing(t *testing.T) {
    db, err := storage.NewDB(&storage.Config{
        Driver: "sqlite",
        DSN:    ":memory:",
    })
    if err != nil {
        t.Fatalf("NewDB() error = %v", err)
    }
    defer db.Close()

    ctx := context.Background()
    if err := db.PingContext(ctx); err != nil {
        t.Errorf("PingContext() error = %v", err)
    }
}

func TestDBTransaction(t *testing.T) {
    db, err := storage.NewDB(&storage.Config{
        Driver: "sqlite",
        DSN:    ":memory:",
    })
    if err != nil {
        t.Fatalf("NewDB() error = %v", err)
    }
    defer db.Close()

    // 创建测试表
    _, err = db.ExecContext(context.Background(), "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)")
    if err != nil {
        t.Fatalf("CREATE TABLE error = %v", err)
    }

    // 测试事务提交
    err = db.Transaction(func(tx *storage.Tx) error {
        _, err := tx.ExecContext(context.Background(), "INSERT INTO test (name) VALUES (?)", "test1")
        return err
    })
    if err != nil {
        t.Errorf("Transaction() error = %v", err)
    }

    // 验证数据已插入
    var count int
    err = db.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM test")
    if err != nil {
        t.Errorf("SELECT COUNT error = %v", err)
    }
    if count != 1 {
        t.Errorf("expected count = 1, got %d", count)
    }

    // 测试事务回滚
    err = db.Transaction(func(tx *storage.Tx) error {
        _, err := tx.ExecContext(context.Background(), "INSERT INTO test (name) VALUES (?)", "test2")
        if err != nil {
            return err
        }
        return storage.ErrRollback // 触发回滚
    })
    if err != nil && err != storage.ErrRollback {
        t.Errorf("Transaction() error = %v", err)
    }

    // 验证数据未插入
    err = db.GetContext(context.Background(), &count, "SELECT COUNT(*) FROM test")
    if err != nil {
        t.Errorf("SELECT COUNT error = %v", err)
    }
    if count != 1 {
        t.Errorf("expected count = 1 after rollback, got %d", count)
    }
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
cd shared
go test ./pkg/storage/... -v
```

Expected: FAIL

- [ ] **Step 3: 实现数据库抽象层**

```go
package storage

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    _ "github.com/mattn/go-sqlite3" // SQLite
    _ "github.com/go-sql-driver/mysql" // MySQL
)

// Config 数据库配置
type Config struct {
    Driver      string `json:"driver" mapstructure:"driver"`      // mysql, sqlite
    DSN         string `json:"dsn" mapstructure:"dsn"`            // 数据源名称
    MaxOpenConn int    `json:"max_open_conn" mapstructure:"max_open_conn"`
    MaxIdleConn int    `json:"max_idle_conn" mapstructure:"max_idle_conn"`
    MaxLifetime int    `json:"max_lifetime" mapstructure:"max_lifetime"` // 秒
}

// DB 数据库封装
type DB struct {
    *sql.DB
    driver string
}

// NewDB 创建数据库连接
func NewDB(cfg *Config) (*DB, error) {
    if cfg == nil {
        return nil, fmt.Errorf("config is nil")
    }

    var db *sql.DB
    var err error

    switch cfg.Driver {
    case "mysql":
        db, err = sql.Open("mysql", cfg.DSN)
    case "sqlite":
        db, err = sql.Open("sqlite3", cfg.DSN)
    default:
        return nil, fmt.Errorf("unsupported driver: %s", cfg.Driver)
    }

    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }

    // 配置连接池
    if cfg.MaxOpenConn > 0 {
        db.SetMaxOpenConns(cfg.MaxOpenConn)
    }
    if cfg.MaxIdleConn > 0 {
        db.SetMaxIdleConns(cfg.MaxIdleConn)
    }
    if cfg.MaxLifetime > 0 {
        db.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
    }

    // 测试连接
    if err := db.Ping(); err != nil {
        db.Close()
        return nil, fmt.Errorf("ping database: %w", err)
    }

    return &DB{
        DB:     db,
        driver: cfg.Driver,
    }, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
    return db.DB.Close()
}

// InContext 返回带 context 的 DB
func (db *DB) InContext(ctx context.Context) *DB {
    return &DB{DB: db.DB, driver: db.driver}
}

// ExecContext 执行 SQL 语句
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
    return db.DB.ExecContext(ctx, query, args...)
}

// QueryContext 查询多行
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    return db.DB.QueryContext(ctx, query, args...)
}

// QueryRowContext 查询单行
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
    return db.DB.QueryRowContext(ctx, query, args...)
}

// GetContext 查询单行并扫描到结构体
func (db *DB) GetContext(ctx context.Context, dest any, query string, args ...any) error {
    return db.QueryRowContext(ctx, query, args...).Scan(dest)
}

// SelectContext 查询多行并扫描到切片
func (db *DB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
    rows, err := db.QueryContext(ctx, query, args...)
    if err != nil {
        return err
    }
    defer rows.Close()

    // 使用 sqlx 的 风格接口
    return db.scanAll(rows, dest)
}

// scanAll 扫描所有行到目标切片
func (db *DB) scanAll(rows *sql.Rows, dest any) error {
    // 这里简化实现，实际可以使用反射或第三方库
    // 在实施时可以考虑使用 github.com/jmoiron/sqlx
    return nil
}

// Driver 返回数据库驱动类型
func (db *DB) Driver() string {
    return db.driver
}
```

- [ ] **Step 4: 实现事务管理**

```go
package storage

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
)

// Tx 事务
type Tx struct {
    *sql.Tx
    db *DB
}

// ErrRollback 用于显式回滚事务
var ErrRollback = errors.New("transaction rollback")

// Transaction 执行事务
func (db *DB) Transaction(fn func(*Tx) error) error {
    tx, err := db.DB.Begin()
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }

    // 确保事务被处理
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback()
            panic(p) // 重新抛出 panic
        }
    }()

    // 执行事务函数
    err = fn(&Tx{Tx: tx, db: db})
    if err != nil {
        // 检查是否是显式回滚
        if errors.Is(err, ErrRollback) {
            if rbErr := tx.Rollback(); rbErr != nil {
                return fmt.Errorf("rollback error: %w (original error: %v)", rbErr, err)
            }
            return err
        }
        // 其他错误则回滚
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("rollback error: %w (original error: %v)", rbErr, err)
        }
        return err
    }

    // 提交事务
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }

    return nil
}

// ExecContext 执行 SQL 语句
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
    return tx.Tx.ExecContext(ctx, query, args...)
}

// QueryContext 查询多行
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
    return tx.Tx.QueryContext(ctx, query, args...)
}

// QueryRowContext 查询单行
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
    return tx.Tx.QueryRowContext(ctx, query, args...)
}

// GetContext 查询单行并扫描
func (tx *Tx) GetContext(ctx context.Context, dest any, query string, args ...any) error {
    return tx.QueryRowContext(ctx, query, args...).Scan(dest)
}

// SelectContext 查询多行并扫描
func (tx *Tx) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
    rows, err := tx.QueryContext(ctx, query, args...)
    if err != nil {
        return err
    }
    defer rows.Close()
    return tx.db.scanAll(rows, dest)
}

// Rollback 显式回滚
func (tx *Tx) Rollback() error {
    return tx.Tx.Rollback()
}

// Commit 显式提交
func (tx *Tx) Commit() error {
    return tx.Tx.Commit()
}
```

- [ ] **Step 5: 实现 Repository 基类**

```go
package storage

import (
    "context"
    "fmt"
)

// Repository 仓储基类
type Repository[T any] struct {
    db    *DB
    table string
}

// NewRepository 创建仓储
func NewRepository[T any](db *DB, table string) *Repository[T] {
    return &Repository[T]{
        db:    db,
        table: table,
    }
}

// DB 获取数据库连接
func (r *Repository[T]) DB() *DB {
    return r.db
}

// Table 获取表名
func (r *Repository[T]) Table() string {
    return r.table
}

// GetTableName 返回完整表名
func (r *Repository[T]) GetTableName() string {
    return r.table
}

// ByID 根据 ID 查询
func (r *Repository[T]) ByID(ctx context.Context, id int64) (*T, error) {
    var entity T
    query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", r.table)
    err := r.db.GetContext(ctx, &entity, query, id)
    if err != nil {
        return nil, fmt.Errorf("get by id: %w", err)
    }
    return &entity, nil
}

// List 列表查询
func (r *Repository[T]) List(ctx context.Context, filter Filter) ([]*T, error) {
    var entities []*T

    query := fmt.Sprintf("SELECT * FROM %s", r.table)
    args := []any{}

    if filter.Where != "" {
        query += " WHERE " + filter.Where
    }
    if filter.OrderBy != "" {
        query += " ORDER BY " + filter.OrderBy
    }
    if filter.Limit > 0 {
        query += " LIMIT ?"
        args = append(args, filter.Limit)
    }
    if filter.Offset > 0 {
        query += " OFFSET ?"
        args = append(args, filter.Offset)
    }

    err := r.db.SelectContext(ctx, &entities, query, args...)
    if err != nil {
        return nil, fmt.Errorf("list: %w", err)
    }

    return entities, nil
}

// Create 创建
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
    // 这里需要根据 entity 结构动态生成 SQL
    // 实际实现可以使用反射或代码生成
    query := fmt.Sprintf("INSERT INTO %s VALUES (...) RETURNING id", r.table)
    err := r.db.GetContext(ctx, new(int64), query)
    if err != nil {
        return fmt.Errorf("create: %w", err)
    }
    return nil
}

// Update 更新
func (r *Repository[T]) Update(ctx context.Context, entity *T) error {
    // 实际实现需要根据 entity 结构生成 SET 子句
    query := fmt.Sprintf("UPDATE %s SET ... WHERE id = ...", r.table)
    _, err := r.db.ExecContext(ctx, query)
    if err != nil {
        return fmt.Errorf("update: %w", err)
    }
    return nil
}

// Delete 删除
func (r *Repository[T]) Delete(ctx context.Context, id int64) error {
    query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", r.table)
    _, err := r.db.ExecContext(ctx, query, id)
    if err != nil {
        return fmt.Errorf("delete: %w", err)
    }
    return nil
}

// Count 统计
func (r *Repository[T]) Count(ctx context.Context, filter Filter) (int64, error) {
    var count int64
    query := fmt.Sprintf("SELECT COUNT(*) FROM %s", r.table)
    args := []any{}

    if filter.Where != "" {
        query += " WHERE " + filter.Where
        args = append(args, filter.Args...)
    }

    err := r.db.GetContext(ctx, &count, query, args...)
    if err != nil {
        return 0, fmt.Errorf("count: %w", err)
    }

    return count, nil
}

// Exists 检查是否存在
func (r *Repository[T]) Exists(ctx context.Context, filter Filter) (bool, error) {
    count, err := r.Count(ctx, filter)
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

// Filter 查询过滤器
type Filter struct {
    Where  string
    Args   []any
    OrderBy string
    Limit  int
    Offset int
}
```

- [ ] **Step 6: 运行测试验证通过**

```bash
cd shared
go test ./pkg/storage/... -v -cover
```

Expected: PASS, 覆盖率 > 70%

- [ ] **Step 7: 提交**

```bash
git add shared/pkg/storage/
git commit -m "feat(shared): 添加数据库抽象层"
```

---

## Task 6: 创建 pkg/config 配置管理

**Files:**
- Create: `shared/pkg/config/config.go`
- Create: `shared/pkg/config/loader.go`
- Create: `shared/pkg/config/validator.go`
- Create: `shared/pkg/config/config_test.go`

- [ ] **Step 1: 编写配置测试**

```go
package config_test

import (
    "os"
    "testing"
    "github.com/quality-gateway/shared/pkg/config"
)

func TestLoadFromEnv(t *testing.T) {
    // 设置环境变量
    os.Setenv("SERVER_PORT", "8080")
    os.Setenv("DB_HOST", "localhost")
    os.Setenv("DB_PORT", "3306")
    defer func() {
        os.Unsetenv("SERVER_PORT")
        os.Unsetenv("DB_HOST")
        os.Unsetenv("DB_PORT")
    }()

    cfg := &config.ServerConfig{}
    err := config.LoadFromEnv(cfg)
    if err != nil {
        t.Fatalf("LoadFromEnv() error = %v", err)
    }

    // 验证配置已加载
    if cfg.Port != 8080 {
        t.Errorf("expected Port = 8080, got %d", cfg.Port)
    }
}

func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        cfg     *config.DatabaseConfig
        wantErr bool
    }{
        {
            name: "valid config",
            cfg: &config.DatabaseConfig{
                Host: "localhost",
                Port: 3306,
                Database: "test",
            },
            wantErr: false,
        },
        {
            name: "missing host",
            cfg: &config.DatabaseConfig{
                Port: 3306,
                Database: "test",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := config.Validate(tt.cfg)
            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

- [ ] **Step 2: 实现配置结构**

```go
package config

import (
    "fmt"
    "strconv"
)

// ServerConfig 服务器配置
type ServerConfig struct {
    Host            string `json:"host" mapstructure:"host" env:"SERVER_HOST"`
    Port            int    `json:"port" mapstructure:"port" env:"SERVER_PORT" validate:"min=1,max=65535"`
    ReadTimeout     int    `json:"read_timeout" mapstructure:"read_timeout" env:"SERVER_READ_TIMEOUT"`
    WriteTimeout    int    `json:"write_timeout" mapstructure:"write_timeout" env:"SERVER_WRITE_TIMEOUT"`
    ShutdownTimeout int    `json:"shutdown_timeout" mapstructure:"shutdown_timeout" env:"SERVER_SHUTDOWN_TIMEOUT"`
    AllowOrigins    string `json:"allow_origins" mapstructure:"allow_origins" env:"SERVER_ALLOW_ORIGINS"`
}

// DefaultServerConfig 返回默认服务器配置
func DefaultServerConfig() *ServerConfig {
    return &ServerConfig{
        Host:            "0.0.0.0",
        Port:            8080,
        ReadTimeout:     30,
        WriteTimeout:    30,
        ShutdownTimeout: 10,
        AllowOrigins:    "*",
    }
}

// GetAddr 返回服务器地址
func (c *ServerConfig) GetAddr() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
    Driver       string `json:"driver" mapstructure:"driver" env:"DB_DRIVER" validate:"required"`
    Host         string `json:"host" mapstructure:"host" env:"DB_HOST" validate:"required"`
    Port         int    `json:"port" mapstructure:"port" env:"DB_PORT" validate:"min=1,max=65535"`
    Username     string `json:"username" mapstructure:"username" env:"DB_USERNAME"`
    Password     string `json:"password" mapstructure:"password" env:"DB_PASSWORD"`
    Database     string `json:"database" mapstructure:"database" env:"DB_DATABASE" validate:"required"`
    MaxOpenConn  int    `json:"max_open_conn" mapstructure:"max_open_conn" env:"DB_MAX_OPEN_CONN"`
    MaxIdleConn  int    `json:"max_idle_conn" mapstructure:"max_idle_conn" env:"DB_MAX_IDLE_CONN"`
    MaxLifetime  int    `json:"max_lifetime" mapstructure:"max_lifetime" env:"DB_MAX_LIFETIME"`
}

// DefaultDatabaseConfig 返回默认数据库配置
func DefaultDatabaseConfig() *DatabaseConfig {
    return &DatabaseConfig{
        Driver:      "mysql",
        Host:        "localhost",
        Port:        3306,
        MaxOpenConn: 25,
        MaxIdleConn: 5,
        MaxLifetime: 300,
    }
}

// GetDSN 生成数据源名称
func (c *DatabaseConfig) GetDSN() string {
    switch c.Driver {
    case "mysql":
        return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
            c.Username, c.Password, c.Host, c.Port, c.Database)
    case "sqlite":
        return c.Database
    default:
        return ""
    }
}

// LoggingConfig 日志配置
type LoggingConfig struct {
    Level  string `json:"level" mapstructure:"level" env:"LOG_LEVEL" validate:"oneof=debug info warn error"`
    Format string `json:"format" mapstructure:"format" env:"LOG_FORMAT" validate:"oneof=json text"`
}

// DefaultLoggingConfig 返回默认日志配置
func DefaultLoggingConfig() *LoggingConfig {
    return &LoggingConfig{
        Level:  "info",
        Format: "json",
    }
}

// Config 总配置
type Config struct {
    Server   *ServerConfig   `json:"server" mapstructure:"server"`
    Database *DatabaseConfig `json:"database" mapstructure:"database"`
    Logging  *LoggingConfig  `json:"logging" mapstructure:"logging"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
    return &Config{
        Server:   DefaultServerConfig(),
        Database: DefaultDatabaseConfig(),
        Logging:  DefaultLoggingConfig(),
    }
}
```

- [ ] **Step 3: 实现配置加载器**

```go
package config

import (
    "fmt"
    "os"
    "reflect"
    "strings"

    "github.com/spf13/viper"
)

// Load 从文件加载配置
func Load(configPath string) (*Config, error) {
    v := viper.New()

    // 设置配置文件
    v.SetConfigFile(configPath)
    v.SetConfigType("yaml")

    // 读取配置文件
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("read config: %w", err)
    }

    cfg := DefaultConfig()
    if err := v.Unmarshal(cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }

    return cfg, nil
}

// LoadFromEnv 从环境变量加载配置
func LoadFromEnv(cfg any) error {
    v := viper.New()
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // 通过反射设置环境变量到结构体
    return loadEnvFromStruct(cfg)
}

// loadEnvFromStruct 通过反射从环境变量加载配置
func loadEnvFromStruct(cfg any) error {
    val := reflect.ValueOf(cfg)
    if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
        return fmt.Errorf("cfg must be a pointer to struct")
    }

    elem := val.Elem()
    typ := elem.Type()

    for i := 0; i < elem.NumField(); i++ {
        field := elem.Field(i)
        fieldType := typ.Field(i)

        // 跳过不可设置的字段
        if !field.CanSet() {
            continue
        }

        // 获取 env tag
        envTag := fieldType.Tag.Get("env")
        if envTag == "" {
            continue
        }

        // 从环境变量读取
        envValue := os.Getenv(envTag)
        if envValue == "" {
            continue
        }

        // 根据字段类型设置值
        switch field.Kind() {
        case reflect.String:
            field.SetString(envValue)
        case reflect.Int:
            intVal, err := strconv.ParseInt(envValue, 10, 64)
            if err != nil {
                return fmt.Errorf("parse %s as int: %w", envTag, err)
            }
            field.SetInt(intVal)
        case reflect.Bool:
            boolVal := envValue == "true" || envValue == "1"
            field.SetBool(boolVal)
        }
    }

    return nil
}
```

- [ ] **Step 4: 实现配置验证器**

```go
package config

import (
    "fmt"
    "strings"

    "github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Validate 验证配置
func Validate(cfg any) error {
    if err := validate.Struct(cfg); err != nil {
        return formatValidationErrors(err)
    }
    return nil
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(err error) error {
    validationErrors, ok := err.(validator.ValidationErrors)
    if !ok {
        return err
    }

    var msgs []string
    for _, e := range validationErrors {
        msg := fmt.Sprintf("%s: failed '%s' validation", e.Field(), e.Tag())
        msgs = append(msgs, msg)
    }

    return fmt.Errorf("validation failed: %s", strings.Join(msgs, "; "))
}
```

- [ ] **Step 5: 添加测试依赖**

```bash
cd shared
go get github.com/spf13/viper
go get github.com/go-playground/validator/v10
go mod tidy
```

- [ ] **Step 6: 运行测试验证通过**

```bash
cd shared
go test ./pkg/config/... -v
```

Expected: PASS

- [ ] **Step 7: 提交**

```bash
git add shared/pkg/config/
git commit -m "feat(shared): 添加配置管理"
```

---

## Task 7: 创建 pkg/testing 测试工具

**Files:**
- Create: `shared/pkg/testing/mock.go`
- Create: `shared/pkg/testing/fixtures.go`
- Create: `shared/pkg/testing/testutil.go`

- [ ] **Step 1: 实现 Mock 服务器**

```go
package testing

import (
    "net/http"
    "net/http/httptest"
    "sync"
)

// MockServer Mock HTTP 服务器
type MockServer struct {
    *httptest.Server
    mu       sync.RWMutex
    handlers map[string]http.HandlerFunc
    requests map[string][]*http.Request
}

// NewMockServer 创建 Mock 服务器
func NewMockServer() *MockServer {
    ms := &MockServer{
        handlers: make(map[string]http.HandlerFunc),
        requests: make(map[string][]*http.Request),
    }

    ms.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ms.mu.RLock()
        handler, ok := ms.handlers[r.URL.Path]
        ms.mu.RUnlock()

        // 记录请求
        ms.mu.Lock()
        ms.requests[r.URL.Path] = append(ms.requests[r.URL.Path], r)
        ms.mu.Unlock()

        if !ok {
            http.NotFound(w, r)
            return
        }

        handler(w, r)
    }))

    return ms
}

// RegisterHandler 注册处理器
func (m *MockServer) RegisterHandler(pattern string, handler http.HandlerFunc) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.handlers[pattern] = handler
}

// HandleJSON 返回 JSON 响应
func (m *MockServer) HandleJSON(pattern string, status int, data any) {
    m.RegisterHandler(pattern, func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        // 简化实现，实际应该用 json.Encoder
    })
}

// GetRequests 获取请求记录
func (m *MockServer) GetRequests(pattern string) []*http.Request {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.requests[pattern]
}

// ClearRequests 清空请求记录
func (m *MockServer) ClearRequests(pattern string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    if pattern == "" {
        m.requests = make(map[string][]*http.Request)
    } else {
        m.requests[pattern] = nil
    }
}

// Reset 重置所有状态
func (m *MockServer) Reset() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.handlers = make(map[string]http.HandlerFunc)
    m.requests = make(map[string][]*http.Request)
}
```

- [ ] **Step 2: 实现测试数据生成器**

```go
package testing

import (
    "time"

    "github.com/quality-gateway/shared/pkg/models"
)

// NewEvent 创建测试事件
func NewEvent(overrides ...func(*models.Event)) *models.Event {
    event := &models.Event{
        ID:          1,
        EventID:     "test-event-123",
        EventType:   models.EventTypePullRequestOpened,
        PRNumber:    123,
        PRTitle:     "Test PR",
        SourceBranch: "feature/test",
        TargetBranch: "main",
        RepoURL:     "https://github.com/test/repo",
        Sender:      "test-user",
        EventStatus: models.EventStatusPending,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    for _, override := range overrides {
        override(event)
    }

    return event
}

// NewTask 创建测试任务
func NewTask(overrides ...func(*models.Task)) *models.Task {
    task := &models.Task{
        ID:         1,
        EventID:    "test-event-123",
        TaskType:   "basic_ci_all",
        TaskStatus: models.TaskStatusPending,
        ResourceID: 1,
        Analyzing:  false,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
    }

    for _, override := range overrides {
        override(task)
    }

    return task
}

// NewResource 创建测试资源
func NewResource(overrides ...func(*models.Resource)) *models.Resource {
    resource := &models.Resource{
        ID:           1,
        UUID:         "test-resource-uuid",
        ResourceType: models.ResourceTypeBasicCI,
        Name:         "Test Resource",
        Description:  "Test resource description",
        AllowSkip:    false,
        Organization: "test-org",
        Project:      "test-project",
        PipelineID:   123,
        IsPublic:     true,
        CreatorID:    1,
        CreatedAt:    time.Now(),
        UpdatedAt:    time.Now(),
    }

    for _, override := range overrides {
        override(resource)
    }

    return resource
}

// NewUser 创建测试用户
func NewUser(overrides ...func(*models.User)) *models.User {
    user := &models.User{
        ID:        1,
        Username:  "testuser",
        Email:     "test@example.com",
        FullName:  "Test User",
        CreatedAt: time.Now(),
    }

    for _, override := range overrides {
        override(user)
    }

    return user
}
```

- [ ] **Step 3: 实现测试辅助函数**

```go
package testing

import (
    "context"
    "database/sql"
    "fmt"
    "testing"
    "time"

    "github.com/quality-gateway/shared/pkg/storage"
)

// SetupTestDB 创建测试数据库
func SetupTestDB(t *testing.T) *storage.DB {
    db, err := storage.NewDB(&storage.Config{
        Driver: "sqlite",
        DSN:    ":memory:",
    })
    if err != nil {
        t.Fatalf("setup test db: %v", err)
    }

    // 创建测试表
    _, err = db.ExecContext(context.Background(), `
        CREATE TABLE IF NOT EXISTS test (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL
        )
    `)
    if err != nil {
        db.Close()
        t.Fatalf("create test table: %v", err)
    }

    return db
}

// TeardownTestDB 清理测试数据库
func TeardownTestDB(t *testing.T, db *storage.DB) {
    if err := db.Close(); err != nil {
        t.Errorf("close test db: %v", err)
    }
}

// WaitForCondition 等待条件满足
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration) {
    deadline := time.Now().Add(timeout)
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    for {
        if time.Now().After(deadline) {
            t.Fatal("condition not met within timeout")
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
func AssertTrue(t *testing.T, condition bool) {
    t.Helper()
    if !condition {
        t.Error("expected true, got false")
    }
}

// AssertFalse 断言为假
func AssertFalse(t *testing.T, condition bool) {
    t.Helper()
    if condition {
        t.Error("expected false, got true")
    }
}

// TruncateTable 清空表
func TruncateTable(t *testing.T, db *storage.DB, table string) {
    _, err := db.ExecContext(context.Background(), fmt.Sprintf("DELETE FROM %s", table))
    if err != nil {
        t.Fatalf("truncate table %s: %v", table, err)
    }
}
```

- [ ] **Step 4: 运行测试验证**

```bash
cd shared
go test ./pkg/testing/... -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add shared/pkg/testing/
git commit -m "feat(shared): 添加测试工具"
```

---

## Task 8: 创建 pkg/api REST API 框架

**Files:**
- Create: `shared/pkg/api/server.go`
- Create: `shared/pkg/api/middleware.go`
- Create: `shared/pkg/api/response.go`
- Create: `shared/pkg/api/client.go`
- Create: `shared/pkg/api/server_test.go`

- [ ] **Step 1: 编写 API 测试**

```go
package api_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/quality-gateway/shared/pkg/api"
)

func TestServer(t *testing.T) {
    srv := api.NewServer(&api.ServerConfig{
        Addr: ":0", // 随机端口
    })

    srv.GET("/health", func(w http.ResponseWriter, r *http.Request) {
        api.SendJSON(w, api.SuccessResponse(map[string]string{"status": "ok"}))
    })

    // 测试 GET 请求
    req := httptest.NewRequest("GET", "/health", nil)
    w := httptest.NewRecorder()
    srv.Handler().ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}

func TestMiddlewareAuth(t *testing.T) {
    token := "test-token-123"

    srv := api.NewServer(&api.ServerConfig{
        Addr: ":0",
    })

    srv.Use(api.AuthMiddleware(token))
    srv.GET("/protected", func(w http.ResponseWriter, r *http.Request) {
        api.SendJSON(w, api.SuccessResponse(nil))
    })

    // 测试无 token
    req := httptest.NewRequest("GET", "/protected", nil)
    w := httptest.NewRecorder()
    srv.Handler().ServeHTTP(w, req)

    if w.Code != http.StatusUnauthorized {
        t.Errorf("expected status 401, got %d", w.Code)
    }

    // 测试有 token
    req = httptest.NewRequest("GET", "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    w = httptest.NewRecorder()
    srv.Handler().ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}
```

- [ ] **Step 2: 实现响应格式**

```go
package api

import "encoding/json"

// Response 统一响应格式
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// SuccessResponse 成功响应
func SuccessResponse(data interface{}) *Response {
    return &Response{
        Success: true,
        Data:    data,
    }
}

// ErrorResponse 错误响应
func ErrorResponse(code, message string) *Response {
    return &Response{
        Success: false,
        Error: &ErrorInfo{
            Code:    code,
            Message: message,
        },
    }
}

// SendJSON 发送 JSON 响应
func SendJSON(w http.ResponseWriter, resp *Response) {
    w.Header().Set("Content-Type", "application/json")

    if !resp.Success {
        switch resp.Error.Code {
        case "ERR_NOT_FOUND":
            w.WriteHeader(http.StatusNotFound)
        case "ERR_INVALID_INPUT":
            w.WriteHeader(http.StatusBadRequest)
        case "ERR_UNAUTHORIZED":
            w.WriteHeader(http.StatusUnauthorized)
        default:
            w.WriteHeader(http.StatusInternalServerError)
        }
    }

    json.NewEncoder(w).Encode(resp)
}
```

- [ ] **Step 3: 实现服务器**

```go
package api

import (
    "context"
    "fmt"
    "net/http"
    "sync"
    "time"

    "github.com/quality-gateway/shared/pkg/logger"
)

// ServerConfig 服务器配置
type ServerConfig struct {
    Addr            string
    ReadTimeout     time.Duration
    WriteTimeout    time.Duration
    ShutdownTimeout time.Duration
}

// Server HTTP 服务器
type Server struct {
    config     *ServerConfig
    router     *Router
    logger     *logger.Logger
    server     *http.Server
    mu         sync.RWMutex
    middleware []Middleware
}

// NewServer 创建服务器
func NewServer(cfg *ServerConfig) *Server {
    if cfg == nil {
        cfg = &ServerConfig{
            Addr:            ":8080",
            ReadTimeout:     30 * time.Second,
            WriteTimeout:    30 * time.Second,
            ShutdownTimeout: 10 * time.Second,
        }
    }

    return &Server{
        config:     cfg,
        router:     NewRouter(),
        logger:     logger.New(&logger.Config{Level: "info", Format: "text"}),
        middleware: []Middleware{},
    }
}

// Use 添加中间件
func (s *Server) Use(mw Middleware) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.middleware = append(s.middleware, mw)
}

// GET 注册 GET 路由
func (s *Server) GET(path string, handler http.HandlerFunc) {
    s.router.Add("GET", path, handler)
}

// POST 注册 POST 路由
func (s *Server) POST(path string, handler http.HandlerFunc) {
    s.router.Add("POST", path, handler)
}

// PUT 注册 PUT 路由
func (s *Server) PUT(path string, handler http.HandlerFunc) {
    s.router.Add("PUT", path, handler)
}

// DELETE 注册 DELETE 路由
func (s *Server) DELETE(path string, handler http.HandlerFunc) {
    s.router.Add("DELETE", path, handler)
}

// Handler 返回 HTTP 处理器
func (s *Server) Handler() http.Handler {
    var handler http.Handler = s.router

    // 应用中间件（反向顺序）
    for i := len(s.middleware) - 1; i >= 0; i-- {
        handler = s.middleware[i](handler)
    }

    return handler
}

// Start 启动服务器
func (s *Server) Start() error {
    s.server = &http.Server{
        Addr:         s.config.Addr,
        Handler:      s.Handler(),
        ReadTimeout:  s.config.ReadTimeout,
        WriteTimeout: s.config.WriteTimeout,
    }

    s.logger.Info("server starting", "addr", s.config.Addr)
    return s.server.ListenAndServe()
}

// Shutdown 优雅关闭
func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info("server shutting down")
    return s.server.Shutdown(ctx)
}
```

- [ ] **Step 4: 实现路由**

```go
package api

import (
    "net/http"
    "strings"
)

// Router 路由器
type Router struct {
    routes map[string]map[string]http.HandlerFunc
}

// NewRouter 创建路由器
func NewRouter() *Router {
    return &Router{
        routes: make(map[string]map[string]http.HandlerFunc),
    }
}

// Add 添加路由
func (r *Router) Add(method, path string, handler http.HandlerFunc) {
    if r.routes[method] == nil {
        r.routes[method] = make(map[string]http.HandlerFunc)
    }
    r.routes[method][path] = handler
}

// GET 注册 GET 路由
func (r *Router) GET(path string, handler http.HandlerFunc) {
    r.Add("GET", path, handler)
}

// POST 注册 POST 路由
func (r *Router) POST(path string, handler http.HandlerFunc) {
    r.Add("POST", path, handler)
}

// PUT 注册 PUT 路由
func (r *Router) PUT(path string, handler http.HandlerFunc) {
    r.Add("PUT", path, handler)
}

// DELETE 注册 DELETE 路由
func (r *Router) DELETE(path string, handler http.HandlerFunc) {
    r.Add("DELETE", path, handler)
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // 简化实现，支持固定路径匹配
    // 实际实现应该支持路径参数
    methodRoutes := r.routes[req.Method]
    if methodRoutes == nil {
        http.NotFound(w, req)
        return
    }

    // 精确匹配
    if handler, ok := methodRoutes[req.URL.Path]; ok {
        handler(w, req)
        return
    }

    // 前缀匹配（用于 /api/ 等前缀）
    for path, handler := range methodRoutes {
        if strings.HasPrefix(req.URL.Path, path) {
            handler(w, req)
            return
        }
    }

    http.NotFound(w, req)
}
```

- [ ] **Step 5: 实现中间件**

```go
package api

import (
    "net/http"
    "strings"

    "github.com/quality-gateway/shared/pkg/logger"
)

// Middleware 中间件类型
type Middleware func(http.Handler) http.Handler

// AuthMiddleware 认证中间件
func AuthMiddleware(token string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                SendJSON(w, ErrorResponse("ERR_UNAUTHORIZED", "missing authorization header"))
                return
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" || parts[1] != token {
                SendJSON(w, ErrorResponse("ERR_UNAUTHORIZED", "invalid token"))
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(log *logger.Logger) Middleware {
    return log.Middleware
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware(log *logger.Logger) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    log.Error("panic recovered", "error", err)
                    SendJSON(w, ErrorResponse("ERR_INTERNAL", "internal server error"))
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}

// CORSMiddleware CORS 中间件
func CORSMiddleware(allowedOrigins []string) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            // 检查 origin 是否允许
            allowed := false
            for _, ao := range allowedOrigins {
                if ao == "*" || ao == origin {
                    allowed = true
                    break
                }
            }

            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Access-Control-Max-Age", "86400")
            }

            // 处理 OPTIONS 预检请求
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

- [ ] **Step 6: 实现服务间 HTTP 客户端**

```go
package api

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/quality-gateway/shared/pkg/logger"
)

// ClientConfig 客户端配置
type ClientConfig struct {
    BaseURL    string
    Token      string
    Timeout    time.Duration
    MaxRetries int
}

// Client HTTP 客户端
type Client struct {
    config     *ClientConfig
    httpClient *http.Client
    logger     *logger.Logger
}

// NewClient 创建客户端
func NewClient(cfg *ClientConfig) *Client {
    if cfg == nil {
        cfg = &ClientConfig{
            Timeout:    30 * time.Second,
            MaxRetries: 3,
        }
    }

    return &Client{
        config: cfg,
        httpClient: &http.Client{
            Timeout: cfg.Timeout,
        },
        logger: logger.New(&logger.Config{Level: "info", Format: "text"}),
    }
}

// Do 执行 HTTP 请求
func (c *Client) Do(req *http.Request) (*Response, error) {
    // 添加认证头
    if c.config.Token != "" {
        req.Header.Set("Authorization", "Bearer "+c.config.Token)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Request-ID", fmt.Sprintf("%d", time.Now().UnixNano()))

    // 执行请求（带重试）
    var resp *http.Response
    var err error

    for i := 0; i <= c.config.MaxRetries; i++ {
        resp, err = c.httpClient.Do(req)
        if err == nil {
            break
        }
        if i < c.config.MaxRetries {
            time.Sleep(time.Duration(i+1) * time.Second)
        }
    }

    if err != nil {
        return nil, fmt.Errorf("http request failed: %w", err)
    }
    defer resp.Body.Close()

    // 解析响应
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read response body: %w", err)
    }

    var apiResp Response
    if err := json.Unmarshal(body, &apiResp); err != nil {
        return nil, fmt.Errorf("unmarshal response: %w", err)
    }

    return &apiResp, nil
}

// Get 执行 GET 请求
func (c *Client) Get(path string) (*Response, error) {
    url := c.config.BaseURL + path
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    return c.Do(req)
}

// Post 执行 POST 请求
func (c *Client) Post(path string, data interface{}) (*Response, error) {
    url := c.config.BaseURL + path
    body, err := json.Marshal(data)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    return c.Do(req)
}

// Put 执行 PUT 请求
func (c *Client) Put(path string, data interface{}) (*Response, error) {
    url := c.config.BaseURL + path
    body, err := json.Marshal(data)
    if err != nil {
        return nil, fmt.Errorf("marshal request: %w", err)
    }

    req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    return c.Do(req)
}

// Delete 执行 DELETE 请求
func (c *Client) Delete(path string) (*Response, error) {
    url := c.config.BaseURL + path
    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }
    return c.Do(req)
}
```

- [ ] **Step 7: 运行测试验证通过**

```bash
cd shared
go test ./pkg/api/... -v -cover
```

Expected: PASS, 覆盖率 > 70%

- [ ] **Step 8: 提交**

```bash
git add shared/pkg/api/
git commit -m "feat(shared): 添加 REST API 框架"
```

---

## Task 9: 创建构建脚本

**Files:**
- Create: `shared/scripts/build.sh`
- Create: `shared/scripts/test.sh`

- [ ] **Step 1: 创建构建脚本**

```bash
#!/bin/bash
# shared/scripts/build.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHARED_DIR="$(dirname "$SCRIPT_DIR")"

echo "Building shared library..."

cd "$SHARED_DIR"

# 格式化代码
echo "Formatting code..."
go fmt ./pkg/...

# 检查
echo "Running go vet..."
go vet ./pkg/...

# 构建
echo "Building..."
go build ./pkg/...

echo "✓ Build complete"
```

- [ ] **Step 2: 创建测试脚本**

```bash
#!/bin/bash
# shared/scripts/test.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHARED_DIR="$(dirname "$SCRIPT_DIR")"

cd "$SHARED_DIR"

echo "Running tests..."

# 运行所有测试
go test -v -race -coverprofile=coverage.out ./pkg/...

# 显示覆盖率
echo ""
echo "Coverage report:"
go tool cover -func=coverage.out | tail -1

echo "✓ Tests complete"
```

- [ ] **Step 3: 添加执行权限**

```bash
chmod +x shared/scripts/build.sh
chmod +x shared/scripts/test.sh
```

- [ ] **Step 4: 验证脚本**

```bash
cd shared
./scripts/build.sh
./scripts/test.sh
```

Expected: 构建和测试都成功

- [ ] **Step 5: 提交**

```bash
git add shared/scripts/
git commit -m "feat(shared): 添加构建和测试脚本"
```

---

## Task 10: 更新 README 和文档

**Files:**
- Modify: `shared/README.md`

- [ ] **Step 1: 更新 README**

```markdown
# Shared Library

双引擎质量网关共享库，为所有微服务提供统一的基础设施。

## 包说明

### pkg/api - REST API 框架

统一的 HTTP 服务器和客户端实现。

\`\`\`go
import "github.com/quality-gateway/shared/pkg/api"

srv := api.NewServer(&api.ServerConfig{Addr: ":4001"})
srv.GET("/health", func(w http.ResponseWriter, r *http.Request) {
    api.SendJSON(w, api.SuccessResponse(map[string]string{"status": "ok"}))
})
srv.Start()
\`\`\`

### pkg/storage - 数据库抽象层

统一的数据库操作接口，支持 MySQL 和 SQLite。

\`\`\`go
import "github.com/quality-gateway/shared/pkg/storage"

db, err := storage.NewDB(&storage.Config{
    Driver: "mysql",
    DSN:    "user:pass@tcp(localhost:3306)/db",
})
err = db.Transaction(func(tx *storage.Tx) error {
    _, err := tx.ExecContext(ctx, "INSERT INTO ...")
    return err
})
\`\`\`

### pkg/models - 共享数据模型

事件、任务、资源等数据模型定义。

### pkg/testing - 测试工具

Mock 服务器、测试数据生成器、测试辅助函数。

\`\`\`go
import "github.com/quality-gateway/shared/pkg/testing"

event := testing.NewEvent()
task := testing.NewTask(func(t *testing.T) {
    t.TaskType = "custom_type"
})
db := testing.SetupTestDB(t)
defer testing.TeardownTestDB(t, db)
\`\`\`

### pkg/config - 配置管理

配置加载和验证，支持 YAML 文件和环境变量。

### pkg/logger - 日志组件

结构化日志，支持 JSON 和文本格式。

\`\`\`go
import "github.com/quality-gateway/shared/pkg/logger"

log := logger.New(&logger.Config{Level: "info", Format: "json"})
log.Info("server starting", "port", 8080)
log.Error("request failed", logger.Error(err))
\`\`\`

### pkg/errors - 错误处理

统一的错误定义和处理。

## 开发

### 构建

\`\`\`bash
./scripts/build.sh
\`\`\`

### 测试

\`\`\`bash
./scripts/test.sh
\`\`\`

## 使用

在服务 go.mod 中引用：

\`\`\`go
require github.com/quality-gateway/shared v1.0.0
\`\`\`

然后 import 使用：

\`\`\`go
import (
    "github.com/quality-gateway/shared/pkg/api"
    "github.com/quality-gateway/shared/pkg/storage"
    "github.com/quality-gateway/shared/pkg/logger"
)
\`\`\`
```

- [ ] **Step 2: 运行最终测试验证**

```bash
cd shared
go test ./... -v -race
./scripts/build.sh
./scripts/test.sh
```

Expected: 全部通过

- [ ] **Step 3: 提交**

```bash
git add shared/README.md
git commit -m "docs(shared): 更新 README 文档"
```

---

## 验收标准

完成所有任务后，验证以下标准：

- [ ] 所有单元测试通过（`go test ./... -v`）
- [ ] 测试覆盖率 > 70%（`go test -cover ./...`）
- [ ] 无竞态条件（`go test -race ./...`）
- [ ] 代码已格式化（`go fmt ./...`）
- [ ] 无 vet 警告（`go vet ./...`）
- [ ] 文档完整（README.md）

---

**计划完成**
