#!/bin/bash
# Event Receiver 远程部署脚本
# 在本地机器 (10.4.174.125) 上通过 SSH 在远程服务器 (10.4.111.141) 部署
#
# 使用方法:
#   ./deploy-quality-remote.sh              # 部署模式（更新容器，保留数据）
#   ./deploy-quality-remote.sh -u          # 升级模式
#   ./deploy-quality-remote.sh -r          # 恢复模式（完全重装，清理数据库）
#   ./deploy-quality-remote.sh -l          # 仅传输镜像，不部署
#   ./deploy-quality-remote.sh -h          # 显示帮助信息

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# ============================================================
# 配置
# ============================================================
# 远程服务器配置
REMOTE_HOST="10.4.111.141"
REMOTE_USER="root"
REMOTE_DIR="/root/dual-mass-engine-gateway"

# Docker 仓库配置
REGISTRY="acr.aishu.cn"
REPOSITORY="dual-mass-engine-gateway"

BACKEND_IMAGE="${REGISTRY}/${REPOSITORY}/quality-server:latest"
FRONTEND_IMAGE="${REGISTRY}/${REPOSITORY}/quality-frontend:latest"

# 网络和容器配置
NETWORK_NAME="quality-network"
MYSQL_ROOT_PASSWORD="root123456"
MYSQL_DATABASE="github_hub"
MYSQL_PORT=3306
QUALITY_SERVER_PORT=5001
FRONTEND_PORT=8081

CONTAINERS=(
    "quality-frontend"
    "quality-server"
    "quality-mysql"
)

# ============================================================
# 参数解析
# ============================================================
MODE="upgrade"
TRANSFER_ONLY=false

while getopts "urlh" opt; do
    case $opt in
        r)
            MODE="recover"
            ;;
        u)
            MODE="upgrade"
            ;;
        l)
            TRANSFER_ONLY=true
            ;;
        h)
            echo "Event Receiver 远程部署脚本"
            echo ""
            echo "使用方法:"
            echo "  ./deploy-quality-remote.sh          # 部署模式（默认）：更新容器，保留数据"
            echo "  ./deploy-quality-remote.sh -u        # 升级模式：更新容器，保留数据"
            echo "  ./deploy-quality-remote.sh -r        # 恢复模式：完全重装，清理数据库"
            echo "  ./deploy-quality-remote.sh -l        # 仅传输镜像，不部署"
            echo "  ./deploy-quality-remote.sh -h        # 显示帮助信息"
            echo ""
            echo "说明:"
            echo "  此脚本通过 SSH 在远程服务器 (${REMOTE_HOST}) 上部署 Event Receiver"
            echo "  镜像从 Docker 仓库拉取，无需传输 tar 文件"
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
echo -e "${BLUE}  Event Receiver 远程部署${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${BLUE}远程服务器:${NC} ${GREEN}${REMOTE_USER}@${REMOTE_HOST}${NC}"
echo -e "${BLUE}远程目录:${NC}   ${GREEN}${REMOTE_DIR}${NC}"
echo ""

if [ "$TRANSFER_ONLY" = true ]; then
    echo -e "${YELLOW}模式: ${GREEN}仅传输镜像${NC}"
elif [ "$MODE" = "recover" ]; then
    echo -e "${YELLOW}模式: ${RED}恢复模式 (完全重装，清理数据库)${NC}"
    echo -e "${RED}警告: 此操作将删除远程服务器上的所有容器和数据库数据！${NC}"
    echo -ne "确认继续? [y/N] "
    read -r confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "操作已取消"
        exit 0
    fi
else
    echo -e "${YELLOW}模式: ${GREEN}升级模式 (更新容器，保留数据)${NC}"
fi
echo ""

# ============================================================
# SSH 辅助函数
# ============================================================
ssh_exec() {
    ssh "${REMOTE_USER}@${REMOTE_HOST}" "$@"
}

# ============================================================
# 检查远程服务器连接
# ============================================================
echo -e "${YELLOW}[1/1] 检查远程服务器连接...${NC}"
if ssh_exec "echo '连接成功'" >/dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} 远程服务器连接成功"
else
    echo -e "  ${RED}✗${NC} 无法连接到远程服务器 ${REMOTE_HOST}"
    echo -e "  ${YELLOW}请确保:${NC}"
    echo -e "    1. 远程服务器可达"
    echo -e "    2. SSH 密钥已配置"
    echo -e "    3. 可以执行: ssh ${REMOTE_USER}@${REMOTE_HOST}"
    exit 1
