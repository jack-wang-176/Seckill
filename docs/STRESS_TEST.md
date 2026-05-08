# 秒杀系统 - 压测工具使用指南

本项目包含完整的秒杀系统实现，以及基于原生 Go Goroutine 的专业压测工具。

## 📁 项目结构

```
.
├── backend/                    # 微服务源代码
│   ├── cmd/
│   │   ├── api/main.go        # API Gateway 入口
│   │   ├── order/main.go      # 订单服务入口
│   │   ├── product/main.go    # 商品服务入口
│   │   └── user/main.go       # 用户服务入口
│   ├── internal/              # 内部实现
│   │   ├── api/               # API Gateway 实现
│   │   ├── order/             # 订单服务实现
│   │   ├── product/           # 商品服务实现
│   │   └── user/              # 用户服务实现
│   └── idl/                   # Thrift IDL 定义
├── test/load/                 # 压测工具
│   ├── seckill_benchmark.go  # 压测脚本源代码
│   ├── bin/                   # 编译输出目录
│   │   └── seckill_benchmark  # 可执行二进制
│   └── README.md              # 详细使用文档
├── docs/                       # 📚 文档
│   ├── QUICK_START.md         # 快速参考卡片
│   ├── STRESS_TEST.md         # 压测指南
│   └── IMPLEMENTATION_SUMMARY.md  # 实现总结
├── config/                     # ⚙️ 配置文件
│   ├── docker-compose.yml     # 容器编排
│   └── kitex_info.yaml        # Kitex 配置
├── script/
│   ├── bootstrap.sh           # 服务启动脚本
│   └── run-benchmark.sh       # 压测快速启动脚本
├── Makefile                   # 构建工具
├── go.mod                     # Go 依赖管理
├── go.sum                     # Go 依赖校验
├── .gitignore                 # Git 忽略
└── README.md                  # 项目主说明
```

## 🚀 快速开始

### 方式一：使用快速启动脚本（推荐）

```bash
# 帮助信息
./script/run-benchmark.sh help

# 轻量级压测（本地开发）
./script/run-benchmark.sh light

# 指定网关端口（避免默认 8081）
./script/run-benchmark.sh light -port 8081

# 指定网关地址并跳过自动启动容器
./script/run-benchmark.sh custom -url http://localhost:8081 -no-auto-start

# 中等压测（预发布环境）
./script/run-benchmark.sh medium

# 重负载压测（生产环境）
./script/run-benchmark.sh heavy

# 自定义参数
./script/run-benchmark.sh custom -concurrency 1000 -duration 60s -users 2000
```

### 方式二：使用 Makefile

```bash
# 编译压测工具
make build-benchmark

# 运行默认压测（100 并发，30 秒）
make benchmark

# 轻量级压测
make benchmark-light

# 中等压测
make benchmark-medium

# 重负载压测
make benchmark-heavy

# 清理编译文件
make clean-benchmark
```

### 方式三：直接运行二进制

```bash
# 编译
go build -o test/load/bin/seckill_benchmark ./test/load

# 运行
./test/load/bin/seckill_benchmark -concurrency 100 -duration 30s

# 使用自定义参数
./test/load/bin/seckill_benchmark \
  -url http://api.example.com:8081 \
  -concurrency 500 \
  -duration 60s \
  -product 1 \
  -users 1000 \
  -rampup 10s
```

## 📊 完整的压测工作流

### 第一步：启动服务

```bash
# 启动所有微服务容器
docker-compose -f config/docker-compose.yml up -d

# 验证服务是否启动
curl http://localhost:8081/api/v1/product/list
```

### 第二步：运行压测

```bash
# 选择合适的压测模式
./script/run-benchmark.sh medium

# 或者自定义参数
./script/run-benchmark.sh custom \
  -concurrency 500 \
  -duration 60s \
  -product 1 \
  -users 1000 \
  -rampup 10s
```

### 第三步：查看结果

压测工具会输出实时进度和最终报告：

```
===========================================
                  秒杀压测报告                    
===========================================

总时长：30.0 秒
总请求数：4500
成功请求：4350
失败请求：150
成功率：96.67%
QPS：150 请求/秒

--- 秒杀路径延迟 ---
平均: 45 ms | P50: 42 ms | P90: 68 ms | P99: 95 ms

--- 下单延迟 ---
平均: 52 ms | P50: 48 ms | P90: 78 ms | P99: 120 ms

--- 结果查询延迟 ---
平均: 38 ms | P50: 35 ms | P90: 55 ms | P99: 80 ms

===========================================
```

## 🔧 压测参数说明

| 参数 | 默认值 | 说明 | 示例 |
|------|------|------|------|
| `-url` | `http://localhost:8081` | API Gateway 地址 | `-url http://api.example.com:8081` |
| `-port` | `8081` | API Gateway 端口（脚本专用） | `-port 8081` |
| `-concurrency` | `100` | 并发数（Goroutine 数量） | `-concurrency 500` |
| `-duration` | `30s` | 压测持续时间 | `-duration 60s` |
| `-product` | `1` | 商品 ID | `-product 2` |
| `-users` | `500` | 测试用户总数 | `-users 1000` |
| `-rampup` | `5s` | 梯度增压时间 | `-rampup 10s` |

> 说明：`-port` 和 `-no-auto-start` 是 `script/run-benchmark.sh` 的脚本级参数，不会传递给压测二进制。

## 📈 压测场景模板

### 场景 1：本地开发验证（5 分钟）

