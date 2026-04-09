#!/bin/bash
# 双引擎质量网关微服务 - 数据库初始化脚本
# 用法: ./init_db.sh [mysql_host] [mysql_port] [mysql_user] [mysql_password]

set -e

# 默认配置
MYSQL_HOST="${1:-localhost}"
MYSQL_PORT="${2:-3306}"
MYSQL_USER="${3:-root}"
MYSQL_PASSWORD="${4:-}"

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
log_info "检查 MySQL 连接..."
if [ -z "$MYSQL_PASSWORD" ]; then
    MYSQL_CMD="mysql -h $MYSQL_HOST -P $MYSQL_PORT -u $MYSQL_USER"
else
    MYSQL_CMD="mysql -h $MYSQL_HOST -P $MYSQL_PORT -u $MYSQL_USER -p$MYSQL_PASSWORD"
fi

if ! $MYSQL_CMD -e "SELECT 1;" &> /dev/null; then
    log_error "无法连接到 MySQL 服务器"
    log_error "请检查: host=$MYSQL_HOST, port=$MYSQL_PORT, user=$MYSQL_USER"
    exit 1
fi

log_info "MySQL 连接成功！"

# 获取脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SQL_FILE="$SCRIPT_DIR/init_all_databases.sql"

if [ ! -f "$SQL_FILE" ]; then
    log_error "找不到 SQL 文件: $SQL_FILE"
    exit 1
fi

# 执行初始化脚本
log_info "开始初始化数据库..."
log_info "SQL 文件: $SQL_FILE"

if $MYSQL_CMD < "$SQL_FILE"; then
    log_info "数据库初始化完成！"
    echo ""
    log_info "已创建的数据库:"
    $MYSQL_CMD -e "SHOW DATABASES LIKE '%_db';" 2>/dev/null || true
    echo ""
    log_info "已创建的表:"
    echo "  - event_store_db: events, quality_checks"
    echo "  - task_scheduler_db: tasks, task_results, task_executions"
    echo "  - resource_manager_db: resources, categories, quota_policies, allocations, testbeds, deployment_tasks, users"
    echo ""
    log_info "默认用户:"
    echo "  - admin / admin@example.com (密码: admin123)"
    echo "  - system / system@example.com (密码: admin123)"
    echo ""
    log_warn "请在生产环境中修改默认密码！"
else
    log_error "数据库初始化失败！"
    exit 1
fi
