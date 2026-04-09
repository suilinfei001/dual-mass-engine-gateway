#!/bin/bash
# Auth Service 部署脚本

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVICE_NAME="auth-service"
CONTAINER_NAME="${SERVICE_NAME}"
IMAGE_NAME="${SERVICE_NAME}:latest"
SERVICE_PORT=4007
NETWORK_NAME="quality-gateway"

MODE="upgrade"

while getopts "ursh" opt; do
    case $opt in
        r) MODE="recover" ;;
        u) MODE="upgrade" ;;
        s) MODE="stop" ;;
        h)
            echo "Auth Service 部署脚本"
            echo "用法: ./deploy.sh [-u|-r|-s|-h]"
            exit 0
            ;;
    esac
done

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Auth Service 部署${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

if [ "$MODE" = "stop" ]; then
    echo -e "${YELLOW}[停止服务]${NC}"
    docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
    echo -e "  ${GREEN}✓${NC} 容器已停止"
    exit 0
fi

# 编译
echo -e "${YELLOW}[步骤 1/3] 编译...${NC}"
mkdir -p "${SCRIPT_DIR}/output"
cd "${SCRIPT_DIR}"
CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o output/auth-service ./cmd/server
echo -e "  ${GREEN}✓${NC} 编译完成"
echo ""

# 复制到根目录供 Dockerfile 使用
cp output/auth-service "${SCRIPT_DIR}/auth-service"

# 构建镜像
echo -e "${YELLOW}[步骤 2/3] 构建镜像...${NC}"
docker build -t "$IMAGE_NAME" "${SCRIPT_DIR}" >/dev/null 2>&1
echo -e "  ${GREEN}✓${NC} 镜像构建完成"
echo ""

# 创建网络
echo -e "${YELLOW}[步骤 3/3] 启动容器...${NC}"
docker network inspect "$NETWORK_NAME" &>/dev/null || docker network create "$NETWORK_NAME"

# 停止并删除旧容器
docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true
docker rm "$CONTAINER_NAME" >/dev/null 2>&1 || true

# 启动新容器
docker run -d \
    --name "$CONTAINER_NAME" \
    --network "$NETWORK_NAME" \
    -p "${SERVICE_PORT}:4007" \
    -e ADMIN_USER=admin \
    -e ADMIN_PASS=admin123 \
    --restart unless-stopped \
    "$IMAGE_NAME"

echo -e "  ${GREEN}✓${NC} 容器已启动"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}默认凭证:${NC}"
echo -e "  用户名: ${GREEN}admin${NC}"
echo -e "  密码:   ${GREEN}admin123${NC}"
echo ""
echo -e "${BLUE}访问地址:${NC}"
echo -e "  Auth API: ${GREEN}http://localhost:${SERVICE_PORT}${NC}"
echo ""