fi
echo ""

# ============================================================
# 传输镜像 tar 文件 (如果使用本地传输)
# ============================================================
if [ "$TRANSFER_ONLY" = true ]; then
    echo -e "${YELLOW}传输镜像到远程服务器...${NC}"
    echo ""

    IMAGE_DIR="${PROJECT_ROOT}/images"

    if [ -f "${IMAGE_DIR}/quality-server.tar.gz" ]; then
        echo -e "  传输后端镜像..."
        scp "${IMAGE_DIR}/quality-server.tar.gz" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/images/"
        echo -e "  ${GREEN}✓${NC} 后端镜像传输完成"
    fi

    if [ -f "${IMAGE_DIR}/quality-frontend.tar.gz" ]; then
        echo -e "  传输前端镜像..."
        scp "${IMAGE_DIR}/quality-frontend.tar.gz" "${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/images/"
        echo -e "  ${GREEN}✓${NC} 前端镜像传输完成"
    fi

    echo ""
    echo -e "${GREEN}镜像传输完成！${NC}"
    echo -e "${YELLOW}在远程服务器上加载镜像:${NC}"
    echo -e "  cd ${REMOTE_DIR}/images"
    echo -e "  docker load < quality-server.tar.gz"
    echo -e "  docker load < quality-frontend.tar.gz"
    echo ""
    exit 0
fi

# ============================================================
# 在远程服务器上执行部署脚本
# ============================================================
echo -e "${YELLOW}在远程服务器上执行部署...${NC}"
echo ""

# 创建远程部署脚本
REMOTE_SCRIPT="/tmp/deploy-quality-remote-$$$.sh"

cat > "/tmp/${REMOTE_SCRIPT##*/}" << 'EOFSCRIPT'
#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
NETWORK_NAME="quality-network"
MYSQL_ROOT_PASSWORD="root123456"
MYSQL_DATABASE="github_hub"
MYSQL_PORT=3306
QUALITY_SERVER_PORT=5001
FRONTEND_PORT=8081

REGISTRY="acr.aishu.cn"
REPOSITORY="dual-mass-engine-gateway"
BACKEND_IMAGE="${REGISTRY}/${REPOSITORY}/quality-server:latest"
FRONTEND_IMAGE="${REGISTRY}/${REPOSITORY}/quality-frontend:latest"

MODE="$1"
PROJECT_ROOT="/root/dual-mass-engine-gateway"

