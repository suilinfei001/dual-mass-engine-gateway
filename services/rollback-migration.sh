#!/bin/bash
# 数据迁移回滚脚本
# 恢复到迁移前的状态

set -e

# 配置
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

# 清空表数据
truncate_table() {
    local database=$1
    local table=$2

    log_info "清空表 $database.$table"

    if [ -n "$NEW_DB_PASS" ]; then
        mysql -h"$NEW_DB_HOST" -P"$NEW_DB_PORT" -u"$NEW_DB_USER" -p"$NEW_DB_PASS" \
            -e "SET FOREIGN_KEY_CHECKS=0; TRUNCATE TABLE $database.$table; SET FOREIGN_KEY_CHECKS=1;"
    else
        mysql -h"$NEW_DB_HOST" -P"$NEW_DB_PORT" -u"$NEW_DB_USER" \
            -e "SET FOREIGN_KEY_CHECKS=0; TRUNCATE TABLE $database.$table; SET FOREIGN_KEY_CHECKS=1;"
    fi
}

log_warn "========== 数据迁移回滚 =========="
log_warn "此操作将清空所有新数据库中的数据"
echo
read -p "确认回滚? (输入 YES 继续): " confirm

if [ "$confirm" != "YES" ]; then
    log_info "回滚已取消"
    exit 0
fi

# 回滚 Event Store 数据
log_info "回滚 Event Store 数据..."
truncate_table "event_store_db" "quality_checks"
truncate_table "event_store_db" "events"

# 回滚 Task Scheduler 数据
log_info "回滚 Task Scheduler 数据..."
truncate_table "task_scheduler_db" "task_executions"
truncate_table "task_scheduler_db" "task_results"
truncate_table "task_scheduler_db" "tasks"

# 回滚 Resource Manager 数据
log_info "回滚 Resource Manager 数据..."
truncate_table "resource_manager_db" "sessions"
truncate_table "resource_manager_db" "deployment_tasks"
truncate_table "resource_manager_db" "allocations"
truncate_table "resource_manager_db" "resource_instances"
truncate_table "resource_manager_db" "testbeds"
truncate_table "resource_manager_db" "quota_policies"
truncate_table "resource_manager_db" "categories"
truncate_table "resource_manager_db" "users"

log_info "========== 回滚完成 =========="
