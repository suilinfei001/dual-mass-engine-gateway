#!/bin/bash
# 双引擎质量网关 - 编译所有服务
# 为 Docker Compose 准备编译好的二进制文件
#
# 使用方法:
#   ./build-all.sh                  # 编译所有服务
#   ./build-all.sh webhook-gateway  # 仅编译指定服务

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ============================================================
# 配置
# ============================================================
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVICES_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# 服务列表（按依赖顺序）
ALL_SERVICES=(
    "event-store"
    "webhook-gateway"
    "executor-service"
    "ai-analyzer"
    "task-scheduler"
    "resource-manager"
)

# ============================================================
# 参数解析
# ============================================================
TARGET_SERVICES=()

if [ $# -eq 0 ]; then
    TARGET_SERVICES=("${ALL_SERVICES[@]}")
else
    TARGET_SERVICES=("$@")
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  编译微服务${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}待编译服务: ${TARGET_SERVICES[*]}${NC}"
echo ""

# ============================================================
# 编译函数
# ============================================================
build_service() {
    local service=$1
    local service_dir="${SERVICES_DIR}/${service}"

    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}✗${NC} 服务目录不存在: $service_dir"
        return 1
    fi

    echo -e "${YELLOW}编译 ${service}...${NC}"

    # 创建 output 目录
    mkdir -p "${service_dir}/output"

    # 设置 Go 环境
    export CGO_ENABLED=0
    export GOOS=linux
    export GOPROXY=off

    # 编译
    cd "$service_dir"
    go build -mod=readonly -ldflags '-extldflags "-static"' \
        -o "output/${service}" ./cmd/server

    if [ $? -eq 0 ]; then
        # 显示二进制文件大小
        local size=$(du -h "output/${service}" | cut -f1)
        echo -e "  ${GREEN}✓${NC} 编译完成 (大小: ${size})"
        return 0
    else
        echo -e "  ${RED}✗${NC} 编译失败"
        return 1
    fi
}

# ============================================================
# 主流程
# ============================================================
FAILED_SERVICES=()
SUCCESS_COUNT=0

for service in "${TARGET_SERVICES[@]}"; do
    if build_service "$service"; then
        ((SUCCESS_COUNT++))
    else
        FAILED_SERVICES+=("$service")
    fi
    echo ""
done

# ============================================================
# 总结
# ============================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  编译总结${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

if [ $SUCCESS_COUNT -gt 0 ]; then
    echo -e "${GREEN}✓ 成功: ${SUCCESS_COUNT} 个服务${NC}"
fi

if [ ${#FAILED_SERVICES[@]} -gt 0 ]; then
    echo -e "${RED}✗ 失败: ${#FAILED_SERVICES[@]} 个服务${NC}"
    echo -e "${RED}  ${FAILED_SERVICES[*]}${NC}"
    echo ""
    exit 1
fi

echo ""
echo -e "${GREEN}所有服务编译完成！${NC}"
echo ""
echo -e "${BLUE}下一步:${NC}"
echo -e "  1. 检查编译结果: ls -la */output/"
echo -e "  2. 启动服务: docker-compose up -d"
echo -e "  3. 查看状态: docker-compose ps"
echo ""
