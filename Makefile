.PHONY: build-api build-order build-product build-user run-api run-order run-product run-user tidy \
         build-benchmark benchmark benchmark-light benchmark-medium benchmark-heavy clean-benchmark

# 微服务构建和运行
build-api:
	go build -o bin/api-server cmd/api/main.go

build-order:
	go build -o bin/order-server cmd/order/main.go

build-product:
	go build -o bin/product-server cmd/product/main.go

build-user:
	go build -o bin/user-server cmd/user/main.go

run-api: build-api
	./bin/api-server

run-order: build-order
	./bin/order-server

run-product: build-product
	./bin/product-server

run-user: build-user
	./bin/user-server

# 压测工具
build-benchmark:
	@mkdir -p test/load/bin
	go build -o test/load/bin/seckill_benchmark ./test/load

benchmark: build-benchmark
	@echo "开始默认配置压测 (100 并发, 30 秒)..."
	./test/load/bin/seckill_benchmark

benchmark-light: build-benchmark
	@echo "开始轻量级压测 (50 并发, 20 秒)..."
	./test/load/bin/seckill_benchmark -concurrency 50 -duration 20s -users 200

benchmark-medium: build-benchmark
	@echo "开始中等压测 (500 并发, 60 秒)..."
	./test/load/bin/seckill_benchmark -concurrency 500 -duration 60s -users 1000 -rampup 10s

benchmark-heavy: build-benchmark
	@echo "开始重负载压测 (2000 并发, 120 秒)..."
	./test/load/bin/seckill_benchmark -concurrency 2000 -duration 120s -users 5000 -rampup 30s

clean-benchmark:
	@rm -f test/load/bin/seckill_benchmark
	@rm -rf test/load/bin
	@echo "清理压测工具完成"

# 通用命令
tidy:
	go mod tidy
