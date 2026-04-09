# 双引擎质量网关微服务化重构设计文档

**日期：** 2026-03-23
**版本：** 1.0
**作者：** Claude
**状态：** 设计中

---

## 1. 概述

### 1.1 重构目标

本重构项目旨在将现有的双引擎质量网关系统从模块化单体架构重构为微服务架构，主要解决以下问题：

1. **代码重复冗余** - 多处相似逻辑，没有统一抽象
2. **测试覆盖不足** - 缺少测试基础设施，改动容易引入 bug
3. **部署运维复杂** - 多个模块、配置分散，部署困难

### 1.2 设计原则

- **功能不减少** - 所有现有功能必须保留
- **独立部署** - 每个服务可独立部署和更新
- **职责单一** - 每个服务只负责一个明确的业务领域
- **代码复用** - 通过共享库消除重复代码

---

## 2. 架构设计

### 2.1 服务划分

系统被拆分为 6 个微服务：

| 服务 | 端口 | 职责 | 部署位置 |
|------|------|------|----------|
| Webhook Gateway | 4001 | 接收 GitHub Webhook 并转发 | 外网 |
| Event Store | 4002 | 事件存储和查询 | 外网 |
| Task Scheduler | 4003 | 任务调度和状态管理 | 内网 |
| Executor Service | 4004 | Azure DevOps Pipeline 执行 | 内网 |
| AI Analyzer | 4005 | AI 日志分析 | 内网 |
| Resource Manager | 4006 | 资源池管理 | 内网 |

### 2.2 架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                           外网环境 (10.4.111.141)                     │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                       Nginx (80/443)                         │    │
│  │    ┌─────────────┐           ┌─────────────┐                │    │
│  │    │ Webhook     │           │ Event Store │                │    │
│  │    │ Gateway     │           │ Frontend    │                │    │
│  │    │ :4001       │           │ :8081       │                │    │
│  │    └──────┬──────┘           └─────────────┘                │    │
│  │           │                       │                          │    │
│  │           │                       ▼                          │    │
│  │           │              ┌─────────────┐                    │    │
│  │           │              │ Event Store │                    │    │
│  │           │              │ API :4002   │                    │    │
│  │           │              └──────┬──────┘                    │    │
│  └───────────┼──────────────────────┼───────────────────────────┘    │
│              │                      │                                │
└──────────────┼──────────────────────┼────────────────────────────────┘
               │                      │
               ▼                      ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         内网环境 (10.4.174.125)                       │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                       Nginx (80/443)                         │    │
│  │    ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │    │
│  │    │ Task         │  │ Executor     │  │ AI           │   │    │
│  │    │ Scheduler    │  │ Service      │  │ Analyzer     │   │    │
│  │    │ :4003        │  │ :4004        │  │ :4005        │   │    │
│  │    └──────┬───────┘  └──────────────┘  └──────────────┘   │    │
│  │           │                                                 │    │
│  │    ┌──────┴───────┐  ┌──────────────┐  ┌──────────────┐   │    │
│  │    │ Resource     │  │ Task         │  │ Admin        │   │    │
│  │    │ Manager      │  │ Scheduler    │  │ Frontend     │   │    │
│  │    │ :4006        │  │ Frontend     │  │ :8082        │   │    │
│  │    └──────────────┘  └──────────────┘  └──────────────┘   │    │
│  └─────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.3 服务间通信

服务间采用 RESTful API 进行同步通信，使用共享的 API Token 进行认证。

---

## 3. 目录结构

```
dual-mass-engine-gateway/
├── services/                        # 微服务目录
│   ├── webhook-gateway/            # Webhook 接收服务
│   │   ├── cmd/server/             # 服务入口
│   │   ├── internal/               # 内部实现
│   │   ├── configs/                # 配置文件
│   │   ├── Dockerfile
│   │   ├── deploy.sh               # 独立部署脚本
│   │   └── go.mod
│   ├── event-store/                # 事件存储服务
│   ├── task-scheduler/             # 任务调度服务
│   ├── executor-service/           # 执行器服务
│   ├── ai-analyzer/                # AI 分析服务
│   └── resource-manager/           # 资源池管理服务
│
├── shared/                          # 共享库
│   ├── pkg/
│   │   ├── api/                    # REST API 框架
│   │   ├── storage/                # 数据库抽象层
│   │   ├── models/                 # 共享数据模型
│   │   ├── testing/                # 测试工具
│   │   ├── config/                 # 配置管理
│   │   ├── logger/                 # 日志组件
│   │   └── errors/                 # 错误处理
│   └── go.mod
│
├── deployments/                    # 部署配置
│   ├── docker-compose/             # Docker Compose 配置
│   ├── kubernetes/                 # K8s 配置（可选）
│   └── nginx/                      # Nginx 配置
│
├── frontend/                       # 统一前端
│   └── admin-ui/                   # 管理后台
│
├── docs/                           # 文档
├── tests/                          # 集成测试
├── CLAUDE.md
├── README.md
└── go.work                         # Go workspace
```

