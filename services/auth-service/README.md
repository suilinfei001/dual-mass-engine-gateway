# Auth Service - 认证服务

独立认证服务，为双引擎质量网关系统提供统一的登录/登出/会话管理功能。

## 功能特性

- 用户名密码认证
- 基于 Cookie 的会话管理（HttpOnly，24小时过期）
- 会话过期自动清理
- CORS 支持跨域访问
- 环境变量配置管理员凭证

## API 端点

支持两种路由格式：
- 标准格式: `/api/login`, `/api/logout`, `/api/check-login`
- 兼容格式: `/api/auth/login`, `/api/auth/logout`, `/api/auth/check-login`, `/api/auth/status`

### POST /api/login (或 /api/auth/login)
用户登录

**请求:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应:**
```json
{
  "success": true,
  "message": "登录成功",
  "username": "admin"
}
```

### GET /api/check-login (或 /api/auth/check-login)
检查登录状态

**响应:**
```json
{
  "is_logged_in": true,
  "username": "admin"
}
```

### POST /api/logout (或 /api/auth/logout)
用户登出

**响应:**
```json
{
  "success": true,
  "message": "登出成功"
}
```

### GET /health
健康检查

**响应:**
```json
{
  "status": "ok"
}
```

## 配置

### 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `4007` | 服务端口 |
| `ADMIN_USER` | `admin` | 管理员用户名 |
| `ADMIN_PASS` | `admin123` | 管理员密码 |
| `LOG_LEVEL` | `info` | 日志级别 |

### CORS 配置

默认允许的源：
- `http://localhost:8081`
- `http://localhost:8082`
- `http://localhost:8083`
- `http://10.4.111.141:8081`
- `http://10.4.111.141`

## 开发

### 编译
```bash
CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o auth-service ./cmd/server
```

### 测试
```bash
go test ./... -v
```

### 运行
```bash
./auth-service
```

## 部署

### 本地部署
```bash
./deploy.sh
```

### Docker 部署
```bash
docker build -t auth-service:latest .
docker run -d \
  --name auth-service \
  --network quality-gateway \
  -p 4007:4007 \
  -e ADMIN_USER=admin \
  -e ADMIN_PASS=admin123 \
  --restart unless-stopped \
  auth-service:latest
```

## 前端集成

前端通过 Vite 代理访问认证服务：

```javascript
// vite.config.js
export default defineConfig({
  server: {
    proxy: {
      '/api/auth': {
        target: 'http://localhost:4007',
        changeOrigin: true
      },
      '/api/login': {
        target: 'http://localhost:4007',
        changeOrigin: true
      },
      '/api/logout': {
        target: 'http://localhost:4007',
        changeOrigin: true
      },
      '/api/check-login': {
        target: 'http://localhost:4007',
        changeOrigin: true
      }
    }
  }
})
```

前端调用示例：

```javascript
// 登录
const response = await fetch('/api/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  credentials: 'include',
  body: JSON.stringify({ username, password })
})

// 检查登录状态
const response = await fetch('/api/check-login', {
  credentials: 'include'
})

// 登出
const response = await fetch('/api/logout', {
  method: 'POST',
  credentials: 'include'
})
```

## 安全说明

1. **Cookie 安全**: 使用 `HttpOnly` 防止 XSS 攻击，`SameSite=Lax` 防止 CSRF 攻击
2. **会话过期**: 会话默认 24 小时过期，自动清理
3. **密码**: 生产环境务必修改默认密码

## 共享库

本服务使用 `shared/pkg/auth` 提供的核心认证功能：
- `SessionStore`: 会话存储
- `ValidateCredentials`: 凭证验证
- `AuthMiddleware`: 认证中间件
