#!/bin/bash
# 数据迁移验证脚本
# 验证迁移后的数据完整性

set -e

# 配置
OLD_DB_HOST="${OLD_DB_HOST:-localhost}"
OLD_DB_PORT="${OLD_DB_PORT:-3306}"
OLD_DB_USER="${OLD_DB_USER:-root}"
OLD_DB_PASS="${OLD_DB_PASS:-}"
OLD_DB_NAME="${OLD_DB_NAME:-quality_db}"

NEW_DB_HOST="${NEW_DB_HOST:-localhost}"
NEW_DB_PORT="${NEW_DB_PORT:-3307}"
NEW_DB_USER="${NEW_DB_USER:-root}"
NEW_DB_PASS="${NEW_DB_PASS:-}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 获取表行数
get_table_count() {
    local host=$1
    local port=$2
    local user=$3
    local pass=$4
    local database=$5
    local table=$6

    if [ -n "$pass" ]; then
        mysql -h"$host" -P"$port" -u"$user" -p"$pass" -N -B -e \
            "SELECT COUNT(*) FROM $database.$table" 2>/dev/null
    else
        mysql -h"$host" -P"$port" -u"$user" -N -B -e \
            "SELECT COUNT(*) FROM $database.$table" 2>/dev/null
    fi
}

# 验证表数据
verify_table() {
    local table=$1
    local old_db=$2
    local new_db=$3

    local old_count=$(get_table_count "$OLD_DB_HOST" "$OLD_DB_PORT" "$OLD_DB_USER" "$OLD_DB_PASS" "$old_db" "$table")
    local new_count=$(get_table_count "$NEW_DB_HOST" "$NEW_DB_PORT" "$NEW_DB_USER" "$NEW_DB_PASS" "$new_db" "$table")

    echo -n "  $table: 旧=$old_count, 新=$new_count ... "

    if [ "$old_count" == "$new_count" ]; then
        echo -e "${GREEN}✓ 匹配${NC}"
        return 0
    else
        echo -e "${RED}✗ 不匹配${NC}"
        return 1
    fi
}

log_info "========== 数据迁移验证 =========="

# 验证 Event Store 数据
log_info "验证 Event Store 数据..."
verify_table "events" "$OLD_DB_NAME" "event_store_db"
verify_table "quality_checks" "$OLD_DB_NAME" "event_store_db"

# 验证 Task Scheduler 数据
log_info "验证 Task Scheduler 数据..."
verify_table "tasks" "$OLD_DB_NAME" "task_scheduler_db"
verify_table "task_results" "$OLD_DB_NAME" "task_scheduler_db"
verify_table "task_executions" "$OLD_DB_NAME" "task_scheduler_db"

# 验证 Resource Manager 数据
log_info "验证 Resource Manager 数据..."
verify_table "users" "$OLD_DB_NAME" "resource_manager_db"
verify_table "categories" "$OLD_DB_NAME" "resource_manager_db"
verify_table "quota_policies" "$OLD_DB_NAME" "resource_manager_db"
verify_table "testbeds" "$OLD_DB_NAME" "resource_manager_db"
verify_table "resource_instances" "$OLD_DB_NAME" "resource_manager_db"
verify_table "allocations" "$OLD_DB_NAME" "resource_manager_db"
verify_table "deployment_tasks" "$OLD_DB_NAME" "resource_manager_db"
verify_table "sessions" "$OLD_DB_NAME" "resource_manager_db"

log_info "========== 验证完成 =========="
