# 秒杀系统压测工具

这是一个基于 Go 原生 Goroutine 实现的秒杀系统压测工具，支持完整的业务流模拟、并发控制和性能指标收集。

## 功能特性

- **完整业务流模拟**：按照真实用户秒杀流程（获取路径 → 提交订单 → 查询结果）
- **并发控制**：支持可配置的并发数，自动管理连接池
- **梯度增压**：支持缓慢升高并发压力，避免瞬间冲击
- **用户池管理**：自动生成和管理测试用户令牌
- **性能指标**：收集关键路径的延迟数据，计算 P50/P90/P99 百分位数
- **实时监控**：压测过程中实时显示 QPS、成功率等关键指标
- **详细报告**：压测结束后生成完整性能报告

## 编译和运行

### 编译

```bash
# 从项目根目录编译
cd /Users/wangyuxin/work/go/full_backend_practice
go build -o test/load/seckill_benchmark ./test/load

# 或者直接运行（自动编译）
go run ./test/load/seckill_benchmark.go [flags]
```

### 基础运行

```bash
# 使用默认参数运行（100 并发，30 秒，本地环境）
go run ./test/load/seckill_benchmark.go

# 或者使用编译后的二进制
./test/load/seckill_benchmark
```

### 参数说明

| 参数 | 默认值 | 说明 | 示例 |
|------|------|------|------|
| `-url` | `http://localhost:8081` | API Gateway 地址 | `-url http://api.example.com:8081` |
| `-concurrency` | `100` | 并发数（Goroutine 数量） | `-concurrency 500` |
| `-duration` | `30s` | 压测持续时间 | `-duration 60s` |
| `-product` | `1` | 商品 ID | `-product 2` |
| `-users` | `500` | 测试用户总数 | `-users 1000` |
| `-rampup` | `5s` | 梯度增压时间 | `-rampup 10s` |

## 使用场景

### 场景1：本地开发环境测试（低压）

```bash
go run ./test/load/seckill_benchmark.go \
   -url http://localhost:8081 \
  -concurrency 50 \
  -duration 20s \
  -users 200
```

### 场景2：预发布环境测试（中压）

```bash
go run ./test/load/seckill_benchmark.go \
  -url http://staging-api.example.com \
  -concurrency 500 \
  -duration 60s \
  -users 1000 \
  -rampup 10s
```

### 场景3：生产环境压力测试（高压）

```bash
go run ./test/load/seckill_benchmark.go \
  -url http://api.example.com \
  -concurrency 2000 \
  -duration 120s \
  -users 5000 \
  -rampup 30s
```

## 压测工作流

1. **用户生成阶段**
   - 自动注册测试用户（username: testuser_0, testuser_1, ...）
   - 通过登录获取 JWT 令牌
   - 构建用户令牌池

2. **梯度增压阶段**
   - 逐步启动 Goroutine worker
   - 均匀分布启动时间，避免瞬间冲击

3. **并发执行阶段**
   - 每个 worker 循环执行秒杀流程：
     - POST `/api/v1/seckill/path` → 获取秒杀路径
     - POST `/api/v1/seckill/order/{path}` → 提交订单
     - GET `/api/v1/seckill/result` → 查询结果（约 1/3 的 worker）
   - 每个请求的延迟被记录到对应的延迟数组

4. **实时监控阶段**
   - 每 5 秒打印一次当前进度（请求数、成功率、QPS）

5. **报告生成阶段**
   - 压测结束后打印完整的性能报告
   - 展示关键指标：QPS、成功率、P50/P90/P99 延迟

## 输出示例

```
开始生成 500 个测试用户...
已生成 50 个用户令牌
已生成 100 个用户令牌
...
共生成 500 个有效用户令牌

开始压测：并发 100 | 持续 30s | 商品 1

已发送请求: 1500 | 成功: 1450 | 失败: 50 | QPS: 500
已发送请求: 3000 | 成功: 2900 | 失败: 100 | QPS: 500
已发送请求: 4500 | 成功: 4350 | 失败: 150 | QPS: 500

========================================
                  秒杀压测报告                    
========================================

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

========================================
```

## 性能调优建议

### 如果 QPS 偏低

1. **增加并发数**
   ```bash
   -concurrency 1000
   ```