CONTAINERS=(
    "quality-frontend"
    "quality-server"
    "quality-mysql"
)

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  远程服务器部署${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 拉取最新镜像
echo -e "${YELLOW}[步骤 1/4] 拉取最新镜像...${NC}"
echo ""

echo -e "  拉取后端镜像..."
docker pull "$BACKEND_IMAGE" >/dev/null 2>&1
echo -e "  ${GREEN}✓${NC} 后端镜像已更新"

echo -e "  拉取前端镜像..."
docker pull "$FRONTEND_IMAGE" >/dev/null 2>&1
echo -e "  ${GREEN}✓${NC} 前端镜像已更新"
echo ""

# 恢复模式：完全清理
if [ "$MODE" = "recover" ]; then
    echo -e "${YELLOW}[恢复模式] 清理现有容器和数据...${NC}"
    echo ""

    for container in "${CONTAINERS[@]}"; do
        if docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
            echo -e "  删除容器 ${YELLOW}${container}${NC}..."
            docker stop "$container" >/dev/null 2>&1 || true
            docker rm "$container" >/dev/null 2>&1
            echo -e "    ${GREEN}✓${NC} 已删除"
        fi
    done

    # 清理数据库数据
    echo -e "  清理数据库数据..."
    if [ -d "${PROJECT_ROOT}/data/quality-mysql" ]; then
        BACKUP_DIR="${PROJECT_ROOT}/data/quality-mysql.backup.$(date +%Y%m%d_%H%M%S)"
        mv "${PROJECT_ROOT}/data/quality-mysql" "$BACKUP_DIR"
        echo -e "    ${GREEN}✓${NC} 数据已备份到: $BACKUP_DIR"
    fi

    mkdir -p "${PROJECT_ROOT}/data/quality-mysql"
    echo -e "    ${GREEN}✓${NC} 数据库目录已重置"
    echo ""

    # 重建网络
    if docker network inspect "$NETWORK_NAME" &>/dev/null; then
        docker network rm "$NETWORK_NAME" >/dev/null 2>&1
        echo -e "  ${GREEN}✓${NC} Docker 网络已删除"
    fi
fi

# 创建 Docker 网络
echo -e "${YELLOW}[步骤 2/4] 创建 Docker 网络...${NC}"

if ! docker network inspect "$NETWORK_NAME" &> /dev/null; then
    docker network create "$NETWORK_NAME"
    echo -e "  ${GREEN}✓${NC} 创建网络: $NETWORK_NAME"
else
    echo -e "  ${GREEN}✓${NC} 网络 $NETWORK_NAME 已存在"
fi
echo ""

# 创建数据目录
echo -e "${YELLOW}[步骤 3/4] 创建数据目录...${NC}"

mkdir -p "${PROJECT_ROOT}/data/quality-mysql"
mkdir -p "${PROJECT_ROOT}/data/quality-server"
echo -e "  ${GREEN}✓${NC} 数据目录准备完成"
echo ""

# 启动容器
echo -e "${YELLOW}[步骤 4/4] 启动容器...${NC}"

# MySQL 容器
echo -e "  启动 ${YELLOW}quality-mysql${NC}..."
if docker ps --format '{{.Names}}' | grep -q "^quality-mysql$"; then
    if [ "$MODE" = "upgrade" ]; then
        echo -e "    ${YELLOW}!${NC} 升级模式：重新创建容器"
        docker stop quality-mysql &> /dev/null || true
        docker rm quality-mysql &> /dev/null || true
        docker run -d \
            --name quality-mysql \
            --network $NETWORK_NAME \
            -p ${MYSQL_PORT}:3306 \
            -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
            -e MYSQL_DATABASE=$MYSQL_DATABASE \
            -v "${PROJECT_ROOT}/data/quality-mysql:/var/lib/mysql" \
            --restart unless-stopped \
            mysql:latest \
            --character-set-server=utf8mb4 \
            --collation-server=utf8mb4_unicode_ci >/dev/null

        # 等待 MySQL 就绪
        echo -ne "    等待 MySQL 就绪"
        for i in {1..30}; do
            if docker exec quality-mysql mysqladmin ping -h localhost -uroot -p"$MYSQL_ROOT_PASSWORD" &> /dev/null; then
                echo -e "\r    ${GREEN}✓${NC} MySQL 已就绪"
                break
            fi
            sleep 1
            echo -ne "."
        done

        # 检查并初始化数据库
        echo -ne "    检查数据库表..."
        TABLE_COUNT=$(docker exec quality-mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$MYSQL_DATABASE';" 2>/dev/null | tail -1)
        if [ "$TABLE_COUNT" = "0" ]; then
            echo -e "\r    ${YELLOW}!${NC} 数据库表不存在，正在初始化..."
            # 创建初始化 SQL
            docker exec quality-mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" << 'EOSQL' 2>/dev/null
CREATE TABLE IF NOT EXISTS github_events (
    id INT AUTO_INCREMENT PRIMARY KEY,
    event_id VARCHAR(36) NOT NULL UNIQUE,
    event_type VARCHAR(50) NOT NULL,
    event_status VARCHAR(50) NOT NULL,
    repository VARCHAR(255) NOT NULL,
    branch VARCHAR(255) NOT NULL,
    target_branch VARCHAR(255),
    commit_sha VARCHAR(255),
    pr_number INT,
    action VARCHAR(50),
    pusher VARCHAR(255),
    author VARCHAR(255),
    payload JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    INDEX idx_event_id (event_id),
    INDEX idx_event_type (event_type),
    INDEX idx_event_status (event_status),
    INDEX idx_repository (repository)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS pr_quality_checks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    github_event_id VARCHAR(36) NOT NULL,
    check_type VARCHAR(50) NOT NULL,
    check_status VARCHAR(50) NOT NULL,
    stage VARCHAR(50) NOT NULL,
    stage_order INT NOT NULL,
    check_order INT NOT NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    duration_seconds DOUBLE,
    error_message TEXT,
    output TEXT,
    extra TEXT,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_github_event_id (github_event_id),
    INDEX idx_check_type (check_type),
    INDEX idx_check_status (check_status),
    INDEX idx_stage (stage),
    FOREIGN KEY (github_event_id) REFERENCES github_events(event_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
EOSQL
            echo -e "\r    ${GREEN}✓${NC} 数据库表初始化完成"
        else
            echo -e "\r    ${GREEN}✓${NC} 数据库表已存在"
        fi
    else
        echo -e "    ${YELLOW}!${NC} 已在运行"
    fi
else
    docker rm -f quality-mysql &> /dev/null || true
    docker run -d \
        --name quality-mysql \
        --network $NETWORK_NAME \
        -p ${MYSQL_PORT}:3306 \
        -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD \
        -e MYSQL_DATABASE=$MYSQL_DATABASE \
        -v "${PROJECT_ROOT}/data/quality-mysql:/var/lib/mysql" \
        --restart unless-stopped \
        mysql:latest \
        --character-set-server=utf8mb4 \
        --collation-server=utf8mb4_unicode_ci >/dev/null

    echo -ne "    等待 MySQL 就绪"
    for i in {1..30}; do
        if docker exec quality-mysql mysqladmin ping -h localhost -uroot -p"$MYSQL_ROOT_PASSWORD" &> /dev/null; then
            echo -e "\r    ${GREEN}✓${NC} MySQL 已就绪"
            break
        fi
        sleep 1
        echo -ne "."
    done

    # 检查并初始化数据库
    echo -ne "    检查数据库表..."
    TABLE_COUNT=$(docker exec quality-mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='$MYSQL_DATABASE';" 2>/dev/null | tail -1)
    if [ "$TABLE_COUNT" = "0" ]; then
        echo -e "\r    ${YELLOW}!${NC} 数据库表不存在，正在初始化..."
        # 创建初始化 SQL
        docker exec quality-mysql mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" << 'EOSQL' 2>/dev/null
CREATE TABLE IF NOT EXISTS github_events (
    id INT AUTO_INCREMENT PRIMARY KEY,
    event_id VARCHAR(36) NOT NULL UNIQUE,
    event_type VARCHAR(50) NOT NULL,
    event_status VARCHAR(50) NOT NULL,
    repository VARCHAR(255) NOT NULL,
    branch VARCHAR(255) NOT NULL,
    target_branch VARCHAR(255),
    commit_sha VARCHAR(255),
    pr_number INT,
    action VARCHAR(50),
    pusher VARCHAR(255),
    author VARCHAR(255),
    payload JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    INDEX idx_event_id (event_id),
    INDEX idx_event_type (event_type),
    INDEX idx_event_status (event_status),
    INDEX idx_repository (repository)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS pr_quality_checks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    github_event_id VARCHAR(36) NOT NULL,
    check_type VARCHAR(50) NOT NULL,
    check_status VARCHAR(50) NOT NULL,
    stage VARCHAR(50) NOT NULL,
    stage_order INT NOT NULL,
    check_order INT NOT NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    duration_seconds DOUBLE,
    error_message TEXT,
    output TEXT,
    extra TEXT,
    retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_github_event_id (github_event_id),
    INDEX idx_check_type (check_type),
    INDEX idx_check_status (check_status),
    INDEX idx_stage (stage),
    FOREIGN KEY (github_event_id) REFERENCES github_events(event_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
EOSQL
        echo -e "\r    ${GREEN}✓${NC} 数据库表初始化完成"
    else
        echo -e "\r    ${GREEN}✓${NC} 数据库表已存在"
    fi
fi

# Quality Server 容器
echo -e "  启动 ${YELLOW}quality-server${NC}..."
if docker ps --format '{{.Names}}' | grep -q "^quality-server$"; then
    if [ "$MODE" = "upgrade" ]; then
        echo -e "    ${YELLOW}!${NC} 升级模式：重新创建容器"
        docker stop quality-server &> /dev/null || true
        docker rm quality-server &> /dev/null || true
        docker run -d \
            --name quality-server \
            --network $NETWORK_NAME \
            -p ${QUALITY_SERVER_PORT}:5001 \
            -v "${PROJECT_ROOT}/data/quality-server:/data" \
            --restart unless-stopped \
            "$BACKEND_IMAGE" \
            /app/quality-server \
            -addr :5001 \
            -db "root:${MYSQL_ROOT_PASSWORD}@tcp(quality-mysql:3306)/${MYSQL_DATABASE}?parseTime=true" \
            -log-level info >/dev/null
        echo -e "    ${GREEN}✓${NC} quality-server 已启动"
    else
        echo -e "    ${YELLOW}!${NC} 已在运行"
    fi
else
    docker rm -f quality-server &> /dev/null || true
    docker run -d \
        --name quality-server \
        --network $NETWORK_NAME \
        -p ${QUALITY_SERVER_PORT}:5001 \
        -v "${PROJECT_ROOT}/data/quality-server:/data" \
        --restart unless-stopped \
        "$BACKEND_IMAGE" \
        /app/quality-server \
        -addr :5001 \
        -db "root:${MYSQL_ROOT_PASSWORD}@tcp(quality-mysql:3306)/${MYSQL_DATABASE}?parseTime=true" \
        -log-level info >/dev/null
    echo -e "    ${GREEN}✓${NC} quality-server 已启动"
fi

# Frontend 容器
echo -e "  启动 ${YELLOW}quality-frontend${NC}..."
if docker ps --format '{{.Names}}' | grep -q "^quality-frontend$"; then
    if [ "$MODE" = "upgrade" ]; then
        echo -e "    ${YELLOW}!${NC} 升级模式：重新创建容器"
        docker stop quality-frontend &> /dev/null || true
        docker rm quality-frontend &> /dev/null || true
        docker run -d \
            --name quality-frontend \
            --network $NETWORK_NAME \
            -p ${FRONTEND_PORT}:80 \
            --restart unless-stopped \
            "$FRONTEND_IMAGE" >/dev/null
        echo -e "    ${GREEN}✓${NC} quality-frontend 已启动"
    else
        echo -e "    ${YELLOW}!${NC} 已在运行"
    fi
else
    docker rm -f quality-frontend &> /dev/null || true
    docker run -d \
        --name quality-frontend \
        --network $NETWORK_NAME \
        -p ${FRONTEND_PORT}:80 \
        --restart unless-stopped \
        "$FRONTEND_IMAGE" >/dev/null
    echo -e "    ${GREEN}✓${NC} quality-frontend 已启动"
fi

echo ""
echo -e "  ${GREEN}✓${NC} 所有容器启动完成"
echo ""

# 显示状态
echo -e "${YELLOW}[服务状态]${NC}"
echo ""
docker ps --filter "network=${NETWORK_NAME}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  远程部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}访问地址:${NC}"
echo -e "  Frontend:       ${GREEN}http://${REMOTE_HOST}:${FRONTEND_PORT}${NC}"
echo -e "  Quality API:    ${GREEN}http://${REMOTE_HOST}:${QUALITY_SERVER_PORT}${NC}"
echo -e "  MySQL:          ${GREEN}localhost:${MYSQL_PORT}${NC}"
echo ""

EOFSCRIPT

# 上传脚本到远程服务器
scp "/tmp/${REMOTE_SCRIPT##*/}" "${REMOTE_USER}@${REMOTE_HOST}:/tmp/"
ssh_exec "chmod +x /tmp/${REMOTE_SCRIPT##*/}"

# 执行远程脚本
ssh_exec "/tmp/${REMOTE_SCRIPT##*/}" "$MODE"

# 清理临时文件
rm -f "/tmp/${REMOTE_SCRIPT##*/}"
ssh_exec "rm -f /tmp/${REMOTE_SCRIPT##*/}"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  远程部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}访问地址:${NC}"
echo -e "  Frontend:    ${GREEN}http://${REMOTE_HOST}:${FRONTEND_PORT}${NC}"
echo -e "  Backend API: ${GREEN}http://${REMOTE_HOST}:${QUALITY_SERVER_PORT}${NC}"
echo ""
