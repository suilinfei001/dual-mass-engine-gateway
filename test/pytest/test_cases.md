# 双引擎质量网关系统 - API 接口测试与端到端测试用例

## 测试环境配置

| 服务名称 | 端口 | 基础 URL |
|---------|------|----------|
| webhook-gateway | 4001 | http://localhost:4001 |
| event-store | 4002 | http://localhost:4002 |
| task-scheduler | 4003 | http://localhost:4003 |
| executor-service | 4004 | http://localhost:4004 |
| ai-analyzer | 4005 | http://localhost:4005 |
| resource-manager | 4006 | http://localhost:4006 |

---

## 一、API 接口测试用例

### 1. Webhook Gateway 服务 (端口 4001)

#### 1.1 健康检查接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| WG-001 | 健康检查正常 | GET | /health | 200 | 返回 status=ok, service=webhook-gateway |
| WG-002 | 服务状态查询 | GET | /api/status | 200 | 返回 service, status, timestamp |

#### 1.2 GitHub Webhook 接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| WG-010 | PR Opened 事件 | POST | /webhook/github | GitHub PR payload | 200 | 返回 event_uuid, status=accepted |
| WG-011 | PR Synchronize 事件 | POST | /webhook/github | GitHub PR sync payload | 200 | 返回 event_uuid |
| WG-012 | Push 事件 | POST | /webhook/github | GitHub push payload | 200 | 返回 event_uuid |
| WG-013 | 空 Payload | POST | /webhook/github | {} | 200/500 | 处理空请求 |
| WG-014 | 无效 JSON | POST | /webhook/github | invalid json | 400/500 | 返回错误 |
| WG-015 | 缺少 Event Header | POST | /webhook/github | valid payload (无 X-GitHub-Event) | 200/400 | 验证 header 必要性 |

#### 1.3 GitLab Webhook 接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| WG-020 | MR Opened 事件 | POST | /webhook/gitlab | GitLab MR payload | 200 | 返回 event_uuid |
| WG-021 | Push 事件 | POST | /webhook/gitlab | GitLab push payload | 200 | 返回 event_uuid |
| WG-022 | 无效 Token | POST | /webhook/gitlab | payload with invalid token | 200/401 | 验证 token 校验 |

---

### 2. Event Store 服务 (端口 4002)

#### 2.1 健康检查接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| ES-001 | 健康检查正常 | GET | /health | 200 | 返回 status=ok, service=event-store |

#### 2.2 事件管理接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| ES-010 | 创建事件 | POST | /api/events | event payload | 200/201 | 返回带 uuid 的事件对象 |
| ES-011 | 列出事件 | GET | /api/events | - | 200 | 返回事件数组 |
| ES-012 | 按 status 过滤 | GET | /api/events?status=pending | - | 200 | 只返回 pending 事件 |
| ES-013 | 按 event_type 过滤 | GET | /api/events?event_type=pull_request.opened | - | 200 | 只返回指定类型事件 |
| ES-014 | 分页查询 | GET | /api/events?limit=10&offset=0 | - | 200 | 返回分页结果 |
| ES-015 | 获取单个事件 | GET | /api/events/{uuid} | - | 200/404 | 返回事件详情或 404 |
| ES-016 | 更新事件状态 | PUT | /api/events/{uuid}/status | {"status":"processing"} | 200 | 返回 status=updated |
| ES-017 | 获取待处理事件 | GET | /api/events/pending | - | 200 | 返回 pending 事件列表 |
| ES-018 | 获取处理中事件 | GET | /api/events/processing | - | 200 | 返回 processing 事件列表 |
| ES-019 | 获取事件统计 | GET | /api/events/statistics | - | 200 | 返回各状态事件数量 |
| ES-020 | 事件不存在 | GET | /api/events/non-existent-uuid | - | 404 | 返回 Event not found |
| ES-021 | 空 Payload 创建 | POST | /api/events | {} | 400/500 | 返回错误 |
| ES-022 | 无效 JSON | POST | /api/events | invalid json | 400 | 返回 Invalid request body |

#### 2.3 PR 相关接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| ES-030 | 获取 PR 事件 | GET | /api/repos/{repo_name}/pulls/{pr_number}/events | 200 | 返回 PR 相关事件 |
| ES-031 | 无效 PR 编号 | GET | /api/repos/test/repo/pulls/invalid | 400 | 返回 Invalid pr_number |

