# Resource Pool Management System - Design Document

## 1. System Overview

### 1.1 Purpose

The Resource Pool Management System manages test environments like a reservoir, providing pre-deployed environments for different testing scenarios:

- **Main Branch Regression Testing**: Less time-sensitive, can tolerate ~40min deployment time
- **PR Testing**: Needs fast feedback, cannot wait for environment provisioning
- **Pre-deployed Testbed Pool**: Maintains a pool of ready-to-use testbeds

### 1.2 Design Metaphor: Water Tank/Reservoir

```
                    ┌─────────────────────────────────────┐
                    │         Resource Pool System         │
                    │                                     │
                    │  ┌───────────────────────────────┐  │
                    │  │     Available Testbeds        │  │
                    │  │  (Ready to allocate)           │  │
                    │  └───────────────────────────────┘  │
                    │                ▼                    │
                    │  ┌───────────────────────────────┐  │
                    │  │      Allocated/In Use          │  │
                    │  │  (Assigned to tasks)           │  │
                    │  └───────────────────────────────┘  │
                    │                ▼                    │
                    │  ┌───────────────────────────────┐  │
                    │  │      Releasing/Cleaning        │  │
                    │  │  (Restoring to base snapshot)  │  │
                    │  └───────────────────────────────┘  │
                    │                ▼                    │
                    │  ┌───────────────────────────────┐  │
                    │  │         Available              │  │
                    │  └───────────────────────────────┘  │
                    └─────────────────────────────────────┘
```

### 1.3 Key Concepts

1. **Categories**: Testbeds grouped by purpose (main, release-001, release-002, etc.)
2. **Quota Management**: Percentage-based allocation across categories
3. **Lifecycle States**: Available → Allocated → In Use → Releasing → Available
4. **Snapshot Management**: Clean state restoration after use
5. **Shared Resources**:
   - **Frontend**: Integrated into event-processor frontend as new menu/pages
   - **Database**: Shared MySQL instance with new tables
   - **Authentication**: Uses event-processor's user login system
6. **Two API Sets**:
   - **Internal API**: No authentication (for event-processor backend)
   - **External API**: Uses event-processor's session-based authentication

## 2. Architecture

### 2.1 Deployment Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  10.4.174.125 (Internal Network)                 │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │               Docker Compose Stack                       │    │
│  ├─────────────────────────────────────────────────────────┤    │
│  │                                                          │    │
│  │  ┌─────────────────┐  ┌─────────────────┐               │    │
│  │  │ event-processor │  │  resource-pool  │               │    │
│  │  │   Backend       │  │   Backend       │               │    │
│  │  │   Port 5002     │  │   Port 5003     │               │    │
│  │  └────────┬────────┘  └────────┬────────┘               │    │
│  │           │                    │                         │    │
│  │           └────────────────────┼─────────────────┐       │    │
│  │                                │                 │       │    │
│  │  ┌─────────────────────────────▼─────────────────▼───┐  │    │
│  │  │          Shared Frontend (Vue 3)                  │  │    │
│  │  │          Port 8082                                 │  │    │
│  │  │  - Event Processor Pages                           │  │    │
│  │  │  - Resource Pool Pages (New Menu)                  │  │    │
│  │  │  - Shared Login/Auth                               │  │    │
│  │  └────────────────────────────────────────────────────┘  │    │
│  │                                │                         │    │
│  │  ┌─────────────────────────────▼───────────────────────┐  │    │
│  │  │          Shared MySQL Database                      │  │    │
│  │  │          Port 3306                                  │  │    │
│  │  │  - event_processor tables                           │  │    │
│  │  │  - resource_pool tables (New)                       │  │    │
│  │  └────────────────────────────────────────────────────┘  │    │
│  │                                                          │    │
│  └───────────────────────────────────────────────────────────┘  │
│                                │                                  │
│                       Internal API Call                          │
│              (No Auth, event-processor → resource-pool)          │
│                                │                                  │
│                       External API Call                          │
│              (Frontend → resource-pool, with session cookie)    │
│                                │                                  │
└─────────────────────────────────────────────────────────────────┘
                                  │
                                  │ GitHub Webhook
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                     10.4.111.141 (External Network)              │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐                                            │
│  │ event-receiver  │                                            │
│  │   Port 5001     │                                            │
│  │   Frontend 8081 │                                            │
│  └─────────────────┘                                            │
└─────────────────────────────────────────────────────────────────┘
```

### 2.2 Module Structure

```
src/modules/resource-pool/                    # 独立后端模块
├── cmd/
│   └── pool-server/
│       └── main.go              # Backend entry point (Port 5003)
├── internal/
│   ├── api/
│   │   ├── server.go            # HTTP server & router
│   │   ├── internal_handler.go  # Internal API (no auth)
│   │   └── admin_handler.go     # Admin management API
│   ├── models/
│   │   ├── testbed.go           # Testbed model
│   │   ├── resource.go          # Resource instance model
│   │   ├── quota.go             # Quota policy model
│   │   ├── allocation.go        # Allocation record model
│   │   └── category.go          # Category model
│   ├── pool/
│   │   ├── manager.go           # Core pool manager
│   │   ├── acquire.go           # Acquire logic
│   │   ├── release.go           # Release logic
│   │   └── replenish.go         # Replenish logic
│   ├── deployer/
│   │   ├── interface.go         # Deployer interface
│   │   └── mock_deployer.go     # Mock implementation (MVP)
│   ├── storage/
│   │   ├── interface.go         # Storage interface
│   │   └── mysql_storage.go     # MySQL implementation
│   └── monitor/
│       ├── metrics.go           # Metrics collection
│   └── alert.go                 # Alerting logic
└── install/
    ├── install-pool.sh          # Backend deployment script
    └── init-mysql.sql           # Database schema (new tables only)

src/modules/event-processor/frontend/           # 共享前端
├── src/
│   ├── pages/                    # 现有页面
│   │   ├── TaskList.jsx
│   │   ├── TaskDetail.jsx
│   │   └── ...
│   └── pages/
│       └── resource-pool/        # 新增资源池页面
│           ├── TestbedList.jsx         # 测试床列表
│           ├── TestbedDetail.jsx       # 测试床详情
│           ├── CategoryManage.jsx       # 类别管理
│           ├── QuotaPolicy.jsx          # 配额策略
│           ├── AllocationHistory.jsx    # 分配历史
│           └── MetricsDashboard.jsx     # 监控仪表盘
├── src/
│   └── components/
│       └── resource-pool/        # 资源池组件
│           ├── TestbedCard.jsx
│           ├── CategoryForm.jsx
│           └── QuotaChart.jsx
```

## 3. Data Models

**设计原则**: 每个模型都包含 `ID`（数据库自增主键）和 `UUID`（应用生成的唯一标识），所有外部关联都使用 UUID。

### 3.1 Testbed (测试床)

```go
type Testbed struct {
    ID                   int        `json:"id"`                        // 数据库自增主键
    UUID                 string     `json:"uuid"`                     // 应用生成的唯一标识
    Name                 string     `json:"name"`                     // e.g., "testbed-main-001"
    CategoryUUID         string     `json:"category_uuid"`            // 关联的 Category UUID
    ResourceInstanceUUID string     `json:"resource_instance_uuid"`  // 关联的 ResourceInstance UUID
    CurrentAllocUUID     *string    `json:"current_alloc_uuid"`       // 当前分配记录的 UUID（如果已分配）
    MariaDBPort          int        `json:"mariadb_port"`             // 产品部署的 MariaDB 端口
    MariaDBUser          string     `json:"mariadb_user"`             // MariaDB 用户名
    MariaDBPasswd        string     `json:"mariadb_passwd"`           // MariaDB 密码 (加密存储)
    Status               string     `json:"status"`                   // available, allocated, in_use, releasing, maintenance
    LastHealthCheck      time.Time  `json:"last_health_check"`
    CreatedAt            time.Time  `json:"created_at"`
    UpdatedAt            time.Time  `json:"updated_at"`
}
```

### 3.2 ResourceInstance (资源实例)

```go
type ResourceInstance struct {
    ID           int        `json:"id"`             // 数据库自增主键
    UUID         string     `json:"uuid"`           // 应用生成的唯一标识
    InstanceType string     `json:"instance_type"`  // VirtualMachine 或 Machine (二选一)
    SnapshotID   *string    `json:"snapshot_id"`    // VirtualMachine 时必填，Machine 时为空
    IPAddress    string     `json:"ip_address"`
    Port         int        `json:"port"`
    Passwd       string     `json:"passwd"`         // 访问密码 (加密存储)
    Description  *string    `json:"description"`     // 描述信息 (非必填)
    IsPublic     bool       `json:"is_public"`      // 是否公开可见
    CreatedBy    string     `json:"created_by"`     // 创建者用户名
    Status       string     `json:"status"`         // active, terminating, terminated
    CreatedAt    time.Time  `json:"created_at"`
    TerminatedAt *time.Time `json:"terminated_at"`
}
```

**InstanceType 说明:**
- `VirtualMachine`: 虚拟机，SnapshotID 必须有值，**可参与资源池管理，支持快照回滚**
- `Machine`: 实体机，SnapshotID 为空，**不参与资源池管理，用户自管理，不支持快照回滚**

**IsPublic 可见性规则:**
| InstanceType | IsPublic 默认值 | 是否可修改 | 可见范围 |
|--------------|-----------------|-----------|----------|
| VirtualMachine | true | ❌ 不可修改 | 所有用户 |
| Machine (公开) | true | ✅ 可修改 | 所有用户 |
| Machine (私有) | false | ✅ 可修改 | 仅创建者 |

**重要区分:**
| 特性 | VirtualMachine | Machine |
|------|----------------|---------|
| 资源池管理 | ✅ 支持 | ❌ 不支持 |
| 快照回滚 | ✅ 支持 | ❌ 不支持 |
| 自动回收 | ✅ 支持 | ❌ 不支持 |
| 管理方式 | 系统管理 | 用户自管理 |
| IsPublic | 强制 true | 用户可选 |

### 3.3 Allocation (分配记录)

```go
type Allocation struct {
    ID              int        `json:"id"`                // 数据库自增主键
    UUID            string     `json:"uuid"`              // 应用生成的唯一标识
    TestbedUUID     string     `json:"testbed_uuid"`      // 关联的 Testbed UUID
    CategoryUUID    string     `json:"category_uuid"`     // 关联的 Category UUID
    Requester       string     `json:"requester"`         // "event-processor", "user:<username>"
    RequesterToken  string     `json:"requester_token"`   // For auth validation
    TaskID          *string    `json:"task_id"`           // Associated task ID
    Purpose         string     `json:"purpose"`           // PR testing, regression, etc.
    Status          string     `json:"status"`            // active, released, expired
    AcquiredAt      time.Time  `json:"acquired_at"`
    ReleasedAt      *time.Time `json:"released_at"`
    ExpiresAt       *time.Time `json:"expires_at"`
}
```

### 3.4 QuotaPolicy (配额策略)

```go
type QuotaPolicy struct {
    ID                  int       `json:"id"`                  // 数据库自增主键
    UUID                string    `json:"uuid"`                // 应用生成的唯一标识
    CategoryUUID        string    `json:"category_uuid"`       // 关联的 Category UUID
    MinInstances        int       `json:"min_instances"`       // Minimum instances to maintain
    MaxInstances        int       `json:"max_instances"`       // Maximum instances allowed
    Priority            int       `json:"priority"`            // 1-10, higher = more important
    AutoReplenish       bool      `json:"auto_replenish"`      // Auto-create new instances
    ReplenishThreshold  int       `json:"replenish_threshold"` // Trigger replenish below this
    MaxLifetimeSeconds  int       `json:"max_lifetime_seconds"` // 最大生命周期(秒)，由管理员设置，如86400=1天
    CreatedAt           time.Time `json:"created_at"`
    UpdatedAt           time.Time `json:"updated_at"`
}
```

### 3.5 Category (类别配置)

```go
type Category struct {
    ID          int        `json:"id"`           // 数据库自增主键
    UUID        string     `json:"uuid"`         // 应用生成的唯一标识
    Name        string     `json:"name"`         // main, release-001, etc.
    Description string     `json:"description"`
    Enabled     bool       `json:"enabled"`
    CreatedAt   time.Time  `json:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at"`
}
```

## 3.6 ResourceInstance → Testbed 部署流程

**概念说明：**

- **ResourceInstance (资源实例)**: 裸的虚拟机或实体机，只有操作系统和基础环境
- **Testbed (测试床)**: 在 ResourceInstance 上部署了产品后的完整测试环境
- **部署 (Provision)**: 将产品部署到 ResourceInstance，创建 Testbed 记录的过程

**关系图：**

```
┌─────────────────────┐         部署 (Provision)        ┌─────────────────────┐
│  ResourceInstance   │ ───────────────────────────────> │      Testbed        │
│  (裸机/基础环境)     │                                 │  (产品已部署)         │
├─────────────────────┤                                 ├─────────────────────┤
│ - IP: 10.0.1.100    │                                 │ - IP: 10.0.1.100     │
│ - Port: 22          │                                 │ - MariaDB Port: 3306 │
│ - Passwd: xxx       │                                 │ - MariaDB User: root │
│ - SnapshotID: snap  │                                 │ - MariaDB Passwd: yyy│
└─────────────────────┘                                 └─────────────────────┘
```

### 3.6.1 部署流程

```
┌─────────────────────────────────────────────────────────────────┐
│              Provision ResourceInstance → Testbed               │
└─────────────────────────────────────────────────────────────────┘

