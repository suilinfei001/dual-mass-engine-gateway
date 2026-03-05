#!/bin/bash
# 后端镜像构建脚本 - 使用本地 go mod cache

set -e

echo "========================================"
echo "  Event Processor 后端镜像构建"
echo "========================================"
echo ""

# 检查本地 go mod cache
if [ ! -d "/root/go/pkg/mod" ]; then
    echo "错误: 本地 go mod cache 不存在"
    exit 1
fi

echo "使用本地 go mod cache: /root/go/pkg/mod"
echo ""

# 使用 BuildKit 的 mount 功能挂载本地缓存
export DOCKER_BUILDKIT=1

docker build \
    --build-arg BUILDKIT_INLINE_CACHE=1 \
    --secret id=gocache,src=/root/go/pkg/mod \
    -f Dockerfile_server \
    -t event-processor-backend:latest \
    .

echo ""
echo "✓ 后端镜像构建完成"
