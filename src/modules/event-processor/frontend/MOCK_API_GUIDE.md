# Resource Pool Mock API 使用指南

## 简介

Resource Pool Mock API 是用于前端开发和测试的模拟后端服务，无需启动实际的后端服务即可测试前端功能。

## 启用 Mock 模式

有两种方式可以启用 Mock 模式：

### 方式一：URL 参数（推荐用于开发测试）

在浏览器地址栏中添加 `?mock=true` 或 `#mock=true` 参数：

```
http://localhost:8082/resource-pool/testbeds?mock=true
```

### 方式二：LocalStorage（持久化）

在浏览器控制台中执行：

```javascript
import { enableMockMode } from './api/resourcePool'
enableMockMode()
```

或者直接设置 localStorage：

```javascript
localStorage.setItem('USE_MOCK_API', 'true')
window.location.reload()
```

## 禁用 Mock 模式

```javascript
import { disableMockMode } from './api/resourcePool'
disableMockMode()
```

或者直接清除 localStorage：

```javascript
localStorage.removeItem('USE_MOCK_API')
window.location.reload()
```

## Mock 数据

### 预置类别

| UUID | 名称 | 描述 |
|------|------|------|
| cat-mysql-001 | MySQL 测试环境 | 用于 MySQL 相关测试的数据库环境 |
| cat-pgsql-001 | PostgreSQL 测试环境 | 用于 PostgreSQL 相关测试的数据库环境 |
| cat-redis-001 | Redis 缓存环境 | 用于 Redis 缓存测试的环境 |

### 预置资源实例

| UUID | 名称 | 主机 | 状态 |
|------|------|------|------|
| res-vm-001 | vm-testbed-01 | 192.168.1.101 | available |
| res-vm-002 | vm-testbed-02 | 192.168.1.102 | available |
| res-vm-003 | vm-testbed-03 | 192.168.1.103 | in_use |
| res-vm-004 | vm-testbed-04 | 192.168.1.104 | available |

### 预置 Testbed

| UUID | 名称 | 类别 | 状态 |
|------|------|------|------|
| tb-mysql-001 | mysql-testbed-01 | MySQL 测试环境 | available |
| tb-mysql-002 | mysql-testbed-02 | MySQL 测试环境 | in_use |
| tb-pgsql-001 | pgsql-testbed-01 | PostgreSQL 测试环境 | available |
| tb-redis-001 | redis-testbed-01 | Redis 缓存环境 | available |

## API 方法支持

### Internal API

- `getTestbeds(params)` - 获取所有 testbed
- `getTestbed(uuid)` - 获取单个 testbed
- `getAvailableTestbeds(params)` - 获取可用的 testbed

### External API

- `acquireTestbed(categoryUUID, options)` - 申请 testbed
- `getMyAllocations(params)` - 获取我的分配
- `getAllocation(uuid)` - 获取分配详情
- `extendAllocation(uuid, extendSeconds)` - 延期分配
- `releaseTestbed(allocationUUID)` - 释放 testbed
- `getCategories()` - 获取所有类别
- `getCategory(uuid)` - 获取类别详情
- `getQuotaPolicy(categoryUUID)` - 获取配额策略

### Admin API

- `getAllTestbeds(params)` - 获取所有 testbed（管理员视图）
- `createTestbed(data)` - 创建 testbed（模拟部署）
- `updateTestbed(uuid, data)` - 更新 testbed
- `deleteTestbed(uuid)` - 删除 testbed
- `getResourceInstances(params)` - 获取资源实例
- `createResourceInstance(data)` - 创建资源实例
- `updateResourceInstance(uuid, data)` - 更新资源实例
- `deleteResourceInstance(uuid)` - 删除资源实例
- `getCategories()` - 获取所有类别
- `createCategory(data)` - 创建类别
- `updateCategory(uuid, data)` - 更新类别
- `deleteCategory(uuid)` - 删除类别
- `getQuotaPolicies()` - 获取所有配额策略
- `getQuotaPolicy(categoryUUID)` - 获取配额策略
- `updateQuotaPolicy(uuid, data)` - 更新配额策略
- `getAllAllocations(params)` - 获取所有分配
- `getAllocationHistory(params)` - 获取分配历史
- `getMetrics()` - 获取指标
- `getUsageStats(params)` - 获取使用统计
- `deployToResourceInstance(resourceInstanceUUID, options)` - 部署到资源实例

## 部署模拟

`deployToResourceInstance` 方法模拟了部署过程：

- **部署时间**: 5 秒
- **进度日志**: 每 1 秒输出一次进度 (20%, 40%, 60%, 80%, 100%)
- **返回结果**: 新创建的 Testbed 对象

### 部署示例

```javascript
import { adminAPI } from './api/resourcePool'

const result = await adminAPI.deployToResourceInstance('res-vm-001', {
  category_uuid: 'cat-mysql-001',
  db_port: 3306,
  db_user: 'root',
  db_password: 'Test@123456'
})

console.log(result.data)
// {
//   uuid: 'tb-mock-1234567890',
//   name: 'vm-testbed-01-deployed',
//   status: 'available',
//   ...
// }
```

## 注意事项

1. **数据不持久化**: Mock 数据存储在内存中，刷新页面后会重置
2. **状态模拟**: 部分状态转换被简化，可能与实际后端行为有差异
3. **并发处理**: Mock API 不支持真正的并发控制
4. **认证跳过**: Mock 模式下不进行真正的认证检查

## 调试

在浏览器控制台中可以查看 Mock API 的日志：

```
[Mock Deploy] Starting deployment to res-vm-001...
[Mock Deploy] Deployment progress: 20%
[Mock Deploy] Deployment progress: 40%
[Mock Deploy] Deployment progress: 60%
[Mock Deploy] Deployment progress: 80%
[Mock Deploy] Deployment progress: 100%
[Mock Deploy] Deployment completed successfully: tb-mock-1234567890
```
