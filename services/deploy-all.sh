#!/bin/bash
# 双引擎质量网关微服务 - 统一部署脚本
# 用法: ./deploy-all.sh [command] [options]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICES_DIR="${SCRIPT_DIR}"
NETWORK_NAME="quality-gateway"

# 服务列表（按依赖顺序）
declare -A SERVICES=(
    ["webhook-gateway"]="4001"
    ["event-store"]="4002"
    ["task-scheduler"]="4003"
    ["executor-service"]="4004"
    ["ai-analyzer"]="4005"
    ["resource-manager"]="4006"
)

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查 Docker
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装"
        exit 1
    fi
}

# 创建网络
create_network() {
    if ! docker network inspect "${NETWORK_NAME}" &> /dev/null; then
        log_info "创建 Docker 网络: ${NETWORK_NAME}"
        docker network create "${NETWORK_NAME}"
    fi
}

# 部署单个服务
deploy_service() {
    local service="$1"
    local service_dir="${SERVICES_DIR}/${service}"

    if [ ! -d "${service_dir}" ]; then
        log_warn "服务目录不存在: ${service_dir}"
        return 1
    fi

    if [ ! -f "${service_dir}/deploy.sh" ]; then
        log_warn "部署脚本不存在: ${service_dir}/deploy.sh"
        return 1
    fi

    log_step "部署服务: ${service}"

    # 尝试不同的部署命令
    (cd "${service_dir}" && bash deploy.sh deploy 2>/dev/null) || \
    (cd "${service_dir}" && bash deploy.sh upgrade 2>/dev/null) || \
    (cd "${service_dir}" && bash deploy.sh reinstall 2>/dev/null) || \
    log_warn "部署 ${service} 失败"
}

# 停止单个服务
stop_service() {
    local service="$1"
    local service_dir="${SERVICES_DIR}/${service}"
    local container_name="${service}"

    if docker ps -a --format '{{.Names}}' | grep -q "^${container_name}$"; then
        log_info "停止服务: ${service}"
        docker stop "${container_name}" || true
        docker rm "${container_name}" || true
    fi
}

# 显示服务状态
show_status() {
    log_step "服务状态:"
    echo ""

    printf "%-20s %-10s %-10s\n" "Service" "Port" "Status"
    printf "%-20s %-10s %-10s\n" "-------" "----" "------"

    for service in "${!SERVICES[@]}"; do
        local port="${SERVICES[$service]}"
        local container_name="${service}"
        local status="stopped"

        if docker ps --format '{{.Names}}' | grep -q "^${container_name}$"; then
            status="running"
        elif docker ps -a --format '{{.Names}}' | grep -q "^${container_NAME}$"; then
            status="exited"
        fi

        printf "%-20s %-10s %-10s\n" "${service}" "${port}" "${status}"
    done

    echo ""
    log_info "容器列表:"
    docker ps --filter "network=${NETWORK_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || true
}

# 初始化数据库
init_database() {
    local schema_dir="${SCRIPT_DIR}/services/deployments/schema"
    local init_script="${schema_dir}/init_db.sh"

    if [ -f "${init_script}" ]; then
        log_step "初始化数据库..."
        bash "${init_script}" \
            "${MYSQL_HOST:-localhost}" \
            "${MYSQL_PORT:-3306}" \
            "${MYSQL_USER:-root}" \
            "${MYSQL_PASSWORD:-}"
    else
        log_warn "数据库初始化脚本不存在"
    fi
}

# 构建所有服务
build_all() {
    log_step "构建所有服务镜像..."

    for service in "${!SERVICES[@]}"; do
        local service_dir="${SERVICES_DIR}/${service}"

        if [ -d "${service_dir}" ] && [ -f "${service_dir}/deploy.sh" ]; then
            log_info "构建: ${service}"
            (cd "${service_dir}" && bash deploy.sh build-only 2>/dev/null || \
             docker build -t "${service}:latest" "${service_dir}" 2>/dev/null || \
             log_warn "构建失败: ${service}")
        fi
    done
}

# 部署所有服务
deploy_all() {
    log_step "部署所有服务..."

    check_docker
    create_network

    for service in "${!SERVICES[@]}"; do
        deploy_service "${service}"
        sleep 2
    done

    log_info "所有服务部署完成！"
    show_status
}

# 停止所有服务
stop_all() {
    log_step "停止所有服务..."

    for service in "${!SERVICES[@]}"; do
        stop_service "${service}"
    done

    log_info "所有服务已停止"
}

# 清理所有服务
clean_all() {
    log_warn "清理所有服务..."
    stop_all

    log_warn "删除网络: ${NETWORK_NAME}"
    docker network rm "${NETWORK_NAME}" 2>/dev/null || true

    log_info "清理完成"
}

# 健康检查
health_check() {
    log_step "健康检查..."

    for service in "${!SERVICES[@]}"; do
        local port="${SERVICES[$service]}"
        local endpoint="http://localhost:${port}/health"

        if curl -sf "${endpoint}" > /dev/null 2>&1; then
            log_info "[OK] ${service} - 端口 ${port}"
        else
            log_warn "[FAIL] ${service} - 端口 ${port}"
        fi
    done
}

# 显示帮助
show_help() {
    cat << EOF
双引擎质量网关微服务 - 统一部署脚本

用法: $0 <command> [options]

命令:
  deploy-all    部署所有服务
  deploy <svc>  部署单个服务 (例如: deploy event-store)
  stop-all      停止所有服务
  stop <svc>    停止单个服务
  status        显示服务状态
  build         构建所有服务镜像
  init-db       初始化数据库
  health        健康检查
  clean         清理所有服务和网络
  help          显示此帮助

服务列表:
EOF

    for service in "${!SERVICES[@]}"; do
        local port="${SERVICES[$service]}"
        echo "  ${service} - 端口 ${port}"
    done

    cat << EOF

环境变量:
  MYSQL_HOST     MySQL 主机地址 (默认: localhost)
  MYSQL_PORT     MySQL 端口 (默认: 3306)
  MYSQL_USER     MySQL 用户 (默认: root)
  MYSQL_PASSWORD MySQL 密码

示例:
  $0 deploy-all          # 部署所有服务
  $0 deploy event-store   # 仅部署 event-store
  $0 status               # 查看状态
  $0 health               # 健康检查
EOF
}

# 主逻辑
case "${1:-help}" in
    deploy-all)
        deploy_all
        ;;
    deploy)
        if [ -z "$2" ]; then
            log_error "请指定服务名称"
            exit 1
        fi
        check_docker
        create_network
        deploy_service "$2"
        ;;
    stop-all)
        stop_all
        ;;
    stop)
        if [ -z "$2" ]; then
            log_error "请指定服务名称"
            exit 1
        fi
        stop_service "$2"
        ;;
    status)
        show_status
        ;;
    build)
        build_all
        ;;
    init-db)
        init_database
        ;;
    health)
        health_check
        ;;
    clean)
        clean_all
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "未知命令: $1"
        show_help
        exit 1
        ;;
esac
