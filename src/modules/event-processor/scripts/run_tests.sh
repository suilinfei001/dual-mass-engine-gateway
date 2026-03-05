#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

VERBOSE=false
COVERAGE=false
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"

print_banner() {
    echo -e "${BLUE}"
    echo "========================================"
    echo "  Event Processor Test Suite"
    echo "========================================"
    echo -e "${NC}"
}

print_usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -v, --verbose     显示详细测试输出"
    echo "  -c, --coverage    生成覆盖率报告"
    echo "  -h, --help        显示帮助信息"
    echo ""
    echo "Examples:"
    echo "  $0                运行所有测试"
    echo "  $0 -v             运行所有测试（详细模式）"
    echo "  $0 -c             运行测试并生成覆盖率报告"
    echo "  $0 -v -c          详细模式 + 覆盖率报告"
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -c|--coverage)
                COVERAGE=true
                shift
                ;;
            -h|--help)
                print_usage
                exit 0
                ;;
            *)
                echo -e "${RED}未知选项: $1${NC}"
                print_usage
                exit 1
                ;;
        esac
    done
}

run_tests() {
    echo -e "${YELLOW}[1/3] 检查测试环境...${NC}"
    
    cd "$PROJECT_ROOT"
    
    if ! command -v go &> /dev/null; then
        echo -e "${RED}错误: Go 未安装${NC}"
        exit 1
    fi
    
    echo -e "  Go 版本: $(go version)"
    echo ""
    
    echo -e "${YELLOW}[2/3] 编译检查...${NC}"
    if ! go build ./...; then
        echo -e "${RED}✗ 编译失败${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ 编译成功${NC}"
    echo ""
    
    echo -e "${YELLOW}[3/3] 运行测试...${NC}"
    echo ""
    
    local test_args=""
    
    if [ "$VERBOSE" = true ]; then
        test_args="-v"
    fi
    
    if [ "$COVERAGE" = true ]; then
        test_args="$test_args -coverprofile=$COVERAGE_FILE -covermode=atomic"
    fi
    
    local start_time=$(date +%s)
    
    if go test ./... $test_args -timeout 5m; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        echo ""
        echo -e "${GREEN}========================================"
        echo -e "  所有测试通过!"
        echo -e "  耗时: ${duration}s"
        echo -e "========================================${NC}"
        
        if [ "$COVERAGE" = true ]; then
            echo ""
            echo -e "${BLUE}覆盖率报告:${NC}"
            go tool cover -func=$COVERAGE_FILE | tail -1
            echo ""
            echo -e "生成 HTML 覆盖率报告: ${COVERAGE_HTML}"
            go tool cover -html=$COVERAGE_FILE -o $COVERAGE_HTML
        fi
        
        echo ""
        echo -e "${BLUE}测试统计:${NC}"
        local total_tests=$(go test ./... -v 2>&1 | grep -c "^=== RUN" || echo "0")
        local passed_tests=$(go test ./... -v 2>&1 | grep -c "^--- PASS" || echo "0")
        echo -e "  总测试数: ${total_tests}"
        echo -e "  通过: ${GREEN}${passed_tests}${NC}"
        
        exit 0
    else
        echo ""
        echo -e "${RED}========================================"
        echo -e "  测试失败!"
        echo -e "========================================${NC}"
        exit 1
    fi
}

main() {
    parse_args "$@"
    print_banner
    run_tests
}

main "$@"
