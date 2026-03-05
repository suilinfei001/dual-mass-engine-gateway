#!/bin/bash
# Event Receiver 本地构建脚本
# 在本地机器 (10.4.174.125) 上构建前后端和 Docker 镜像
#
# 使用方法:
#   ./build-quality.sh                    # 构建所有镜像
#   ./build-quality.sh -s                # 跳过前端构建
#   ./build-quality.sh -b                # 仅构建后端
#   ./build-quality.sh -f                # 仅构建前端

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================================
# 配置
# ============================================================
# 获取项目根目录
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

# 模块路径
MODULE_DIR="${PROJECT_ROOT}/src/modules/event-receiver"

# Docker 仓库配置
REGISTRY="acr.aishu.cn"
REPOSITORY="dual-mass-engine-gateway"

# 镜像名称
BACKEND_IMAGE="${REGISTRY}/${REPOSITORY}/quality-server:latest"
FRONTEND_IMAGE="${REGISTRY}/${REPOSITORY}/quality-frontend:latest"

# ============================================================
# 参数解析
# ============================================================
SKIP_FRONTEND=false
BACKEND_ONLY=false
FRONTEND_ONLY=false

while getopts "sbfh" opt; do
    case $opt in
        s)
            SKIP_FRONTEND=true
            ;;
        b)
            BACKEND_ONLY=true
            ;;
        f)
            FRONTEND_ONLY=true
            ;;
        h)
            echo "Event Receiver 本地构建脚本"
            echo ""
            echo "使用方法:"
            echo "  ./build-quality.sh            # 构建所有镜像"
            echo "  ./build-quality.sh -s        # 跳过前端构建"
            echo "  ./build-quality.sh -b        # 仅构建后端"
            echo "  ./build-quality.sh -f        # 仅构建前端"
            echo ""
            exit 0
            ;;
        \?)
            echo -e "${RED}无效选项: -$OPTARG${NC}"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Event Receiver 本地构建${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${BLUE}目标镜像:${NC}"
echo -e "  后端: ${GREEN}${BACKEND_IMAGE}${NC}"
echo -e "  前端: ${GREEN}${FRONTEND_IMAGE}${NC}"
echo ""

# ============================================================
# 检查环境
# ============================================================
echo -e "${YELLOW}[1/5] 检查构建环境...${NC}"

if [ ! -d "$MODULE_DIR" ]; then
    echo -e "  ${RED}✗${NC} event-receiver 目录不存在: $MODULE_DIR"
    exit 1
fi

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "  ${RED}✗${NC} Docker 未安装"
    exit 1
fi
echo -e "  ${GREEN}✓${NC} Docker 已安装"

# 检查 Go (仅后端)
if [ "$FRONTEND_ONLY" != true ]; then
    if ! command -v go &> /dev/null; then
        echo -e "  ${RED}✗${NC} Go 未安装"
        exit 1
    fi
    echo -e "  ${GREEN}✓${NC} Go 已安装: $(go version | awk '{print $3}')"
fi

# 检查 Node.js (仅前端)
if [ "$BACKEND_ONLY" != true ] && [ "$SKIP_FRONTEND" != true ]; then
    if ! command -v node &> /dev/null; then
        echo -e "  ${RED}✗${NC} Node.js 未安装"
        exit 1
    fi
    echo -e "  ${GREEN}✓${NC} Node.js 已安装: $(node -v)"
fi
echo ""

# ============================================================
# 构建后端
# ============================================================
if [ "$FRONTEND_ONLY" != true ]; then
    echo -e "${YELLOW}[2/5] 构建后端...${NC}"
    echo ""

    cd "$MODULE_DIR"

    # 构建 Go 二进制
    echo -e "  构建 Go 二进制文件..."
    build_output=$(CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o quality-server ./cmd/quality-server 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} quality-server 二进制文件构建完成"
    else
        echo "$build_output" | while IFS= read -r line; do
            echo -e "    $line"
        done
        echo -e "  ${RED}✗${NC} quality-server 二进制文件构建失败"
        exit 1
    fi

    # 构建后端 Docker 镜像
    echo -e "  构建后端 Docker 镜像..."
    build_output=$(docker build -f Dockerfile_quality_server -t "$BACKEND_IMAGE" . 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} 后端镜像构建完成: ${BACKEND_IMAGE}"
    else
        echo "$build_output" | while IFS= read -r line; do
            echo -e "    $line"
        done
        echo -e "  ${RED}✗${NC} 后端镜像构建失败"
        exit 1
    fi

    cd "$PROJECT_ROOT"
    echo ""
fi

# ============================================================
# 构建前端
# ============================================================
if [ "$BACKEND_ONLY" != true ] && [ "$SKIP_FRONTEND" != true ]; then
    echo -e "${YELLOW}[3/5] 构建前端...${NC}"
    echo ""

    cd "$MODULE_DIR/frontend"

    # 配置 npm 镜像
    npm config set registry https://registry.npmmirror.com

    # 安装依赖
    if [ ! -d "node_modules" ]; then
        echo -e "  安装 npm 依赖..."
        install_output=$(npm install 2>&1)
        if [ $? -ne 0 ]; then
            echo "$install_output" | while IFS= read -r line; do
                echo -e "    $line"
            done
            echo -e "  ${RED}✗${NC} npm install 失败"
            exit 1
        fi
        echo -e "  ${GREEN}✓${NC} 依赖安装完成"
    else
        echo -e "  ${GREEN}✓${NC} 依赖已存在，跳过安装"
    fi

    # 构建前端
    echo -e "  构建前端..."
    build_output=$(npm run build 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} 前端构建完成"
    else
        echo "$build_output" | while IFS= read -r line; do
            echo -e "    $line"
        done
        echo -e "  ${RED}✗${NC} 前端构建失败"
        exit 1
    fi

    cd "$MODULE_DIR"

    # 构建前端 Docker 镜像
    echo -e "  构建前端 Docker 镜像..."
    build_output=$(docker build -f Dockerfile_frontend -t "$FRONTEND_IMAGE" . 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} 前端镜像构建完成: ${FRONTEND_IMAGE}"
    else
        echo "$build_output" | while IFS= read -r line; do
            echo -e "    $line"
        done
        echo -e "  ${RED}✗${NC} 前端镜像构建失败"
        exit 1
    fi

    cd "$PROJECT_ROOT"
    echo ""
fi

# ============================================================
# 推送镜像到仓库
# ============================================================
echo -e "${YELLOW}[4/4] 推送镜像到仓库...${NC}"
echo ""

if [ "$FRONTEND_ONLY" != true ]; then
    echo -e "  推送后端镜像..."
    push_output=$(docker push "$BACKEND_IMAGE" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} 后端镜像推送完成"
    else
        echo "$push_output" | while IFS= read -r line; do
            echo -e "    $line"
        done
        echo -e "  ${YELLOW}!${NC} 后端镜像推送失败，可能需要登录仓库"
        echo -e "  ${YELLOW}请运行: docker login ${REGISTRY}${NC}"
    fi
fi

if [ "$BACKEND_ONLY" != true ] && [ "$SKIP_FRONTEND" != true ]; then
    echo -e "  推送前端镜像..."
    push_output=$(docker push "$FRONTEND_IMAGE" 2>&1)
    if [ $? -eq 0 ]; then
        echo -e "  ${GREEN}✓${NC} 前端镜像推送完成"
    else
        echo "$push_output" | while IFS= read -r line; do
            echo -e "    $line"
        done
        echo -e "  ${YELLOW}!${NC} 前端镜像推送失败，可能需要登录仓库"
        echo -e "  ${YELLOW}请运行: docker login ${REGISTRY}${NC}"
    fi
fi
echo ""

# ============================================================
# 完成
# ============================================================
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  构建完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}镜像已推送到仓库:${NC}"
if [ "$FRONTEND_ONLY" != true ]; then
    echo -e "  ${GREEN}${BACKEND_IMAGE}${NC}"
fi
if [ "$BACKEND_ONLY" != true ] && [ "$SKIP_FRONTEND" != true ]; then
    echo -e "  ${GREEN}${FRONTEND_IMAGE}${NC}"
fi
echo ""
echo -e "${BLUE}下一步: 在目标服务器 (10.4.111.141) 上运行部署脚本${NC}"
echo -e "  ${YELLOW}./deploy-quality-remote.sh${NC}"
echo ""
