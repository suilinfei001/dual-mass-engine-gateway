#!/bin/bash
# Event Processor 部署脚本
# 使用本地 Docker 镜像，不使用 docker-compose
#
# 使用方法:
#   ./deploy-event-processor.sh              # 首次部署或升级
#   ./deploy-event-processor.sh -r           # 完全重装（删除旧容器）
#   ./deploy-event-processor.sh -h           # 显示帮助信息

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ============================================================
# 配置
# ============================================================
# 获取脚本所在目录（event-processor 模块目录）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
MODULE_DIR="$SCRIPT_DIR"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
EVENT_RECEIVER_API="${EVENT_RECEIVER_API:-}"

BACKEND_PORT=5003
FRONTEND_PORT=8082
MYSQL_PORT=3307

CONTAINERS=(
    "event-processor-mysql"
    "event-processor-server"
    "event-processor-frontend"
)

LEGACY_CONTAINERS=(
    "processor-mysql"
    "processor-server"
    "processor-frontend"
)

IMAGES=(
    "acr.aishu.cn/dual-mass-engine-gateway/mysql:latest"
    "event-processor-backend:latest"
    "event-processor-frontend:latest"
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
            echo "Event Processor 部署脚本"
            echo ""
            echo "使用方法:"
            echo "  ./deploy-event-processor.sh              # 升级模式（默认）：更新容器，保留数据"
            echo "  ./deploy-event-processor.sh -u        # 升级模式：更新容器，保留数据"
            echo "  ./deploy-event-processor.sh -r        # 恢复模式：完全重装，删除旧容器"
            echo "  ./deploy-event-processor.sh -h        # 显示帮助信息"
            echo ""
            echo "说明:"
            echo "  此脚本部署 Event Processor 到本地 Docker 环境"
            echo "  Event Processor 通过 REST API 与 Event Receiver 通信"
            echo "  环境变量 EVENT_RECEIVER_API 指定 Event Receiver 的 API 地址"
            echo "  默认: ${EVENT_RECEIVER_API}"
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
echo -e "${BLUE}  Event Processor 部署${NC}"
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

echo -e "${BLUE}Event Receiver API: ${GREEN}${EVENT_RECEIVER_API}${NC}"
echo ""

# ============================================================
# 步骤 1: 编译 Go 代码和 Vue 前端
# ============================================================
echo -e "${YELLOW}[步骤 1/4] 编译 Go 代码和 Vue 前端...${NC}"
echo ""

echo -e "  编译 Vue 前端..."
cd "${MODULE_DIR}/frontend"

# 检测 Node.js 安装路径
NODE_PATH=""
if [ -x "/root/node-v24.8.0-linux-x64/bin/node" ]; then
    NODE_PATH="/root/node-v24.8.0-linux-x64/bin"
elif [ -x "/tmp/node-v20.11.0-linux-x64/bin/node" ]; then
    NODE_PATH="/tmp/node-v20.11.0-linux-x64/bin"
elif command -v node &> /dev/null; then
    NODE_PATH=$(dirname $(command -v node))
fi

if [ -z "$NODE_PATH" ]; then
    echo -e "  ${RED}✗${NC} Node.js 未安装，请先安装"
    exit 1
fi

echo -e "  使用 Node.js: ${NODE_PATH}/node"
export PATH="${NODE_PATH}:$PATH"
npm install --registry=https://registry.npmmirror.com
if [ $? -ne 0 ]; then
    echo -e "  ${RED}✗${NC} npm install 失败"
    exit 1
fi
npm run build
if [ $? -ne 0 ]; then
    echo -e "  ${RED}✗${NC} Vue 前端构建失败"
    exit 1
fi
echo -e "  ${GREEN}✓${NC} 前端编译完成"
cd "${MODULE_DIR}"

echo ""

# 编译 Go 后端（在宿主机上编译，避免 Docker 内下载依赖）
echo -e "  编译 Go 后端..."
mkdir -p output

# 设置 Go 环境
export CGO_ENABLED=0
export GOOS=linux
export GOPROXY=off

# 使用本地 go mod cache 编译
go build -mod=readonly -ldflags '-extldflags "-static"' -o output/event-processor ./cmd/server
if [ $? -ne 0 ]; then
    echo -e "  ${RED}✗${NC} Go 后端编译失败"
    exit 1
fi
echo -e "  ${GREEN}✓${NC} Go 后端编译完成"

echo ""

# ============================================================
# 步骤 2: 构建 Docker 镜像
# ============================================================
echo -e "${YELLOW}[步骤 2/4] 构建 Docker 镜像...${NC}"
echo ""

# 检查必要文件
if [ ! -f "${MODULE_DIR}/Dockerfile_server" ]; then
    echo -e "  ${RED}✗${NC} Dockerfile_server 不存在"
    exit 1
fi

if [ ! -f "${MODULE_DIR}/Dockerfile_frontend_simple" ]; then
    echo -e "  ${RED}✗${NC} Dockerfile_frontend_simple 不存在"
    exit 1
fi

# 预先拉取基础镜像
echo -e "  ${YELLOW}预先拉取基础镜像...${NC}"
docker pull acr.aishu.cn/dual-mass-engine-gateway/alpine:3.19 >/dev/null 2>&1 || true
docker pull acr.aishu.cn/dual-mass-engine-gateway/nginx:latest >/dev/null 2>&1 || true
docker pull acr.aishu.cn/dual-mass-engine-gateway/mysql:latest >/dev/null 2>&1 || true
echo -e "  ${GREEN}✓${NC} 基础镜像拉取完成"
echo ""

# 构建后端镜像（使用预编译的二进制文件）
echo -e "  构建 event-processor-backend 镜像..."
build_output=$(docker build -f "${MODULE_DIR}/Dockerfile_server" -t event-processor-backend:latest "${MODULE_DIR}" 2>&1)
if [ $? -eq 0 ]; then
    echo "$build_output" | while IFS= read -r line; do echo -e "    $line"; done
    echo -e "  ${GREEN}✓${NC} 后端镜像构建完成"
else
    echo "$build_output" | while IFS= read -r line; do echo -e "    $line"; done
    echo -e "  ${RED}✗${NC} 后端镜像构建失败"
    exit 1
fi

# 构建前端镜像（使用预编译的 dist 目录）
echo -e "  构建 event-processor-frontend 镜像..."
build_output=$(docker build -f "${MODULE_DIR}/Dockerfile_frontend_simple" -t event-processor-frontend:latest "${MODULE_DIR}" 2>&1)
if [ $? -eq 0 ]; then
    echo "$build_output" | while IFS= read -r line; do echo -e "    $line"; done
    echo -e "  ${GREEN}✓${NC} 前端镜像构建完成"
else
    echo "$build_output" | while IFS= read -r line; do echo -e "    $line"; done
    echo -e "  ${RED}✗${NC} 前端镜像构建失败"
    exit 1
fi

echo ""
echo -e "  ${GREEN}✓${NC} 所有镜像构建完成"
echo ""

# ============================================================
# 步骤 2: 创建 Docker 网络
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
# 恢复模式：删除旧容器和数据卷
# ============================================================
if [ "$MODE" = "recover" ]; then
    echo -e "${YELLOW}[恢复模式] 清理现有容器和数据卷...${NC}"
    echo ""

    for container in "${CONTAINERS[@]}"; do
        if docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
            echo -e "  删除容器 ${YELLOW}${container}${NC}..."
            docker stop "$container" >/dev/null 2>&1 || true
            docker rm "$container" >/dev/null 2>&1
            echo -e "    ${GREEN}✓${NC} 已删除"
        fi
    done

    # 删除 MySQL 数据卷（确保重新初始化所有表）
    echo -e "  删除 MySQL 数据卷 ${YELLOW}event-processor-mysql-data${NC}..."
    docker volume rm event-processor-mysql-data >/dev/null 2>&1 || true
    echo -e "    ${GREEN}✓${NC} 数据卷已删除"

    # 删除并重建网络
    if docker network inspect "$NETWORK_NAME" &>/dev/null; then
        docker network rm "$NETWORK_NAME" >/dev/null 2>&1
        docker network create "$NETWORK_NAME"
        echo -e "  ${GREEN}✓${NC} Docker 网络已重建"
    fi
    echo ""
fi

# ============================================================
# 步骤 4: 启动容器
# ============================================================
echo -e "${YELLOW}[步骤 4/4] 启动容器...${NC}"
echo ""

# 升级模式：停止并删除旧容器（避免端口冲突）
if [ "$MODE" = "upgrade" ]; then
    echo -e "  ${YELLOW}检查并停止旧容器...${NC}"
    
    # 清理新命名方式的容器
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
    
    # 清理旧命名方式的容器（兼容历史版本）
    for container in "${LEGACY_CONTAINERS[@]}"; do
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

# 启动 MySQL 容器
echo -e "  启动 ${YELLOW}event-processor-mysql${NC}..."
docker run -d \
    --name event-processor-mysql \
    --network "$NETWORK_NAME" \
    -p ${MYSQL_PORT}:3306 \
    -e MYSQL_ROOT_PASSWORD=root123456 \
    -e MYSQL_DATABASE=event_processor \
    -v "${PROJECT_ROOT}/install/scripts:/docker-entrypoint-initdb.d:ro" \
    -v event-processor-mysql-data:/var/lib/mysql \
    --restart unless-stopped \
    acr.aishu.cn/dual-mass-engine-gateway/mysql:latest

echo -e "    ${GREEN}✓${NC} event-processor-mysql 已启动"

# 等待 MySQL 就绪
echo -e "  ${YELLOW}等待 MySQL 就绪...${NC}"
for i in {1..30}; do
    if docker exec event-processor-mysql mysql -uroot -proot123456 -e "SELECT 1" &>/dev/null; then
        echo -e "    ${GREEN}✓${NC} MySQL 已就绪"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "    ${RED}✗${NC} MySQL 启动超时"
        exit 1
    fi
    sleep 1
done
echo ""

# 启动后端容器
echo -e "  启动 ${YELLOW}event-processor-server${NC}..."
docker run -d \
    --name event-processor-server \
    --network "$NETWORK_NAME" \
    -p ${BACKEND_PORT}:5002 \
    -e EVENT_RECEIVER_API="$EVENT_RECEIVER_API" \
    -e DB_DSN="root:root123456@tcp(event-processor-mysql:3306)/event_processor?parseTime=true" \
    -v /tmp:/tmp \
    --restart unless-stopped \
    event-processor-backend:latest

echo -e "    ${GREEN}✓${NC} event-processor-server 已启动"

# 启动前端容器
echo -e "  启动 ${YELLOW}event-processor-frontend${NC}..."
docker run -d \
    --name event-processor-frontend \
    --network "$NETWORK_NAME" \
    -p ${FRONTEND_PORT}:80 \
    --restart unless-stopped \
    event-processor-frontend:latest

echo -e "    ${GREEN}✓${NC} event-processor-frontend 已启动"
echo ""

# ============================================================
# 显示状态
# ============================================================
echo -e "${YELLOW}[容器状态]${NC}"
echo ""
docker ps --filter "network=${NETWORK_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}访问地址:${NC}"
echo -e "  Backend API:  ${GREEN}http://localhost:${BACKEND_PORT}${NC}"
echo -e "  Frontend:     ${GREEN}http://localhost:${FRONTEND_PORT}${NC}"
echo -e "  MySQL:        ${GREEN}localhost:${MYSQL_PORT}${NC}"
echo ""
echo -e "${BLUE}配置信息:${NC}"
echo -e "  Event Receiver API: ${GREEN}${EVENT_RECEIVER_API}${NC}"
echo -e "  Network:           ${GREEN}${NETWORK_NAME}${NC}"
echo -e "  MySQL Database:     ${GREEN}event_processor${NC}"
echo ""
echo -e "${BLUE}常用命令:${NC}"
echo -e "  查看日志:       docker logs -f <container-name>"
echo -e "  停止服务:       docker stop event-processor-mysql event-processor-server event-processor-frontend"
echo -e "  重启服务:       $0"
echo -e "  完全重装:       $0 -r"
echo ""
