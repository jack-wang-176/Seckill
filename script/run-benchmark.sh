#!/bin/bash

# 秒杀系统压测快速启动脚本
# 用法: ./script/run-benchmark.sh [light|medium|heavy|custom] [参数]

set -e

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
PROJECT_ROOT=$(cd "$SCRIPT_DIR/.." && pwd)
BENCHMARK_BINARY="$PROJECT_ROOT/test/load/bin/seckill_benchmark"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

# 检查服务是否运行
check_service() {
    local url=$1
    local max_retries=30
    local retry=0

    while [ $retry -lt $max_retries ]; do
        if curl -s "$url/api/v1/product/list" > /dev/null 2>&1; then
            print_success "服务已启动: $url"
            return 0
        fi
        retry=$((retry + 1))
        echo -ne "\r等待服务启动... ($retry/$max_retries)"
        sleep 1
    done

    echo ""
    print_error "服务无响应: $url"
    return 1
}

# 编译压测工具
build_benchmark() {
    print_info "编译压测工具..."
    mkdir -p "$PROJECT_ROOT/test/load/bin"
    
    if cd "$PROJECT_ROOT" && go build -o "$BENCHMARK_BINARY" ./test/load; then
        print_success "压测工具编译完成"
        return 0
    else
        print_error "压测工具编译失败"
        return 1
    fi
}

# 显示使用帮助
show_help() {
    cat <<EOF
${GREEN}秒杀系统压测快速启动脚本${NC}

使用方法:
  $0 [模式] [选项]

可用模式:
  ${BLUE}light${NC}     - 轻量级压测 (50 并发, 20 秒, 200 用户)
  ${BLUE}medium${NC}    - 中等压测   (500 并发, 60 秒, 1000 用户)
  ${BLUE}heavy${NC}     - 重负载压测 (2000 并发, 120 秒, 5000 用户)
  ${BLUE}custom${NC}    - 自定义参数 (需要提供所有参数)
  ${BLUE}help${NC}      - 显示本帮助信息

自定义参数:
  -url          API Gateway 地址 (默认: http://localhost:8080)
  -concurrency  并发数 (默认: 100)
  -duration     持续时间 (默认: 30s)
  -product      商品 ID (默认: 1)
  -users        测试用户数 (默认: 500)
  -rampup       梯度增压时间 (默认: 5s)

示例:
  # 运行轻量级压测
  $0 light

  # 运行中等压测
  $0 medium

  # 自定义参数运行
  $0 custom -concurrency 1000 -duration 60s -users 2000

  # 指定目标 URL
  $0 light -url http://staging.example.com:8080
EOF
}

# 运行压测
run_benchmark() {
    local url="$1"
    shift
    local args=("$@")

    # 检查服务
    if ! check_service "$url"; then
        print_warning "请确保服务已启动:"
        echo "  docker-compose up -d"
        return 1
    fi

    echo ""
    print_info "启动压测..."
    echo "参数:"
    echo "  URL: $url"
    echo "  其他: ${args[@]}"
    echo ""

    if "$BENCHMARK_BINARY" -url "$url" "${args[@]}"; then
        print_success "压测完成"
        return 0
    else
        print_error "压测执行出错"
        return 1
    fi
}

# 检查并启动 Docker 服务
auto_start_services() {
    print_info "检查 Docker 服务状态..."

    # 检查 Docker 是否运行
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker 未运行，请先启动 Docker"
        return 1
    fi

    # 检查容器是否运行
    if ! docker-compose -f "$PROJECT_ROOT/config/docker-compose.yml" ps | grep -q "Up"; then
        print_warning "服务容器未运行，即将启动..."
        if cd "$PROJECT_ROOT" && docker-compose -f config/docker-compose.yml up -d; then
            print_success "服务容器已启动"
            sleep 5
            return 0
        else
            print_error "启动服务容器失败"
            return 1
        fi
    else
        print_success "服务容器已运行"
        return 0
    fi
}

# 主函数
main() {
    local mode="${1:-help}"
    shift || true

    print_info "秒杀系统压测工具 v1.0"
    echo ""

    # 处理帮助命令
    if [ "$mode" = "help" ] || [ "$mode" = "-h" ] || [ "$mode" = "--help" ]; then
        show_help
        return 0
    fi

    # 编译压测工具
    if ! build_benchmark; then
        return 1
    fi

    echo ""

    # 自动启动服务（可选）
    if [ "$mode" != "custom" ] || [[ " $@ " =~ " -no-auto-start " ]]; then
        if ! auto_start_services; then
            print_warning "将使用已运行的服务继续压测..."
        fi
    fi

    echo ""

    # 根据模式执行
    case "$mode" in
        light)
            run_benchmark "http://localhost:8080" -concurrency 50 -duration 20s -users 200 "$@"
            ;;
        medium)
            run_benchmark "http://localhost:8080" -concurrency 500 -duration 60s -users 1000 -rampup 10s "$@"
            ;;
        heavy)
            run_benchmark "http://localhost:8080" -concurrency 2000 -duration 120s -users 5000 -rampup 30s "$@"
            ;;
        custom)
            # 自定义模式，直接传递参数
            run_benchmark "http://localhost:8080" "$@"
            ;;
        *)
            print_error "未知的模式: $mode"
            echo ""
            show_help
            return 1
            ;;
    esac
}

# 执行主函数
main "$@"