#### 2.4 质量检查接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| ES-040 | 创建质量检查 | POST | /api/events/{uuid}/quality-checks | check payload | 200 | 返回检查对象 |
| ES-041 | 获取质量检查列表 | GET | /api/events/{uuid}/quality-checks | - | 200 | 返回检查数组 |
| ES-042 | 更新质量检查 | PUT | /api/quality-checks/{id} | update payload | 200 | 返回 status=updated |
| ES-043 | 按类型更新检查 | PUT | /api/events/{uuid}/quality-checks/{type} | update payload | 200 | 返回 status=updated |
| ES-044 | 获取检查统计 | GET | /api/quality-checks/statistics | - | 200 | 返回统计数据 |

---

### 3. Task Scheduler 服务 (端口 4003)

#### 3.1 健康检查接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| TS-001 | 健康检查正常 | GET | /health | 200 | 返回 status=healthy |

#### 3.2 任务管理接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| TS-010 | 列出任务 | GET | /api/tasks | - | 200 | 返回 tasks 数组和 total |
| TS-011 | 分页查询 | GET | /api/tasks?limit=10&offset=0 | - | 200 | 返回分页结果 |
| TS-012 | 获取任务详情 | GET | /api/tasks/{id} | - | 200/404 | 返回任务详情 |
| TS-013 | 启动任务 | POST | /api/tasks/{id}/start | - | 200/404 | 返回更新后的任务 |
| TS-014 | 完成任务 | POST | /api/tasks/{id}/complete | {"results":[...]} | 200 | 返回成功消息 |
| TS-015 | 标记任务失败 | POST | /api/tasks/{id}/fail | {"reason":"error"} | 200 | 返回成功消息 |
| TS-016 | 取消任务 | POST | /api/tasks/{id}/cancel | {"reason":"cancelled"} | 200 | 返回成功消息 |
| TS-017 | 任务不存在 | GET | /api/tasks/999999 | - | 404 | 返回 Task not found |
| TS-018 | 无效任务 ID | GET | /api/tasks/invalid | - | 400 | 返回 Invalid task ID |
| TS-019 | 负数任务 ID | GET | /api/tasks/-1 | - | 400/404 | 返回错误 |

#### 3.3 事件任务取消接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| TS-020 | 取消事件所有任务 | POST | /api/events/{event-id}/cancel | {"reason":"PR sync"} | 200 | 返回取消数量 count |

---

### 4. Executor Service 服务 (端口 4004)

#### 4.1 健康检查接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| EX-001 | 健康检查正常 | GET | /health | 200 | 返回 status=ok |

#### 4.2 执行管理接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| EX-010 | 执行任务 | POST | /api/execute | execution request | 200/202 | 返回执行结果 |
| EX-011 | 缺少 task_uuid | POST | /api/execute | {} | 400 | 返回 task_uuid is required |
| EX-012 | 获取执行状态 | GET | /api/executions/{id} | - | 200/404 | 返回执行状态 |
| EX-013 | 获取执行日志 | GET | /api/executions/{id}/logs | - | 200 | 返回日志数组 |
| EX-014 | 取消执行 | DELETE | /api/executions/{id} | - | 200 | 返回 status=canceled |
| EX-015 | 列出执行记录 | GET | /api/executions | - | 200 | 返回 executions 数组 |
| EX-016 | 执行不存在 | GET | /api/executions/non-existent | - | 400/404 | 返回错误 |

#### 4.3 Pipeline 兼容接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| EX-020 | 获取 Pipeline 状态 | GET | /api/status/{buildId} | 200/500 | 返回 build_id, status, result |
| EX-021 | 获取 Pipeline 日志 | GET | /api/logs/{buildId} | 200/500 | 返回 logs |

---

### 5. AI Analyzer 服务 (端口 4005)

#### 5.1 健康检查接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| AI-001 | 健康检查正常 | GET | /health | 200 | 返回 status=healthy |
| AI-002 | API 健康检查 | GET | /api/health | 200 | 返回 status=healthy |

#### 5.2 日志分析接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| AI-010 | 分析日志 | POST | /api/analyze | {"log_content":"...","task_name":"basic_ci_all"} | 200/500 | 返回分析结果 |
| AI-011 | 空 log_content | POST | /api/analyze | {"log_content":""} | 400 | 返回 log_content is required |
| AI-012 | 缺少 log_content | POST | /api/analyze | {} | 400 | 返回错误 |
| AI-013 | 批量分析 | POST | /api/analyze/batch | {"log_contents":[...]} | 200/500 | 返回批量结果 |
| AI-014 | 空批量请求 | POST | /api/analyze/batch | {"log_contents":[]} | 400 | 返回错误 |

