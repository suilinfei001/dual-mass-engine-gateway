#!/bin/bash
# Resource Pool 部署脚本
# 使用本地 Docker 镜像，不使用 docker-compose
#
# 使用方法:
#   ./deploy-resource-pool.sh              # 升级模式（默认）
#   ./deploy-resource-pool.sh -u           # 升级模式
#   ./deploy-resource-pool.sh -r           # 完全重装（删除容器）
#   ./deploy-resource-pool.sh -h           # 帮助信息

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ============================================================
# 配置
# ============================================================
# 获取脚本所在目录（resource-pool 模块目录）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MODULE_DIR="$SCRIPT_DIR"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

BACKEND_PORT=5004

CONTAINERS=(
    "resource-pool-server"
)

IMAGES=(
    "resource-pool-backend:latest"
)

NETWORK_NAME="processor-network"

# ============================================================
# 参数解析
# ============================================================
MODE="upgrade"

while getopts "ruh" opt; do
    case $opt in
        r)
            MODE="recover"
            ;;
        u)
            MODE="upgrade"
            ;;
        h)
            echo "Resource Pool 部署脚本"
            echo ""
            echo "使用方法:"
            echo "  ./deploy-resource-pool.sh              # 升级模式（默认）：更新容器，保留数据"
            echo "  ./deploy-resource-pool.sh -u        # 升级模式：更新容器，保留数据"
            echo "  ./deploy-resource-pool.sh -r        # 恢复模式：完全重装，删除旧容器"
            echo "  ./deploy-resource-pool.sh -h        # 显示帮助信息"
            echo ""
            echo "说明:"
            echo "  此脚本部署 Resource Pool 到本地 Docker 环境"
            echo "  Resource Pool 共享 event-processor 的 MySQL 数据库和用户系统"
            exit 0
            ;;
        \?)
            echo -e "${RED}无效选项: -$OPTARG${NC}"
            echo "使用 -h 查看帮助信息"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Resource Pool 部署${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

if [ "$MODE" = "recover" ]; then
    echo -e "${YELLOW}模式: ${RED}恢复模式 (完全重装)${NC}"
    echo -e "${RED}警告: 此操作将删除所有容器！${NC}"
    echo -ne "确认继续? [y/N] "
    read -r confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "操作已取消"
        exit 0
    fi
else
    echo -e "${YELLOW}模式: ${GREEN}升级模式 (更新容器)${NC}"
fi
echo ""

# ============================================================
# 步骤 1: 编译 Go 代码
# ============================================================
echo -e "${YELLOW}[步骤 1/3] 编译 Go 代码...${NC}"
echo ""

mkdir -p "${MODULE_DIR}/output"

# 设置 Go 环境
export CGO_ENABLED=0
export GOOS=linux
export GOPROXY=off

# 使用本地 go mod cache 编译
echo -e "  编译 Go 后端..."
cd "${MODULE_DIR}"
go build -mod=readonly -ldflags '-extldflags "-static"' -o output/resource-pool ./cmd/server
if [ $? -ne 0 ]; then
    echo -e "  ${RED}✗${NC} Go 后端编译失败"
    exit 1
fi
echo -e "  ${GREEN}✓${NC} Go 后端编译完成"

echo ""

# ============================================================
# 步骤 2: 构建 Docker 镜像
# ============================================================
echo -e "${YELLOW}[步骤 2/3] 构建 Docker 镜像...${NC}"
echo ""

# 检查必要文件
if [ ! -f "${MODULE_DIR}/Dockerfile_server" ]; then
    echo -e "  ${RED}✗${NC} Dockerfile_server 不存在"
    exit 1
fi

# 预先拉取基础镜像
echo -e "  ${YELLOW}预先拉取基础镜像...${NC}"
docker pull acr.aishu.cn/dual-mass-engine-gateway/alpine:3.19 >/dev/null 2>&1 || true
echo -e "  ${GREEN}✓${NC} 基础镜像拉取完成"
echo ""

# 构建后端镜像
echo -e "  构建 resource-pool-backend 镜像..."
build_output=$(docker build -f "${MODULE_DIR}/Dockerfile_server" -t resource-pool-backend:latest "${MODULE_DIR}" 2>&1)
if [ $? -eq 0 ]; then
    echo "$build_output" | while IFS= read -r line; do echo -e "    $line"; done
    echo -e "  ${GREEN}✓${NC} 后端镜像构建完成"
else
    echo "$build_output" | while IFS= read -r line; do echo -e "    $line"; done
    echo -e "  ${RED}✗${NC} 后端镜像构建失败"
    exit 1
fi

echo ""
echo -e "  ${GREEN}✓${NC} 所有镜像构建完成"
echo ""

# ============================================================
# 步骤 3: 创建 Docker 网络
# ============================================================
echo -e "${YELLOW}[步骤 3/3] 创建 Docker 网络...${NC}"

if ! docker network inspect "$NETWORK_NAME" &> /dev/null; then
    docker network create "$NETWORK_NAME"
    echo -e "  ${GREEN}✓${NC} 创建网络: $NETWORK_NAME"
else
    echo -e "  ${GREEN}✓${NC} 网络 $NETWORK_NAME 已存在"
fi
echo ""

# ============================================================
# 恢复模式：删除旧容器
# ============================================================
if [ "$MODE" = "recover" ]; then
    echo -e "${YELLOW}[恢复模式] 清理现有容器...${NC}"
    echo ""

    for container in "${CONTAINERS[@]}"; do
        if docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
            echo -e "  删除容器 ${YELLOW}${container}${NC}..."
            docker stop "$container" >/dev/null 2>&1 || true
            docker rm "$container" >/dev/null 2>&1
            echo -e "    ${GREEN}✓${NC} 已删除"
        fi
    done
    echo ""
fi

# ============================================================
# 升级模式：停止并删除旧容器
# ============================================================
if [ "$MODE" = "upgrade" ]; then
    echo -e "  ${YELLOW}检查并停止旧容器...${NC}"

    for container in "${CONTAINERS[@]}"; do
        if docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
            echo -e "    停止运行中的容器 ${YELLOW}${container}${NC}..."
            docker stop "$container" >/dev/null 2>&1 || true
        fi
        if docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
            echo -e "    删除旧容器 ${YELLOW}${container}${NC}..."
            docker rm "$container" >/dev/null 2>&1 || true
        fi
    done

    echo -e "  ${GREEN}✓${NC} 旧容器已清理"
    echo ""
fi

# ============================================================
# 步骤 4: 启动容器
# ============================================================
echo -e "${YELLOW}[步骤 4/4] 启动容器...${NC}"
echo ""

# 启动后端容器
echo -e "  启动 ${YELLOW}resource-pool-server${NC}..."
docker run -d \
    --name resource-pool-server \
    --network "$NETWORK_NAME" \
    -p ${BACKEND_PORT}:5003 \
    -e DB_DSN="root:root123456@tcp(event-processor-mysql:3306)/event_processor?parseTime=true&loc=Local" \
    -e DEPLOYER_TYPE=ssh \
    --restart unless-stopped \
    resource-pool-backend:latest

echo -e "    ${GREEN}✓${NC} resource-pool-server 已启动"
echo ""

# ============================================================
# 显示状态
# ============================================================
echo -e "${YELLOW}[容器状态]${NC}"
echo ""
docker ps --filter "name=resource-pool" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}访问地址:${NC}"
echo -e "  Backend API:  ${GREEN}http://localhost:${BACKEND_PORT}${NC}"
echo ""
echo -e "${BLUE}配置信息:${NC}"
echo -e "  Network:      ${GREEN}${NETWORK_NAME}${NC}"
echo -e "  MySQL:        共享 event-processor-mysql"
echo ""
echo -e "${BLUE}常用命令:${NC}"
echo -e "  查看日志:       docker logs -f resource-pool-server"
echo -e "  停止服务:       docker stop resource-pool-server"
echo -e "  重启服务:       $0"
echo -e "  完全重装:       $0 -r"
echo ""