触发方式:
├─ 自动补充 (Replenish Pool): 根据 QuotaPolicy 自动触发
└─ 手动部署 (Manual Provision): 管理员手动触发

Step 1: 准备 ResourceInstance
└─ 查找可用的 ResourceInstance
    └─ 查询: SELECT * FROM resource_instances
             WHERE instance_type='VirtualMachine'
               AND status='active'
               AND NOT EXISTS (SELECT 1 FROM testbeds
                              WHERE resource_instance_uuid = resource_instances.uuid)

Step 2: 部署产品到 ResourceInstance
├─ 调用部署服务: deployer.DeployProduct(resource_instance, product_config)
├─ 等待部署完成
├─ 部署内容:
│   ├─ 安装应用依赖
│   ├─ 初始化数据库
│   ├─ 配置服务端口
│   └─ 启动应用服务
└─ 获取部署结果:
    ├─ mariadb_port: 3306 (或动态分配的端口)
    ├─ mariadb_user: "root" (或自动生成的用户)
    └─ mariadb_passwd: "随机生成或配置指定"

Step 3: 创建 Testbed 记录
├─ INSERT INTO testbeds (
│     uuid,
│     name,
│     category_uuid,
│     resource_instance_uuid,
│     mariadb_port,
│     mariadb_user,
│     mariadb_passwd,
│     status
│   ) VALUES (
│     generate_uuid(),
│     '{category}-{auto_increment}',
│     category_uuid,
│     resource_instance_uuid,
│     deployment_result.mariadb_port,
│     deployment_result.mariadb_user,
│     encrypt(deployment_result.mariadb_passwd),
│     'available'
│   )
└─ 标记 Testbed 状态为 'available'

Step 4: 健康检查
├─ 调用: health_checker.Check(testbed)
├─ 检查项:
│   ├─ SSH 连接: ssh root@ip
│   ├─ 应用端口: nc -zv ip port
│   ├─ MariaDB: mysql -h ip -P port -u user -p
│   └─ 应用接口: GET http://ip:port/health
└─ 如果健康检查失败:
    ├─ 标记 status = 'maintenance'
    └─ 记录失败原因

Step 5: 完成
└─ 返回 Testbed 信息
```

### 3.6.2 部署服务接口定义

```go
// DeployService 产品部署服务接口
type DeployService interface {
    // DeployProduct 将产品部署到指定的资源实例
    DeployProduct(ctx context.Context, req DeployRequest) (*DeployResult, error)

    // RestoreSnapshot 将资源实例回滚到指定快照
    RestoreSnapshot(ctx context.Context, resourceUUID, snapshotID string) error
}

// DeployRequest 部署请求
type DeployRequest struct {
    ResourceInstanceUUID string            `json:"resource_instance_uuid"`
    ProductConfig        ProductConfig     `json:"product_config"`
    Timeout              time.Duration     `json:"timeout"`
}

// ProductConfig 产品配置
type ProductConfig struct {
    Version        string            `json:"version"`         // 产品版本
    ConfigFile     string            `json:"config_file"`     // 配置文件内容
    EnvVars        map[string]string `json:"env_vars"`        // 环境变量
}

// DeployResult 部署结果
type DeployResult struct {
    Success        bool   `json:"success"`
    MariaDBPort    int    `json:"mariadb_port"`
    MariaDBUser    string `json:"mariadb_user"`
    MariaDBPasswd  string `json:"mariadb_passwd"`
    AppPort        int    `json:"app_port"`
    ErrorMessage   string `json:"error_message,omitempty"`
    LogURL         string `json:"log_url,omitempty"`
}
```

### 3.6.3 自动补充部署流程

**唯一方式: 自动补充 (Auto Replenish)**

```
触发条件:
├─ 定时任务: 每5分钟检查一次
└─ 配额触发: 可用数量 < min_instances

完整流程:

1. 检查是否需要补充
   ├─ 查询: COUNT(*) FROM testbeds
   │          WHERE category_uuid = ? AND status = 'available'
   └─ IF count < quota_policy.min_instances:
       └─ 触发补充

2. 检查配额限制
   ├─ 查询: COUNT(*) FROM testbeds
   │          WHERE category_uuid = ?
   └─ IF total >= quota_policy.max_instances:
       └─ 跳过补充 (已达上限)

3. 计算需要创建的数量
   └─ needed = min_instances - current_available

4. 创建新的 ResourceInstance
   FOR i = 1 TO needed:
       a. 从快照创建虚拟机
          └─ cloud_api.CreateInstance(
                  snapshot_id = category.snapshot_id,
                  instance_type = "VirtualMachine"
              )

       b. 等待虚拟机就绪
          └─ 轮询实例状态直到 status='running'

       c. 记录 ResourceInstance
          └─ INSERT INTO resource_instances (
                  uuid, instance_type, snapshot_id,
                  ip_address, port, passwd,
                  is_public = true,
                  created_by = 'system',
                  status = 'active'
              )

5. 部署产品到 ResourceInstance
   FOR EACH new_resource_instance:
       a. 调用部署服务
          └─ deployer.DeployProduct(
                  resource_instance = new_resource_instance,
                  product_config = category.product_config
              )

       b. 等待部署完成
          └─ 超时时间: 180分钟

       c. 部署失败处理
          └─ 删除 ResourceInstance
              └─ 记录失败日志
              └─ 发送告警

6. 创建 Testbed 记录
   FOR EACH successful_deployment:
       └─ INSERT INTO testbeds (
                  uuid, name, category_uuid,
                  resource_instance_uuid,
                  mariadb_port, mariadb_user, mariadb_passwd,
                  status = 'available'
              )

7. 健康检查
   FOR EACH new_testbed:
       └─ health_checker.Check(testbed)
           └─ 失败则标记为 'maintenance'
```

**重要说明:**
- ❌ 不支持手动注册 Testbed
- ✅ 只能通过自动补充机制创建 Testbed
- ✅ 补充由后台定时任务自动触发

### 3.6.4 部署状态机

```
┌──────────────┐
│  provisioning │ ← 部署中
└──────┬───────┘
       │ 部署成功
       ▼
┌──────────────┐
│  available   │ ← 可用，可分配
└──────┬───────┘
       │ 被分配
       ▼
┌──────────────┐
│   allocated  │ ← 已分配
└──────┬───────┘
       │ 用户开始使用
       ▼
┌──────────────┐
│    in_use    │ ← 使用中
└──────┬───────┘
       │ 释放或过期
       ▼
┌──────────────┐
│   releasing  │ ← 释放中，恢复快照
└──────┬───────┘
       │ 恢复完成
       ▼
┌──────────────┐
│  available   │ ← 回到可用状态
└──────────────┘

异常状态:
┌──────────────┐
│ maintenance  │ ← 维护中，不可分配
└──────────────┘
```

## 4. Core Algorithms

### 4.1 Acquire Testbed (获取测试床)

```
┌─────────────────────────────────────────────────────────────────┐
│                        Acquire Flow                              │
└─────────────────────────────────────────────────────────────────┘

1. Validate Request
   ├─ Check authentication (external API only)
   ├─ Validate category exists
   └─ Check quota availability

2. Find Available Testbed
   ├─ Query: SELECT * FROM testbeds
   │           WHERE category_uuid = ? AND status = 'available'
   │           ORDER BY updated_at ASC
   │           LIMIT 1
   └─ If none found:
       ├─ Check quota: Can we provision new?
       └─ If yes, trigger provisioning (async)

3. Mark as Allocated
   ├─ UPDATE testbeds SET status = 'allocated',
   │                           current_alloc_uuid = ?
   │           WHERE uuid = ? AND status = 'available'
   └─ Create allocation record with:
       ├─ expires_at = NOW() + quota_policy.max_lifetime_seconds
       └─ status = 'active'

4. Return Access Information
   ├─ Testbed UUID, Name
   ├─ Resource Instance details (IP, port, passwd)
   ├─ Allocation UUID
   └─ Expires At (基于配额策略的最大生命周期)

Error Handling:
- No quota available: Return 409 Conflict
- Category disabled: Return 400 Bad Request
- Provisioning needed: Return 202 Accepted with async operation ID
```

### 4.2 Release Testbed (释放测试床)

```
┌─────────────────────────────────────────────────────────────────┐
│                        Release Flow                              │
└─────────────────────────────────────────────────────────────────┘

1. Validate Request
   ├─ Check allocation exists
   ├─ Verify requester owns the allocation
   └─ Check allocation is still active

2. Mark as Releasing
   ├─ UPDATE testbeds SET status = 'releasing'
   │           WHERE id = ? AND current_alloc_uuid = ?
   └─ UPDATE allocations SET status = 'released',
                              released_at = NOW()
           WHERE id = ?

