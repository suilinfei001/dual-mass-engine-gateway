# 双引擎质量网关系统规则

## 项目概述

双引擎质量网关系统，包含两个独立模块：
- **事件接收器 (Event Receiver)** - 部署在外网 (10.4.111.141)，接收 GitHub Webhook 事件
- **事件处理器 (Event Processor)** - 部署在内网 (10.4.174.125)，处理质量检查任务

## 开发规范

### 重要规则

**事件接收器已部署运行**
- 部署地址: `http://10.4.111.141:5001`
- event-processor 需要请求 event-receiver 相关的接口时，直接访问 `http://10.4.111.141:5001`
- 每次回答我，先说 老板 你好，再回答问题问题

### 测试规范（重要）

**在修改代码后，必须先运行测试用例确保测试通过，再进行部署。**

- 所有新增功能必须配套编写测试用例
- 修改现有代码后，必须运行相关测试确保没有引入回归问题
- 测试文件命名规范：`xxx_test.go`
- 运行测试命令：`go test ./...`

### Bug 追踪规范（重要）

**每次对话开始前，必须先阅读根目录的 `bug.txt` 文件，优先处理未解决的 Bug。**

- Bug 文件位置: `/root/dev/dual-mass-engine-gateway/bug.txt`
- Bug 格式: `bug_id:bug_description:is_fixed:root_cause`
  - `bug_id`: Bug 唯一标识
  - `bug_description`: Bug 描述
  - `is_fixed`: 是否已修复 (true/false)
  - `root_cause`: 根因分析

**规则**:
1. 每次对话前先 review `bug.txt`，检查是否有未修复的 Bug (`is_fixed=false`)
2. 如果有未修复 Bug，必须优先修复后再处理新需求
3. 新发现的 Bug 立即记录到 `bug.txt`
4. Bug 修复后更新 `is_fixed=true` 并注明修复方式

**示例**:
```
bug_001:mockTaskStorage缺少UpdateTaskAnalyzing方法:true:接口新增但mock未实现
bug_002:用户注册邮箱重复检查逻辑错误:true:测试用例假设错误，实际需先创建用户
```

### 部署规则
#### Event Receiver 部署（远程服务器 10.4.111.141）

Event Receiver 部署在外网服务器，与 Event Processor 分离部署。

**部署方式**：在本地 (10.4.174.125) 构建镜像，通过 SSH 在远程服务器 (10.4.111.141) 部署

```bash
# 使用统一更新脚本（推荐）
cd /root/dev/dual-mass-engine-gateway/src/modules/event-receiver
./update.sh              # 完整更新（构建+部署）
./update.sh -b          # 仅构建
./update.sh -d          # 仅部署
./update.sh -r          # 恢复模式部署
```

**访问地址**：
- Frontend: http://10.4.111.141:8081
- Backend API: http://10.4.111.141:5001
#### event-processor 代码更新后的部署流程

当修改了 event-processor 模块的代码后，**必须先运行测试确保通过，然后才能部署**。

#### 部署命令

```bash
cd /root/dev/dual-mass-engine-gateway/src/modules/event-processor
./deploy-event-processor.sh
```

#### 部署脚本说明

部署脚本会自动执行以下步骤：
1. **编译 Go 代码** - 使用静态链接编译后端二进制文件
2. **构建 Docker 镜像** - 构建前端和后端镜像
3. **创建 Docker 网络** - 确保容器网络正确配置
4. **启动容器** - 启动 MySQL、后端和前端服务

- **脚本位置**: `/root/dev/dual-mass-engine-gateway/src/modules/event-processor/deploy-event-processor.sh`
- **默认模式**: 升级模式（保留数据，更新容器）
- **完全重装**: `./deploy-event-processor.sh -r`（会删除所有容器和数据）

#### 服务访问地址

- **Backend API**: http://localhost:5003
- **Frontend**: http://localhost:8082
- **MySQL**: localhost:3307

### 代码结构

