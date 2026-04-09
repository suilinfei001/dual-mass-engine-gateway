# deployment_deployment 质量节点改造方案

## 背景说明

当前 `deployment_deployment` 质量节点通过 AI 匹配可执行资源的方式存在以下问题：

1. **AI 匹配对于 deployment 不适用**：deployment 类型是系统中唯一且公用的可执行资源，不同微服务只是流水线参数中的 `chart_url_1` 字段值不同
2. **缺少 Testbed 获取逻辑**：deployment 执行需要实际的服务器资源（host、ssh_user、ssh_password），这些信息应该从 resource-pool 获取
3. **缺少资源释放机制**：任务失败或完成后没有释放已占用的 Testbed

## 改造目标

1. 简化 deployment 任务的资源匹配逻辑，无需 AI 匹配
2. 在执行 deployment 任务前自动获取 Testbed
3. 实现任务失败/完成后的自动释放机制
4. 确保整个质量流程结束后 Testbed 被正确释放

## 详细设计方案

### 1. 可执行资源结构

系统中 deployment 类型可执行资源的流水线参数结构如下：

```json
{
  "chart_url_1": "to_be_replace_chart",
  "chart_url_2": "-",
  "chart_url_3": "-",
  "enable_full_deploy": "False",
  "enable_helm_upgrade": "True",
  "host": "to_be_replace_host",
  "ssh_password": "to_be_replace_ssh_password",
  "ssh_user": "to_be_replace_ssh_user"
}
```

参数说明：

| 参数 | 来源 | 说明 |
|------|------|------|
| chart_url_1 | basic_ci_all 任务结果 | 微服务的 chart 地址 |
| chart_url_2 | 固定值 "-" | 保留字段 |
| chart_url_3 | 固定值 "-" | 保留字段 |
| enable_full_deploy | 固定值 "False" | 是否完整部署 |
| enable_helm_upgrade | 固定值 "True" | 是否使用 helm upgrade |
| host | resource-pool | 目标服务器 IP |
| ssh_user | resource-pool | SSH 用户名 |
| ssh_password | resource-pool | SSH 密码 |

### 2. 任务创建阶段改造

#### 2.1 跳过 AI 资源匹配

对于 `deployment_deployment` 任务，直接使用系统中唯一的 deployment 类型资源，无需 AI 匹配：

```go
// 在 TaskCreator.getResourceURL 中处理
if taskName == "deployment_deployment" {
    // 直接获取 deployment 类型的资源
    resource, err := tc.resourceStorage.GetResourceByType(models.ResourceTypeDeployment)
    if err != nil {
        return nil, fmt.Errorf("%w: no deployment resource available", ErrNoResourceMatched)
    }
    return &ResourceInfo{
        ResourceID: resource.ID,
        RequestURL: buildAzureURL(resource),
    }, nil
}
```

#### 2.2 从 basic_ci_all 获取 chart_url_1

在创建 deployment 任务时，需要从同一 event 的 basic_ci_all 任务结果中获取 chart 值：

```go
// 在创建 deployment 任务前
basicCiResults, err := tc.taskStorage.GetTaskResultsByEventAndTaskName(eventID, "basic_ci_all")
chartURL := extractChartFromResults(basicCiResults)
```

### 3. 任务执行阶段改造

#### 3.1 获取 Testbed

在 `executeNextTask` 中处理 `deployment_deployment` 任务时：

```go
if task.TaskName == "deployment_deployment" {
    // 重试获取 Testbed 直到成功
    var testbed *models.Testbed
    for {
        testbed, err = acquireTestbedWithRetry(ctx, resourcePoolAPI, categoryUUID, "robot")
        if err == nil && testbed != nil {
            break
        }
        log.Printf("Failed to acquire testbed, retrying in 1 minute: %v", err)
        time.Sleep(1 * time.Minute)
    }
    
    // 将 testbed 信息保存到任务中
    task.TestbedID = testbed.ID
    task.TestbedIP = testbed.IPAddress
    task.SSHUser = testbed.SSHUser
    task.SSHPassword = testbed.SSHPassword
}
```

#### 3.2 构建流水线参数

获取 Testbed 后，填充 deployment 资源的流水线参数：

```go
// 在 ExecuteTask 中处理 deployment 任务
if task.TaskName == "deployment_deployment" {
    params := map[string]interface{}{
        "chart_url_1":       task.ChartURL,         // 从 basic_ci_all 获取
        "chart_url_2":       "-",
        "chart_url_3":       "-",
        "enable_full_deploy": "False",
        "enable_helm_upgrade": "True",
        "host":              task.TestbedIP,
        "ssh_user":          task.SSHUser,
        "ssh_password":      task.SSHPassword,
    }
    // 调用 Azure Pipeline
}
```

### 4. 资源释放机制

#### 4.1 失败时释放 Testbed

在以下情况下需要释放 Testbed：

1. **任务执行失败**（FailTask）
2. **任务被取消**（CancelTask）
3. **任务被跳过**（SkipTask）

