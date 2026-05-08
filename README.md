# 秒杀系统 - 完整微服务实现

一个基于 Go 语言的完整秒杀系统实现，包含微服务架构、数据库设计、缓存管理。



##  项目概览

### 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                   Client (Web/Mobile)                       │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│              API Gateway (Port 8081)                        │
│         - 请求路由                                          │
│         - 认证 (JWT)                                        │
│         - 用户上下文管理                                    │
└──────┬──────────┬──────────────┬───────────────────────────┘
       │          │              │
┌──────▼───┐ ┌───▼──────┐ ┌─────▼──────┐ ┌──────────┐
│   User   │ │ Product  │ │   Order    │ │  Redis  │
│ Service  │ │ Service  │ │  Service   │ │ (Cache) │
│ (Port    │ │ (Port    │ │ (Port      │ │ (Port   │
│ 8889)    │ │ 8890)    │ │ 8891)      │ │ 6379)   │
└──────────┘ └──────────┘ └────────────┘ └────────┬─┘
       │          │              │          │
       └──────────┴──────────────┴──────────┘
              (Kitex RPC)
       │
┌──────▼────────────────────────────┐
│      MySQL (Port 3306)            │
│  - 用户表                         │
│  - 商品表                         │
│  - 订单表                         │
└───────────────────────────────────┘
```

### 核心模块

| 模块 | 描述 | 端口 | 技术栈 |
|------|------|------|-------|
| **API Gateway** | HTTP 入口层，请求路由和认证 | 8081 | Hertz + JWT |
| **User Service** | 用户管理（注册、登录、认证） | 8888 | Kitex RPC + BCrypt |
| **Product Service** | 商品管理（列表、详情、创建） | 8889 | Kitex RPC + MySQL |
| **Order Service** | 秒杀流程（路径生成、库存、结果） | 8890 | Kitex RPC + Redis |
| **Infrastructure** | 缓存、数据库、消息队列 | - | Redis + MySQL + RabbitMQ |

##  快速开始

### 前置要求

- Go 1.18+
- Docker & Docker Compose
- 8GB+ 内存
- 空闲端口：8081, 8888, 8889, 8890, 3306, 6379

### 3 步启动

####  启动基础设施

```bash
# 启动 Docker 容器（MySQL、Redis、RabbitMQ）
docker-compose -f config/docker-compose.yml up -d

# 验证服务状态
docker-compose -f config/docker-compose.yml ps
```

####  启动微服务

```bash
# 编译所有服务
make build-api && make build-order && make build-product && make build-user

# 或者在不同终端分别运行
make run-api
make run-order
make run-product
make run-user
```

####  运行压测

```bash
# 轻量级压测
./script/run-benchmark.sh light

# 或使用 Makefile
make benchmark-light
```

##  项目结构

```
.
├── backend/                      # 微服务核心代码
│   ├── cmd/                      # 服务入口
│   │   ├── api/main.go           # API Gateway
│   │   ├── order/main.go         # Order Service
│   │   ├── product/main.go       # Product Service
│   │   └── user/main.go          # User Service
│   ├── internal/                 # 内部实现
│   │   ├── api/                  # Gateway 实现
│   │   │   ├── handler/          # HTTP 处理器
│   │   │   ├── middleware/       # 中间件（认证）
│   │   │   └── router/           # 路由配置
│   │   ├── order/                # Order 实现
│   │   ├── product/              # Product 实现
│   │   ├── user/                 # User 实现
│   │   └── rpc/                  # RPC 客户端
│   ├── idl/                      # Thrift IDL 定义
│   │   ├── base.thrift
│   │   ├── order.thrift
│   │   ├── product.thrift
│   │   └── user.thrift
│   └── pkg/                      # 共享包
│       ├── database/             # 数据库驱动
│       ├── mq/                   # 消息队列
│       └── ...
│
├── test/load/                    # 压测工具
│   ├── seckill_benchmark.go     # 压测源代码
│   ├── bin/                      # 编译输出
│   │   └── seckill_benchmark    # 可执行文件
│   └── README.md                 # 详细文档
│
├── docs/                         #  文档
│   ├── QUICK_START.md            # 快速参考（2 分钟）
│   ├── STRESS_TEST.md            # 压测指南（10 分钟）
│   └── IMPLEMENTATION_SUMMARY.md # 实现总结（15 分钟）
│
├── config/                       #  配置
│   ├── docker-compose.yml        # 容器编排
│   └── kitex_info.yaml           # Kitex 配置
│
├── script/                       # 脚本
│   ├── bootstrap.sh              # 服务启动脚本
│   └── run-benchmark.sh          # 压测启动脚本
│
├── Makefile                      # 构建工具
├── go.mod                        # 依赖管理
├── go.sum                        # 依赖校验
├── .gitignore                    # Git 忽略
└── README.md                     # 本文件
```

##  核心功能

### 1. 用户系统

```bash
# 注册用户
POST /api/v1/user/register
{
  "username": "user123",
  "password": "password123"
}