2. **减少梯度增压时间**
   ```bash
   -rampup 1s
   ```

3. **检查网络延迟**
   - 在靠近服务器的机器上运行压测
   - 减少网络跳转次数

### 如果失败率高

1. **减少并发数**
   ```bash
   -concurrency 200
   ```

2. **增加用户池大小**
   ```bash
   -users 2000
   ```
   - 避免同一用户重复下单导致业务逻辑失败

3. **检查服务端日志**
   - 查看是否有 500 错误
   - 检查数据库连接池是否充足

4. **增加超时时间**
   - 修改代码中的 `Timeout: 10 * time.Second` 为更大的值

## 高级用法

### 多次测试对比

```bash
# 测试1：基线
go run ./test/load/seckill_benchmark.go -concurrency 100 -duration 30s > baseline.txt

# 测试2：优化后
go run ./test/load/seckill_benchmark.go -concurrency 100 -duration 30s > optimized.txt

# 对比两个报告
diff baseline.txt optimized.txt
```

### 长时间稳定性测试

```bash
# 运行 10 分钟，观察系统是否存在内存泄漏
go run ./test/load/seckill_benchmark.go \
  -concurrency 200 \
  -duration 600s \
  -users 2000
```

### 梯度压力测试

```bash
# 模拟用户逐步增加的场景
for concurrency in 100 200 500 1000; do
  echo "Testing with concurrency: $concurrency"
  go run ./test/load/seckill_benchmark.go \
    -concurrency $concurrency \
    -duration 60s \
    -rampup 10s
  sleep 5
done
```

## 故障排除

### 错误：无法生成测试用户

**原因**：API Gateway 不可达或用户服务异常

**解决**：
1. 检查 API Gateway 是否启动
   ```bash
   curl http://localhost:8081/api/v1/user/register
   ```

2. 检查用户服务是否可用
   ```bash
   docker-compose logs user-service
   ```

### 错误：所有请求都失败

**原因**：商品 ID 不存在或秒杀配置异常

**解决**：
1. 检查商品是否存在
   ```bash
   curl http://localhost:8081/api/v1/product/list
   ```

2. 确认秒杀时间窗口有效
   ```bash
   # 修改 -product 参数为有效的商品 ID
   -product 1
   ```

### 内存占用过高

**原因**：并发数过高，单次延迟采样过多

**解决**：
1. 减少并发数
2. 优化内存采样间隔
3. 定期清理压测工具中的历史数据

## 扩展功能

如需扩展功能，可修改以下部分：

- **自定义请求头**：修改 `doRequest` 方法中的 header 设置
- **复杂业务流**：在 `Run` 方法中添加额外的请求步骤
- **动态参数**：支持从 CSV 或数据库读取参数
- **结果导出**：将指标导出为 JSON 或 CSV 格式

## 注意事项

1. **不要在生产环境直接运行**
   - 压测会产生大量测试数据
   - 建议在独立的测试环境运行

2. **用户隔离**
   - 测试用户使用 `testuser_` 前缀，便于事后清理
   - 可定期清理测试数据

3. **并发限制**
   - 本工具默认采用系统连接数限制
   - 如需超高并发，需调整 OS 文件描述符限制
   ```bash
   ulimit -n 65536
   ```

4. **测试环境准备**
   - 确保商品表中存在相应商品
   - 确保秒杀时间窗口有效（start_time ≤ now ≤ end_time）
   - 预准备足够的库存

## 相关命令

### 快速启动压测环境

```bash
# 1. 启动 Docker 容器
docker-compose up -d

# 2. 等待服务启动
sleep 10

# 3. 运行压测
go run ./test/load/seckill_benchmark.go -concurrency 100 -duration 30s

# 4. 停止服务
docker-compose down
```

### 查看实时日志

```bash
# 在另一个终端查看 API Gateway 日志
docker-compose logs -f api-gateway
```

## 性能基准参考

基于标准配置（100 并发，30 秒，8GB 内存机器）的预期性能：

| 指标 | 预期值 |
|------|--------|
| QPS | 500-2000 |
| 成功率 | > 95% |
| P50 延迟 | < 100ms |
| P99 延迟 | < 300ms |
| 平均 GC 暂停 | < 50ms |

实际性能取决于：
- 网络延迟
- 数据库性能
- Redis 性能
- 服务器资源配置