```
src/modules/
├── event-receiver/     # 事件接收器（已部署，无需本地编译）
│   ├── cmd/quality-server/      # 后端服务入口
│   ├── internal/quality/        # 核心业务逻辑
│   │   ├── api/                 # REST API 处理器
│   │   ├── handlers/            # Webhook 事件处理器
│   │   ├── models/              # 数据模型（event.go, time.go, enums.go）
│   │   ├── storage/             # MySQL 数据存储
│   │   └── logger/              # 日志组件
│   └── frontend/                # Vue 3 前端
│
└── event-processor/    # 事件处理器（开发重点）
    ├── cmd/server/              # 后端服务入口
    ├── internal/
    │   ├── api/                 # Event Receiver API 客户端
    │   ├── models/              # 任务模型
    │   ├── scheduler/           # 调度器（含 PR 处理逻辑）
    │   ├── executor/            # Azure DevOps 执行器
    │   ├── storage/             # 任务存储
    │   └── mock/                # Mock 测试服务器
    ├── frontend/                # Vue 3 前端
    └── test/e2e/                # 端到端测试
```

### 开发工作流

1. **事件处理器开发**
   - 修改 `src/modules/event-processor/` 下的代码
   - 调用 Event Receiver API: `http://10.4.111.141:5001/api/*`
   - 本地测试时无需启动 event-receiver

2. **API 通信**
   ```go
   // Event Processor 调用 Event Receiver 示例
   const eventReceiverAPI = "http://10.4.111.141:5001"

   // 获取所有事件
   GET http://10.4.111.141:5001/api/events

   // 获取待处理事件
   GET http://10.4.111.141:5001/api/events?status=pending

   // 更新事件状态
   POST http://10.4.111.141:5001/api/events/{id}/status

   // Mock API (测试用)
   POST http://10.4.111.141:5001/api/mock/simulate/pull_request.opened
   POST http://10.4.111.141:5001/api/mock/simulate/pull_request.synchronize
   ```

### 生产环境部署说明

- **Event Receiver**: 已部署在 `10.4.111.141:5001`
  - 使用 `./install/install_quality.sh` 部署/更新
  - 前端端口: 8081
  - 后端端口: 5001
  - MySQL 端口: 3306

- **Event Processor**: 部署在内网 `10.4.174.125`
  - 使用 `./install/install_event_processor.sh` 部署
  - 前端端口: 8082
  - 后端端口: 5002

## 关键功能实现

### PR 取消逻辑 (PR Cancel)

当 PR 发生 synchronize 事件时，需要取消之前未完成的事件：
- **实现位置**: `src/modules/event-processor/internal/scheduler/pr_handler.go`
- **关键函数**:
  - `handlePRSynchronize()`: 处理 PR 同步事件
  - `cancelEvent()`: 取消事件及其关联任务
  - `completeEvent()`: 完成事件

**取消逻辑**:
1. 获取同一 PR 的所有相关事件
2. 根据事件状态 (pending/processing) 和创建时间筛选
3. 将过期事件标记为 `cancelled`
4. 将过期的任务也标记为 `cancelled`

### 时间处理 (LocalTime)

- **实现位置**: `src/modules/event-receiver/internal/quality/models/time.go`
- **时区**: Asia/Shanghai (UTC+8)
- **支持格式**: RFC3339, 自定义格式
- **关键函数**:
  - `FromTime(t time.Time) LocalTime`: 从标准 time.Time 创建 LocalTime
  - `MarshalJSON/UnmarshalJSON`: JSON 序列化支持
  - `Value/Scan`: 数据库存储支持

### MySQL 时间戳处理

- **问题**: MySQL 驱动返回 `[]uint8` 而非 `time.Time`
- **解决方案**: 在 `mysql_storage.go` 中手动解析二进制时间戳格式
- **实现位置**: `src/modules/event-receiver/internal/quality/storage/mysql_storage.go`

## 测试

### 端到端测试

- **测试目录**: `src/modules/event-processor/test/e2e/`
- **测试内容**: PR 取消流程验证
  1. 发送 PR opened 事件
  2. 等待 30 秒
  3. 发送 PR synchronize 事件
  4. 验证事件状态变为 `cancelled`
  5. 验证 quality_checks 状态变为 `cancelled`

### 运行测试

```bash
cd /root/dev/dual-mass-engine-gateway

# 运行所有单元测试
./tests/test.sh

# 运行 E2E 测试
cd src/modules/event-processor/test/e2e
go test -v
```

## 已知问题和限制

1. **EventStatusCancelled 解析**: `enums.go` 中的 `ParseEventStatus` 需要包含 `EventStatusCancelled` case
2. **MySQL 时间戳**: 需要手动处理 `[]uint8` 类型的时间戳数据
3. **测试覆盖**: E2E 测试框架已搭建，部分功能逻辑仍需完善

## 相关文档

- [项目结构说明](docs/PROJECT_STRUCTURE.md)
- [重构需求文档](docs/project_refactor.txt)