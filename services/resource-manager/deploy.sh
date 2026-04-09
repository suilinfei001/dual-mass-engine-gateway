#!/bin/bash
# Resource Manager Service 部署脚本
# 使用本地编译，Dockerfile 直接使用编译后的二进制
#
# 使用方法:
#   ./deploy.sh                    # 升级模式（默认）
#   ./deploy.sh -u                 # 升级模式
#   ./deploy.sh -r                 # 完全重装（删除容器）
#   ./deploy.sh -s                 # 仅停止
#   ./deploy.sh -h                 # 帮助信息

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
SERVICE_NAME="resource-manager"
CONTAINER_NAME="${SERVICE_NAME}"
IMAGE_NAME="${SERVICE_NAME}:latest"
SERVICE_PORT=4006
NETWORK_NAME="quality-network"

# ============================================================
# 参数解析
# ============================================================
MODE="upgrade"

while getopts "ursh" opt; do
    case $opt in
        r)
            MODE="recover"
            ;;
        u)
            MODE="upgrade"
            ;;
        s)
            MODE="stop"
            ;;
        h)
            echo "Resource Manager 部署脚本"
            echo ""
            echo "使用方法:"
            echo "  ./deploy.sh              # 升级模式（默认）：更新容器，保留数据"
            echo "  ./deploy.sh -u           # 升级模式：更新容器，保留数据"
            echo "  ./deploy.sh -r           # 恢复模式：完全重装，删除旧容器"
            echo "  ./deploy.sh -s           # 停止服务"
            echo "  ./deploy.sh -h           # 显示帮助信息"
            echo ""
            echo "说明:"
            echo "  此脚本在本地编译 Go 代码，然后构建 Docker 镜像"
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
echo -e "${BLUE}  Resource Manager 部署${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# ============================================================
# 停止模式
# ============================================================
if [ "$MODE" = "stop" ]; then
    echo -e "${YELLOW}[停止服务]${NC}"
    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo -e "  停止容器 ${YELLOW}${CONTAINER_NAME}${NC}..."
        docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
        echo -e "  ${GREEN}✓${NC} 容器已停止"
    else
        echo -e "  ${YELLOW}容器未运行${NC}"
    fi
    exit 0
fi

if [ "$MODE" = "recover" ]; then
    echo -e "${YELLOW}模式: ${RED}恢复模式 (完全重装)${NC}"
    echo -e "${RED}警告: 此操作将删除容器！${NC}"
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
echo -e "${YELLOW}[步骤 1/4] 编译 Go 代码...${NC}"
echo ""

mkdir -p "${SCRIPT_DIR}/output"

# 设置 Go 环境
export CGO_ENABLED=0
export GOOS=linux
export GOPROXY=off

# 使用本地 go mod cache 编译
echo -e "  编译 ${SERVICE_NAME}..."
cd "${SCRIPT_DIR}"
go build -mod=readonly -ldflags '-extldflags "-static"' -o output/resource-manager ./cmd/server
if [ $? -ne 0 ]; then
    echo -e "  ${RED}✗${NC} 编译失败"
    exit 1
fi
echo -e "  ${GREEN}✓${NC} 编译完成"

echo ""

# ============================================================
# 步骤 2: 构建 Docker 镜像
# ============================================================
echo -e "${YELLOW}[步骤 2/4] 构建 Docker 镜像...${NC}"
echo ""

# 预先拉取基础镜像
echo -e "  ${YELLOW}预先拉取基础镜像...${NC}"
docker pull acr.aishu.cn/dual-mass-engine-gateway/alpine:3.19 >/dev/null 2>&1 || true
echo -e "  ${GREEN}✓${NC} 基础镜像准备完成"
echo ""

# 构建镜像
echo -e "  构建 ${IMAGE_NAME} 镜像..."
build_output=$(docker build -f "${SCRIPT_DIR}/Dockerfile" -t "$IMAGE_NAME" "${SCRIPT_DIR}" 2>&1)
if [ $? -eq 0 ]; then
    echo -e "  ${GREEN}✓${NC} 镜像构建完成"
else
    echo "$build_output" | while IFS= read -r line; do echo -e "    $line"; done
    echo -e "  ${RED}✗${NC} 镜像构建失败"
    exit 1
fi

echo ""

# ============================================================
# 步骤 3: 创建 Docker 网络
# ============================================================
echo -e "${YELLOW}[步骤 3/4] 创建 Docker 网络...${NC}"

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

    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo -e "  删除容器 ${YELLOW}${CONTAINER_NAME}${NC}..."
        docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
        docker rm "$CONTAINER_NAME" >/dev/null 2>&1
        echo -e "    ${GREEN}✓${NC} 已删除"
    fi
    echo ""
fi

# ============================================================
# 升级模式：停止并删除旧容器
# ============================================================
if [ "$MODE" = "upgrade" ]; then
    echo -e "  ${YELLOW}检查并停止旧容器...${NC}"

    if docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo -e "    停止运行中的容器 ${YELLOW}${CONTAINER_NAME}${NC}..."
        docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
    fi
    if docker ps -a --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
        echo -e "    删除旧容器 ${YELLOW}${CONTAINER_NAME}${NC}..."
        docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true
    fi

    echo -e "  ${GREEN}✓${NC} 旧容器已清理"
    echo ""
fi

# ============================================================
# 步骤 4: 启动容器
# ============================================================
echo -e "${YELLOW}[步骤 4/4] 启动容器...${NC}"
echo ""

echo -e "  启动 ${YELLOW}${CONTAINER_NAME}${NC}..."
docker run -d \
    --name "$CONTAINER_NAME" \
    --network "$NETWORK_NAME" \
    -p "${SERVICE_PORT}:4006" \
    -e DB_HOST="resource-manager-mysql" \
    -e DB_PORT="3306" \
    -e DB_USER="root" \
    -e DB_PASSWORD="root123456" \
    -e DB_NAME="resource_manager" \
    -e LOG_LEVEL=info \
    --restart unless-stopped \
    "$IMAGE_NAME"

echo -e "    ${GREEN}✓${NC} 容器已启动"
echo ""

# ============================================================
# 显示状态
# ============================================================
echo -e "${YELLOW}[容器状态]${NC}"
echo ""
docker ps --filter "name=${CONTAINER_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}访问地址:${NC}"
echo -e "  Backend API:  ${GREEN}http://localhost:${SERVICE_PORT}${NC}"
echo ""
echo -e "${BLUE}配置信息:${NC}"
echo -e "  Network:      ${GREEN}${NETWORK_NAME}${NC}"
echo -e "  Database:     host.docker.internal:3307/resource_manager"
echo ""
echo -e "${BLUE}常用命令:${NC}"
echo -e "  查看日志:       docker logs -f ${CONTAINER_NAME}"
echo -e "  停止服务:       $0 -s"
echo -e "  重启服务:       $0"
echo -e "  完全重装:       $0 -r"
echo ""