---

## 4. 共享库设计

### 4.1 API 框架 (shared/pkg/api/)

统一所有服务的 HTTP 层实现：

```go
// server.go - 通用 HTTP 服务器
type Server struct {
    config    *Config
    router    *Router
    logger    *logger.Logger
}

// middleware.go - 内置中间件
func AuthMiddleware(token string) Handler
func LoggingMiddleware(logger *logger.Logger) Handler
func RecoveryMiddleware(logger *logger.Logger) Handler
func CORSMiddleware(allowedOrigins []string) Handler

// response.go - 统一响应格式
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *ErrorInfo  `json:"error,omitempty"`
}
```

### 4.2 数据库抽象层 (shared/pkg/storage/)

统一数据库操作：

```go
// db.go - 数据库连接管理
type DB struct {
    *sql.DB
    logger *logger.Logger
}

// repository.go - Repository 模式基类
type Repository[T any] struct {
    db     *DB
    table  string
}
```

### 4.3 测试工具 (shared/pkg/testing/)

```go
// mock.go - Mock 生成器
type MockServer struct {
    *httptest.Server
    handlers map[string]http.HandlerFunc
}

// fixtures.go - 测试数据生成
func NewEvent(overrides ...func(*Event)) *Event
func NewTask(overrides ...func(*Task)) *Task
```

---

## 5. API 设计

所有服务遵循统一的 RESTful API 风格。

### 5.1 统一响应格式

```json
{
  "success": true,
  "data": { ... },
  "error": null
}
```

### 5.2 Event Store API (4002)

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/events | 创建新事件 |
| GET | /api/events | 获取事件列表 |
| GET | /api/events/:id | 获取事件详情 |
| PUT | /api/events/:id/status | 更新事件状态 |
| DELETE | /api/events | 删除所有事件 |
| GET | /api/events/:id/quality-checks | 获取质量检查列表 |
| PUT | /api/events/:id/quality-checks/batch | 批量更新质量检查 |

### 5.3 Task Scheduler API (4003)

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/tasks | 获取任务列表 |
| GET | /api/tasks/:id | 获取任务详情 |
| POST | /api/tasks/:id/start | 启动任务 |
| POST | /api/tasks/:id/complete | 完成任务 |
| POST | /api/tasks/:id/fail | 标记任务失败 |
| POST | /api/tasks/:id/cancel | 取消任务 |
| POST | /api/events/:event-id/cancel | 取消事件的所有任务 |

### 5.4 Executor Service API (4004)

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/execute | 执行 Pipeline |
| GET | /api/status/:buildId | 获取 Pipeline 状态 |
| GET | /api/logs/:buildId | 获取构建日志 |

### 5.5 AI Analyzer API (4005)

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/analyze | 分析单个日志 |
| POST | /api/analyze/batch | 批量分析日志 |
| POST | /api/config/pool-size | 设置请求池大小 |

### 5.6 Resource Manager API (4006)

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/resources | 获取资源列表 |
| POST | /api/resources | 创建资源 |
| PUT | /api/resources/:id | 更新资源 |
| DELETE | /api/resources/:id | 删除资源 |
| POST | /api/resources/match | 匹配最优资源 |
| GET | /api/quota-policies | 获取配额策略列表 |
| GET | /api/allocations | 获取分配历史 |
| GET | /api/testbeds | 获取 Testbed 列表 |
| POST | /api/deployments | 创建部署任务 |

---

## 6. 数据存储设计

### 6.1 数据库分布

| 服务 | 数据库 | 表 |
|------|--------|-----|
| Event Store | event_store_db | events, quality_checks |
| Task Scheduler | task_scheduler_db | tasks, task_results, task_executions |
| Resource Manager | resource_manager_db | resources, quota_policies, allocations, testbeds, deployment_tasks, categories, users |

### 6.2 Event Store 数据库

