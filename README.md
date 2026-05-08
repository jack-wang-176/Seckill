# 秒杀系统 - 微服务实现

基于 Go + CloudWeGo (Hertz/Kitex) 的秒杀系统，采用微服务架构 + Redis 原子库存 + RabbitMQ 异步下单，支撑高并发秒杀场景。

---

## 架构概览

```
用户 → API Gateway (Hertz, 8081)
        ├── /api/v1/user/*    → User Service (Kitex RPC, 8889)
        ├── /api/v1/product/* → Product Service (Kitex RPC, 8890)
        └── /api/v1/seckill/* → Order Service (Kitex RPC, 8891)
                                 ├── Redis (库存原子扣减 / 动态路径防刷)
                                 └── RabbitMQ (异步下单) → MySQL
```

## 秒杀业务流程

```
┌────────────────────────────────────────────┐
│  Step 1: 获取秒杀路径                       │
│  POST /api/v1/seckill/path                 │
│  需 Authorization: Bearer <token>           │
│  基于 userId + productId + 盐值生成MD5      │
│  存入Redis，用于后续校验，防止直接下单       │
│  返回: {"path":"a1b2c3d4..."}              │
└───────────────────┬────────────────────────┘
                    │
┌───────────────────▼────────────────────────┐
│  Step 2: 提交秒杀订单                       │
│  POST /api/v1/seckill/order/:path          │
│  流程:                                      │
│  ① 校验路径 → Redis GET 对比是否一致        │
│  ② 扣库存 → Redis Lua 脚本原子扣减          │
│  ③ 发消息 → RabbitMQ 异步推送订单           │
│  ④ 消费者落库 → MySQL 事务更新库存+创建订单 │
│  ⑤ 标记结果 → Redis SET 用户秒杀成功状态    │
│  返回: {"message":"seckill success"}        │
└───────────────────┬────────────────────────┘
                    │
┌───────────────────▼────────────────────────┐
│  Step 3: 查询秒杀结果                       │
│  POST /api/v1/seckill/result               │
│  轮询 Redis 查看秒杀是否成功                │
└────────────────────────────────────────────┘
```

### 秒杀核心设计

| 设计 | 目的 |
|------|------|
| **动态路径** | 每次请求生成 MD5 路径，无法提前或直接下单 |
| **Redis Lua 原子扣减** | 保证高并发下库存不超卖 |
| **RabbitMQ 异步落库** | 秒杀接口只操作 Redis，订单异步写入 MySQL |
| **结果标记** | Redis 记录用户抢购结果，支持轮询查询 |

---

## 快速启动

```bash
# 1. 启动基础设施 (MySQL, Redis, RabbitMQ, Etcd)
docker compose -f config/docker-compose.yml up -d

# 2. 启动微服务 (4个终端分别执行)
make run-user      # 用户服务 (8889)
make run-product   # 商品服务 (8890)
make run-order     # 订单服务 (8891)
make run-api       # API Gateway (8081)

# 3. 创建商品并预热缓存
curl -X POST http://localhost:8081/api/v1/product/create \
  -H "Content-Type: application/json" \
  -d '{"name":"秒杀商品","price":99.99,"seckill_prict":49.99,"stock":1000,"version":1,"start_time":1746662400,"end_time":1746835199}'
curl -X POST http://localhost:8081/api/v1/product/heat

# 4. 运行压测
./script/run-benchmark.sh light
```

---

## 压测结果

### 环境
- MacBook Air (Apple Silicon) · light 模式 (50 并发, 20 秒, 200 用户)

```
========================================
                  秒杀压测报告
========================================
总时长：20.0 秒
总请求数：177,610
成功请求：177,610
失败请求：0
成功率：100.00%
QPS：8,879 请求/秒
--- 秒杀路径延迟 ---
平均: 1 ms | P50: 2 ms | P90: 2 ms | P99: 4 ms
--- 下单延迟 ---
平均: 4 ms | P50: 4 ms | P90: 5 ms | P99: 8 ms
========================================
```

### 更多压测模式

```bash
./script/run-benchmark.sh medium         # 500并发 60秒 1000用户
./script/run-benchmark.sh heavy          # 2000并发 120秒 5000用户
./script/run-benchmark.sh custom \
  -concurrency 100 -duration 30s -users 500   # 自定义参数
```

---

## 项目结构

```
├── cmd/              # 服务入口 (api/user/product/order)
├── internal/         # 核心实现
│   ├── api/          # Gateway: handler / middleware / router
│   ├── user/         # 用户: 注册登录 / JWT 签发
│   ├── product/      # 商品: CRUD / 缓存预热 / RabbitMQ 更新
│   ├── order/        # 秒杀: 路径生成 / 库存扣减 / 异步消费
│   └── rpc/          # Kitex RPC 客户端 (Etcd 服务发现)
├── infrastructure/   # 基础设施: MySQL / Redis / RabbitMQ / Logger
├── pkg/              # 共享: 响应封装 / JWT 工具
├── kitex_gen/        # Thrift 自动生成代码
├── idl/              # Thrift IDL 定义
├── test/load/        # 压测工具 (Go 实现)
├── config/           # Docker Compose
└── script/           # 启动脚本
```

## API 参考

### 用户 (无需认证)

```bash
POST /api/v1/user/register  {"username":"...","password":"..."}
POST /api/v1/user/login     {"username":"...","password":"..."}
```

### 商品 (无需认证)

```bash
GET  /api/v1/product/list                            # 商品列表
GET  /api/v1/product/:id                             # 商品详情
POST /api/v1/product/create  {"name":"...","price":99,"stock":1000,...}
POST /api/v1/product/heat                            # 预热缓存到 Redis
```

### 秒杀 (需 Authorization: Bearer <token>)

```bash
POST /api/v1/seckill/path         {"product_id":"1"}       # 获取路径
POST /api/v1/seckill/order/:path  {"product_id":"1"}       # 提交订单
POST /api/v1/seckill/result       {"product_id":"1"}       # 查询结果
```

---

## 技术栈

| 类别 | 技术 |
|------|------|
| 框架 | Hertz (HTTP), Kitex (RPC) |
| 序列化 | Thrift |
| 存储 | MySQL (GORM), Redis |
| 消息 | RabbitMQ |
| 注册中心 | Etcd |
| 认证 | JWT |
| 依赖注入 | dig |
| 日志 | zap |


