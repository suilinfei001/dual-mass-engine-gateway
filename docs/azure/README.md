# Azure DevOps Pipeline Functions

基于 Azure DevOps REST API 实现的 Python 管道操作函数库。

## 功能

提供了一系列函数来操作 Azure DevOps Pipeline，包括：
- 运行管道
- 取消正在运行的管道
- 获取管道运行状态
- 获取管道时间线（任务列表和状态）
- 获取管道运行日志

## 依赖

```bash
pip install requests
```

## 函数说明

### 1. run_pipeline

运行 Azure DevOps 管道。

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `organization` | str | 是 | - | 组织名称，例如 "AISHUDevOps" |
| `project` | str | 是 | - | 项目名称，例如 "DIP" |
| `pipeline_id` | int | 是 | - | 管道 ID |
| `pat` | str | 是 | - | Personal Access Token |
| `branch` | str | 否 | "refs/heads/main" | 分支名称 |
| `template_parameters` | dict | 否 | None | 模板参数字典 |
| `api_version` | str | 否 | "6.0-preview.1" | API 版本 |
| `verify_ssl` | bool | 否 | True | 是否验证 SSL 证书 |

**返回值：**

`Dict[str, Any]` - API 响应的 JSON 数据，包含：
- `id`: 运行 ID (Build ID)
- `name`: 运行名称
- `state`: 状态 (inProgress, completed 等)
- `pipeline`: 管道信息
- `_links`: 相关链接

**示例：**

```python
from azure_devops_pipeline import run_pipeline

result = run_pipeline(
    organization="AISHUDevOps",
    project="DIP",
    pipeline_id=6903,
    pat="your_pat_token",
    branch="refs/heads/develop",
    template_parameters={
        "TRIVY_EXIT_CODE": "1",
        "SKIP_SONARQUBE": "0",
        "BUILD_TYPE": "opensource"
    }
)

print(f"Build ID: {result['id']}")
print(f"State: {result['state']}")
```

---

### 2. cancel_pipeline

取消正在运行的 Azure DevOps 管道。

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `organization` | str | 是 | - | 组织名称 |
| `project` | str | 是 | - | 项目名称 |
| `build_id` | int | 是 | - | Build ID（从 run_pipeline 返回的 id） |
| `pat` | str | 是 | - | Personal Access Token |
| `api_version` | str | 否 | "6.0" | API 版本 |
| `verify_ssl` | bool | 否 | True | 是否验证 SSL 证书 |

**返回值：**

`Dict[str, Any]` - API 响应的 JSON 数据，包含：
- `id`: Build ID
- `status`: 状态 (cancelling, canceled 等)
- `buildNumber`: 构建编号

**示例：**

```python
from azure_devops_pipeline import cancel_pipeline

cancel_result = cancel_pipeline(
    organization="AISHUDevOps",
    project="DIP",
    build_id=1752871,
    pat="your_pat_token"
)

print(f"Status: {cancel_result['status']}")
```

---

### 3. get_pipeline_status

获取管道运行状态。

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `organization` | str | 是 | - | 组织名称 |
| `project` | str | 是 | - | 项目名称 |
| `build_id` | int | 是 | - | Build ID |
| `pat` | str | 是 | - | Personal Access Token |
| `api_version` | str | 否 | "6.0" | API 版本 |
| `verify_ssl` | bool | 否 | True | 是否验证 SSL 证书 |

**返回值：**

`Dict[str, Any]` - API 响应的 JSON 数据，包含：
- `status`: 状态 (inProgress, completed, canceled, failed 等)
- `result`: 结果 (succeeded, failed, canceled 等) - 仅在完成后有值
- `finishTime`: 完成时间
- `buildNumber`: 构建编号

**状态说明：**

| 状态 | 说明 |
|------|------|
| `inProgress` | 运行中 |
| `notStarted` | 未开始 |
| `completed` | 已完成 |
| `canceled` | 已取消 |
| `failed` | 失败 |

**示例：**

```python
from azure_devops_pipeline import get_pipeline_status

status_info = get_pipeline_status(
    organization="AISHUDevOps",
    project="DIP",
    build_id=1752871,
    pat="your_pat_token"
)

print(f"Status: {status_info['status']}")
print(f"Result: {status_info.get('result', 'Running...')}")
```

---

### 4. get_pipeline_timeline

获取管道运行的时间线（任务列表和状态）。

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `organization` | str | 是 | - | 组织名称 |
| `project` | str | 是 | - | 项目名称 |
| `build_id` | int | 是 | - | Build ID |
| `pat` | str | 是 | - | Personal Access Token |
| `api_version` | str | 否 | "6.0" | API 版本 |
| `verify_ssl` | bool | 否 | True | 是否验证 SSL 证书 |

**返回值：**

`Dict[str, Any]` - API 响应的 JSON 数据，包含：
- `records`: 任务记录列表，每个记录包含：
  - `name`: 任务名称
  - `type`: 任务类型 (Task, Job, Stage, Phase 等)
  - `state`: 状态 (pending, inProgress, completed 等)
  - `result`: 结果 (succeeded, failed, canceled, abandoned 等)

**示例：**

```python
from azure_devops_pipeline import get_pipeline_timeline

timeline = get_pipeline_timeline(
    organization="AISHUDevOps",
    project="DIP",
    build_id=1752871,
    pat="your_pat_token"
)

for record in timeline['records']:
    task_name = record['name']
    task_type = record['type']
    state = record['state']
    result = record.get('result', 'N/A')
    print(f"[{task_type}] {task_name}: {state} (result: {result})")
```