```bash
./script/run-benchmark.sh custom \
  -url http://localhost:8081 \
  -concurrency 50 \
  -duration 60s \
  -users 200 \
  -rampup 2s
```

**预期指标**：QPS > 100, 成功率 > 95%

### 场景 2：预发布环境验证（10 分钟）

```bash
./script/run-benchmark.sh medium
```

**预期指标**：QPS > 1000, 成功率 > 99%

### 场景 3：生产环境压力测试（30 分钟）

```bash
./script/run-benchmark.sh heavy
```

**预期指标**：QPS > 5000, 成功率 > 98%

### 场景 4：长时间稳定性测试（1 小时）

```bash
./script/run-benchmark.sh custom \
  -concurrency 200 \
  -duration 3600s \
  -users 2000 \
  -rampup 30s
```

**监测内容**：
- 内存泄漏（应保持平稳）
- 连接泄漏（应保持平稳）
- 延迟漂移（不应显著增加）

## 🎯 关键性能指标（KPI）

### 响应时间

| 分位数 | 本地环境 | 预发布环境 | 生产环境 |
|-------|--------|----------|--------|
| P50 | < 50ms | < 100ms | < 200ms |
| P90 | < 100ms | < 200ms | < 500ms |
| P99 | < 200ms | < 500ms | < 1000ms |

### 系统容量

| 指标 | 目标值 |
|------|--------|
| 最大 QPS | > 5000 |
| 成功率 | > 98% |
| 平均延迟 | < 100ms |
| 内存占用 | < 500MB |

## 🐛 故障排除

### 问题 1：无法连接到 API Gateway

```bash
# 检查服务是否运行
docker-compose ps

# 检查 API Gateway 日志
docker-compose logs api-gateway

# 尝试手动请求
curl http://localhost:8081/api/v1/product/list
```

**解决方案**：
1. 启动服务：`docker-compose up -d`
2. 等待服务启动：`sleep 10`
3. 重新运行压测

### 问题 2：用户生成失败

```bash
# 检查用户服务日志
docker-compose logs user-service

# 手动注册用户
curl -X POST http://localhost:8081/api/v1/user/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "password123"}'
```

**解决方案**：
1. 确保用户服务正常运行
2. 检查数据库连接
3. 增加 `-users` 参数

### 问题 3：所有请求都失败

```bash
# 检查商品是否存在
curl http://localhost:8081/api/v1/product/list

# 检查秒杀是否在时间窗口内
# 查看 product 表中 start_time 和 end_time
```

**解决方案**：
1. 创建有效的商品
2. 确保秒杀时间窗口包含当前时间
3. 使用正确的商品 ID：`-product 1`

### 问题 4：性能指标异常

| 症状 | 原因 | 解决方案 |
|------|------|--------|
| QPS 很低 | 网络延迟或服务器性能不足 | 靠近服务器运行，减少并发 |
| 失败率高 | 并发过高或用户不足 | 增加 `-users` 参数，减少 `-concurrency` |
| 内存占用高 | 采样数据过多 | 减少并发数或压测时间 |
| 延迟急剧上升 | 服务器过载或垃圾回收 | 减少并发，检查服务器日志 |

## 📚 相关命令速查

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看实时日志
docker-compose logs -f api-gateway

# 停止所有服务
docker-compose down

# 重启特定服务
docker-compose restart api-gateway

# 编译所有微服务
make build-api make build-order make build-product make build-user

# 编译压测工具
make build-benchmark

# 清理编译文件
make clean-benchmark

# 检查代码质量
go vet ./...

# 格式化代码
go fmt ./...

# 运行单元测试
go test ./...
```

## 📖 进阶使用

### 梯度加压测试

模拟用户逐步增加的情景：

```bash
for i in 100 200 500 1000 2000; do
  echo "========== 压测并发数: $i =========="
  ./test/load/bin/seckill_benchmark \
    -concurrency $i \
    -duration 30s \
    -users $((i * 2)) \
    -rampup 10s
  sleep 10
done
```

### 长时间稳定性测试

```bash
# 运行 1 小时，监控内存泄漏
./test/load/bin/seckill_benchmark \
  -concurrency 200 \
  -duration 3600s \
  -users 2000
```

### 多商品测试

```bash
# 分别测试不同商品
for product_id in 1 2 3 4 5; do
  ./test/load/bin/seckill_benchmark \
    -product $product_id \
    -concurrency 100 \
    -duration 30s \
    -users 500
done
```

### 远程环境测试

```bash
# 连接到远程服务器进行压测
./test/load/bin/seckill_benchmark \
  -url http://staging.example.com:8081 \
  -concurrency 500 \
  -duration 60s \
  -users 1000
```

## 💡 最佳实践

1. **始终从轻量级开始**
   - 先用 `light` 模式验证基础功能
   - 确保没有异常后再提升压力

2. **观察实时指标**
   - 监控 CPU、内存使用率
   - 关注服务日志中的错误信息

3. **记录每次测试结果**
   - 建立性能基准
   - 对比优化前后的差异

4. **压测后清理数据**
   - 删除测试用户（名称以 `testuser_` 开头）
   - 清理测试订单数据

5. **避免在生产环境直接运行**
   - 使用独立的测试环境
   - 或在低流量时段运行

## 📞 支持

如遇到问题，请检查：

1. [test/load/README.md](test/load/README.md) - 详细的压测工具文档
2. 项目日志 - `docker-compose logs`
3. 服务健康状态 - `/api/v1/product/list`

## 📄 许可

本项目遵循相关开源许可协议。
