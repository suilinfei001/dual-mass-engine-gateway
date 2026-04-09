"""
Azure DevOps Pipeline Functions
基于 Azure DevOps REST API 实现的管道操作函数
"""

import base64
import os
import requests
from typing import Dict, Any, Optional
from datetime import datetime
from pathlib import Path


def run_pipeline(
    organization: str,
    project: str,
    pipeline_id: int,
    pat: str,
    branch: str = "refs/heads/main",
    template_parameters: Optional[Dict[str, str]] = None,
    api_version: str = "6.0-preview.1",
    verify_ssl: bool = True
) -> Dict[str, Any]:
    """
    运行 Azure DevOps 管道

    Args:
        organization: 组织名称，例如 "AISHUDevOps"
        project: 项目名称，例如 "DIP"
        pipeline_id: 管道 ID
        pat: Personal Access Token
        branch: 分支名称，默认为 "refs/heads/main"
        template_parameters: 模板参数字典
        api_version: API 版本
        verify_ssl: 是否验证 SSL 证书

    Returns:
        API 响应的 JSON 数据
    """
    # 构建 URL
    url = f"https://devops.aishu.cn/{organization}/{project}/_apis/pipelines/{pipeline_id}/runs"

    # 构建认证头
    auth_string = f":{pat}"
    auth_bytes = auth_string.encode('ascii')
    base64_auth = base64.b64encode(auth_bytes).decode('ascii')

    headers = {
        'Content-Type': 'application/json',
        'Authorization': f'Basic {base64_auth}'
    }

    # 构建请求体
    payload = {
        "resources": {
            "repositories": {
                "self": {
                    "refName": branch
                }
            }
        }
    }

    # 添加模板参数（如果提供）
    if template_parameters:
        payload["templateParameters"] = template_parameters

    # 发送请求
    # 禁用 requests 的自动认证（防止使用 Windows 凭据管理器中的旧凭据）
    session = requests.Session()
    session.trust_env = False  # 禁用 .netrc 和代理认证

    response = session.post(
        url,
        headers=headers,
        json=payload,
        params={'api-version': api_version},
        verify=verify_ssl
    )

    # 确保请求成功
    response.raise_for_status()

    return response.json()


def cancel_pipeline(
    organization: str,
    project: str,
    build_id: int,
    pat: str,
    api_version: str = "6.0",
    verify_ssl: bool = True
) -> Dict[str, Any]:
    """
    取消正在运行的 Azure DevOps 管道

    注意：使用 Builds API 而不是 Pipelines API

    Args:
        organization: 组织名称，例如 "AISHUDevOps"
        project: 项目名称，例如 "DIP"
        build_id: Build ID（从 run_pipeline 返回的 id）
        pat: Personal Access Token
        api_version: API 版本
        verify_ssl: 是否验证 SSL 证书

    Returns:
        API 响应的 JSON 数据
    """
    # 构建 URL - 使用 Builds API
    url = f"https://devops.aishu.cn/{organization}/{project}/_apis/build/builds/{build_id}"

    # 构建认证头
    auth_string = f":{pat}"
    auth_bytes = auth_string.encode('ascii')
    base64_auth = base64.b64encode(auth_bytes).decode('ascii')

    headers = {
        'Content-Type': 'application/json',
        'Authorization': f'Basic {base64_auth}'
    }

    # 构建请求体 - 设置状态为 canceling (注意是 canceling 不是 canceled)
    payload = {
        "status": "cancelling"
    }

    # 发送 PATCH 请求
    session = requests.Session()
    session.trust_env = False

    response = session.patch(
        url,
        headers=headers,
        json=payload,
        params={'api-version': api_version},
        verify=verify_ssl
    )

    # 确保请求成功
    response.raise_for_status()

    return response.json()


