# 前端 API 配置

微服务架构的统一前端 API 配置模块。

## 安装

```bash
npm install @microservices/frontend-config
```

## 使用方式

### 方式一：直接导入配置

```javascript
import { SERVICES, EventStoreAPI, TaskSchedulerAPI } from '@microservices/frontend-config'

// 获取事件列表
const response = await fetch(EventStoreAPI.events())
```

### 方式二：使用 URL 转换（兼容旧代码）

```javascript
import { getApiUrl } from '@microservices/frontend-config'

// 旧路径自动转换为新的微服务端点
const oldUrl = '/api/events'
const newUrl = getApiUrl(oldUrl)  // -> 'http://localhost:4002/api/events'
```

### 方式三：更新 Vite 配置

```javascript
// vite.config.js
import { defineConfig } from 'vite'
import { SERVICES } from '@microservices/frontend-config'

export default defineConfig({
  server: {
    proxy: {
      '/api/events': {
        target: SERVICES.eventStore,
        changeOrigin: true
      },
      '/api/tasks': {
        target: SERVICES.taskScheduler,
        changeOrigin: true
      },
      '/api/resources': {
        target: SERVICES.resourceManager,
        changeOrigin: true
      }
    }
  }
})
```

## 服务端点

| 服务 | 端口 | 说明 |
|------|------|------|
| Event Store | 4002 | 事件存储和查询 |
| Task Scheduler | 4003 | 任务调度 |
| Executor Service | 4004 | Pipeline 执行 |
| AI Analyzer | 4005 | AI 日志分析 |
| Resource Manager | 4006 | 资源池管理 |
| Webhook Gateway | 4001 | Webhook 接收 |

## 环境配置

- **开发环境**: 直接连接 `localhost:*`
- **生产环境**: 通过 Nginx 代理或内网地址

## API 迁移映射

| 旧路径 | 新服务 |
|--------|--------|
| `/api/events` | Event Store (4002) |
| `/api/tasks` | Task Scheduler (4003) |
| `/api/resource-pool` | Resource Manager (4006) |
| `/api/analyze` | AI Analyzer (4005) |
| `/api/execute` | Executor Service (4004) |