3. Trigger Restore Process (Async)
   ├─ 检查 ResourceInstance.InstanceType
   ├─ IF InstanceType == 'VirtualMachine':
   │   ├─ 获取 ResourceInstance.SnapshotID
   │   ├─ 调用快照回滚: deployer.RestoreSnapshot(resource_instance.uuid, snapshot_id)
   │   │   **注意**: 初期实现时先 mock 此操作，记录日志即可
   │   ├─ 等待回滚完成
   │   └─ 成功后: UPDATE testbeds SET status = 'available'
   ├─ ELSE (Machine 类型):
   │   └─ 直接标记为 available，不做回滚
   └─ 记录恢复操作日志

4. Return Confirmation
   └─ 200 OK with allocation details

Error Handling:
- Allocation not found: Return 404 Not Found
- Not authorized: Return 403 Forbidden
- Already released: Return 409 Conflict
```

### 4.3 Auto Expire & Reclaim (自动过期回收)

**后台定时任务** (每分钟执行一次)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Auto Expire Flow                              │
└─────────────────────────────────────────────────────────────────┘

1. 查找过期分配
   SELECT * FROM allocations
   WHERE status = 'active'
     AND expires_at IS NOT NULL
     AND expires_at < NOW()

2. 批量处理过期分配
   FOR EACH expired_allocation:
       a. 标记分配为已过期
          UPDATE allocations SET status = 'expired'
          WHERE uuid = ?

       b. 释放关联的 Testbed
          UPDATE testbeds SET status = 'releasing',
                              current_alloc_uuid = NULL
          WHERE current_alloc_uuid = ?

       c. 记录过期日志
          INSERT INTO release_logs (allocation_uuid, reason, timestamp)
          VALUES (?, 'auto_expired', NOW())

3. 触发 Testbed 恢复流程
   └─ 调用 Restore Process (同 Release Testbed 步骤3)

4. 通知用户 (可选)
   └─ 发送通知: "您的 testbed xxx 已因超时被系统回收"
```

**管理员配置** (通过控制台):
- 设置每个 Category 的 `MaxLifetimeSeconds`
  - 短期测试: 3600 (1小时)
  - PR 测试: 7200 (2小时)
  - 长期调试: 86400 (1天)
  - 特殊场景: 604800 (7天)

### 4.4 Replenish Pool (补充池)

```
┌─────────────────────────────────────────────────────────────────┐
│                       Replenish Flow                             │
└─────────────────────────────────────────────────────────────────┘

Trigger: Periodic (every 5 min) or on-demand

1. For Each Category:
   ├─ Get quota policy
   ├─ Count available instances
   └─ If count < min_instances:
       ├─ Calculate needed: min_instances - count
       ├─ Check: current + needed <= max_instances?
       └─ If yes, provision needed instances

2. Provision New Instance
   ├─ Call deployer.Provision(category, snapshot_id)
   ├─ Wait for ready state
   ├─ Create testbed record (status: 'available')
   ├─ Create resource_instance record
   └─ Add to pool

3. Health Check
   ├─ Ping each available testbed
   ├─ Mark unhealthy as 'maintenance'
   └─ Alert if too many unhealthy

Monitoring:
- Log replenish actions
- Alert if provisioning fails
- Alert if pool consistently low
```

### 4.4 Quota Allocation (配额分配)

```
Quota Calculation:

Total Pool Size = N instances (global limit)

Category Allocations:
- main:        40% (max 4 instances)
- release-001: 30% (max 3 instances)
- release-002: 20% (max 2 instances)
- general:     10% (max 1 instance)

Priority-Based Borrowing:
If category A needs more but at max:
1. Check if lower priority category has spare
2. Temporarily "borrow" from lower priority
3. Return when lower priority needs it

Example:
main (priority 10) needs 5, max is 4
release-001 (priority 5) has 2/3 (1 spare)
→ main borrows 1 from release-001
→ main now has 5 (4 owned + 1 borrowed)
```

### 4.5 Implementation Notes (实现注意事项)

#### 4.5.1 Mock 快照回滚操作 (初期实现)

**快照回滚操作暂时 Mock：**

```go
// MockSnapshotRollback 初期实现
func MockSnapshotRollback(resourceInstanceUUID, snapshotID string) error {
    // TODO: 实际调用云平台 API 执行快照回滚
    // 当前实现: 仅记录日志，模拟成功

    log.Printf("[MOCK] Snapshot rollback called: resource=%s, snapshot=%s",
        resourceInstanceUUID, snapshotID)

    // 模拟延迟
    time.Sleep(1 * time.Second)

    log.Printf("[MOCK] Snapshot rollback completed: resource=%s", resourceInstanceUUID)
    return nil
}
```

**实际实现时的接口定义：**

```go
// SnapshotService 快照服务接口
type SnapshotService interface {
    // RestoreSnapshot 回滚到指定快照
    RestoreSnapshot(resourceUUID, snapshotID string) error

    // GetSnapshotStatus 获取快照状态
    GetSnapshotStatus(snapshotID string) (string, error)
}
```

#### 4.5.2 验证规则

**Testbed 注册验证：**

```go
func ValidateTestbedRegistration(testbed Testbed, instance ResourceInstance) error {
    // 只允许 VirtualMachine 类型的 ResourceInstance 注册到资源池
    if instance.InstanceType != "VirtualMachine" {
        return errors.New("只有 VirtualMachine 类型可以注册到资源池")
    }

    // VirtualMachine 必须有 SnapshotID
    if instance.InstanceType == "VirtualMachine" && instance.SnapshotID == nil {
        return errors.New("VirtualMachine 类型必须指定 SnapshotID")
    }

    return nil
}
```

**资源池操作验证：**

| 操作 | VirtualMachine | Machine |
|------|----------------|---------|
| 注册到资源池 | ✅ 允许 | ❌ 拒绝 |
| 分配获取 | ✅ 允许 | ❌ 拒绝 |
| 快照回滚 | ✅ 支持 | ❌ 不支持 |

## 5. Workflows (工作流程)

### 5.1 System Initialization Workflow (系统初始化流程)

```
┌─────────────────────────────────────────────────────────────────┐
│                     系统首次安装流程                              │
└─────────────────────────────────────────────────────────────────┘

Step 1: 初始化数据库
├─ 执行 resource-pool/install/init-mysql.sql
├─ 在现有 event-processor 数据库中创建新表
├─ 表包括: categories, quota_policies, testbeds, resource_instances, allocations
└─ 不创建新的用户表 (使用 event-processor 的 users 表)

Step 2: 部署后端服务
├─ 执行 install/install-pool.sh
├─ 构建 Docker 镜像
├─ 启动 resource-pool 容器 (Port 5003)
└─ 验证健康检查端点

Step 3: 更新前端 (添加资源池菜单)
├─ 在 event-processor/frontend 中添加资源池相关页面
├─ 重新构建前端
├─ 重启 event-processor 前端容器
└─ 验证新菜单可见

Step 4: 系统就绪
├─ 数据库表已创建
├─ 后端 API 端点可用 (http://10.4.174.125:5003)
├─ 前端菜单已更新 (http://10.4.174.125:8082)
└─ 使用 event-processor 的用户登录系统
```

### 5.2 Admin Configuration Workflow (管理员配置流程)

```
┌─────────────────────────────────────────────────────────────────┐
│                     管理员配置完整流程                            │
└─────────────────────────────────────────────────────────────────┘

Workflow A: 创建类别 (Create Category)

1. 管理员登录前端 (使用 event-processor 账户)
   ├─ 访问 http://10.4.174.125:8082
   └─ 使用 event-processor 的用户登录 (已存在的管理员账户)

2. 进入资源池管理页面
   ├─ 点击侧边栏 "Resource Pool" 菜单 (新增)
   └─ 点击 "Categories" 子菜单

3. 创建新类别
   ├─ 点击 "New Category" 按钮

3. 填写类别信息
   POST /api/v1/admin/categories
   {
       "name": "main",
       "description": "主分支回归测试环境",
       "enabled": true
   }

4. 系统响应
   {
       "id": 1,
       "uuid": "cat-main-uuid-xxxx",
       "name": "main",
       "description": "主分支回归测试环境",
       "enabled": true,
       "created_at": "2026-03-06T10:00:00Z"
   }

5. 重复创建其他类别
   ├─ release-001 (发版分支001)
   ├─ release-002 (发版分支002)
   └─ pr-testing (PR测试环境)


Workflow B: 创建配额策略 (Create Quota Policy)

1. 进入策略管理页面
   ├─ 点击 "Quota Policies" 菜单
   └─ 点击 "New Policy"

2. 为每个类别配置策略
   PUT /api/v1/admin/quota/{category_uuid}

   示例 - main 类别:
   {
       "category_uuid": "cat-main-uuid-xxxx",
       "min_instances": 2,                    // 最少保持2个可用
       "max_instances": 5,                    // 最多5个实例
       "priority": 10,                        // 高优先级
       "auto_replenish": true,                // 自动补充
       "replenish_threshold": 1,              // 低于1个时触发补充
       "max_lifetime_seconds": 28800          // 最大生命周期8小时
   }

   示例 - release-001 类别:
   {
       "category_uuid": "cat-rel001-uuid-xxxx",
       "min_instances": 1,
       "max_instances": 3,
       "priority": 8,
       "auto_replenish": true,
       "replenish_threshold": 1,
       "max_lifetime_seconds": 86400          // 最大生命周期1天
   }

3. 生命周期建议值
   ├─ 快速测试: 3600 (1小时)
   ├─ PR 测试: 7200 (2小时)
   ├─ 日常调试: 28800 (8小时)
   ├─ 长期任务: 86400 (1天)
   └─ 特殊场景: 604800 (7天)

3. 策略生效
   ├─ 系统开始监控各类别可用数量
   └─ 低于阈值时自动触发补充


Workflow C: 注册机器/环境 (Register Machine)

方式一: 自动注册 (推荐，Phase 2)

1. 配置自动部署参数
   POST /api/v1/admin/auto-provision
   {
       "category": "main",
       "count": 2,                    // 创建2个环境
       "snapshot_id": "snap-main-v1.0"
   }

2. 系统自动调用 DevOps API
   ├─ 创建云主机实例
   ├─ 部署基础环境
   ├─ 配置网络和安全组
   └─ 记录环境信息到数据库

3. 环境就绪
   {
       "created_testbeds": [
           {"id": 1, "name": "testbed-main-001", "status": "available"},
           {"id": 2, "name": "testbed-main-002", "status": "available"}
       ]
   }


Workflow D: 触发初始补充 (Initial Replenish)

1. 确认类别和策略已配置
   GET /api/v1/admin/categories
   GET /api/v1/admin/quota-policies

2. 手动触发补充
   POST /api/v1/admin/replenish
   {
       "category": "main"    // 留空表示全部类别
   }

3. 系统执行补充
   ├─ 检查每个类别的当前数量
   ├─ 对比 min_instances
   ├─ 创建缺失的实例
   └─ 更新环境状态为 available

4. 验证结果
   GET /api/v1/admin/metrics
   {
       "categories": {
           "main": {
               "total": 2,
               "available": 2,
               "allocated": 0,
               "in_use": 0
           }
       }
   }
```

