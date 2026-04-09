# ResourceInstance 与 Testbed 生命周期管理

## 概述

资源池系统中包含两个核心实体：
- **ResourceInstance（资源实例）**：持久化的物理机或虚拟机，可反复创建 Testbed
- **Testbed（测试床）**：一次性使用的测试环境，用完即焚

## 核心设计理念

```
ResourceInstance (持久)
        │
        │ 部署产品 (Provisioning)
        ▼
    Testbed (一次性)
        │
        │ 使用后释放/过期
        ▼
     回滚快照
        │
        ▼
ResourceInstance (恢复干净，可再次创建 Testbed)
```

## ResourceInstance 生命周期

### 状态定义

| 状态 | 说明 | 触发条件 |
|------|------|----------|
| `pending` | 新创建，尚未激活 | 初始创建 |
| `active` | 健康，可用于创建 Testbed | 健康检查成功 |
| `unreachable` | 不可达 | 健康检查失败 |

### 状态转换图

```
    ┌─────────┐
    │ pending │ 新创建
    └────┬────┘
         │ 首次健康检查成功
         ▼
    ┌─────────┐
    │ active  │ 健康，可用
    └────┬────┘
         │ 健康检查失败
         ▼
    ┌─────────────┐
    │unreachable  │ 不可达
    └────┬────────┘
         │ 健康检查恢复
         ▼
    ┌─────────┐
    │ active  │
    └─────────┘
```

### 关键特性

- **持久化**：ResourceInstance 不会被删除，可以反复使用
- **快照隔离**：每个 VM 类型实例有初始快照，回滚后恢复干净状态
- **健康监控**：HealthCheckJob 每 5 分钟检查一次健康状态

## Testbed 生命周期

### 状态定义

| 状态 | 说明 | 触发条件 |
|------|------|----------|
| `available` | 可用，可被分配 | 创建成功 |
| `allocated` | 已分配，等待使用 | 用户获取 |
| `in_use` | 使用中 | 用户开始使用 |
| `releasing` | 释放中，快照回滚 | 主动释放或过期 |
| `deleted` | 已删除，不可恢复 | 回滚完成 |

### 状态转换图

```
    创建 (Provisioning)
         │
         ▼
    ┌──────────┐
    │available │ 可被分配
    └────┬─────┘
         │ 用户获取
         ▼
    ┌──────────┐
    │allocated │ 已分配
    └────┬─────┘
         │ 开始使用
         ▼
    ┌──────────┐
    │  in_use  │ 使用中
    └────┬─────┘
         │
    ┌────┴────────┐
    ▼               ▼
主动释放        过期
    │               │
    └───────┬───────┘
            ▼
      ┌──────────┐
      │releasing │ 快照回滚中
      └────┬─────┘
           │ 回滚完成
           ▼
      ┌──────────┐
      │ deleted  │ 终态，不可逆
      └──────────┘
```

### 关键特性

- **一次性**：Testbed 使用后即删除，不会回到 `available` 状态
- **隔离性**：每个 Testbed 有独立的环境和数据
- **自动回收**：过期后由 AutoExpireJob 自动处理

## 转化流程

### ResourceInstance → Testbed (Provisioning)

```
┌─────────────────────────────────────────────────────────────────┐
│                       Provisioning 流程                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  1. ReplenishJob 检测到可用 Testbed 数量低于阈值                        │
│     或                                                          │
│     2. 管理员手动创建                                              │
│                                                                 │
│     ↓                                                           │
│                                                                 │
│  ┌─────────────────┐     ┌──────────────────┐                      │
│  │ ResourceInstance │     │   Category       │                      │
│  │   (active)       │◄────│ (service_target) │                      │
│  │  type: VM        │     │  robot/normal    │                      │
│  │  有 snapshot_id   │     └──────────────────┘                      │
│  └────────┬─────────┘                                             │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │  DeployTask      │                                           │
│  │  type: deploy    │                                           │
│  └────────┬─────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐                                           │
│  │  DeployService  │                                           │
│  │  SSH + 部署产品  │                                           │
│  └────────┬─────────┘                                           │
│           │                                                     │
│           ▼                                                     │
│  ┌─────────────────┐     ┌──────────────────┐                     │
│  │    Testbed       │────►│ ResourceInstance │                     │
│  │   (available)    │     │   (仍 active)    │                     │
│  │                   │     │   被占用          │                     │
│  │ 包含连接信息等     │     └──────────────────┘                     │
│  └─────────────────┘                                           │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Testbed → ResourceInstance (Release/Restore)

```
┌─────────────────────────────────────────────────────────────────┐
│                     Release/Restore 流程                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐                                          │
│  │    Testbed       │                                          │
│  │   (in_use)        │                                          │
│  └────────┬─────────┘                                          │
│           │                                                     │
│           │ 主动释放 或 过期                                      │
│           ▼                                                     │
│  ┌──────────────────┐     ┌──────────────────┐                     │
│  │   Allocation     │     │  RollbackTask    │                     │
│  │ released/expired │     │     (async)      │                     │
│  └──────────────────┘     └──────────────────┘                     │
│           │                                                     │
│           ▼                                                     │
│  ┌──────────────────┐                                          │
│  │    Testbed       │                                          │
│  │   (releasing)    │                                          │
│  └────────┬─────────┘                                          │
│           │                                                     │
│           │ 快照回滚 (VM) 或直接完成                               │
│           ▼                                                     │
│  ┌──────────────────┐     ┌──────────────────┐                     │
│  │  DeployService   │────►│ ResourceInstance │                     │
│  │ RestoreSnapshot  │     │   恢复干净       │                     │
│  └──────────────────┘     │   可再次使用      │                     │
│                           └──────────────────┘                     │
│           │                                                     │
│           ▼                                                     │
│  ┌──────────────────┐                                          │
│  │    Testbed       │                                          │
│  │    (deleted)     │ ◄───┐ 终态，不可逆                          │
│  └──────────────────┘     │                                      │
│                             │ Testbed 从列表中移除                │
│                             │ 可通过筛选查看历史记录              │
│                             └──────────────────────────────────┘│
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## 后台任务

