# 数据迁移指南

本文档说明如何从旧系统数据库迁移数据到新的微服务数据库。

## 环境配置

### 旧数据库（Event Processor / Resource Pool）
- 主机: `localhost` (或 `10.4.174.125`)
- 端口: `3307`
- 数据库: `quality_db`
- 用户: `root`

### 新数据库（微服务）
- 主机: `localhost`
- 端口: `3307`
- 数据库:
  - `event_store_db` - Event Store 服务
  - `task_scheduler_db` - Task Scheduler 服务
  - `resource_manager_db` - Resource Manager 服务

## 迁移步骤

### 1. 准备阶段

确保所有新微服务的数据库已创建并初始化 schema：

```bash
# 启动各服务以自动创建数据库和表
cd /root/dev/dual-mass-engine-gateway/services

# 启动 Event Store
cd event-store && ./deploy.sh start

# 启动 Task Scheduler
cd ../task-scheduler && ./deploy.sh start

# 启动 Resource Manager
cd ../resource-manager && ./deploy.sh start
```

### 2. 执行迁移

设置环境变量（可选）：

```bash
# 旧数据库配置
export OLD_DB_HOST="localhost"
export OLD_DB_PORT="3307"
export OLD_DB_USER="root"
export OLD_DB_PASS=""
export OLD_DB_NAME="quality_db"

# 新数据库配置
export NEW_DB_HOST="localhost"
export NEW_DB_PORT="3307"
export NEW_DB_USER="root"
export NEW_DB_PASS=""
```

运行迁移脚本：

```bash
cd /root/dev/dual-mass-engine-gateway/services
./migrate-data.sh
```

### 3. 验证数据

运行验证脚本检查数据完整性：

```bash
./verify-migration.sh
```

预期输出：
```
events: 旧=100, 新=100 ... ✓ 匹配
quality_checks: 旧=500, 新=500 ... ✓ 匹配
...
```

### 4. 回滚（如需要）

如果迁移出现问题，可以回滚：

```bash
./rollback-migration.sh
```

## 数据映射

### Event Store (event_store_db)

| 旧表 | 新表 | 说明 |
|------|------|------|
| events | events | 直接映射 |
| quality_checks | quality_checks | 直接映射 |

### Task Scheduler (task_scheduler_db)

| 旧表 | 新表 | 说明 |
|------|------|------|
| tasks | tasks | 直接映射 |
| task_results | task_results | 直接映射 |
| task_executions | task_executions | 直接映射 |

### Resource Manager (resource_manager_db)

| 旧表 | 新表 | 说明 |
|------|------|------|
| users | users | 直接映射 |
| resource_pool_categories | categories | 字段调整 |
| resource_pool_quota_policies | quota_policies | 直接映射 |
| resource_pool_testbeds | testbeds | 直接映射 |
| resource_pool_instances | resource_instances | 直接映射 |
| resource_pool_allocations | allocations | 直接映射 |
| deployment_tasks | deployment_tasks | 直接映射 |
| sessions | sessions | 直接映射 |

## 备份

迁移脚本会自动在 `./migrations/backup/<timestamp>/` 目录下创建 SQL 备份文件。

备份文件命名：
- `events.sql`
- `quality_checks.sql`
- `tasks.sql`
- `task_results.sql`
- `task_executions.sql`
- `users.sql`
- `categories.sql`
- `quota_policies.sql`
- `testbeds.sql`
- `resource_instances.sql`
- `allocations.sql`
- `deployment_tasks.sql`
- `sessions.sql`

## 故障排查

### 问题: 无法连接到数据库

检查数据库是否运行：
```bash
docker ps | grep mysql
# 或
systemctl status mysql
```

### 问题: 表不存在

确保服务已启动并创建了 schema：
```bash
cd services/<service>
./deploy.sh status
```

### 问题: 外键约束错误

迁移脚本会自动处理外键约束，如果仍有问题，请检查：
1. 表的创建顺序
2. 外键约束的定义