### 5.3 Daily Operations Workflow (日常运维流程)

```
┌─────────────────────────────────────────────────────────────────┐
│                     日常使用流程                                  │
└─────────────────────────────────────────────────────────────────┘

Workflow A: Event-Processor 自动获取环境

┌──────────────────────────────────────────────────────────────────┐
│  Event-Processor                                                  │
│                                                                  │
│  1. 接收到 PR 任务                                                │
│  ├─ task.category = "main"                                       │
│  └─ task.type = "specialized_tests_agent_e2e"                   │
│                                                                  │
│  2. 调用 Resource Pool API                                        │
│  POST /internal/v1/testbeds/acquire                          │
│  {                                                               │
│      "category": "main",                                         │
│      "requester": "event-processor",                             │
│      "task_id": "task-123",                                      │
│      "purpose": "PR testing for PR #456"                         │
│  }                                                               │
│                                                                  │
│  3. 获取环境分配                                                  │
│  {                                                               │
│      "status": "acquired",                                       │
│      "allocation_uuid": "alloc-789-uuid-xxxx",                  │
│      "testbed": {                                                │
│          "id": 1,                                                │
│          "uuid": "testbed-1-uuid-xxxx",                          │
│          "name": "testbed-main-001",                             │
│          "category": "main"                                      │
│      },                                                          │
│      "resource": {                                               │
│          "uuid": "res-uuid-xxxx",                                │
│          "instance_type": "VirtualMachine",                       │
│          "snapshot_id": "snap-main-v1.0",                        │
│          "ip_address": "10.0.1.100",                             │
│          "port": 8080                                            │
│      },                                                          │
│      "expires_at": "2026-03-06T14:00:00Z"                        │
│  }                                                               │
│                                                                  │
│  4. 在获取的环境上执行测试                                         │
│  ├─ 部署代码到环境                                               │
│  ├─ 运行测试脚本                                                 │
│  └─ 收集测试结果                                                 │
│                                                                  │
│  5. 测试完成，释放环境                                            │
│  POST /internal/v1/testbeds/release                          │
│  {                                                               │
│      "allocation_uuid": "alloc-789-uuid-xxxx",                  │
│      "requester": "event-processor"                              │
│  }                                                               │
│                                                                  │
│  6. 环境自动恢复                                                  │
│  ├─ 状态变更为 "releasing"                                       │
│  ├─ 恢复到基础快照                                               │
│  └─ 状态变更为 "available"                                       │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘


Workflow B: 用户手动获取环境

1. 用户登录 (使用 event-processor 账户)
   ├─ 访问 http://10.4.174.125:8082
   └─ 使用 event-processor 账户登录

2. 进入资源池管理
   ├─ 点击侧边栏 "Resource Pool" 菜单
   └─ 点击 "Request Resource" 子菜单

3. 提交申请
   POST /api/v1/testbeds/acquire
   Authorization: Bearer {user_token}
   {
       "category": "main",
       "purpose": "手动验证 PR #456 的修复"
   }

4. 获取分配
   ├─ 系统分配可用环境
   ├─ 显示连接信息
   └─ 记录分配历史

5. 使用环境
   ├─ 通过 SSH/浏览器访问
   ├─ 进行测试/调试
   └─ 完成后释放

6. 释放环境
   POST /api/v1/testbeds/release
   {
       "allocation_uuid": "alloc-790-uuid-xxxx"
   }


Workflow C: 监控和告警处理

1. 实时监控
   GET /api/v1/admin/metrics

   仪表盘显示:
   ├─ 每个类别的资源使用情况
   ├─ 当前分配数量
   ├─ 等待队列长度
   └─ 健康状态统计

2. 告警处理流程

   告警: Pool below minimum threshold
   ├─ 检查 main 类别可用 < 1
   ├─ 手动触发补充
   │   POST /api/v1/admin/replenish {"category": "main"}
   └─ 检查补充结果

   告警: Testbed stuck in releasing
   ├─ 查找卡住的环境
   │   GET /api/v1/admin/testbeds?status=releasing
   ├─ 强制恢复环境
   │   POST /api/v1/admin/testbeds/{id}/force-reset
   └─ 检查恢复结果

   告警: High failure rate
   ├─ 查看失败日志
   ├─ 识别失败原因
   └─ 修复或标记环境为 maintenance


Workflow D: 用户查看和管理我的分配 (User View My Allocations)

1. 查看我的所有分配
   GET /api/v1/my/allocations
   Cookie: session_id={session_id}

   响应示例:
   {
       "allocations": [
           {
               "uuid": "alloc-001-uuid-xxxx",
               "testbed": {
                   "uuid": "testbed-1-uuid-xxxx",
                   "name": "testbed-main-001",
                   "ip_address": "10.0.1.100",
                   "mariadb_port": 3306,
                   "mariadb_user": "root"
               },
               "category": "main",
               "status": "active",
               "acquired_at": "2026-03-06T10:00:00Z",
               "expires_at": "2026-03-06T18:00:00Z",    // 过期时间，基于配额策略
               "remaining_time": 28800                   // 剩余秒数
           },
           {
               "uuid": "alloc-002-uuid-xxxx",
               "testbed": {
                   "uuid": "testbed-2-uuid-xxxx",
                   "name": "testbed-main-002",
                   "ip_address": "10.0.1.101",
                   "mariadb_port": 3306,
                   "mariadb_user": "root"
               },
               "category": "main",
               "status": "active",
               "acquired_at": "2026-03-06T09:30:00Z",
               "expires_at": "2026-03-06T17:30:00Z",
               "remaining_time": 25200
           }
       ]
   }

2. 主动释放某个分配
   POST /api/v1/my/allocations/{allocation_uuid}/release
   Cookie: session_id={session_id}

   请求:
   {
       "reason": "测试完成"  // 可选
   }

   响应:
   {
       "status": "released",
       "allocation_uuid": "alloc-001-uuid-xxxx",
       "released_at": "2026-03-06T14:30:00Z"
   }

3. 续期申请 (可选功能)
   POST /api/v1/my/allocations/{allocation_uuid}/extend
   Cookie: session_id={session_id}

   请求:
   {
       "extend_seconds": 3600  // 续期1小时
   }

   响应 (成功):
   {
       "status": "extended",
       "new_expires_at": "2026-03-06T19:00:00Z"
   }

   响应 (拒绝):
   {
       "error": "extension_denied",
       "reason": "已达到最大生命周期限制"
   }

前端页面显示:
├─ 分配列表表格
│  ├─ Testbed 名称
│  ├─ 获取时间
│  ├─ 过期时间 (倒计时显示)
│  ├─ 剩余时间 (进度条可视化)
│  └─ 操作按钮 (释放/续期)
└─ 过期提醒
   ├─ 剩余 < 30分钟: 黄色警告
   └─ 剩余 < 10分钟: 红色警告
```

### 5.4 Testbed Lifecycle Workflow (测试床生命周期)

```
┌─────────────────────────────────────────────────────────────────┐
│              测试床状态流转完整图                                 │
└─────────────────────────────────────────────────────────────────┘

    ┌──────────┐
    │ Created  │  ← 测试床被创建
    └─────┬────┘
          │
          ▼
   ┌──────────────┐
   │ provisioning │  ← 正在部署/初始化
   └─────┬────────┘
         │
         ▼
   ┌──────────────┐
   │  available   │  ← 可用，等待分配
   └─────┬────────┘
         │ acquire()
         ▼
   ┌──────────────┐
   │  allocated   │  ← 已分配，等待使用
   └─────┬────────┘
         │ user confirms / auto-start
         ▼
   ┌──────────────┐
   │   in_use     │  ← 正在使用中
   └─────┬────────┘
         │ release() or timeout
         ▼
   ┌──────────────┐
   │  releasing   │  ← 正在恢复快照
   └─────┬────────┘
         │ restore complete
         ▼
   ┌──────────────┐
   │  available   │  ← 回到可用状态
   └──────────────┘

                    ┌──────────┐
                    │ Failed   │  ← 错误状态
                    └─────┬────┘
                          │
                    admin recovers
                          │
                    ┌─────▼────┐
                    │ available│
                    └──────────┘

                    ┌─────────────┐
                    │ maintenance │  ← 维护状态
                    └──────┬──────┘
                           │
                     admin fixes
                           │
                     ┌─────▼────┐
                     │ available│
                     └──────────┘
```

### 5.5 Troubleshooting Workflow (故障处理流程)

```
┌─────────────────────────────────────────────────────────────────┐
│                     常见问题处理流程                              │
└─────────────────────────────────────────────────────────────────┘

Problem 1: 无法获取环境 (Acquire Failed)

症状:
  POST /internal/v1/testbeds/acquire
  返回 409 Conflict: no_available_resources

诊断步骤:
1. 检查该类别的环境数量
   GET /api/v1/admin/testbeds?category=main

2. 检查配额策略
   GET /api/v1/admin/quota/main

3. 检查环境状态分布
   GET /api/v1/admin/metrics

可能原因和解决方案:
├─ 原因A: 环境都在使用中
│  └─ 方案: 等待环境释放，或调大 max_instances
├─ 原因B: 环境卡在 releasing 状态
│  └─ 方案: 强制重置卡住的环境
├─ 原因C: 达到配额上限
│  └─ 方案: 调整配额策略或降低其他类别使用
└─ 原因D: 环境处于 maintenance 状态
   └─ 方案: 修复环境并标记为 available


Problem 2: 环境卡在 releasing 状态

症状:
  环境状态长时间显示为 "releasing"

诊断步骤:
1. 查看释放操作日志
2. 检查快照恢复服务状态
3. 验证网络连接

解决方案:
├─ 方案A: 强制重置环境
│  POST /api/v1/admin/testbeds/{id}/force-reset
│  {
│      "reason": "stuck in releasing state",
│      "force_available": true
│  }
└─ 方案B: 标记为维护状态
   POST /api/v1/admin/testbeds/{id}/maintenance
   {
       "reason": "需要手动修复"
   }


Problem 3: 健康检查失败

症状:
  环境被标记为 unhealthy

诊断步骤:
1. 检查环境连通性
   GET /api/v1/admin/testbeds/{id}/health

2. 查看资源实例状态
   GET /api/v1/admin/testbeds/{id}/resources

解决方案:
├─ 方案A: 重新注册环境信息
│  PUT /api/v1/admin/testbeds/{id}
│  {
│      "ip_address": "10.0.1.101",  // 更新IP
│      "port": 8080
│  }
└─ 方案B: 删除并重新创建环境
   DELETE /api/v1/admin/testbeds/{id}
   POST /api/v1/admin/testbeds/provision


Problem 4: 补充失败 (Replenish Failed)

症状:
  自动补充任务失败

诊断步骤:
1. 查看补充日志
2. 检查 DevOps API 连接
3. 验证配额策略配置

解决方案:
├─ 方案A: 手动补充
│  POST /api/v1/admin/replenish
├─ 方案B: 检查并修复配额策略
│  PUT /api/v1/admin/quota/{category}
└─ 方案C: 禁用自动补充，转为手动管理
   PUT /api/v1/admin/quota/{category}
   {
       "auto_replenish": false
   }
```

