# 双引擎质量网关系统 - 项目结构

## 概述

本项目是一个双引擎质量网关系统，用于保证 GitHub 开源项目的代码质量。系统由两个独立模块组成：

1. **事件接收器 (Event Receiver)** - 部署在外网，接收 GitHub Webhook 事件
2. **事件处理器 (Event Processor)** - 部署在内网，处理质量检查任务

## 目录结构

```
dual-mass-engine-gateway/
├── src/
│   └── modules/
│       ├── event-receiver/          # 事件接收器模块
│       │   ├── cmd/                  # 后端入口
│       │   │   └── quality-server/   # Go 后端服务
│       │   ├── internal/             # 内部代码
│       │   │   └── quality/          # 质量检查相关代码
│       │   │       ├── api/          # API 处理器
│       │   │       ├── handlers/     # Webhook 处理器
│       │   │       ├── models/       # 数据模型
│       │   │       ├── storage/      # 数据存储
│       │   │       └── logger/       # 日志系统
│       │   ├── frontend/             # 前端代码
│       │   │   ├── src/
│       │   │   ├── dist/             # 构建输出
│       │   │   ├── package.json
│       │   │   └── vite.config.js
│       │   ├── configs/              # 配置文件
│       │   ├── Dockerfile_frontend   # 前端 Dockerfile
│       │   ├── Dockerfile_quality_server  # 后端 Dockerfile
│       │   ├── nginx.conf            # Nginx 配置
│       │   └── go.mod                # Go 模块定义
│       │
│       └── event-processor/          # 事件处理器模块
│           ├── cmd/                  # 后端入口（待开发）
│           ├── internal/             # 内部代码（待开发）
│           ├── frontend/             # 前端代码
│           ├── configs/              # 配置文件
│           ├── Dockerfile_frontend   # 前端 Dockerfile
│           ├── Dockerfile_server     # 后端 Dockerfile
│           └── go.mod                # Go 模块定义
│
├── install/                          # 安装脚本
│   ├── install_quality.sh            # 事件接收器部署脚本
│   ├── install_event_processor.sh    # 事件处理器部署脚本
│   └── scripts/                      # 辅助脚本
│       └── init-mysql.sql            # MySQL 初始化脚本
│
├── tests/                            # 测试脚本
│   ├── test.sh                       # 单元测试脚本
│   ├── loadtest.sh                   # 压力测试脚本
│   └── loadtest/                     # 压力测试代码
│
├── docs/                             # 文档
│   ├── project_refactor.txt          # 重构需求文档
│   └── PROJECT_STRUCTURE.md          # 本文档
│
├── data/                             # 数据目录（运行时生成）
│   ├── quality-mysql/                # MySQL 数据
│   └── quality-server/               # 服务数据
│
├── go.mod                            # 根 Go 模块（保留）
├── go.sum                            # 依赖锁定
├── README.md                         # 项目说明
└── README.zh.md                      # 中文说明
```

## 模块说明

### 事件接收器 (Event Receiver)

**部署位置**: 外网 (10.4.111.141)

**功能**:
- 提供 Webhook 接收 GitHub Action 事件
- 接收主线 push 事件和 PR 事件
- 创建质量卡点并存储到 MySQL 数据库
- 提供网站查看事件和质量卡点状态

**技术栈**:
- 后端: Go 1.21+
- 前端: Vue 3 + Vite
- 数据库: MySQL
- Web 服务器: Nginx

**端口**:
- 前端: 8081
- 后端 API: 5001
- MySQL: 3306

### 事件处理器 (Event Processor)

**部署位置**: 内网

**功能**:
- 从事件接收器获取待处理任务
- 执行质量检查（待开发）
- 将结果返回给事件接收器（待开发）

**技术栈**:
- 后端: Go 1.21+ (待开发)
- 前端: Vue 3 + Vite
- API: REST API 与事件接收器通信

**端口**:
- 前端: 8082
- 后端 API: 5002

## 部署说明

### 事件接收器部署

```bash
cd /root/dev/dual-mass-engine-gateway
./install/install_quality.sh              # 升级模式
./install/install_quality.sh -u           # 升级模式
./install/install_quality.sh -r           # 恢复模式（完全重装）
./install/install_quality.sh -h           # 帮助
```

### 事件处理器部署

```bash
cd /root/dev/dual-mass-engine-gateway
./install/install_event_processor.sh      # 部署事件处理器
```

## 测试

### 运行单元测试

```bash
cd /root/dev/dual-mass-engine-gateway
./tests/test.sh                           # 运行所有测试
./tests/test.sh -v                        # 详细输出
./tests/test.sh -r -c                     # 竞态检测 + 覆盖率
```

### 运行压力测试

```bash
cd /root/dev/dual-mass-engine-gateway
./tests/loadtest.sh
```

## 开发指南

### 事件接收器开发

```bash
cd src/modules/event-receiver

# 后端开发
cd cmd/quality-server
go run main.go -db "user:pass@tcp(localhost:3306)/db"

# 前端开发
cd frontend
npm install
npm run dev        # 开发模式
npm run build      # 构建生产版本
```

### 事件处理器开发

```bash
cd src/modules/event-processor

# 前端开发
cd frontend
npm install
npm run dev        # 开发模式
npm run build      # 构建生产版本

# 后端开发（待实现）
# cd cmd/server
# go run main.go
```

## 网络架构

```
┌─────────────────────────────────────────────────────────────┐
│                        外网环境                               │
│  ┌──────────────────┐         ┌──────────────────┐          │
│  │   GitHub         │────────>│ Event Receiver   │          │
│  │   (Webhook)      │         │  :5001           │          │
│  └──────────────────┘         │  MySQL:3306      │          │
│                               │  Frontend:8081   │          │
│                               └──────────────────┘          │
│                                        │                     │
│                                        │ REST API            │
│                                        ▼                     │
└─────────────────────────────────────────────────────────────┘
                          │
                          │ 内网通信
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                        内网环境                               │
│  ┌──────────────────┐                                       │
│  │ Event Processor  │                                       │
│  │  :5002           │                                       │
│  │  Frontend:8082   │                                       │
│  └──────────────────┘                                       │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## 注意事项

1. 事件接收器和事件处理器使用不同的数据库
2. 事件处理器只能通过 REST API 与事件接收器通信
3. 两个模块共享同一个代码仓库，但独立部署
4. 配置文件分别存放在各自的 `configs/` 目录下
