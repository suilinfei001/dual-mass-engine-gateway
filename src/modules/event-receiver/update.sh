#!/bin/bash
# Event Receiver 统一更新脚本
# 在本地 (10.4.174.125) 执行，完成构建和远程部署
#
# 使用方法:
#   ./update.sh                          # 完整更新（构建+部署）
#   ./update.sh -b                      # 仅构建
#   ./update.sh -d                      # 仅部署
#   ./update.sh -r                      # 恢复模式部署

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================================
# 配置
# ============================================================
# 获取项目根目录（从 event-receiver 向上三级）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
INSTALL_DIR="${PROJECT_ROOT}/install"

BUILD_ONLY=false
DEPLOY_ONLY=false
DEPLOY_MODE="upgrade"

while getopts "bdrh" opt; do
    case $opt in
        b)
            BUILD_ONLY=true
            ;;
        d)
            DEPLOY_ONLY=true
            ;;
        r)
            DEPLOY_MODE="recover"
            ;;
        h)
            echo "Event Receiver 统一更新脚本"
            echo ""
            echo "使用方法:"
            echo "  ./update.sh              # 完整更新（构建+部署）"
            echo "  ./update.sh -b          # 仅构建"
            echo "  ./update.sh -d          # 仅部署"
            echo "  ./update.sh -r          # 恢复模式部署"
            echo ""
            exit 0
            ;;
        \?)
            echo -e "${RED}无效选项: -$OPTARG${NC}"
            exit 1
            ;;
    esac
done

# ============================================================
# 打印标题
# ============================================================
print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  $1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

# ============================================================
# 构建步骤
# ============================================================
if [ "$DEPLOY_ONLY" != true ]; then
    print_header "步骤 1/2: 构建镜像"

    if [ -f "$INSTALL_DIR/build-quality.sh" ]; then
        cd "$INSTALL_DIR"
        ./build-quality.sh
        cd "$SCRIPT_DIR"
    else
        echo -e "${RED}✗${NC} build-quality.sh 不存在 ($INSTALL_DIR)"
        exit 1
    fi
fi

# ============================================================
# 部署步骤
# ============================================================
if [ "$BUILD_ONLY" != true ]; then
    print_header "步骤 2/2: 部署到远程"

    if [ "$DEPLOY_MODE" = "recover" ]; then
        echo -e "${YELLOW}警告: 恢复模式将清空数据库！${NC}"
        echo -ne "确认继续? [y/N] "
        read -r confirm
        if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
            echo "操作已取消"
            exit 0
        fi
    fi

    if [ -f "$INSTALL_DIR/deploy-quality-remote.sh" ]; then
        cd "$INSTALL_DIR"
        if [ "$DEPLOY_MODE" = "recover" ]; then
            ./deploy-quality-remote.sh -r
        else
            ./deploy-quality-remote.sh
        fi
        cd "$SCRIPT_DIR"
    else
        echo -e "${RED}✗${NC} deploy-quality-remote.sh 不存在 ($INSTALL_DIR)"
        exit 1
    fi
fi

# ============================================================
# 完成
# ============================================================
print_header "更新完成！"

echo -e "${BLUE}访问地址:${NC}"
echo -e "  Frontend:    ${GREEN}http://10.4.111.141:8081${NC}"
echo -e "  Backend API: ${GREEN}http://10.4.111.141:5001${NC}"
echo ""