### 5.6 Admin Operation Checklist (管理员操作清单)

```
┌─────────────────────────────────────────────────────────────────┐
│                     离职/交接清单                                 │
└─────────────────────────────────────────────────────────────────┘

初次安装后必做:
□ 1. 确认 event-processor 已有管理员账户 (复用现有账户)
□ 2. 创建至少一个类别 (category)
□ 3. 为每个类别配置配额策略 (quota policy)
□ 4. 手动注册或自动创建环境
□ 5. 触发初始补充，确保环境可用
□ 6. 验证 acquire/release API 正常工作

日常运维 (每日):
□ 1. 查看监控仪表盘
□ 2. 检查告警信息
□ 3. 验证各类别可用数量 >= min_instances
□ 4. 处理卡在 releasing 状态的环境

定期运维 (每周):
□ 1. 审查分配历史，识别异常使用
□ 2. 清理过期的 allocation 记录
□ 3. 评估配额策略是否需要调整
□ 4. 检查环境健康状态
□ 5. 验证快照恢复功能

变更操作:
□ 1. 新增类别: 先创建 category，再配置 quota
□ 2. 调整配额: 先确认有足够资源，再修改策略
□ 3. 下线环境: 确保无活跃分配，再删除
□ 4. 更新快照: 通过快照管理接口更新 (独立于 Category)
```

## 6. API Design

### 5.1 Internal API (No Authentication)

Used by event-processor and trusted internal services.

#### Acquire Testbed

```
POST /internal/v1/testbeds/acquire

Request:
{
    "category": "main",
    "requester": "event-processor",
    "task_id": "task-123",
    "purpose": "PR testing"
}

Response 200 OK:
{
    "status": "acquired",
    "allocation_uuid": "alloc-456-uuid-xxxx",
    "testbed": {
        "id": 78,
        "uuid": "testbed-78-uuid-xxxx",
        "name": "testbed-main-001",
        "category": "main"
    },
    "resource": {
        "uuid": "res-uuid-xxxx",
        "instance_type": "VirtualMachine",
        "ip_address": "10.0.1.100",
        "port": 8080
    },
    "expires_at": "2026-03-05T12:00:00Z"
}

Response 202 Accepted (provisioning):
{
    "status": "provisioning",
    "operation_id": "op-789",
    "estimated_ready": "2026-03-05T11:10:00Z"
}

Response 409 Conflict:
{
    "error": "no_available_resources",
    "message": "No resources available for category 'main'",
    "category": "main",
    "queue_position": 3
}
```

#### Release Testbed

```
POST /internal/v1/testbeds/release

Request:
{
    "allocation_uuid": "alloc-456-uuid-xxxx",
    "requester": "event-processor"
}

Response 200 OK:
{
    "status": "released",
    "allocation_uuid": "alloc-456-uuid-xxxx",
    "testbed_uuid": "testbed-78-uuid-xxxx",
    "restored": true
}
```

#### Check Allocation Status

```
GET /internal/v1/allocations/{allocation_uuid}

Response 200 OK:
{
    "id": 456,
    "uuid": "alloc-456-uuid-xxxx",
    "testbed_uuid": "testbed-78-uuid-xxxx",
    "status": "active",
    "acquired_at": "2026-03-05T10:00:00Z",
    "expires_at": "2026-03-05T12:00:00Z"
}
```

### 5.2 External API (With Session Authentication)

Used by frontend. Uses event-processor's session-based authentication.

#### Acquire Testbed (External)

```
POST /api/v1/testbeds/acquire
Cookie: session_id={session_id}  # From event-processor login

Request:
{
    "category": "main",
    "purpose": "Manual testing"
}

Response: Same as internal API
```

#### Release Testbed (External)

```
POST /api/v1/testbeds/release
Cookie: session_id={session_id}

Request:
{
    "allocation_uuid": "alloc-456-uuid-xxxx"
}

Response: Same as internal API
```

#### List My Allocations

```
GET /api/v1/my/allocations
Cookie: session_id={session_id}

Response 200 OK:
{
    "allocations": [
        {
            "id": 456,
            "uuid": "alloc-456-uuid-xxxx",
            "testbed": {
                "uuid": "testbed-78-uuid-xxxx",
                "name": "testbed-main-001",
                "ip_address": "10.0.1.100",
                "mariadb_port": 3306,
                "mariadb_user": "root"
            },
            "category": "main",
            "status": "active",
            "acquired_at": "2026-03-05T10:00:00Z",
            "expires_at": "2026-03-05T18:00:00Z",
            "remaining_seconds": 28800
        }
    ]
}
```

#### Release My Allocation

```
POST /api/v1/my/allocations/{allocation_uuid}/release
Cookie: session_id={session_id}

Request:
{
    "reason": "测试完成"  // 可选
}

Response 200 OK:
{
    "status": "released",
    "allocation_uuid": "alloc-456-uuid-xxxx",
    "released_at": "2026-03-05T14:30:00Z"
}

Response 404 Not Found:
{
    "error": "allocation_not_found",
    "message": "分配记录不存在或已过期"
}

Response 403 Forbidden:
{
    "error": "not_owner",
    "message": "无权释放此分配"
}
```

#### Extend Allocation (可选功能)

```
POST /api/v1/my/allocations/{allocation_uuid}/extend
Cookie: session_id={session_id}

Request:
{
    "extend_seconds": 3600  // 续期秒数
}

Response 200 OK:
{
    "status": "extended",
    "allocation_uuid": "alloc-456-uuid-xxxx",
    "new_expires_at": "2026-03-05T19:00:00Z"
}

Response 400 Bad Request:
{
    "error": "extension_denied",
    "message": "已达到最大生命周期限制",
    "max_lifetime_seconds": 86400
}
```
```

### 5.3 Admin API

Requires session authentication. Validates session with event-processor.

#### List All Testbeds

```
GET /api/v1/admin/testbeds

Response 200 OK:
{
    "testbeds": [
        {
            "id": 78,
            "name": "testbed-main-001",
            "category": "main",
            "status": "available",
            "current_alloc_uuid": null
        }
    ]
}
```

#### Update Quota Policy

```
PUT /api/v1/admin/quota/{category_uuid}

Request:
{
    "min_instances": 2,
    "max_instances": 5,
    "priority": 10,
    "auto_replenish": true,
    "max_lifetime_seconds": 86400  // 最大生命周期(秒)，1天
}

Response 200 OK:
{
    "uuid": "quota-uuid-xxxx",
    "category_uuid": "cat-main-uuid-xxxx",
    "max_lifetime_seconds": 86400,
    "updated_at": "2026-03-06T10:00:00Z"
}
```

#### Trigger Replenish

```
POST /api/v1/admin/replenish

Request:
{
    "category_uuid": "cat-main-uuid-xxxx"  // empty for all categories
}

Response 200 OK:
{
    "status": "replenishing",
    "provisioning_count": 2
}
```

#### Monitoring Metrics

```
GET /api/v1/admin/metrics

Response 200 OK:
{
    "categories": {
        "main": {
            "total": 4,
            "available": 2,
            "allocated": 1,
            "in_use": 1,
            "releasing": 0
        }
    },
    "system": {
        "total_testbeds": 10,
        "overall_utilization": 0.6
    }
}
```

#### Filter Testbeds

```
GET /api/v1/admin/testbeds?category_uuid=cat-main-uuid-xxxx&status=available

Response 200 OK:
{
    "testbeds": [
        {
            "id": 1,
            "uuid": "testbed-1-uuid-xxxx",
            "name": "testbed-main-001",
            "category": "main",
            "status": "available"
        }
    ]
}
```

#### Auto Provision Testbeds (手动触发补充)

```
POST /api/v1/admin/auto-provision

Request:
{
    "category_uuid": "cat-main-uuid-xxxx",
    "count": 2,
    "snapshot_id": "snap-main-v1.0"
}

Response 202 Accepted:
{
    "operation_id": "op-001",
    "status": "provisioning",
    "estimated_ready": "2026-03-06T11:00:00Z"
}
```

#### Force Reset Testbed

```
POST /api/v1/admin/testbeds/{id}/force-reset

Request:
{
    "reason": "stuck in releasing state",
    "force_available": true
}

Response 200 OK:
{
    "id": 1,
    "status": "available",
    "reset_at": "2026-03-06T10:30:00Z"
}
```

#### Set Testbed Maintenance Mode

```
POST /api/v1/admin/testbeds/{id}/maintenance

Request:
{
    "reason": "需要手动修复网络配置"
}

Response 200 OK:
{
    "id": 1,
    "status": "maintenance",
    "reason": "需要手动修复网络配置"
}
```

#### Delete Testbed

```
DELETE /api/v1/admin/testbeds/{id}

Response 200 OK:
{
    "status": "deleted",
    "id": 1
}
```

#### Get Testbed Health

```
GET /api/v1/admin/testbeds/{testbed_uuid}/health

Response 200 OK:
{
    "testbed_uuid": "testbed-1-uuid-xxxx",
    "healthy": true,
    "last_check": "2026-03-06T10:25:00Z",
    "checks": {
        "network": "ok",
        "ssh": "ok",
        "http": "ok"
    }
}
```

#### Get Testbed Resources

```
GET /api/v1/admin/testbeds/{testbed_uuid}/resources

Response 200 OK:
{
    "testbed_uuid": "testbed-1-uuid-xxxx",
    "resources": [
        {
            "id": 1,
            "uuid": "res-1-uuid-xxxx",
            "instance_type": "VirtualMachine",
            "snapshot_id": "snap-main-v1.0",
            "ip_address": "10.0.1.100",
            "port": 8080,
            "status": "active"
        }
    ]
}
```

#### Create Category

```
POST /api/v1/admin/categories

Request:
{
    "name": "main",
    "description": "主分支回归测试环境",
    "enabled": true
}

Response 201 Created:
{
    "id": 1,
    "uuid": "cat-main-uuid-xxxx",
    "name": "main",
    "description": "主分支回归测试环境",
    "enabled": true,
    "created_at": "2026-03-06T10:00:00Z"
}
```

#### List Categories

```
GET /api/v1/admin/categories

Response 200 OK:
{
    "categories": [
        {
            "id": 1,
            "name": "main",
            "description": "主分支回归测试环境",
            "enabled": true
        },
        {
            "id": 2,
            "name": "release-001",
            "description": "发版分支001",
            "enabled": true
        }
    ]
}
```

#### Get Quota Policy

```
GET /api/v1/admin/quota/{category}

Response 200 OK:
{
    "category": "main",
    "min_instances": 2,
    "max_instances": 5,
    "priority": 10,
    "auto_replenish": true,
    "replenish_threshold": 1
}
```

#### Provision Testbed

```
POST /api/v1/admin/testbeds/provision

Request:
{
    "category_uuid": "cat-main-uuid-xxxx",
    "snapshot_id": "snap-main-v1.0"
}