```sql
CREATE TABLE events (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    pr_number INT,
    pr_title VARCHAR(255),
    source_branch VARCHAR(100),
    target_branch VARCHAR(100),
    repo_url VARCHAR(500),
    sender VARCHAR(100),
    event_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_event_status (event_status)
);

CREATE TABLE quality_checks (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    event_id BIGINT NOT NULL,
    check_type VARCHAR(50) NOT NULL,
    check_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    stage_order INT NOT NULL,
    result JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    UNIQUE KEY unique_event_check (event_id, check_type)
);
```

---

## 7. 部署设计

### 7.1 独立部署脚本

每个服务都有独立的部署脚本：

```bash
#!/bin/bash
# services/xxx/deploy.sh
# Usage: ./deploy.sh [upgrade|reinstall|stop|status]
```

### 7.2 统一部署脚本

根目录提供一键部署/更新所有服务的脚本：

```bash
#!/bin/bash
# deployments/deploy-all.sh
```

### 7.3 Docker Compose 配置

开发环境一键启动所有服务：

```yaml
# deployments/docker-compose/docker-compose.dev.yml
version: '3.8'

services:
  webhook-gateway:
    build: ../../services/webhook-gateway
    ports: ["4001:4001"]
    # ...

  event-store:
    build: ../../services/event-store
    ports: ["4002:4002"]
    # ...

  # ... 其他服务
```

---

## 8. 迁移计划

### 8.1 迁移阶段

```
阶段 1: 准备阶段 ────────────────────────────────────────────▶
    创建新目录结构
    创建共享库
    创建部署脚本

阶段 2: 共享库迁移 ──────────────────────────────────────────▶
    pkg/api, pkg/storage, pkg/config
    pkg/logger, pkg/testing, pkg/models, pkg/errors

阶段 3: 服务迁移 (逐步) ─────────────────────────────────────▶
    3.1 Resource Manager (优先)
    3.2 Event Store
    3.3 Webhook Gateway
    3.4 Executor Service
    3.5 AI Analyzer
    3.6 Task Scheduler (最后)

阶段 4: 前端适配 ─────────────────────────────────────────────▶
    更新 API 端点
    更新前端配置

阶段 5: 测试验证 ─────────────────────────────────────────────▶
    单元测试、集成测试、E2E 测试

阶段 6: 数据迁移 ─────────────────────────────────────────────▶
    旧数据库导出、新数据库导入、数据验证

阶段 7: 切换上线 ─────────────────────────────────────────────▶ ✅
    蓝绿部署、验证、清理旧代码
```

### 8.2 服务迁移顺序

1. **Resource Manager** - 独立性强，依赖最少
2. **Event Store** - 核心数据服务
3. **Webhook Gateway** - 入口服务
4. **Executor Service** - 执行层
5. **AI Analyzer** - 分析层
6. **Task Scheduler** - 调度层，依赖最多

---

## 9. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 服务间通信失败 | 任务无法执行 | 实现重试机制和熔断器 |
| 数据迁移丢失 | 数据不可恢复 | 充分备份，迁移后验证 |
| 并发问题回归 | 任务状态混乱 | 使用 CAS 操作，充分测试 |
| 部署复杂度增加 | 运维困难 | 提供一键部署脚本 |

---

## 10. 验收标准

1. 所有现有功能正常工作
2. 每个服务可独立部署和更新
3. 单元测试覆盖率 > 70%
4. 集成测试全部通过
5. E2E 测试全部通过
6. 无已知 bug
7. 文档完整

---

## 附录

### A. 端口分配

| 服务 | 内部端口 | 外部端口 |
|------|----------|----------|
| Webhook Gateway | 4001 | - |
| Event Store | 4002 | 4002 (外网) |
| Task Scheduler | 4003 | 4003 (内网) |
| Executor Service | 4004 | - |
| AI Analyzer | 4005 | - |
| Resource Manager | 4006 | 4006 (内网) |

### B. 源码映射

| 新服务 | 源码位置 |
|--------|----------|
| webhook-gateway | src/modules/event-receiver/internal/quality/handlers/ |
| event-store | src/modules/event-receiver/internal/quality/ |
| task-scheduler | src/modules/event-processor/internal/scheduler/ |
| executor-service | src/modules/event-processor/internal/executor/ |
| ai-analyzer | src/modules/event-processor/internal/ai/ |
| resource-manager | src/modules/resource-pool/ |

---

**文档结束**