def get_pipeline_timeline(
    organization: str,
    project: str,
    build_id: int,
    pat: str,
    api_version: str = "6.0",
    verify_ssl: bool = True
) -> Dict[str, Any]:
    """
    获取管道运行的时间线（任务列表和状态）

    Args:
        organization: 组织名称，例如 "AISHUDevOps"
        project: 项目名称，例如 "DIP"
        build_id: Build ID
        pat: Personal Access Token
        api_version: API 版本
        verify_ssl: 是否验证 SSL 证书

    Returns:
        API 响应的 JSON 数据，包含任务列表和状态
    """
    # 构建 URL
    url = f"https://devops.aishu.cn/{organization}/{project}/_apis/build/builds/{build_id}/timeline"

    # 构建认证头
    auth_string = f":{pat}"
    auth_bytes = auth_string.encode('ascii')
    base64_auth = base64.b64encode(auth_bytes).decode('ascii')

    headers = {
        'Content-Type': 'application/json',
        'Authorization': f'Basic {base64_auth}'
    }

    # 发送 GET 请求
    session = requests.Session()
    session.trust_env = False

    response = session.get(
        url,
        headers=headers,
        params={'api-version': api_version},
        verify=verify_ssl
    )

    # 确保请求成功
    response.raise_for_status()

    return response.json()


def get_pipeline_logs(
    organization: str,
    project: str,
    build_id: int,
    pat: str,
    log_id: Optional[int] = None,
    start_line: Optional[int] = None,
    end_line: Optional[int] = None,
    api_version: str = "6.0",
    verify_ssl: bool = True
) -> str:
    """
    获取管道运行的日志

    Args:
        organization: 组织名称，例如 "AISHUDevOps"
        project: 项目名称，例如 "DIP"
        build_id: Build ID
        pat: Personal Access Token
        log_id: 日志 ID，如果为 None 则获取所有日志的列表
        start_line: 起始行号（可选）
        end_line: 结束行号（可选）
        api_version: API 版本
        verify_ssl: 是否验证 SSL 证书

    Returns:
        如果指定 log_id，返回日志内容（文本）
        如果未指定 log_id，返回日志列表的 JSON
    """
    # 构建 URL
    if log_id is None:
        url = f"https://devops.aishu.cn/{organization}/{project}/_apis/build/builds/{build_id}/logs"
    else:
        url = f"https://devops.aishu.cn/{organization}/{project}/_apis/build/builds/{build_id}/logs/{log_id}"

    # 构建认证头
    auth_string = f":{pat}"
    auth_bytes = auth_string.encode('ascii')
    base64_auth = base64.b64encode(auth_bytes).decode('ascii')

    headers = {
        'Accept': 'text/plain',
        'Authorization': f'Basic {base64_auth}'
    }

    # 构建查询参数
    params = {'api-version': api_version}
    if start_line is not None:
        params['startLine'] = start_line
    if end_line is not None:
        params['endLine'] = end_line

    # 发送 GET 请求
    session = requests.Session()
    session.trust_env = False

    # 如果没有指定 log_id，使用 JSON 接受类型获取日志列表
    if log_id is None:
        headers['Accept'] = 'application/json'

    response = session.get(
        url,
        headers=headers,
        params=params,
        verify=verify_ssl
    )

    # 确保请求成功
    response.raise_for_status()

    # 如果是获取日志列表，返回 JSON
    if log_id is None:
        return response.json()

    # 否则返回文本内容
    return response.text


def get_pipeline_status(
    organization: str,
    project: str,
    build_id: int,
    pat: str,
    api_version: str = "6.0",
    verify_ssl: bool = True
) -> Dict[str, Any]:
    """
    获取管道运行状态

    Args:
        organization: 组织名称，例如 "AISHUDevOps"
        project: 项目名称，例如 "DIP"
        build_id: Build ID
        pat: Personal Access Token
        api_version: API 版本
        verify_ssl: 是否验证 SSL 证书

    Returns:
        API 响应的 JSON 数据，包含状态信息
        status 字段: inProgress, completed, canceled, failed 等
        result 字段: succeeded, failed, canceled 等 (仅在完成后)
    """
    # 构建 URL
    url = f"https://devops.aishu.cn/{organization}/{project}/_apis/build/builds/{build_id}"

    # 构建认证头
    auth_string = f":{pat}"
    auth_bytes = auth_string.encode('ascii')
    base64_auth = base64.b64encode(auth_bytes).decode('ascii')

    headers = {
        'Content-Type': 'application/json',
        'Authorization': f'Basic {base64_auth}'
    }

    # 发送 GET 请求
    session = requests.Session()
    session.trust_env = False

    response = session.get(
        url,
        headers=headers,
        params={'api-version': api_version},
        verify=verify_ssl
    )

    # 确保请求成功
    response.raise_for_status()

    return response.json()