```go
// 在 FailTask/CancelTask/SkipTask 中
func (s *SchedulerWithStorage) releaseTestbedIfNeeded(task *models.Task) error {
    if task.TestbedID > 0 {
        // 调用 resource-pool 释放接口
        err := releaseTestbed(task.TestbedID)
        if err != nil {
            log.Printf("Failed to release testbed %d: %v", task.TestbedID, err)
            return err
        }
        log.Printf("Released testbed %d for task %s", task.TestbedID, task.TaskName)
    }
    return nil
}
```

#### 4.2 整个流程结束后释放

当事件的所有质量检查完成后，需要检查并释放可能残留的 Testbed：

```go
// 在 CompleteTask 中，当任务为最后一个任务时
if shouldCreateNext == false && isLastTask {
    // 检查是否有未释放的 testbed
    s.releaseTestbedForEvent(task.EventID)
}
```

或者在事件状态变为 completed/failed 时：

```go
// 在事件状态更新回调中
func OnEventStatusChanged(eventID int, status string) {
    if status == "completed" || status == "failed" {
        // 释放该事件关联的所有 testbed
        releaseAllTestbedsForEvent(eventID)
    }
}
```

### 5. 核心代码修改点

| 模块 | 文件 | 修改内容 |
|------|------|----------|
| TaskCreator | scheduler/creator.go | 跳过 deployment 的 AI 匹配逻辑 |
| Task 模型 | models/task.go | 添加 TestbedID、TestbedIP、SSHUser、SSHPassword、ChartURL 字段 |
| 执行器 | executor/service.go | 在 ExecuteTask 中获取 Testbed 并构建流水线参数 |
| 调度器 | scheduler/scheduler_db.go | 在 FailTask/CancelTask/SkipTask 中添加释放 Testbed 逻辑 |
| 任务监控 | monitor/monitor.go | 在任务失败时触发释放 |
| 主程序 | cmd/server/main.go | 添加重试获取 Testbed 的循环逻辑 |

### 6. API 调用设计

#### 6.1 获取 Testbed

```
POST /internal/testbeds/acquire
Content-Type: application/json

{
    "category_uuid": "deployment-category-uuid",
    "requester": "robot"
}
```

响应：

```json
{
    "allocation_uuid": "alloc-uuid",
    "testbed": {
        "uuid": "testbed-uuid",
        "ip_address": "10.0.0.1",
        "ssh_user": "root",
        "ssh_password": "password"
    }
}
```

#### 6.2 释放 Testbed

```
POST /internal/testbeds/{uuid}/release
```

### 7. 数据流图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         质量检查流程                                      │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  1. 事件触发 (PR opened/synchronize)                                     │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  2. 创建任务                                                             │
│     ├── basic_ci_all (stage_order=1)                                    │
│     ├── deployment_deployment (stage_order=2)                           │
│     └── specialized_tests_* (stage_order=3)                              │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  3. 执行 basic_ci_all                                                    │
│     ├── 运行 Azure Pipeline                                             │
│     ├── AI 分析日志                                                      │
│     └── 提取 chart_url_1                                                │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  4. 执行 deployment_deployment                                           │
│     │                                                                   │
│     ├── 4.1 尝试获取 Testbed (retry every 1 min until success)           │
│     │    POST /internal/testbeds/acquire                                │
│     │                                                                   │
│     ├── 4.2 构建流水线参数                                               │
│     │    chart_url_1 <- basic_ci_all 结果                               │
│     │    host/ssh_user/ssh_password <- Testbed                          │
│     │                                                                   │
│     ├── 4.3 运行 Azure Pipeline                                         │
│     │                                                                   │
│     └── 4.4 AI 分析日志                                                  │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  5. 执行后续质量节点 (specialized_tests_*)                               │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  6. 流程结束 / 失败                                                      │
│     │                                                                   │
│     ├── 6.1 释放 Testbed                                                 │
│     │    POST /internal/testbeds/{uuid}/release                        │
│     │                                                                   │
│     └── 6.2 更新 Event 状态                                             │
└─────────────────────────────────────────────────────────────────────────┘
```

### 8. 错误处理

| 场景 | 处理方式 |
|------|----------|
| 获取 Testbed 超时 | 每 1 分钟重试一次，直到获取成功 |
| Azure Pipeline 执行失败 | 标记任务为 failed，释放 Testbed |
| 任务被取消 | 释放 Testbed |
| 任务被跳过 | 释放 Testbed |
| 整个流程结束 | 释放残留的 Testbed |

### 9. 测试要点

1. **单元测试**
   - TaskCreator 跳过 deployment AI 匹配
   - 流水线参数正确构建
   - Testbed 获取和释放逻辑

2. **集成测试**
   - deployment 任务正确获取 Testbed
   - 任务失败时 Testbed 被释放
   - 流程结束后 Testbed 被释放

3. **E2E 测试**
   - 完整的 PR 质量检查流程
   - Testbed 生命周期验证
