# 压测工具 - 快速参考卡片

## 🚀 3 秒快速开始

```bash
# 一行命令启动
./script/run-benchmark.sh light
```

## 🎯 常用命令

### 使用脚本启动（推荐）

```bash
./script/run-benchmark.sh light      # 轻量级 (50 并发, 20 秒)
./script/run-benchmark.sh medium     # 中等   (500 并发, 60 秒)
./script/run-benchmark.sh heavy      # 重负载 (2000 并发, 120 秒)
./script/run-benchmark.sh help       # 显示帮助
```

### 使用 Makefile 启动

```bash
make build-benchmark         # 编译
make benchmark              # 默认压测 (100 并发)
make benchmark-light        # 轻量级
make benchmark-medium       # 中等
make benchmark-heavy        # 重负载
make clean-benchmark        # 清理
```

### 直接运行（完全自定义）

```bash
./test/load/bin/seckill_benchmark \
  -concurrency 500 \
  -duration 60s \
  -product 1 \
  -users 1000
```

## 📊 参数速查

| 参数 | 默认值 | 示例 |
|------|--------|------|
| `-url` | `http://localhost:8080` | `-url http://api.example.com` |
| `-concurrency` | `100` | `-concurrency 500` |
| `-duration` | `30s` | `-duration 60s` |
| `-product` | `1` | `-product 2` |
| `-users` | `500` | `-users 1000` |
| `-rampup` | `5s` | `-rampup 10s` |

## 📋 预设场景

### Light（本地开发）
- **并发数**：50
- **持续时间**：20 秒
- **测试用户**：200
- **命令**：`./script/run-benchmark.sh light`

### Medium（预发布）
- **并发数**：500
- **持续时间**：60 秒
- **测试用户**：1000
- **增压时间**：10 秒
- **命令**：`./script/run-benchmark.sh medium`

### Heavy（生产压力）
- **并发数**：2000
- **持续时间**：120 秒
- **测试用户**：5000
- **增压时间**：30 秒
- **命令**：`./script/run-benchmark.sh heavy`

## 🔧 启动服务

```bash
# 启动 Docker 容器
docker-compose up -d

# 验证服务
curl http://localhost:8080/api/v1/product/list

# 查看日志
docker-compose logs -f api-gateway
```

## 📈 性能指标解读

```
总请求数        - 发送的 HTTP 请求总数
成功请求        - 返回 200 OK 的请求数
失败请求        - 返回非 200 或超时的请求数
成功率          - (成功请求 / 总请求) × 100%
QPS             - 每秒请求数

平均延迟        - 所有请求的平均响应时间
P50             - 50% 的请求延迟（中位数）
P90             - 90% 的请求延迟
P99             - 99% 的请求延迟（尾部延迟）
```

## ⚡ 快速场景

```bash
# 场景1：本地验证（5 分钟）
./script/run-benchmark.sh light

# 场景2：提交前测试（10 分钟）
./script/run-benchmark.sh medium

# 场景3：发布前压力测试（30 分钟）
./script/run-benchmark.sh heavy

# 场景4：自定义压测
./script/run-benchmark.sh custom \
  -concurrency 1000 \
  -duration 60s \
  -users 2000
```

## 🐛 常见问题

| 问题 | 解决方案 |
|------|---------|
| 无法连接到服务 | `docker-compose up -d` |
| 用户生成失败 | 检查用户服务：`docker-compose logs user-service` |
| 请求全部失败 | 验证商品存在：`curl http://localhost:8080/api/v1/product/list` |
| QPS 很低 | 增加并发数：`-concurrency 1000` |
| 失败率高 | 增加用户数：`-users 2000` |

## 📚 相关文档

- **详细说明**：[test/load/README.md](../test/load/README.md)
- **快速指南**：[STRESS_TEST.md](STRESS_TEST.md)
- **实现总结**：[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)

## 🎓 三种使用方式对比

| 方式 | 命令 | 场景 | 难度 |
|------|------|------|------|
| **脚本** | `./script/run-benchmark.sh light` | 日常使用 | ⭐ |
| **Makefile** | `make benchmark-medium` | 快速启动 | ⭐⭐ |
| **直接运行** | `./test/load/bin/seckill_benchmark -concurrency 500` | 完全自定义 | ⭐⭐⭐ |

## 💡 最佳实践

1. ✅ 从 `light` 开始验证基础功能
2. ✅ 观察 CPU 和内存使用率
3. ✅ 记录每次压测的结果
4. ✅ 对比优化前后的性能差异
5. ✅ 避免在生产环境直接运行

## 📱 快速命令速记

```bash
# 开发环境验证
make benchmark-light

# 预发布测试
make benchmark-medium

# 生产压力测试
make benchmark-heavy

# 自定义 1000 并发, 60 秒
./test/load/bin/seckill_benchmark -concurrency 1000 -duration 60s
```

---

**更多帮助**：
- `./script/run-benchmark.sh help` - 详细使用说明
- 查看 [test/load/README.md](../test/load/README.md) 获取完整文档
