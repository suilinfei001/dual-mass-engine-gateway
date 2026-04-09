#!/bin/bash
# 数据迁移脚本
# 从旧系统数据库迁移数据到新的微服务数据库

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
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 MySQL 连接
check_mysql_connection() {
    local host=$1
    local port=$2
    local user=$3
    local pass=$4

    if [ -n "$pass" ]; then
        mysql -h"$host" -P"$port" -u"$user" -p"$pass" -e "SELECT 1" &>/dev/null
    else
        mysql -h"$host" -P"$port" -u"$user" -e "SELECT 1" &>/dev/null
    fi

    return $?
}

# 导出数据
export_data() {
    local table=$1
    local output_file=$2

    log_info "导出表 $table ..."

    if [ -n "$OLD_DB_PASS" ]; then
        mysqldump -h"$OLD_DB_HOST" -P"$OLD_DB_PORT" -u"$OLD_DB_USER" -p"$OLD_DB_PASS" \
            --single-transaction --quick --lock-tables=false \
            "$OLD_DB_NAME" "$table" > "$output_file"
    else
        mysqldump -h"$OLD_DB_HOST" -P"$OLD_DB_PORT" -u"$OLD_DB_USER" \
            --single-transaction --quick --lock-tables=false \
            "$OLD_DB_NAME" "$table" > "$output_file"
    fi

    if [ $? -eq 0 ]; then
        log_info "导出 $table 成功"
    else
        log_error "导出 $table 失败"
        return 1
    fi
}

# 导入数据到新数据库
import_data() {
    local database=$1
    local input_file=$2

    log_info "导入数据到 $database ..."

    if [ -n "$NEW_DB_PASS" ]; then
        mysql -h"$NEW_DB_HOST" -P"$NEW_DB_PORT" -u"$NEW_DB_USER" -p"$NEW_DB_PASS" "$database" < "$input_file"
    else
        mysql -h"$NEW_DB_HOST" -P"$NEW_DB_PORT" -u"$NEW_DB_USER" "$database" < "$input_file"
    fi

    if [ $? -eq 0 ]; then
        log_info "导入到 $database 成功"
    else
        log_error "导入到 $database 失败"
        return 1
    fi
}

# 创建备份目录
BACKUP_DIR="./migrations/backup/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

log_info "备份目录: $BACKUP_DIR"
log_info "开始数据迁移..."

# 检查旧数据库连接
log_info "检查旧数据库连接 ($OLD_DB_HOST:$OLD_DB_PORT)..."
if ! check_mysql_connection "$OLD_DB_HOST" "$OLD_DB_PORT" "$OLD_DB_USER" "$OLD_DB_PASS"; then
    log_error "无法连接到旧数据库"
    exit 1
fi

# 检查新数据库连接
log_info "检查新数据库连接 ($NEW_DB_HOST:$NEW_DB_PORT)..."
if ! check_mysql_connection "$NEW_DB_HOST" "$NEW_DB_PORT" "$NEW_DB_USER" "$NEW_DB_PASS"; then
    log_error "无法连接到新数据库"
    exit 1
fi

# 迁移 Event Store 数据
log_info "========== 迁移 Event Store 数据 =========="
export_data "events" "$BACKUP_DIR/events.sql"
export_data "quality_checks" "$BACKUP_DIR/quality_checks.sql"

import_data "event_store_db" "$BACKUP_DIR/events.sql"
import_data "event_store_db" "$BACKUP_DIR/quality_checks.sql"

# 迁移 Task Scheduler 数据
log_info "========== 迁移 Task Scheduler 数据 =========="
export_data "tasks" "$BACKUP_DIR/tasks.sql"
export_data "task_results" "$BACKUP_DIR/task_results.sql"
export_data "task_executions" "$BACKUP_DIR/task_executions.sql"

import_data "task_scheduler_db" "$BACKUP_DIR/tasks.sql"
import_data "task_scheduler_db" "$BACKUP_DIR/task_results.sql"
import_data "task_scheduler_db" "$BACKUP_DIR/task_executions.sql"

# 迁移 Resource Manager 数据
log_info "========== 迁移 Resource Manager 数据 =========="
export_data "users" "$BACKUP_DIR/users.sql"
export_data "categories" "$BACKUP_DIR/categories.sql"
export_data "quota_policies" "$BACKUP_DIR/quota_policies.sql"
export_data "testbeds" "$BACKUP_DIR/testbeds.sql"
export_data "resource_instances" "$BACKUP_DIR/resource_instances.sql"
export_data "allocations" "$BACKUP_DIR/allocations.sql"
export_data "deployment_tasks" "$BACKUP_DIR/deployment_tasks.sql"
export_data "sessions" "$BACKUP_DIR/sessions.sql"

import_data "resource_manager_db" "$BACKUP_DIR/users.sql"
import_data "resource_manager_db" "$BACKUP_DIR/categories.sql"
import_data "resource_manager_db" "$BACKUP_DIR/quota_policies.sql"
import_data "resource_manager_db" "$BACKUP_DIR/testbeds.sql"
import_data "resource_manager_db" "$BACKUP_DIR/resource_instances.sql"
import_data "resource_manager_db" "$BACKUP_DIR/allocations.sql"
import_data "resource_manager_db" "$BACKUP_DIR/deployment_tasks.sql"
import_data "resource_manager_db" "$BACKUP_DIR/sessions.sql"

log_info "========== 数据迁移完成 =========="
log_info "备份文件保存在: $BACKUP_DIR"
