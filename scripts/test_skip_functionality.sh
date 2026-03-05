#!/bin/bash
# 测试任务跳过功能
# 用于验证 basic_ci_all、deployment_deployment 和 specialized_tests 任务跳过时，
# 所有相关的质量检查项都正确标记为 skipped 状态

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================================
# 配置
# ============================================================
EVENT_RECEIVER_API="http://10.4.111.141:5001"

# ============================================================
# 辅助函数
# ============================================================
print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}→${NC} $1"
}

# 获取最新的事件 ID
get_latest_event_id() {
    curl -s "${EVENT_RECEIVER_API}/api/events" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    events = data.get('data', [])
    if events:
        print(max(e['id'] for e in events))
    else:
        print('0')
except Exception:
    print('0')
"
}

# 删除指定 ID 的事件
delete_event() {
    local event_id=$1
    curl -s -X DELETE "${EVENT_RECEIVER_API}/api/events/${event_id}" > /dev/null 2>&1 || true
}

# 触发 mock push 事件
trigger_push_event() {
    curl -s -X POST "${EVENT_RECEIVER_API}/api/mock/simulate/push" \
        -H "Content-Type: application/json" > /dev/null 2>&1
}

# 等待事件处理完成
wait_for_event_completion() {
    local event_id=$1
    local max_wait=90
    local elapsed=0

    while [ $elapsed -lt $max_wait ]; do
        local status=$(curl -s "${EVENT_RECEIVER_API}/api/events/${event_id}" 2>/dev/null | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(data.get('data', {}).get('event_status', 'unknown'))
except:
    print('unknown')
")

        # 如果事件不存在（返回空或unknown），等待更长时间
        if [ -z "$status" ] || [ "$status" = "unknown" ]; then
            sleep 5
            elapsed=$((elapsed + 5))
            echo -n "?"
            continue
        fi

        if [ "$status" = "completed" ] || [ "$status" = "failed" ]; then
            return 0
        fi
        sleep 5
        elapsed=$((elapsed + 5))
        echo -n "."
    done
    echo ""
    return 1
}

# 检查质量检查状态
check_quality_checks() {
    local event_id=$1

    # 获取事件详情并保存到临时文件
    local temp_file=$(mktemp)
    curl -s "${EVENT_RECEIVER_API}/api/events/${event_id}" > "$temp_file"

    # 使用 Python 脚本分析
    python3 << PYTHON_SCRIPT
import sys
import json

try:
    with open('$temp_file', 'r') as f:
        data = json.load(f)

    event = data.get('data', {})
    if not event:
        print("ERROR: Event not found")
        sys.exit(1)

    checks = event.get('quality_checks', [])
    event_status = event.get('event_status', 'unknown')

    print(f"Event ID: {event.get('id')}")
    print(f"Event status: {event_status}")
    print(f"Total quality checks: {len(checks)}")
    print("")

    # 按阶段分组检查
    from collections import defaultdict, Counter
    by_stage = defaultdict(list)
    for c in checks:
        by_stage[c['stage']].append(c)

    all_passed = True
    total_checks = 0
    skipped_checks = 0

    for stage in ['basic_ci', 'deployment', 'specialized_tests']:
        if stage not in by_stage:
            continue
        print(f"{stage.upper()} checks:")
        for c in by_stage[stage]:
            status = c['check_status']
            total_checks += 1
            if status == 'skipped':
                skipped_checks += 1
                print(f"  ✓ {c['check_type']:35} status={status}")
            else:
                all_passed = False
                print(f"  ✗ {c['check_type']:35} status={status} (expected: skipped)")
        print("")

    # 统计
    status_count = Counter(c['check_status'] for c in checks)
    print(f"Status summary: {dict(status_count)}")
    print("")

    if all_passed and total_checks == skipped_checks:
        print("RESULT: PASS")
        sys.exit(0)
    else:
        print("RESULT: FAIL")
        sys.exit(1)

except Exception as e:
    print(f"ERROR: {e}")
    import traceback
    traceback.print_exc()
    sys.exit(1)
PYTHON_SCRIPT

    local result=$?
    rm -f "$temp_file"
    return $result
}

# ============================================================
# 主流程
# ============================================================
main() {
    print_header "测试任务跳过功能"

    # 步骤 1: 清理旧的测试事件
    print_info "步骤 1: 清理旧的测试事件 (保留最新的3个事件)"
    LATEST_ID=$(get_latest_event_id)
    if [ "$LATEST_ID" -gt 3 ]; then
        START_ID=$((LATEST_ID - 2))
        for ((id=START_ID; id<=LATEST_ID; id++)); do
            delete_event "$id"
        done
        print_success "已保留前3个事件，删除事件 $START_ID-$LATEST_ID"
    else
        print_info "没有需要清理的旧事件"
    fi

    # 步骤 2: 触发新的测试事件
    print_info "步骤 2: 触发新的 push 事件"
    OLD_EVENT_ID=$(get_latest_event_id)
    trigger_push_event
    sleep 3  # 等待事件创建
    NEW_EVENT_ID=$(get_latest_event_id)

    # 检查新事件是否创建成功
    if [ "$NEW_EVENT_ID" = "$OLD_EVENT_ID" ] || [ "$NEW_EVENT_ID" = "0" ]; then
        print_error "新事件创建失败"
        return 1
    fi
    print_success "事件已触发，新事件 ID: $NEW_EVENT_ID"

    # 步骤 3: 等待事件处理
    print_info "步骤 3: 等待事件处理 (最多等待60秒)"
    if wait_for_event_completion "$NEW_EVENT_ID"; then
        print_success "事件处理完成"
    else
        print_error "事件处理超时"
        return 1
    fi

    # 步骤 4: 检查质量检查状态
    print_header "步骤 4: 验证质量检查状态"
    sleep 2  # 额外等待确保所有检查都已更新

    if check_quality_checks "$NEW_EVENT_ID"; then
        print_success "测试通过: 所有质量检查项都正确标记为 skipped"
        RESULT=0
    else
        print_error "测试失败: 部分质量检查项未正确标记为 skipped"
        RESULT=1
    fi

    # 步骤 5: 显示详细信息
    print_header "步骤 5: 事件详情"
    curl -s "${EVENT_RECEIVER_API}/api/events/${NEW_EVENT_ID}" | python3 << 'PYTHON_SCRIPT'
import sys, json
data = json.load(sys.stdin)
event = data.get('data', {})
checks = event.get('quality_checks', [])
print(f"Event ID: {event.get('id')}")
print(f"Repository: {event.get('repository')}")
print(f"Branch: {event.get('branch')}")
print(f"Status: {event.get('event_status')}")
print("")
print("All quality checks:")
for c in checks:
    print(f"  {c['check_type']:35} stage={c['stage']:20} status={c['check_status']:12}")
PYTHON_SCRIPT

    return $RESULT
}

# 执行主流程
main "$@"
