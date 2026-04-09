#!/bin/bash
# 双引擎质量网关 - 停止所有服务
#
# 使用方法:
#   ./stop-all.sh                  # 停止所有服务（保留数据）
#   ./stop-all.sh --clean         # 停止并删除容器和网络
#   ./stop-all.sh --purge         # 停止并删除容器、网络和数据卷

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
MODE="stop"

# ============================================================
# 参数解析
# ============================================================
for arg in "$@"; do
    case $arg in
        --clean)
            MODE="clean"
            shift
            ;;
        --purge)
            MODE="purge"
            shift
            ;;
        *)
            ;;
    esac
done

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  双引擎质量网关 - 停止服务${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

cd "$SCRIPT_DIR"

if [ "$MODE" = "stop" ]; then
    echo -e "${YELLOW}模式: 停止服务（保留数据和容器）${NC}"
    echo ""
    docker-compose stop
    echo ""
    echo -e "${GREEN}✓ 服务已停止${NC}"
    echo ""
    echo -e "${BLUE}提示: 使用 ./start-all.sh --no-build 重新启动${NC}"

elif [ "$MODE" = "clean" ]; then
    echo -e "${YELLOW}模式: 停止并删除容器（保留数据卷）${NC}"
    echo ""
    echo -e "${RED}警告: 此操作将删除所有容器！${NC}"
    echo -ne "确认继续? [y/N] "
    read -r confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "操作已取消"
        exit 0
    fi
    echo ""
    docker-compose down
    echo ""
    echo -e "${GREEN}✓ 容器和网络已删除${NC}"
    echo ""
    echo -e "${BLUE}提示: 数据卷已保留，使用 ./start-all.sh 重新启动${NC}"

elif [ "$MODE" = "purge" ]; then
    echo -e "${YELLOW}模式: 停止并删除所有内容（包括数据）${NC}"
    echo ""
    echo -e "${RED}警告: 此操作将删除所有容器、网络和数据卷！${NC}"
    echo -e "${RED}所有数据将丢失！${NC}"
    echo -ne "确认继续? [y/N] "
    read -r confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "操作已取消"
        exit 0
    fi
    echo ""
    docker-compose down -v
    echo ""
    echo -e "${GREEN}✓ 所有资源已删除${NC}"
    echo ""
    echo -e "${BLUE}提示: 下次启动将重新初始化数据库${NC}"
fi

echo ""