Response 202 Accepted:
{
    "operation_id": "op-002",
    "status": "provisioning"
}
```

#### Update Testbed

```
PUT /api/v1/admin/testbeds/{testbed_uuid}

Request:
{
    "resource_instance_uuid": "123e4567-e89b-12d3-a456-426614174000"
}

Response 200 OK:
{
    "id": 1,
    "uuid": "testbed-1-uuid-xxxx",
    "name": "testbed-main-001",
    "category": "main",
    "resource_instance_uuid": "123e4567-e89b-12d3-a456-426614174000",
    "status": "available"
}
```

#### Get ResourceInstance Health

```
GET /api/v1/admin/instances/{instance_uuid}/health

Response 200 OK:
{
    "uuid": "res-1-uuid-xxxx",
    "healthy": true,
    "last_check": "2026-03-06T10:30:00Z",
    "checks": {
        "network": "ok",
        "ssh": "ok",
        "mariadb": "ok"
    }
}
```

### 5.4 ResourceInstance API (资源实例 API)

#### List All ResourceInstances (公开实例)

```
GET /api/v1/instances
Cookie: session_id={session_id}

Query Parameters:
- type: filter by instance_type (optional)
- status: filter by status (optional)

Response 200 OK:
{
    "instances": [
        {
            "uuid": "res-1-uuid-xxxx",
            "instance_type": "VirtualMachine",
            "snapshot_id": "snap-main-v1.0",
            "ip_address": "10.0.1.100",
            "port": 8080,
            "is_public": true,
            "created_by": "admin",
            "status": "active"
        },
        {
            "uuid": "res-2-uuid-xxxx",
            "instance_type": "Machine",
            "snapshot_id": null,
            "ip_address": "10.0.1.101",
            "port": 8080,
            "is_public": true,
            "created_by": "user1",
            "status": "active"
        }
    ]
}
```

**查询逻辑:**
```sql
SELECT * FROM resource_instances
WHERE (is_public = true OR instance_type = 'VirtualMachine')
  AND status = 'active'
ORDER BY created_at DESC
```

#### List My ResourceInstances

```
GET /api/v1/my/instances
Cookie: session_id={session_id}

Query Parameters:
- type: filter by instance_type (optional)

Response 200 OK:
{
    "instances": [
        {
            "uuid": "res-2-uuid-xxxx",
            "instance_type": "Machine",
            "snapshot_id": null,
            "ip_address": "10.0.1.101",
            "port": 8080,
            "is_public": false,
            "status": "active",
            "created_at": "2026-03-06T09:00:00Z"
        },
        {
            "uuid": "res-3-uuid-xxxx",
            "instance_type": "Machine",
            "snapshot_id": null,
            "ip_address": "10.0.1.102",
            "port": 8080,
            "is_public": true,
            "status": "active",
            "created_at": "2026-03-05T14:00:00Z"
        }
    ]
}
```

**查询逻辑:**
```sql
SELECT * FROM resource_instances
WHERE created_by = {current_user}
ORDER BY created_at DESC
```

#### Create ResourceInstance (仅 Machine 类型)

```
POST /api/v1/my/instances
Cookie: session_id={session_id}

Request:
{
    "instance_type": "Machine",
    "ip_address": "10.0.1.103",
    "port": 8080,
    "passwd": "my-passwd-123",
    "is_public": true
}

Response 201 Created:
{
    "uuid": "res-new-uuid-xxxx",
    "instance_type": "Machine",
    "ip_address": "10.0.1.103",
    "port": 8080,
    "is_public": true,
    "created_by": "current_user",
    "status": "active"
}
```

#### Update ResourceInstance

```
PUT /api/v1/my/instances/{instance_uuid}
Cookie: session_id={session_id}

Request:
{
    "ip_address": "10.0.1.104",
    "port": 8080,
    "passwd": "new-passwd-456",
    "is_public": false
}

Response 200 OK:
{
    "uuid": "res-2-uuid-xxxx",
    "instance_type": "Machine",
    "ip_address": "10.0.1.104",
    "port": 8080,
    "is_public": false,
    "status": "active"
}
```

#### Delete ResourceInstance

```
DELETE /api/v1/my/instances/{instance_uuid}
Cookie: session_id={session_id}

Response 200 OK:
{
    "uuid": "res-2-uuid-xxxx",
    "status": "terminated",
    "terminated_at": "2026-03-06T11:00:00Z"
}
```

**验证规则:**
- 只能删除自己创建的实例
- VirtualMachine 类型不能通过此 API 删除（由管理员管理）
- Machine 类型可以删除

## 7. Database Schema

**说明**: 这些表添加到现有的 `event_processor` 数据库中，不创建新的数据库。

```sql
-- ============================================================
-- Resource Pool Tables (添加到 event_processor 数据库)
-- 设计原则: 每个表都有 id (自增主键) 和 uuid (应用生成的唯一标识)
-- 所有外键关联都使用 uuid
-- ============================================================

-- Categories (类别配置)
CREATE TABLE IF NOT EXISTS categories (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,          -- 应用生成的唯一标识
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Quota Policies
CREATE TABLE quota_policies (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,          -- 应用生成的唯一标识
    category_uuid CHAR(36) NOT NULL,        -- 关联 categories.uuid
    min_instances INT NOT NULL DEFAULT 1,
    max_instances INT NOT NULL DEFAULT 5,
    priority INT NOT NULL DEFAULT 5,
    auto_replenish BOOLEAN DEFAULT TRUE,
    replenish_threshold INT NOT NULL DEFAULT 1,
    max_lifetime_seconds INT NOT NULL DEFAULT 86400,  -- 最大生命周期(秒)，默认1天
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_uuid) REFERENCES categories(uuid)
);

-- Testbeds
CREATE TABLE testbeds (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,          -- 应用生成的唯一标识
    name VARCHAR(100) UNIQUE NOT NULL,
    category_uuid CHAR(36) NOT NULL,        -- 关联 categories.uuid
    resource_instance_uuid CHAR(36) NOT NULL,  -- 关联 resource_instances.uuid
    current_alloc_uuid CHAR(36),            -- 当前分配记录的 uuid (关联 allocations.uuid)
    mariadb_port INT,                       -- 产品部署的 MariaDB 端口
    mariadb_user VARCHAR(100),              -- MariaDB 用户名
    mariadb_passwd VARCHAR(255),            -- MariaDB 密码 (加密存储)
    status ENUM('available', 'allocated', 'in_use', 'releasing', 'maintenance') DEFAULT 'available',
    last_health_check TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (category_uuid) REFERENCES categories(uuid),
    FOREIGN KEY (resource_instance_uuid) REFERENCES resource_instances(uuid),
    FOREIGN KEY (current_alloc_uuid) REFERENCES allocations(uuid)
);

-- Resource Instances
CREATE TABLE resource_instances (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,          -- 应用生成的唯一标识
    instance_type ENUM('VirtualMachine', 'Machine') NOT NULL,
    snapshot_id VARCHAR(255),               -- VirtualMachine 时必填，Machine 时为 NULL
    ip_address VARCHAR(50),
    port INT,
    passwd VARCHAR(255),                    -- 访问密码 (加密存储)
    description TEXT,                       -- 描述信息 (非必填)
    is_public BOOLEAN DEFAULT TRUE,         -- 是否公开可见
    created_by VARCHAR(100) NOT NULL,       -- 创建者用户名
    status ENUM('active', 'terminating', 'terminated') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    terminated_at TIMESTAMP NULL
);

-- Allocations
CREATE TABLE allocations (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,          -- 应用生成的唯一标识
    testbed_uuid CHAR(36) NOT NULL,         -- 关联 testbeds.uuid
    category_uuid CHAR(36) NOT NULL,        -- 关联 categories.uuid
    requester VARCHAR(255) NOT NULL,
    requester_token VARCHAR(255),
    task_id VARCHAR(255),
    purpose TEXT,
    status ENUM('active', 'released', 'expired') DEFAULT 'active',
    acquired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    released_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    FOREIGN KEY (testbed_uuid) REFERENCES testbeds(uuid),
    FOREIGN KEY (category_uuid) REFERENCES categories(uuid)
);

-- Indexes
CREATE INDEX idx_testbeds_category_status ON testbeds(category_uuid, status);
CREATE INDEX idx_allocations_status ON allocations(status);
CREATE INDEX idx_allocations_requester ON allocations(requester);
CREATE INDEX idx_allocations_testbed_uuid ON allocations(testbed_uuid);
CREATE INDEX idx_quota_policies_category_uuid ON quota_policies(category_uuid);
```

## 8. Monitor & Alert

### 7.1 Metrics to Track

1. **Pool Utilization**
   - Available vs allocated per category
   - Overall utilization percentage
   - Time to acquire (latency)

2. **Provisioning Metrics**
   - Provision success rate
   - Average provision time
   - Failed provisions

3. **Health Status**
   - Healthy vs unhealthy testbeds
   - Health check failures

4. **Alert Conditions**
   - Pool below minimum threshold
   - High failure rate (> 10%)
   - Provisioning timeout
   - Testbed stuck in "releasing" state

### 7.2 Alert Channels

- Log to file with severity levels
- Webhook notifications (integrate with existing alert system)
- Dashboard metrics display

## 9. Integration Points

### 9.1 Event-Processor Backend Integration

Event-processor 后端调用 resource-pool 内部 API (无需认证):

```go
// In event-processor, when executing a task that needs a testbed:

type TaskExecutor struct {
    poolClient *ResourcePoolClient
}