#### 5.3 池管理接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| AI-020 | 获取池统计 | GET | /api/pool/stats | - | 200 | 返回 total_size, available, in_use |
| AI-021 | 获取池大小 | GET | /api/config/pool-size | - | 200 | 返回 total_size |
| AI-022 | 设置池大小 | POST | /api/config/pool-size | {"size":5} | 200 | 返回更新后的 size |
| AI-023 | 无效池大小 | POST | /api/config/pool-size | {"size":0} | 400 | 返回 size must be greater than 0 |

---

### 6. Resource Manager 服务 (端口 4006)

#### 6.1 健康检查接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| RM-001 | 健康检查正常 | GET | /health | 200 | 返回 status=ok |

#### 6.2 资源管理接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| RM-010 | 列出资源 | GET | /api/resources | - | 200 | 返回资源数组 |
| RM-011 | 创建资源 | POST | /api/resources | resource payload | 200 | 返回带 uuid 的资源 |
| RM-012 | 获取资源 | GET | /api/resources/{uuid} | - | 200/404 | 返回资源详情 |
| RM-013 | 更新资源 | PUT | /api/resources/{uuid} | update payload | 200 | 返回更新后的资源 |
| RM-014 | 删除资源 | DELETE | /api/resources/{uuid} | - | 204 | 无内容返回 |
| RM-015 | 资源不存在 | GET | /api/resources/non-existent | - | 404 | 返回 Resource not found |
| RM-016 | 按类别列出资源 | GET | /api/categories/{id}/resources | - | 200 | 返回该类别资源 |

#### 6.3 分类管理接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| RM-020 | 列出分类 | GET | /api/categories | - | 200 | 返回分类数组 |
| RM-021 | 创建分类 | POST | /api/categories | category payload | 200 | 返回分类对象 |
| RM-022 | 获取分类 | GET | /api/categories/{id} | - | 200/404 | 返回分类详情 |
| RM-023 | 更新分类 | PUT | /api/categories/{id} | update payload | 200 | 返回更新后的分类 |
| RM-024 | 删除分类 | DELETE | /api/categories/{id} | - | 204 | 无内容返回 |
| RM-025 | 分类不存在 | GET | /api/categories/999999 | - | 404 | 返回 Category not found |
| RM-026 | 空名称创建 | POST | /api/categories | {"name":""} | 400/500 | 返回错误 |

#### 6.4 资源匹配接口

| 用例ID | 测试场景 | 方法 | 端点 | 请求体 | 预期状态码 | 验证点 |
|--------|---------|------|------|--------|-----------|--------|
| RM-030 | 匹配资源 | POST | /api/resources/match | match request | 200/500 | 返回匹配结果 |
| RM-031 | 释放资源 | POST | /api/resources/{uuid}/release | - | 200 | 返回 status=released |
| RM-032 | 获取分配信息 | GET | /api/resources/{uuid}/allocation | - | 200 | 返回分配状态 |

#### 6.5 Testbed 接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| RM-040 | 列出 Testbed | GET | /api/testbeds | 200 | 返回 testbed 数组 |

#### 6.6 配额策略接口

| 用例ID | 测试场景 | 方法 | 端点 | 预期状态码 | 验证点 |
|--------|---------|------|------|-----------|--------|
| RM-050 | 列出配额策略 | GET | /api/categories/{id}/quota-policies | 200 | 返回策略数组 |

---

## 二、端到端测试用例 (E2E)

### 场景 1: GitHub PR 质量检查完整流程

| 步骤 | 操作 | 验证点 |
|------|------|--------|
| 1 | 发送 GitHub PR Opened Webhook | Webhook Gateway 返回 event_uuid |
| 2 | 查询事件状态 | Event Store 返回事件详情，status=pending |
| 3 | 获取待处理事件列表 | 事件在 pending 列表中 |
| 4 | 启动任务执行 | Task Scheduler 返回任务已启动 |
| 5 | 执行任务 | Executor Service 返回执行 ID |
| 6 | 获取执行状态 | 返回执行进度 |
| 7 | 完成任务 | Task Scheduler 返回成功 |
| 8 | 更新事件状态 | Event Store 返回更新成功 |
| 9 | 验证事件统计 | 统计数据已更新 |

### 场景 2: PR Synchronize 取消流程

| 步骤 | 操作 | 验证点 |
|------|------|--------|
| 1 | 发送 PR Opened Webhook | 创建事件 E1 |
| 2 | 发送 PR Synchronize Webhook | 创建事件 E2 |
| 3 | 验证 E1 状态 | E1 状态为 cancelled |
| 4 | 验证 E2 状态 | E2 状态为 pending |
| 5 | 验证任务状态 | E1 关联任务已取消 |

### 场景 3: 资源分配与释放流程

