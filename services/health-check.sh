#!/bin/bash
# 微服务健康检查脚本

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 服务端点配置
declare -A SERVICES=(
    ["Webhook Gateway"]="http://localhost:4001/health"
    ["Event Store"]="http://localhost:4002/health"
    ["Task Scheduler"]="http://localhost:4003/health"
    ["Executor Service"]="http://localhost:4004/health"
    ["AI Analyzer"]="http://localhost:4005/health"
    ["Resource Manager"]="http://localhost:4006/health"
)

# 超时设置（秒）
TIMEOUT=5

# 检查单个服务
check_service() {
    local name=$1
    local url=$2

    # 使用 curl 检查健康状态
    response=$(curl -s -o /dev/null -w "%{http_code}" --max-time $TIMEOUT "$url" 2>/dev/null)
    curl_exit=$?

    if [ $curl_exit -eq 0 ]; then
        if [ "$response" = "200" ]; then
            echo -e "${GREEN}✓${NC} $name - 健康 (HTTP $response)"
            return 0
        else
            echo -e "${YELLOW}⚠${NC} $name - HTTP $response"
            return 1
        fi
    else
        echo -e "${RED}✗${NC} $name - 无法连接"
        return 2
    fi
}

# 检查 Docker 容器状态
check_containers() {
    echo
    echo "=== Docker 容器状态 ==="

    docker ps --filter "name=webhook-gateway|event-store|task-scheduler|executor-service|ai-analyzer|resource-manager" \
        --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || \
        echo "无法获取容器状态"
}

# 检查数据库连接
check_database() {
    echo
    echo "=== 数据库连接检查 ==="

    local mysql_host="${MYSQL_HOST:-localhost}"
    local mysql_port="${MYSQL_PORT:-3307}"

    if docker ps | grep -q "mysql"; then
        echo -e "${GREEN}✓${NC} MySQL 容器运行中"
    elif mysql -h"$mysql_host" -P"$mysql_port" -u"${MYSQL_USER:-root}" -e "SELECT 1" &>/dev/null; then
        echo -e "${GREEN}✓${NC} MySQL 连接成功 ($mysql_host:$mysql_port)"
    else
        echo -e "${RED}✗${NC} MySQL 连接失败"
    fi
}

# 主函数
main() {
    echo "=== 微服务健康检查 ==="
    echo "时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo

    local total=${#SERVICES[@]}
    local healthy=0
    local warning=0
    local down=0

    for service in "${!SERVICES[@]}"; do
        check_service "$service" "${SERVICES[$service]}"
        case $? in
            0) healthy=$((healthy + 1)) ;;
            1) warning=$((warning + 1)) ;;
            2) down=$((down + 1)) ;;
        esac
    done

    echo
    echo "=== 汇总 ==="
    echo -e "健康: ${GREEN}$healthy${NC}/$total"
    [ $warning -gt 0 ] && echo -e "警告: ${YELLOW}$warning${NC}/$total"
    [ $down -gt 0 ] && echo -e "离线: ${RED}$down${NC}/$total"

    # 显示容器状态
    check_containers

    # 检查数据库
    check_database

    # 返回退出码
    if [ $down -gt 0 ]; then
        exit 2
    elif [ $warning -gt 0 ]; then
        exit 1
    else
        exit 0
    fi
}

main "$@"