def test_pipeline_functions():
    """
    测试 Azure DevOps Pipeline 函数
    """
    import time
    import json

    # 从环境变量读取 PAT
    pat = os.environ.get("AZURE_DEVOPS_PAT")
    if not pat:
        raise ValueError("请设置环境变量 AZURE_DEVOPS_PAT")

    test_config = {
        "organization": "AISHUDevOps",
        "project": "DIP",
        "pipeline_id": 6903,
        "pat": pat,
        "branch": "refs/heads/develop",
        "template_parameters": {
            "TRIVY_EXIT_CODE": "1",
            "SKIP_SONARQUBE": "0",
            "SOURCE_CODE_BRANCH": "main",
            "BUILD_TYPE": "opensource",
            "BUILD_ARM": "false"
        }
    }

    # ===== 测试 1: 运行管道 =====
    print("=" * 50)
    print("测试 1: 运行管道")
    print("=" * 50)

    try:
        result = run_pipeline(
            organization=test_config["organization"],
            project=test_config["project"],
            pipeline_id=test_config["pipeline_id"],
            pat=test_config["pat"],
            branch=test_config["branch"],
            template_parameters=test_config["template_parameters"]
        )

        print("\n运行成功！")
        print(f"Pipeline: {result['pipeline']['name']}")
        print(f"Run ID: {result['id']}")
        print(f"Run Name: {result['name']}")
        print(f"State: {result['state']}")
        print(f"Web URL: {result['_links']['web']['href']}")

        build_id = result['id']
        build_number = result['name']
        pipeline_name = result['pipeline']['name']

        # 创建 logs 目录
        logs_dir = Path("logs")
        logs_dir.mkdir(exist_ok=True)

        # 创建以 pipeline_name_build_id_timestamp 命名的子目录
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        build_log_dir = logs_dir / f"{pipeline_name}_{build_number}_{timestamp}"
        build_log_dir.mkdir(exist_ok=True)

        print(f"\n日志将保存到: {build_log_dir}")

        # ===== 测试 2: 轮询管道状态直到完成 =====
        print("\n" + "=" * 50)
        print("测试 2: 轮询管道状态")
        print("=" * 50)

        poll_interval = 30  # 每 30 秒查询一次
        max_polls = 120  # 最多轮询 120 次（1 小时）

        for poll_count in range(max_polls):
            status_info = get_pipeline_status(
                organization=test_config["organization"],
                project=test_config["project"],
                build_id=build_id,
                pat=test_config["pat"]
            )

            status = status_info.get('status', 'unknown')
            result_status = status_info.get('result', None)
            finish_time = status_info.get('finishTime', None)

            current_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            print(f"[{current_time}] 状态: {status}, 结果: {result_status or '运行中...'}")

            # 检查是否完成（status 必须是 completed、canceled、failed 等最终状态）
            # inProgress 和 notStart 都不是完成状态
            completed_statuses = ['completed', 'canceled', 'failed']
            if status in completed_statuses:
                print(f"\n管道已完成！")
                print(f"最终状态: {status}")
                print(f"结果: {result_status}")
                break

            # 等待 30 秒后再次查询
            if poll_count < max_polls - 1:
                print(f"等待 {poll_interval} 秒后再次查询...")
                time.sleep(poll_interval)

        # ===== 测试 3: 获取 Timeline =====
        print("\n" + "=" * 50)
        print("测试 3: 获取 Timeline")
        print("=" * 50)

        timeline = get_pipeline_timeline(
            organization=test_config["organization"],
            project=test_config["project"],
            build_id=build_id,
            pat=test_config["pat"]
        )

        print(f"\nTimeline 获取成功！")
        print(f"任务数量: {len(timeline.get('records', []))}")

        # 保存 timeline 到文件
        timeline_file = build_log_dir / "timeline.json"
        with open(timeline_file, 'w', encoding='utf-8') as f:
            json.dump(timeline, f, ensure_ascii=False, indent=2)
        print(f"Timeline 已保存到: {timeline_file}")

        # 显示任务状态
        print("\n任务列表:")
        for record in timeline.get('records', []):
            task_name = record.get('name', 'Unknown')
            task_type = record.get('type', 'Unknown')
            state = record.get('state', 'Unknown')
            result_status = record.get('result', 'N/A')
            print(f"  - [{task_type}] {task_name}: {state} (result: {result_status})")

        # 保存任务列表到文件
        tasks_file = build_log_dir / "tasks_list.txt"
        with open(tasks_file, 'w', encoding='utf-8') as f:
            f.write("任务列表:\n")
            for record in timeline.get('records', []):
                task_name = record.get('name', 'Unknown')
                task_type = record.get('type', 'Unknown')
                state = record.get('state', 'Unknown')
                result_status = record.get('result', 'N/A')
                f.write(f"  - [{task_type}] {task_name}: {state} (result: {result_status})\n")
        print(f"任务列表已保存到: {tasks_file}")

        # ===== 测试 4: 获取并合并所有日志 =====
        print("\n" + "=" * 50)
        print("测试 4: 获取并合并所有日志")
        print("=" * 50)

        logs = get_pipeline_logs(
            organization=test_config["organization"],
            project=test_config["project"],
            build_id=build_id,
            pat=test_config["pat"]
        )

        print(f"\n日志列表获取成功！")
        print(f"日志数量: {len(logs.get('value', []))}")

        # 合并所有日志到一个文件
        merged_log_file = build_log_dir / "full_logs.txt"

        with open(merged_log_file, 'w', encoding='utf-8') as merged_file:
            # 写入任务列表
            merged_file.write("=" * 60 + "\n")
            merged_file.write("任务列表\n")
            merged_file.write("=" * 60 + "\n\n")
            for record in timeline.get('records', []):
                task_name = record.get('name', 'Unknown')
                task_type = record.get('type', 'Unknown')
                state = record.get('state', 'Unknown')
                result_status = record.get('result', 'N/A')
                merged_file.write(f"  - [{task_type}] {task_name}: {state} (result: {result_status})\n")
            merged_file.write("\n")

            # 写入日志内容
            for log_entry in logs.get('value', []):
                log_id = log_entry.get('id')
                line_count = log_entry.get('lineCount', 0)

                print(f"正在获取 Log ID: {log_id} ({line_count} 行)...")

                log_content = get_pipeline_logs(
                    organization=test_config["organization"],
                    project=test_config["project"],
                    build_id=build_id,
                    log_id=log_id,
                    pat=test_config["pat"]
                )

                # 写入分隔符和日志内容
                merged_file.write("=" * 60 + "\n")
                merged_file.write(f"Log ID: {log_id} ({line_count} 行)\n")
                merged_file.write("=" * 60 + "\n\n")
                merged_file.write(log_content)
                merged_file.write("\n\n")

        print(f"\n所有日志已合并保存到: {merged_log_file}")
        print(f"\n所有日志已保存到目录: {build_log_dir}")

    except requests.exceptions.HTTPError as e:
        print(f"\n请求失败: {e}")
        print("可能的原因:")
        print("  1. PAT Token 已过期或无效")
        print("  2. 管道 ID 不存在")
        print("  3. 权限不足")
        print("\n请检查配置后重试。")


