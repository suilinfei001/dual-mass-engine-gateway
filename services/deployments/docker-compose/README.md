# 双引擎质量网关 - Docker Compose 部署

## 概述

这是双引擎质量网关微服务架构的 Docker Compose 开发环境配置。

## 服务列表

| 服务 | 端口 | 描述 |
|------|------|------|
| webhook-gateway | 4001 | 接收 GitHub/GitLab Webhook |
| event-store | 4002 | 事件存储和查询 |
| task-scheduler | 4003 | 任务调度和状态管理 |
| executor-service | 4004 | Azure DevOps Pipeline 执行 |
| ai-analyzer | 4005 | AI 日志分析 |
| resource-manager | 4006 | 资源池管理 |

## 数据库端口

| 数据库 | 端口 | 描述 |
|--------|------|------|
| event-store-mysql | 3308 | Event Store 数据库 |
| task-scheduler-mysql | 3309 | Task Scheduler 数据库 |
| resource-manager-mysql | 3310 | Resource Manager 数据库 |

## 快速开始

### 一键启动所有服务

```bash
cd services/deployments/docker-compose
./start-all.sh
```

这将：
1. 编译所有微服务（本地编译，静态链接）
2. 拉取所需的 Docker 镜像
3. 启动所有服务和数据库

### 仅启动（跳过编译）

```bash
./start-all.sh --no-build
```

### 停止服务

```bash
# 停止服务（保留数据和容器）
./stop-all.sh

# 停止并删除容器（保留数据）
./stop-all.sh --clean

# 停止并删除所有内容（包括数据）
./stop-all.sh --purge
```

## 单独操作

### 编译指定服务

```bash
./build-all.sh webhook-gateway
./build-all.sh event-store task-scheduler
```

### 查看服务状态

```bash
docker-compose ps
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看指定服务日志
docker-compose logs -f webhook-gateway
docker-compose logs -f task-scheduler
```

### 重启指定服务

```bash
docker-compose restart webhook-gateway
```

## 目录结构

```
docker-compose/
├── docker-compose.yml    # Docker Compose 配置
├── build-all.sh          # 编译所有服务
├── start-all.sh          # 一键启动脚本
├── stop-all.sh           # 停止脚本
└── README.md             # 本文档
```

## 开发流程

1. **修改代码后重新部署**

   ```bash
   # 方式 1: 完全重新启动
   ./start-all.sh

   # 方式 2: 仅重新编译并重启指定服务
   ./build-all.sh webhook-gateway
   docker-compose up -d webhook-gateway
   ```

2. **查看服务健康状态**

   ```bash
   docker-compose ps
   curl http://localhost:4001/api/health
   curl http://localhost:4002/api/health
   curl http://localhost:4003/api/health
   ```

3. **进入容器调试**

   ```bash
   docker-compose exec webhook-gateway sh
   docker-compose exec event-store sh
   ```

## 网络架构

所有服务都在 `quality-network` Docker 网络中，可以通过服务名互相访问：

- `webhook-gateway` 访问 `http://event-store:4002`
- `task-scheduler` 访问 `http://event-store:4002`
- `task-scheduler` 访问 `http://executor-service:4004`
- `task-scheduler` 访问 `http://ai-analyzer:4005`

## 数据持久化

数据库数据存储在 Docker 卷中：

- `event-store-mysql-data`
- `task-scheduler-mysql-data`
- `resource-manager-mysql-data`

使用 `./stop-all.sh --purge` 可以删除这些卷。

## 故障排查

### 服务启动失败

```bash
# 查看详细日志
docker-compose logs [service-name]

# 检查服务是否在运行
docker-compose ps
```

### 数据库连接失败

```bash
# 检查数据库是否健康
docker-compose exec event-store-mysql mysqladmin ping -h localhost -uroot -proot123456
```

### 完全重置环境

```bash
# 停止并删除所有内容
./stop-all.sh --purge

# 重新启动
./start-all.sh
```
