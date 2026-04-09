#!/bin/bash
# 双引擎质量网关 - 一键启动所有服务
#
# 使用方法:
#   ./start-all.sh                  # 编译并启动所有服务
#   ./start-all.sh --no-build       # 跳过编译，直接启动

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

# ============================================================
# 参数解析
# ============================================================
SKIP_BUILD=false

for arg in "$@"; do
    case $arg in
        --no-build)
            SKIP_BUILD=true
            shift
            ;;
        *)
            ;;
    esac
done

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  双引擎质量网关 - 微服务启动${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# ============================================================
# 步骤 1: 编译所有服务
# ============================================================
if [ "$SKIP_BUILD" = false ]; then
    echo -e "${YELLOW}[步骤 1/3] 编译所有服务...${NC}"
    echo ""
    "${SCRIPT_DIR}/build-all.sh"
    echo ""
else
    echo -e "${YELLOW}[步骤 1/3] 跳过编译${NC}"
    echo ""
fi

# ============================================================
# 步骤 2: 拉取基础镜像
# ============================================================
echo -e "${YELLOW}[步骤 2/3] 拉取基础镜像...${NC}"
echo ""

echo -e "  拉取 ${BLUE}alpine${NC}..."
docker pull acr.aishu.cn/dual-mass-engine-gateway/alpine:3.19 >/dev/null 2>&1 || true
echo -e "  ${GREEN}✓${NC} alpine"

echo -e "  拉取 ${BLUE}mysql${NC}..."
docker pull acr.aishu.cn/dual-mass-engine-gateway/mysql >/dev/null 2>&1 || true
echo -e "  ${GREEN}✓${NC} mysql"

echo ""

# ============================================================
# 步骤 3: 启动服务
# ============================================================
echo -e "${YELLOW}[步骤 3/3] 启动服务...${NC}"
echo ""

cd "$SCRIPT_DIR"
docker-compose up -d

echo ""

# 等待服务启动
echo -e "${YELLOW}等待服务启动...${NC}"
sleep 5

# ============================================================
# 显示状态
# ============================================================
echo ""
echo -e "${YELLOW}[服务状态]${NC}"
echo ""
docker-compose ps

echo ""

# ============================================================
# 显示访问地址
# ============================================================
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  启动完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}访问地址:${NC}"
echo -e "  Webhook Gateway:    ${GREEN}http://localhost:4001${NC}"
echo -e "  Event Store:        ${GREEN}http://localhost:4002${NC}"
echo -e "  Task Scheduler:     ${GREEN}http://localhost:4003${NC}"
echo -e "  Executor Service:   ${GREEN}http://localhost:4004${NC}"
echo -e "  AI Analyzer:        ${GREEN}http://localhost:4005${NC}"
echo -e "  Resource Manager:   ${GREEN}http://localhost:4006${NC}"
echo ""
echo -e "${BLUE}数据库端口:${NC}"
echo -e "  Event Store DB:     ${GREEN}3308${NC}"
echo -e "  Task Scheduler DB:  ${GREEN}3309${NC}"
echo -e "  Resource Manager:   ${GREEN}3310${NC}"
echo ""
echo -e "${BLUE}常用命令:${NC}"
echo -e "  查看日志:       docker-compose logs -f [service]"
echo -e "  停止所有:       docker-compose stop"
echo -e "  停止并删除:     docker-compose down"
echo -e "  停止并删除数据: docker-compose down -v"
echo ""