---

### 5. get_pipeline_logs

获取管道运行的日志。

**参数：**

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `organization` | str | 是 | - | 组织名称 |
| `project` | str | 是 | - | 项目名称 |
| `build_id` | int | 是 | - | Build ID |
| `pat` | str | 是 | - | Personal Access Token |
| `log_id` | int | 否 | None | 日志 ID，None 时获取日志列表 |
| `start_line` | int | 否 | None | 起始行号（可选） |
| `end_line` | int | 否 | None | 结束行号（可选） |
| `api_version` | str | 否 | "6.0" | API 版本 |
| `verify_ssl` | bool | 否 | True | 是否验证 SSL 证书 |

**返回值：**

- 当 `log_id` 为 `None` 时：返回日志列表的 JSON (`Dict`)
- 当指定 `log_id` 时：返回日志内容的文本 (`str`)

**示例：**

```python
from azure_devops_pipeline import get_pipeline_logs

# 获取日志列表
logs_list = get_pipeline_logs(
    organization="AISHUDevOps",
    project="DIP",
    build_id=1752871,
    pat="your_pat_token"
)

for log_entry in logs_list['value']:
    log_id = log_entry['id']
    line_count = log_entry['lineCount']
    print(f"Log ID: {log_id}, Lines: {line_count}")

# 获取具体日志内容
log_content = get_pipeline_logs(
    organization="AISHUDevOps",
    project="DIP",
    build_id=1752871,
    log_id=1,
    pat="your_pat_token"
)

print(log_content)
```

---

## 完整使用示例

### 运行管道并等待完成

```python
import time
from azure_devops_pipeline import run_pipeline, get_pipeline_status

# 运行管道
result = run_pipeline(
    organization="AISHUDevOps",
    project="DIP",
    pipeline_id=6903,
    pat="your_pat_token",
    branch="refs/heads/develop"
)

build_id = result['id']
print(f"Pipeline started, Build ID: {build_id}")

# 轮询状态直到完成
while True:
    status_info = get_pipeline_status(
        organization="AISHUDevOps",
        project="DIP",
        build_id=build_id,
        pat="your_pat_token"
    )

    status = status_info['status']
    result_status = status_info.get('result', 'Running...')

    print(f"Status: {status}, Result: {result_status}")

    # 检查是否完成
    if status in ['completed', 'canceled', 'failed']:
        print(f"Pipeline finished with result: {result_status}")
        break

    time.sleep(30)  # 每 30 秒查询一次
```

### 运行管道并获取完整日志

```python
import time
from azure_devops_pipeline import (
    run_pipeline,
    get_pipeline_status,
    get_pipeline_timeline,
    get_pipeline_logs
)

# 1. 运行管道
result = run_pipeline(
    organization="AISHUDevOps",
    project="DIP",
    pipeline_id=6903,
    pat="your_pat_token",
    branch="refs/heads/develop"
)

build_id = result['id']
build_number = result['name']
pipeline_name = result['pipeline']['name']

# 2. 等待完成
while True:
    status_info = get_pipeline_status(
        organization="AISHUDevOps",
        project="DIP",
        build_id=build_id,
        pat="your_pat_token"
    )
    if status_info['status'] in ['completed', 'canceled', 'failed']:
        break
    time.sleep(30)

# 3. 获取时间线
timeline = get_pipeline_timeline(
    organization="AISHUDevOps",
    project="DIP",
    build_id=build_id,
    pat="your_pat_token"
)

# 打印任务列表
print("\n任务列表:")
for record in timeline['records']:
    print(f"  - [{record['type']}] {record['name']}: {record['state']}")

# 4. 获取并保存日志
logs_list = get_pipeline_logs(
    organization="AISHUDevOps",
    project="DIP",
    build_id=build_id,
    pat="your_pat_token"
)

with open(f"{pipeline_name}_{build_number}_logs.txt", 'w', encoding='utf-8') as f:
    for log_entry in logs_list['value']:
        log_id = log_entry['id']
        log_content = get_pipeline_logs(
            organization="AISHUDevOps",
            project="DIP",
            build_id=build_id,
            log_id=log_id,
            pat="your_pat_token"
        )
        f.write(f"{'='*60}\n")
        f.write(f"Log ID: {log_id}\n")
        f.write(f"{'='*60}\n\n")
        f.write(log_content)
        f.write("\n\n")

print("日志已保存!")
```

### 取消正在运行的管道

```python
from azure_devops_pipeline import run_pipeline, cancel_pipeline

# 运行管道
result = run_pipeline(
    organization="AISHUDevOps",
    project="DIP",
    pipeline_id=6903,
    pat="your_pat_token",
    branch="refs/heads/develop"
)

build_id = result['id']

# 取消管道
cancel_result = cancel_pipeline(
    organization="AISHUDevOps",
    project="DIP",
    build_id=build_id,
    pat="your_pat_token"
)

print(f"Cancel status: {cancel_result['status']}")
```

## 注意事项

1. **Windows 凭据问题**：代码中使用 `session.trust_env = False` 来防止 requests 库自动使用 Windows 凭据管理器中的旧凭据。

2. **PAT Token**：请确保您的 PAT Token 有足够的权限：
   - `Build: Execute` - 运行和取消构建
   - `Build: Read` - 查看构建状态和日志

3. **状态轮询**：建议使用适当的轮询间隔（如 30 秒），避免频繁请求 API。

4. **SSL 验证**：如果使用自签名证书，可以将 `verify_ssl` 设置为 `False`。