def test_pipeline_functions_2():
    """
    测试 Azure DevOps Pipeline 函数 - AnyShareFamily 项目
    """
    import time
    import json

    # 从环境变量读取 PAT
    pat = os.environ.get("AZURE_DEVOPS_PAT")
    if not pat:
        raise ValueError("请设置环境变量 AZURE_DEVOPS_PAT")

    test_config = {
        "organization": "AISHUDevOps",
        "project": "AnyShareFamily",
        "pipeline_id": 3875,
        "pat": pat,
        "branch": "refs/heads/MISSION",
        "template_parameters": {
            "ExecuteATWithAPI": "False",
            "host": "10.4.176.24",
            "password": "eisoo.com123",
            "client_host": "10.4.132.46",
            "db_host": "10.4.176.24",
            "eceph_host": "10.2.177.66",
            "eceph_pwd": "eisoo.com123",
            "db_type": "mariadb",
            "db_user": "Anyshare",
            "db_password": "asAlqlTkWU0zqfxrLTed",
            "db_port": "3320",
            "Service": "ThirdpartyMessagePlugin",
            "rule": "--reruns 1 --reruns-delay 2",
            "mark": "not reliability and not manual and not time and not nsq and not eacplog"
        }
    }

    # ===== 测试 1: 运行管道 =====
    print("=" * 50)
    print("测试 1: 运行管道 (AnyShareFamily)")
    print("=" * 50)

    try:
        result = run_pipeline(
            organization=test_config["organization"],
            project=test_config["project"],
            pipeline_id=test_config["pipeline_id"],
            pat=test_config["pat"],
            branch=test_config["branch"],
            template_parameters=test_config["template_parameters"]
        )

        print("\n运行成功！")
        print(f"Pipeline: {result['pipeline']['name']}")
        print(f"Run ID: {result['id']}")
        print(f"Run Name: {result['name']}")
        print(f"State: {result['state']}")
        print(f"Web URL: {result['_links']['web']['href']}")

        build_id = result['id']
        build_number = result['name']
        pipeline_name = result['pipeline']['name']

        # 创建 logs 目录
        logs_dir = Path("logs")
        logs_dir.mkdir(exist_ok=True)

        # 创建以 pipeline_name_build_id_timestamp 命名的子目录
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        build_log_dir = logs_dir / f"{pipeline_name}_{build_number}_{timestamp}"
        build_log_dir.mkdir(exist_ok=True)

        print(f"\n日志将保存到: {build_log_dir}")

        # ===== 测试 2: 轮询管道状态直到完成 =====
        print("\n" + "=" * 50)
        print("测试 2: 轮询管道状态")
        print("=" * 50)

        poll_interval = 30  # 每 30 秒查询一次
        max_polls = 120  # 最多轮询 120 次（1 小时）

        for poll_count in range(max_polls):
            status_info = get_pipeline_status(
                organization=test_config["organization"],
                project=test_config["project"],
                build_id=build_id,
                pat=test_config["pat"]
            )

            status = status_info.get('status', 'unknown')
            result_status = status_info.get('result', None)
            finish_time = status_info.get('finishTime', None)

            current_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
            print(f"[{current_time}] 状态: {status}, 结果: {result_status or '运行中...'}")

            # 检查是否完成（status 必须是 completed、canceled、failed 等最终状态）
            # inProgress 和 notStart 都不是完成状态
            completed_statuses = ['completed', 'canceled', 'failed']
            if status in completed_statuses:
                print(f"\n管道已完成！")
                print(f"最终状态: {status}")
                print(f"结果: {result_status}")
                break

            # 等待 30 秒后再次查询
            if poll_count < max_polls - 1:
                print(f"等待 {poll_interval} 秒后再次查询...")
                time.sleep(poll_interval)

        # ===== 测试 3: 获取 Timeline =====
        print("\n" + "=" * 50)
        print("测试 3: 获取 Timeline")
        print("=" * 50)

        timeline = get_pipeline_timeline(
            organization=test_config["organization"],
            project=test_config["project"],
            build_id=build_id,
            pat=test_config["pat"]
        )

        print(f"\nTimeline 获取成功！")
        print(f"任务数量: {len(timeline.get('records', []))}")

        # 保存 timeline 到文件
        timeline_file = build_log_dir / "timeline.json"
        with open(timeline_file, 'w', encoding='utf-8') as f:
            json.dump(timeline, f, ensure_ascii=False, indent=2)
        print(f"Timeline 已保存到: {timeline_file}")

        # 显示任务状态
        print("\n任务列表:")
        for record in timeline.get('records', []):
            task_name = record.get('name', 'Unknown')
            task_type = record.get('type', 'Unknown')
            state = record.get('state', 'Unknown')
            result_status = record.get('result', 'N/A')
            print(f"  - [{task_type}] {task_name}: {state} (result: {result_status})")

        # 保存任务列表到文件
        tasks_file = build_log_dir / "tasks_list.txt"
        with open(tasks_file, 'w', encoding='utf-8') as f:
            f.write("任务列表:\n")
            for record in timeline.get('records', []):
                task_name = record.get('name', 'Unknown')
                task_type = record.get('type', 'Unknown')
                state = record.get('state', 'Unknown')
                result_status = record.get('result', 'N/A')
                f.write(f"  - [{task_type}] {task_name}: {state} (result: {result_status})\n")
        print(f"任务列表已保存到: {tasks_file}")

        # ===== 测试 4: 获取并合并所有日志 =====
        print("\n" + "=" * 50)
        print("测试 4: 获取并合并所有日志")
        print("=" * 50)

        logs = get_pipeline_logs(
            organization=test_config["organization"],
            project=test_config["project"],
            build_id=build_id,
            pat=test_config["pat"]
        )

        print(f"\n日志列表获取成功！")
        print(f"日志数量: {len(logs.get('value', []))}")

        # 合并所有日志到一个文件
        merged_log_file = build_log_dir / "full_logs.txt"

        with open(merged_log_file, 'w', encoding='utf-8') as merged_file:
            # 写入任务列表
            merged_file.write("=" * 60 + "\n")
            merged_file.write("任务列表\n")
            merged_file.write("=" * 60 + "\n\n")
            for record in timeline.get('records', []):
                task_name = record.get('name', 'Unknown')
                task_type = record.get('type', 'Unknown')
                state = record.get('state', 'Unknown')
                result_status = record.get('result', 'N/A')
                merged_file.write(f"  - [{task_type}] {task_name}: {state} (result: {result_status})\n")
            merged_file.write("\n")

            # 写入日志内容
            for log_entry in logs.get('value', []):
                log_id = log_entry.get('id')
                line_count = log_entry.get('lineCount', 0)

                print(f"正在获取 Log ID: {log_id} ({line_count} 行)...")

                log_content = get_pipeline_logs(
                    organization=test_config["organization"],
                    project=test_config["project"],
                    build_id=build_id,
                    log_id=log_id,
                    pat=test_config["pat"]
                )

                # 写入分隔符和日志内容
                merged_file.write("=" * 60 + "\n")
                merged_file.write(f"Log ID: {log_id} ({line_count} 行)\n")
                merged_file.write("=" * 60 + "\n\n")
                merged_file.write(log_content)
                merged_file.write("\n\n")

        print(f"\n所有日志已合并保存到: {merged_log_file}")
        print(f"\n所有日志已保存到目录: {build_log_dir}")

    except requests.exceptions.HTTPError as e:
        print(f"\n请求失败: {e}")
        print("可能的原因:")
        print("  1. PAT Token 已过期或无效")
        print("  2. 管道 ID 不存在")
        print("  3. 权限不足")
        print("\n请检查配置后重试。")


if __name__ == "__main__":
    import sys

    # 根据命令行参数选择测试用例
    if len(sys.argv) > 1:
        test_case = sys.argv[1]
        if test_case == "2":
            test_pipeline_functions_2()
        else:
            test_pipeline_functions()
    else:
        # 默认运行第一个测试用例
        test_pipeline_functions()