func (e *TaskExecutor) ExecuteTask(task *Task) error {
    // 1. Acquire testbed via internal API
    alloc, err := e.poolClient.Acquire(pool.AcquireRequest{
        Category: task.Category,
        Requester: "event-processor",
        TaskID: task.ID,
        Purpose: "PR testing",
    })
    if err != nil {
        return fmt.Errorf("failed to acquire testbed: %w", err)
    }
    defer e.poolClient.Release(alloc.ID)

    // 2. Use testbed for testing
    envURL := fmt.Sprintf("http://%s:%d", alloc.Resource.IPAddress, alloc.Resource.Port)
    err = e.runTests(envURL, task)

    // 3. Testbed automatically released on defer
    return err
}
```

### 9.2 Frontend Integration (共享前端)

Resource Pool 页面作为 event-processor 前端的新菜单:

```javascript
// event-processor/frontend/src/router.js
{
    path: '/resource-pool',
    component: ResourcePoolLayout,
    meta: { requiresAuth: true },
    children: [
        // ====== Testbed 相关页面 ======
        { path: 'testbeds', component: TestbedList },           // 所有 Testbed (管理员)
        { path: 'my-testbeds', component: MyTestbedList },     // 我申请的 Testbed
        { path: 'testbeds/:id', component: TestbedDetail },

        // ====== ResourceInstance 相关页面 ======
        { path: 'instances', component: ResourceInstanceList },        // 所有资源实例 (公开)
        { path: 'my-instances', component: MyResourceInstanceList },  // 我的资源实例

        // ====== 管理员配置页面 ======
        { path: 'categories', component: CategoryManage },
        { path: 'quota', component: QuotaPolicy },
        { path: 'allocations', component: AllocationHistory },
        { path: 'metrics', component: MetricsDashboard },
    ]
}
```

### 9.3 Frontend Pages (前端页面说明)

#### 9.3.1 Testbed 页面

| 页面 | 路径 | 权限 | 说明 |
|------|------|------|------|
| 所有 Testbed | `/resource-pool/testbeds` | 管理员 | **仅显示可用状态**的 Testbed (status='available') |
| 我的 Testbed | `/resource-pool/my-testbeds` | 登录用户 | 查看自己申请的 Testbed，显示过期倒计时，支持主动释放 |
| Testbed 详情 | `/resource-pool/testbeds/:id` | 相关用户 | 查看 Testbed 详细信息 |

**所有 Testbed 页面规则：**
- **只显示 `status = 'available'` 的 Testbed**
- 已被占用的 Testbed 不在此页面显示
- 查询条件: `WHERE status = 'available'`

**密码显示规则：**
| 页面 | MariaDBPasswd 显示 |
|------|-------------------|
| 所有 Testbed | `****` (星号掩码) |
| Testbed 详情 | `****` (星号掩码) |
| 我的 Testbed | 明文显示 (因为是自己申请的) |

**我的 Testbed 页面功能：**
- 显示当前用户所有分配中的 Testbed
- 每行显示: Testbed 名称、IP、MariaDB 端口、**MariaDB 密码(明文)**、获取时间、过期时间
- 过期时间倒计时显示
- 剩余时间 < 30分钟：黄色警告
- 剩余时间 < 10分钟：红色警告
- 操作按钮: [释放] [续期]

#### 9.3.2 ResourceInstance 页面

| 页面 | 路径 | 权限 | 说明 |
|------|------|------|------|
| 所有资源实例 | `/resource-pool/instances` | 登录用户 | 显示所有公开的资源实例 (VirtualMachine + 公开的 Machine) |
| 我的资源实例 | `/resource-pool/my-instances` | 登录用户 | 显示自己创建的所有资源实例 (包括私有的 Machine) |

**所有资源实例页面 (`/resource-pool/instances`)：**
- 查询条件: `WHERE is_public = true OR instance_type = 'VirtualMachine'`
- 显示内容: 实例类型、IP、端口、**Passwd(掩码)**、快照ID、创建者、状态
- 用户可以查看但不能修改公开的实例

**密码显示规则：**
| 页面 | Passwd 显示 |
|------|-------------|
| 所有资源实例 | `****` (星号掩码) |
| 我的资源实例 | 明文显示 (因为是自己创建的) |

**我的资源实例页面 (`/resource-pool/my-instances`)：**
- 查询条件: `WHERE created_by = current_user`
- 显示内容: 实例类型、是否公开、IP、端口、**Passwd(明文)**、状态
- 操作:
  - VirtualMachine: 查看详情
  - Machine: [编辑] [删除] [切换公开/私有]

**IsPublic 字段说明：**
- **VirtualMachine**: 强制 `is_public = true`，不可修改
- **Machine**: 用户可选择 `is_public = true/false`
  - `true`: 在"所有资源实例"页面可见
  - `false`: 仅在创建者的"我的资源实例"页面可见

#### 9.3.3 密码掩码规则 (Password Masking)

**前端页面密码显示规则：**

| 场景 | MariaDBPasswd (Testbed) | Passwd (ResourceInstance) |
|------|------------------------|--------------------------|
| 所有 Testbed 页面 | `****` | - |
| Testbed 详情页 | `****` | - |
| 我的 Testbed 页面 | 明文 | - |
| 所有资源实例页面 | - | `****` |
| 我的资源实例页面 | - | 明文 |

**后端 API 密码返回规则：**

| API 端点 | 请求者 | 密码字段返回 |
|----------|--------|------------|
| `GET /api/v1/admin/testbeds` | 管理员 | `mariadb_passwd: "****"` |
| `GET /api/v1/my/testbeds` | 用户 | `mariadb_passwd: "明文"` (仅自己的) |
| `GET /api/v1/instances` | 用户 | `passwd: "****"` 或不返回 |
| `GET /api/v1/my/instances` | 用户 | `passwd: "明文"` (仅自己的) |

**实现逻辑：**
```go
// 密码掩码函数
func maskPassword(password string, show bool) string {
    if show {
        return password
    }
    return "****"
}