### AutoExpireJob（自动过期任务）

- **频率**：每 1 分钟
- **职责**：
  1. 查找所有过期的 Allocation
  2. 标记 Allocation 为 `expired`
  3. 标记关联 Testbed 为 `releasing`
  4. 异步执行快照回滚
  5. 回滚完成后标记 Testbed 为 `deleted`

### ReplenishJob（自动补充任务）

- **频率**：每 1 分钟（启动时立即执行一次）
- **职责**：
  1. 检查每个类别（按 service_target 分组）的可用 Testbed 数量
  2. 当数量低于 `replenish_threshold` 时触发补充
  3. 使用可用的 ResourceInstance 创建新 Testbed
  4. 遵守 `max_instances` 配额限制

### HealthCheckJob（健康检查任务）

- **频率**：每 5 分钟
- **职责**：
  1. 检查所有 ResourceInstance 的健康状态
  2. 通过 SSH/TCP 连接测试可达性
  3. 更新状态为 `active` 或 `unreachable`
  4. 最大并发数：100

## 数据库 Schema

### ResourceInstance 表

```sql
CREATE TABLE resource_instances (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    instance_type ENUM('virtual_machine', 'machine') NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    port INT NOT NULL,
    ssh_user VARCHAR(100),
    ssh_passwd VARCHAR(255),
    snapshot_id CHAR(36),
    status ENUM('pending', 'active', 'unreachable') DEFAULT 'pending',
    ...
);
```

### Testbed 表

```sql
CREATE TABLE testbeds (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,
    name VARCHAR(100) UNIQUE NOT NULL,
    category_uuid CHAR(36) NOT NULL,
    service_target ENUM('robot', 'normal') NOT NULL DEFAULT 'normal',
    resource_instance_uuid CHAR(36) NOT NULL,
    status ENUM('available', 'allocated', 'in_use', 'releasing', 'deleted') DEFAULT 'available',
    ...
);
```

### Allocation 表

```sql
CREATE TABLE allocations (
    id INT PRIMARY KEY AUTO_INCREMENT,
    uuid CHAR(36) UNIQUE NOT NULL,
    testbed_uuid CHAR(36) NOT NULL,
    requester VARCHAR(100) NOT NULL,
    status ENUM('pending', 'active', 'released', 'expired') DEFAULT 'pending',
    expires_at TIMESTAMP NULL,
    released_at TIMESTAMP NULL,
    ...
);
```

## 独立资源池设计

```
                    ┌─────────────────────┐
                    │     Category        │
                    └──────────┬──────────┘
                               │
              ┌────────────────┴────────────────┐
              ▼                                 ▼
    ┌─────────────────────┐         ┌─────────────────────┐
    │ service_target=robot │         │service_target=normal │
    │   (Robot 用户池)     │         │   (普通用户池)       │
    ├─────────────────────┤         ├─────────────────────┤
    │ Testbeds (robot)    │         │ Testbeds (normal)   │
    │ 独立配额管理         │         │ 独立配额管理         │
    │ 互不影响            │         │ 互不影响            │
    └─────────────────────┘         └─────────────────────┘
```

## API 端点

### Testbed 管理

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/admin/testbeds` | 列出所有 Testbed |
| POST | `/admin/testbeds` | 创建 Testbed |
| GET | `/admin/testbeds/{uuid}` | 获取 Testbed 详情 |
| PUT | `/admin/testbeds/{uuid}` | 更新 Testbed |
| DELETE | `/admin/testbeds/{uuid}` | 删除 Testbed |
| GET | `/external/testbeds/{uuid}` | 获取 Testbed 详情（外部） |

### Allocation 管理

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/external/allocations` | 获取我的 Allocation |
| POST | `/external/allocations` | 获取 Testbed |
| GET | `/external/allocations/{uuid}` | 获取 Allocation 详情 |
| DELETE | `/external/allocations/{uuid}` | 释放 Testbed |
| POST | `/external/allocations/{uuid}/extend` | 延长 Allocation |

## 迁移指南

### 从旧版本升级

如果数据库中存在 `maintenance` 状态的 Testbed，运行迁移脚本：

```bash
mysql -u root -p event_processor < install/migrations/001_update_testbed_status.sql
```

这将：
1. 更新 `testbeds.status` 字段，将 `maintenance` 替换为 `deleted`
2. 将现有 `maintenance` 状态的 Testbed 更新为 `available`

## 关键设计优势

1. **资源隔离**：Robot 和普通用户使用完全独立的资源池
2. **用完即焚**：Testbed 一次性使用，避免状态污染
3. **快照隔离**：每个 Testbed 有独立快照，确保环境干净
4. **自动化管理**：后台任务自动处理过期、补充和健康检查
5. **原子操作**：使用 CAS 避免并发竞态条件
6. **异步处理**：耗时的部署/回滚操作异步执行