# 登录获取令牌
POST /api/v1/user/login
{
  "username": "user123",
  "password": "password123"
}
# 返回: { "token": "eyJhbGc..." }
```

### 2. 商品管理

```bash
# 查询商品列表
GET /api/v1/product/list

# 获取商品详情
GET /api/v1/product/:id

# 创建商品
POST /api/v1/product/create
{
  "name": "商品名称",
  "price": 99.99,
  "stock": 1000,
  "start_time": "2026-05-08T10:00:00Z",
  "end_time": "2026-05-08T11:00:00Z"
}
```

### 3. 秒杀流程

```bash
# 步骤1: 获取秒杀路径
POST /api/v1/seckill/path (需要认证)
{
  "product_id": "1"
}
# 返回: { "path": "a1b2c3d4e5f6..." }

# 步骤2: 提交订单
POST /api/v1/seckill/order/:path (需要认证)
{
  "product_id": "1"
}
# 返回: { "message": "success" }

# 步骤3: 查询秒杀结果
GET /api/v1/seckill/result?product_id=1 (需要认证)
# 返回: { "success": true, "message": "秒杀成功" }
```

##  压测工具

### 快速启动

```bash
# 查看帮助
./script/run-benchmark.sh help

# 轻量级（50 并发，20 秒）
./script/run-benchmark.sh light

# 中等（500 并发，60 秒）
./script/run-benchmark.sh medium

# 重负载（2000 并发，120 秒）
./script/run-benchmark.sh heavy
```

### 性能指标

压测工具会输出详细的性能报告：

```
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

========================================
```

##  构建命令

### 编译

```bash
# 编译所有服务
make build-api build-order make build-product make build-user

# 编译压测工具
make build-benchmark

# 清理编译文件
make clean-benchmark
```

### 运行

```bash
# 启动单个服务
make run-api       # API Gateway
make run-order     # Order Service
make run-product   # Product Service
make run-user      # User Service

# 运行压测
make benchmark             # 默认配置
make benchmark-light       # 轻量级
make benchmark-medium      # 中等
make benchmark-heavy       # 重负载
```

### 其他命令

```bash
# 依赖管理
make tidy

# 查看所有可用命令
make help
```

##  认证机制

系统使用 JWT (JSON Web Token) 进行身份验证。

### 认证流程

1. **注册/登录** → 获取 JWT 令牌
2. **使用令牌** → 在请求头中添加 `Authorization: Bearer <token>`
3. **验证** → 中间件验证令牌的有效性和用户身份

### 示例

```bash
# 获取令牌
TOKEN=$(curl -s -X POST http://localhost:8081/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user123","password":"password123"}' \
  | jq -r '.token')

# 使用令牌
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8081/api/v1/seckill/path \
  -d '{"product_id":"1"}'
```

##  性能优化特性

### 1. 缓存策略

- **Redis 缓存**：存储商品信息和秒杀路径
- **生产者-消费者**：异步更新缓存
- **TTL 管理**：自动过期清理

### 2. 并发控制

- **库存原子性**：使用 Redis 原子操作
- **路径验证**：防止直接下单绕过
- **用户令牌**：防止重复秒杀

### 3. 异步处理

- **RabbitMQ 队列**：异步处理订单
- **非阻塞 I/O**：快速响应客户端

##  故障排除

### 问题 1：服务无法启动

```bash
# 检查端口是否被占用
lsof -i :8081
lsof -i :8888

# 检查 Docker 容器状态
docker-compose -f config/docker-compose.yml ps

# 查看容器日志
docker-compose -f config/docker-compose.yml logs -f mysql
```

### 问题 2：连接超时

```bash
# 检查 Redis 连接
redis-cli -h 127.0.0.1 -p 6379 ping

# 检查 MySQL 连接
mysql -h 127.0.0.1 -u root -p

# 检查网络
curl http://localhost:8081/api/v1/product/list
```

### 问题 3：秒杀失败

```bash
# 查看服务日志
docker-compose -f config/docker-compose.yml logs -f order-service

# 验证商品是否存在
curl http://localhost:8081/api/v1/product/list

# 验证秒杀时间窗口
# 确保 start_time <= now <= end_time
```



##  相关技术栈

### 后端框架
- **Hertz** - 高性能 HTTP 框架（API Gateway）
- **Kitex** - 高性能 RPC 框架（微服务通信）
- **Thrift** - IDL 定义和序列化

### 存储层
- **MySQL** - 关系数据库（持久化存储）
- **Redis** - 内存缓存（性能优化）
- **RabbitMQ** - 消息队列（异步处理）

### 工具库
- **go.uber.org/dig** - 依赖注入
- **go.uber.org/zap** - 结构化日志
- **golang-jwt/jwt** - JWT 认证

### 开发工具
- **Docker & Docker Compose** - 容器化部署
- **Go 1.18+** - 编程语言


