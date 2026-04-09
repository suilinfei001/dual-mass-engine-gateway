#!/bin/bash
# shared/scripts/test.sh
# 共享库测试脚本

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHARED_DIR="$(dirname "$SCRIPT_DIR")"

cd "$SHARED_DIR"

echo "Running tests..."

# 运行所有测试
go test -v -race -coverprofile=coverage.out ./pkg/...

# 显示覆盖率
echo ""
echo "Coverage report:"
go tool cover -func=coverage.out | tail -1

echo "✓ Tests complete"
