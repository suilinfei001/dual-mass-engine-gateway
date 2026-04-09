# Shared Library

双引擎质量网关共享库，为所有微服务提供统一的基础设施。

## 包说明

### pkg/api - REST API 框架

统一的 HTTP 服务器和客户端实现。

```go
import "github.com/quality-gateway/shared/pkg/api"

srv := api.NewServer(&api.ServerConfig{Addr: ":4001"})
srv.GET("/health", func(w http.ResponseWriter, r *http.Request) {
    api.SendJSON(w, api.SuccessResponse(map[string]string{"status": "ok"}))
})
srv.Start()
```

### pkg/storage - 数据库抽象层

统一的数据库操作接口，支持 MySQL 和 SQLite。

```go
import "github.com/quality-gateway/shared/pkg/storage"

db, err := storage.NewDB(&storage.Config{
    Driver: "mysql",
    DSN:    "user:pass@tcp(localhost:3306)/db",
})
err = db.Transaction(func(tx *storage.Tx) error {
    _, err := tx.ExecContext(ctx, "INSERT INTO ...")
    return err
})
```

### pkg/models - 共享数据模型

事件、任务、资源等数据模型定义。

### pkg/testing - 测试工具

Mock 服务器、测试数据生成器、测试辅助函数。

```go
import "github.com/quality-gateway/shared/pkg/testing"

event := testing.NewEvent()
task := testing.NewTask(func(t *testing.T) {
    t.TaskType = "custom_type"
})
db := testing.SetupTestDB(t)
defer testing.TeardownTestDB(t, db)
```

### pkg/config - 配置管理

配置加载和验证，支持 YAML 文件和环境变量。

### pkg/logger - 日志组件

结构化日志，支持 JSON 和文本格式。

```go
import "github.com/quality-gateway/shared/pkg/logger"

log := logger.New(&logger.Config{Level: "info", Format: "json"})
log.Info("server starting", "port", 8080)
log.Error("request failed", logger.Error(err))
```

### pkg/errors - 错误处理

统一的错误定义和处理。

## 开发

### 构建

```bash
./scripts/build.sh
```

### 测试

```bash
./scripts/test.sh
```

## 使用

在服务 go.mod 中引用：

```go
require github.com/quality-gateway/shared v1.0.0
```

然后 import 使用：

```go
import (
    "github.com/quality-gateway/shared/pkg/api"
    "github.com/quality-gateway/shared/pkg/storage"
    "github.com/quality-gateway/shared/pkg/logger"
)
```