| 步骤 | 操作 | 验证点 |
|------|------|--------|
| 1 | 创建资源分类 | 返回分类 ID |
| 2 | 创建资源实例 | 返回资源 UUID |
| 3 | 匹配资源 | 返回匹配的资源 |
| 4 | 验证资源状态 | 状态变为 in_use |
| 5 | 释放资源 | 返回成功 |
| 6 | 验证资源状态 | 状态恢复为 available |

### 场景 4: AI 日志分析流程

| 步骤 | 操作 | 验证点 |
|------|------|--------|
| 1 | 获取池状态 | 返回池统计信息 |
| 2 | 提交日志分析请求 | 返回分析结果 |
| 3 | 验证结果格式 | 包含错误类型、建议等 |
| 4 | 批量分析日志 | 返回批量结果 |

### 场景 5: 服务健康检查

| 步骤 | 操作 | 验证点 |
|------|------|--------|
| 1 | 检查所有服务健康状态 | 所有服务返回 status=ok/healthy |
| 2 | 验证响应时间 | 响应时间 < 1s |

---

## 三、边界条件测试用例

### 3.1 无效输入测试

| 用例ID | 服务 | 测试场景 | 预期结果 |
|--------|------|---------|---------|
| BC-001 | All | 空 JSON Body | 400 Bad Request |
| BC-002 | All | 无效 JSON 格式 | 400 Bad Request |
| BC-003 | Event Store | 缺少必填字段 | 400/500 |
| BC-004 | Resource Manager | 缺少 name 字段 | 400/500 |
| BC-005 | AI Analyzer | 空 log_content | 400 |

### 3.2 资源不存在测试

| 用例ID | 服务 | 测试场景 | 预期结果 |
|--------|------|---------|---------|
| BC-010 | Event Store | 事件不存在 | 404 Not Found |
| BC-011 | Task Scheduler | 任务不存在 | 404 Not Found |
| BC-012 | Executor | 执行不存在 | 400/404 |
| BC-013 | Resource Manager | 资源不存在 | 404 Not Found |
| BC-014 | Resource Manager | 分类不存在 | 404 Not Found |

### 3.3 参数边界测试

| 用例ID | 服务 | 测试场景 | 预期结果 |
|--------|------|---------|---------|
| BC-020 | Event Store | 无效 status 值 | 400/500 |
| BC-021 | Task Scheduler | 无效任务 ID (字符串) | 400 |
| BC-022 | Task Scheduler | 负数任务 ID | 400/404 |
| BC-023 | Resource Manager | 无效 IP 地址 | 200 (不做校验) |
| BC-024 | Resource Manager | 负数端口 | 200 (不做校验) |
| BC-025 | AI Analyzer | 无效池大小 (0 或负数) | 400 |

### 3.4 特殊字符测试

| 用例ID | 服务 | 测试场景 | 预期结果 |
|--------|------|---------|---------|
| BC-030 | Event Store | XSS 攻击字符串 | 200 (转义存储) |
| BC-031 | Resource Manager | SQL 注入字符串 | 200 (参数化查询) |
| BC-032 | Category | 特殊字符名称 | 200 |

### 3.5 大数据量测试

| 用例ID | 服务 | 测试场景 | 预期结果 |
|--------|------|---------|---------|
| BC-040 | Event Store | 1MB+ Payload | 200/413 |
| BC-041 | AI Analyzer | 超长日志内容 | 200/500 |
| BC-042 | All | 大 limit 值 | 200 (限制返回) |

### 3.6 分页参数测试

| 用例ID | 服务 | 测试场景 | 预期结果 |
|--------|------|---------|---------|
| BC-050 | Event Store | limit=-1 | 200 (使用默认值) |
| BC-051 | Event Store | limit=0 | 200 (使用默认值) |
| BC-052 | Event Store | limit=abc | 200 (使用默认值) |
| BC-053 | Task Scheduler | offset=-1 | 200 (使用默认值) |

---

## 四、测试统计

| 分类 | 用例数量 |
|------|---------|
| API 接口测试 | 86 |
| 端到端测试 | 5 场景 (约 25 步骤) |
| 边界条件测试 | 24 |
| **总计** | **约 135 个测试点** |

---

## 五、测试优先级

### P0 - 核心功能 (必须通过)
- 所有健康检查接口
- 事件创建和查询
- 任务启动、完成、取消
- 资源创建和匹配

### P1 - 重要功能
- Webhook 接收和处理
- 质量检查管理
- 执行状态查询
- 分类管理

### P2 - 辅助功能
- 统计接口
- 池管理
- 分页查询

### P3 - 边界条件
- 无效输入处理
- 资源不存在处理
- 特殊字符处理

---

**请审核以上测试用例文档，确认后我将使用 pytest + Python 实现。**
