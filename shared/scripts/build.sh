#!/bin/bash
# shared/scripts/build.sh
# 共享库构建脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHARED_DIR="$(dirname "$SCRIPT_DIR")"

echo "Building shared library..."

cd "$SHARED_DIR"

# 格式化代码
echo "Formatting code..."
go fmt ./pkg/...

# 检查
echo "Running go vet..."
go vet ./pkg/...

# 构建
echo "Building..."
go build ./pkg/...

echo "✓ Build complete"