// API 响应处理
func formatTestbedForAPI(testbed Testbed, requester string) TestbedResponse {
    response := TestbedResponse{
        UUID: testbed.UUID,
        Name: testbed.Name,
        // ... 其他字段
    }

    // 只有拥有者可以看到明文密码
    isOwner := (testbed.CurrentAllocation.Requester == requester)
    response.MariaDBPasswd = maskPassword(testbed.MariaDBPasswd, isOwner)

    return response
}
```

### 9.4 Session Authentication

Resource Pool 后端验证来自 event-processor 的 session:

```go
// resource-pool 内部的 session 验证中间件
func validateSession(sessionID string) (string, error) {
    // 调用 event-processor 验证 session
    resp, err := http.Get(
        "http://event-processor:5002/api/user/validate?session_id=" + sessionID)
    if err != nil || resp.StatusCode != 200 {
        return "", errors.New("invalid session")
    }

    var result struct {
        Username string `json:"username"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.Username, nil
}

// 外部 API 中间件
func sessionAuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        sessionID, err := r.Cookie("session_id")
        if err != nil {
            http.Error(w, "unauthorized", http.StatusUnauthorized)
            return
        }

        username, err := validateSession(sessionID.Value)
        if err != nil {
            http.Error(w, "invalid session", http.StatusUnauthorized)
            return
        }

        // 将用户信息添加到 context
        ctx := context.WithValue(r.Context(), "username", username)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## 10. Implementation Phases

### Phase 1: MVP (Mock Deployer)
**后端**:
- [ ] Basic CRUD for testbeds and allocations
- [ ] Internal acquire/release API (no auth)
- [ ] Mock deployer (no real provisioning)
- [ ] Manual testbed registration
- [ ] Session validation middleware (calls event-processor)

**前端** (添加到 event-processor/frontend):
- [ ] Resource Pool 菜单项
- [ ] 环境列表页面
- [ ] 类别管理页面
- [ ] 基础监控仪表盘

### Phase 2: Real Deployer Integration
**后端**:
- [ ] Implement Tencent DevOps API deployer
- [ ] Auto-provisioning on acquire
- [ ] Snapshot restoration on release
- [ ] Health check system

**前端**:
- [ ] 环境详情页面
- [ ] 分配历史页面
- [ ] 配额策略管理页面

### Phase 3: Advanced Features
**后端**:
- [ ] Quota management with priorities
- [ ] Auto-replenishment
- [ ] Advanced monitoring and alerting
- [ ] Queue system for resource contention

**前端**:
- [ ] 实时监控仪表盘
- [ ] 告警通知
- [ ] 高级分析图表

## 11. Testing Requirements (测试要求)

### 11.1 Testing Strategy (测试策略)

**测试金字塔：**

```
        ┌─────────────┐
        │   E2E Tests │  少量，关键流程验证
        │    (10%)    │
        ├─────────────┤
        │Integration  │  中等，模块间交互
        │   Tests     │  (30%)
        │    (30%)    │
        ├─────────────┤
        │  Unit Tests │  大量，单个函数/方法
        │    (60%)    │
        └─────────────┘
```

**测试原则：**
1. **所有新功能必须配套测试用例**
2. **测试先行** (TDD): 先写测试，再写实现
3. **CI/CD 集成**: 代码提交前必须通过测试
4. **测试隔离**: 每个测试独立运行，不依赖其他测试
5. **可重复性**: 测试结果应稳定，不因环境变化而不同

### 11.2 Unit Testing (单元测试)

**测试范围：**

| 模块 | 测试文件 | 覆盖内容 | 最低覆盖率 |
|------|----------|----------|-----------|
| Models | `models_test.go` | 数据模型验证、序列化/反序列化 | 80% |
| Storage | `storage_test.go` | CRUD 操作、事务处理 | 85% |
| Service | `service_test.go` | 业务逻辑、状态转换 | 85% |
| API Handlers | `handlers_test.go` | 请求/响应处理、错误处理 | 80% |
| Deployer | `deployer_test.go` | 接口实现、Mock 行为 | 90% |

**关键测试用例：**

```go
// models_test.go 示例
func TestTestbedStatusTransition(t *testing.T) {
    tests := []struct {
        name     string
        from     string
        to       string
        allowed bool
    }{
        {"available to allocated", "available", "allocated", true},
        {"allocated to in_use", "allocated", "in_use", true},
        {"in_use to releasing", "in_use", "releasing", true},
        {"releasing to available", "releasing", "available", true},
        {"invalid: available to in_use", "available", "in_use", false},
        {"invalid: in_use to available", "in_use", "available", false},
    }
    // ... 测试实现
}

// storage_test.go 示例
func TestAcquireTestbed_Concurrent(t *testing.T) {
    // 测试并发获取同一个 Testbed 的竞态条件
    // 应该只有一个请求成功，其他返回错误
}

// service_test.go 示例
func TestReleaseTestbed_SnapshotRestore(t *testing.T) {
    // 测试 VirtualMachine 释放后触发快照回滚
    // 测试 Machine 释放不触发回滚
}

// service_test.go 示例
func TestAutoExpire_Allocation(t *testing.T) {
    // 测试过期配额自动释放
    // 验证状态转换和时间戳
}
```

### 11.3 Integration Testing (集成测试)

**测试范围：**

| 场景 | 测试内容 | 验证点 |
|------|----------|--------|
| 数据库集成 | 表创建、索引、外键约束 | Schema 正确性 |
| API 集成 | Internal/External API 调用链 | 端到端请求处理 |
| Event-Processor 集成 | Session 验证 | 外部服务调用 |
| Deployer 集成 | 部署/回滚接口 | 接口契约 |

**集成测试配置：**

```go
// integration_test.go
func TestMain(m *testing.M) {
    // 使用测试数据库
    os.Setenv("POOL_DB_NAME", "resource_pool_test")

    // 运行测试前迁移
    setupTestDB()

    // 运行测试
    code := m.Run()

    // 清理
    teardownTestDB()
    os.Exit(code)
}

func setupTestDB() {
    // 创建测试数据库
    // 运行 migrations
}

func teardownTestDB() {
    // 删除测试数据库
}
```

**关键集成测试用例：**

```go
func TestAcquireTestbed_FullFlow(t *testing.T) {
    // 1. 创建 Category
    // 2. 创建 QuotaPolicy
    // 3. 创建 ResourceInstance
    // 4. 创建 Testbed
    // 5. 调用 Acquire API
    // 6. 验证 Allocation 记录
    // 7. 验证 Testbed 状态变为 'in_use'
}

func TestReleaseTestbed_FullFlow(t *testing.T) {
    // 1. 准备已分配的 Testbed
    // 2. 调用 Release API
    // 3. 验证 Allocation 状态变为 'released'
    // 4. 验证快照回滚被调用 (Mock)
    // 5. 验证 Testbed 状态变为 'available'
}
```

### 11.4 E2E Testing (端到端测试)

**测试流程：**

```
┌─────────────────────────────────────────────────────────────┐
│                     E2E Test Flow                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. System Initialization                                   │
│     ├─ 启动所有服务                                          │
│     ├─ 运行数据库迁移                                        │
│     └─ 创建初始数据                                          │
│                                                              │
│  2. Admin Configuration                                     │
│     ├─ 创建 Category                                         │
│     ├─ 设置 QuotaPolicy                                      │
│     └─ 触发补充资源池                                        │
│                                                              │
│  3. User Allocation                                         │
│     ├─ 用户请求分配 Testbed                                  │
│     ├─ 验证分配成功                                          │
│     └─ 验证 Testbed 状态                                     │
│                                                              │
│  4. Lifecycle Management                                    │
│     ├─ 等待过期                                              │
│     ├─ 验证自动释放                                          │
│     └─ 验证快照回滚                                          │
│                                                              │
│  5. Cleanup                                                  │
│     └─ 清理测试数据                                          │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**E2E 测试用例：**

```go
// test/e2e/lifecycle_test.go
func TestTestbedLifecycle_E2E(t *testing.T) {
    // 完整生命周期测试
    // 初始化 → 配置 → 分配 → 使用 → 释放 → 回滚
}

// test/e2e/concurrent_test.go
func TestConcurrentAllocation_E2E(t *testing.T) {
    // 多用户并发分配测试
    // 验证配额限制正确生效
}

// test/e2e/auto_expire_test.go
func TestAutoExpire_E2E(t *testing.T) {
    // 自动过期测试
    // 验证后台定时任务正确运行
}
```

### 11.5 Performance Testing (性能测试)

**测试指标：**

| 指标 | 目标值 | 测试方法 |
|------|--------|----------|
| API 响应时间 | < 200ms (P95) | 压力测试 |
| 并发分配 | 100 req/s | 负载测试 |
| 数据库查询 | < 50ms (P95) | 查询分析 |
| 快照回滚 | < 5min | 端到端计时 |

**性能测试用例：**

```go
// test/performance/benchmark_test.go
func BenchmarkAcquireTestbed(b *testing.B) {
    // 基准测试: 分配操作性能
}

func BenchmarkConcurrentAcquire(b *testing.B) {
    // 并发基准测试
}

// test/performance/load_test.go
func TestLoadAllocation(t *testing.T) {
    // 模拟 100 并发用户持续分配
    // 持续 10 分钟
    // 监控响应时间和错误率
}
```

### 11.6 Test Coverage Requirements (覆盖率要求)

**覆盖率目标：**

| 类型 | 语句覆盖率 | 分支覆盖率 | 函数覆盖率 |
|------|-----------|-----------|-----------|
| 总体 | > 75% | > 65% | > 80% |
| 核心业务逻辑 | > 85% | > 75% | > 90% |
| API Handlers | > 80% | > 70% | > 85% |
| Models | > 80% | > 70% | > 85% |
| Storage | > 85% | > 75% | > 90% |

**覆盖率检查：**

```bash
# 运行测试并生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 检查覆盖率是否达标
go test ./... -coverprofile=coverage.out | grep TOTAL
```

### 11.7 Mock/Stub Strategy (Mock 策略)

**需要 Mock 的外部依赖：**

| 依赖 | Mock 方式 | 使用场景 |
|------|-----------|----------|
| 云平台 API | Interface + MockImpl | 单元测试、集成测试 |
| Event-Processor API | HTTP Test Server | Session 验证测试 |
| Deployer | MockDeployer | 部署流程测试 |
| 定时任务 | Manual Trigger | 自动过期测试 |

**Mock 接口定义：**

```go
// mock_deployer_test.go
type MockDeployer struct {
    mock.Mock
}

func (m *MockDeployer) DeployProduct(ctx context.Context, req DeployRequest) (*DeployResult, error) {
    args := m.Called(ctx, req)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*DeployResult), args.Error(1)
}

func (m *MockDeployer) RestoreSnapshot(ctx context.Context, resourceUUID, snapshotID string) error {
    args := m.Called(ctx, resourceUUID, snapshotID)
    return args.Error(0)
}

// 使用示例
func TestReleaseTestbed_WithMockDeployer(t *testing.T) {
    mockDeployer := new(MockDeployer)
    mockDeployer.On("RestoreSnapshot", mock.Anything, "uuid-123", "snap-456").Return(nil)

    // 运行测试
    // ...

    mockDeployer.AssertExpectations(t)
}
```

### 11.8 Test Data Management (测试数据管理)

**测试数据原则：**
1. **独立性**: 每个测试使用独立数据，不共享
2. **清理**: 测试结束后清理数据
3. **可重复**: 使用固定数据，保证结果可重复
4. **隔离**: 使用独立测试数据库

**测试数据工厂：**

```go
// test/factory/factory.go
package factory

type TestFixture struct {
    DB *sql.DB
}

func (f *TestFixture) CreateCategory(name string) *Category {
    category := &Category{
        UUID: generateUUID(),
        Name: name,
        Description: "Test category",
        Enabled: true,
    }
    f.DB.Insert(category)
    return category
}

func (f *TestFixture) CreateTestbed(categoryUUID string, status string) *Testbed {
    // 创建测试用 Testbed
}

func (f *TestFixture) CreateAllocation(testbedUUID, requester string) *Allocation {
    // 创建测试用 Allocation
}

// 使用示例
func TestAcquireTestbed(t *testing.T) {
    fixture := NewTestFixture(testDB)
    defer fixture.Cleanup()

    category := fixture.CreateCategory("test-main")
    testbed := fixture.CreateTestbed(category.UUID, "available")

    // 运行测试
    // ...
}
```

### 11.9 CI/CD Integration (持续集成)

**测试流水线：**

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run unit tests
        run: go test ./... -short -coverprofile=coverage.out

      - name: Check coverage
        run: |
          coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          if (( $(echo "$coverage < 75" | bc -l) )); then
            echo "Coverage $coverage% is below 75%"
            exit 1
          fi

      - name: Run integration tests
        run: go test ./... -tags=integration

      - name: Upload coverage
        uses: codecov/codecov-action@v3
```

### 11.10 Testing Checklist (测试检查清单)

**开发前：**
- [ ] 明确功能需求和边界条件
- [ ] 设计测试用例覆盖正常/异常流程
- [ ] 确定测试数据和 Mock 策略

**开发中：**
- [ ] 遵循 TDD，先写测试再写实现
- [ ] 每个函数/方法都有对应测试
- [ ] 异常流程都有测试覆盖

**开发后：**
- [ ] 所有测试通过
- [ ] 覆盖率达到要求
- [ ] 集成测试通过
- [ ] 代码审查通过

**发布前：**
- [ ] E2E 测试通过
- [ ] 性能测试通过
- [ ] 安全测试通过
- [ ] 回归测试通过

## 12. Configuration

### Port Allocation

| 服务 | 容器端口 | 主机端口 | 说明 |
|------|---------|---------|------|
| Resource Pool Backend | 5003 | 5003 | 独立容器 |
| Event-Processor Backend | 5002 | 5002 | 共享后端 |
| Shared Frontend | 80 | 8082 | 共享前端 |
| Shared MySQL | 3306 | 3307 | 共享数据库 |

### Environment Variables

```bash
# Resource Pool Server
POOL_SERVER_PORT=5003

# 数据库连接（通过 Docker 网络访问 event-processor-mysql）
DB_DSN=root:root123456@tcp(event-processor-mysql:3306)/event_processor?parseTime=true

# Event-Processor 集成（通过 Docker 网络访问）
EVENT_PROCESSOR_API=http://event-processor-server:5002

# Deployer 配置
DEPLOYER_TYPE=mock  # mock or tencent
TENCENT_API_KEY=xxx
TENCENT_SECRET_KEY=yyy

# 监控配置
LOG_LEVEL=info
METRICS_ENABLED=true
```

### Deployment (部署方式)

**部署架构：** 与 event-processor 相同，使用单独的 `docker run` 命令部署，不使用 docker-compose。

**容器列表：**

| 容器名称 | 镜像 | 端口映射 | 说明 |
|---------|------|---------|------|
| event-processor-mysql | acr.aishu.cn/dual-mass-engine-gateway/mysql:latest | 3307:3306 | 共享数据库 |
| event-processor-server | event-processor-backend:latest | 5002:5002 | Event-Processor 后端 |
| event-processor-frontend | event-processor-frontend:latest | 8082:80 | 共享前端 |
| resource-pool-server | resource-pool-backend:latest | 5003:5003 | Resource-Pool 后端 |

**部署脚本：** `src/modules/resource-pool/deploy-resource-pool.sh`

**使用方法：**
```bash
cd src/modules/resource-pool
./deploy-resource-pool.sh              # 升级模式（默认）
./deploy-resource-pool.sh -u           # 升级模式
./deploy-resource-pool.sh -r           # 完全重装（删除容器）
./deploy-resource-pool.sh -h           # 帮助信息
```

**部署流程：**
```
1. 编译 Go 代码（静态链接）
   └─ go build -ldflags '-extldflags "-static"' -o output/resource-pool ./cmd/server

2. 构建 Vue 前端
   ├─ npm install
   └─ npm run build

3. 构建 Docker 镜像
   ├─ docker build -f Dockerfile_server -t resource-pool-backend:latest
   └─ docker build -f Dockerfile_frontend -t resource-pool-frontend:latest

4. 启动容器
   ├─ docker run --name resource-pool-server ...
   └─ docker run --name resource-pool-frontend ...
```

**Docker 网络配置：**
```bash
# 共享网络（与 event-processor 同一网络）
NETWORK_NAME="processor-network"
docker network create "$NETWORK_NAME" 2>/dev/null || true

# Resource-Pool 容器加入该网络
docker run --network "$NETWORK_NAME" ...
```

**环境变量配置：**
```bash
# 数据库连接（通过 Docker 网络访问）
DB_DSN="root:root123456@tcp(event-processor-mysql:3306)/event_processor?parseTime=true"

# Event-Processor API（通过 Docker 网络访问）
EVENT_PROCESSOR_API="http://event-processor-server:5002"

# Deployer 配置
DEPLOYER_TYPE=mock  # or tencent
TENCENT_API_KEY=xxx
TENCENT_SECRET_KEY=yyy
```

## 13. Risk & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Resource exhaustion | High | Quota limits, queue system, alerts |
| Stuck testbeds | Medium | Auto-cleanup after timeout, health checks |
| Concurrent acquire conflicts | Medium | Database transactions, retry logic |
| Provisioning failures | High | Fallback to manual recovery, alerts |
| Snapshot restore failures | Medium | Mark testbed for manual review |

## 14. Open Questions

1. **Snapshot Management**: How are base snapshots created and updated?
2. **Access Key Security**: How to encrypt/decrypt access keys securely?
3. **Session Validation**: event-processor 的 session 验证 API 端点详情
4. **Cleanup Policy**: How long to keep allocation history?
5. **Borrowing Logic**: Should borrowing be implemented in MVP or Phase 2?
