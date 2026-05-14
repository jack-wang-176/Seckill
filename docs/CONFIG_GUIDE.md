# 环境配置指南

本项目使用模块化的配置管理方式，每个配置模块都有独立的文件。

## 配置文件结构

### pkg/config/ 目录

配置系统以模块化的方式组织在 `pkg/config/` 目录下：

| 文件 | 描述 | 主要配置 |
|------|------|--------|
| `mysql.go` | MySQL 数据库配置 | `MYSQL_DSN` |
| `redis.go` | Redis 缓存配置 | `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` |
| `rabbitmq.go` | RabbitMQ 消息队列配置 | `RABBITMQ_URL`, `RABBITMQ_QUEUES` |
| `etcd.go` | etcd 注册中心配置 | `ETCD_ENDPOINTS` |
| `env.go` | 环境变量工具函数 | 通用辅助函数 |
| `loader.go` | 统一配置加载入口 | 加载所有配置 |
| `container.go` | 依赖注入提供者 | dig 容器配置 |

## 环境变量配置

### 快速开始

```bash
# 1. 复制示例文件
cp env.example .env

# 2. 编辑 .env 文件，修改配置（如需要）
# 大多数情况下，默认配置可直接使用本地开发环境
nano .env  # 或使用你喜欢的编辑器

# 3. 启动应用时，应用会自动从 .env 读取配置
# （需要使用 godotenv 等工具加载 .env 文件）
```

### 环境变量列表

#### MySQL 配置
```env
# MySQL 数据源名称
MYSQL_DSN=root:root@tcp(127.0.0.1:13306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local
```

#### Redis 配置
```env
# Redis 服务器地址
REDIS_ADDR=127.0.0.1:6379

# Redis 密码（可选）
REDIS_PASSWORD=

# Redis 数据库编号 (0-15)
REDIS_DB=0
```

#### RabbitMQ 配置
```env
# RabbitMQ 连接 URL
RABBITMQ_URL=amqp://guest:guest@127.0.0.1:5672/

# RabbitMQ 队列列表（逗号分隔）
RABBITMQ_QUEUES=order_seckill,product_list,product_single
```

#### etcd 配置
```env
# etcd 服务端点（逗号分隔，支持多个端点）
ETCD_ENDPOINTS=127.0.0.1:2379
```

#### 服务端口配置
```env
API_GATEWAY_PORT=8081
USER_RPC_PORT=8889
PRODUCT_RPC_PORT=8890
ORDER_RPC_PORT=8891
```

## 配置加载优先级

1. **环境变量** - 系统环境变量中的值（优先级最高）
2. **.env 文件** - 项目根目录的 `.env` 文件（需要通过 godotenv 或类似工具加载）
3. **默认值** - 各配置模块中定义的默认值（优先级最低）

## 在代码中使用配置

### 方法 1: 直接加载（兼容旧代码）

```go
import "full_backend_practice/pkg/config"

// 加载所有配置
mysql, redis, rabbitmq, etcd := config.LoadConfigFromEnv()

// 使用配置
dsn := mysql.DSN
addr := redis.Addr
```

### 方法 2: 使用结构体（推荐）

```go
import "full_backend_practice/pkg/config"

// 加载所有配置到结构体
allConfig := config.LoadAllConfig()

// 使用配置
mysqlDsn := allConfig.MySQL.DSN
redisAddr := allConfig.Redis.Addr
```

### 方法 3: 在 dig 容器中使用

```go
import (
    "full_backend_practice/pkg/config"
    "go.uber.org/dig"
)

c := dig.New()

// 提供配置（兼容旧代码）
config.ProvideConfigs(c)

// 或使用新方式
config.RegisterConfigProviders(c)

// 现在可以在函数中注入配置
c.Provide(func(cfg config.MySQLConfig) {
    // 使用 cfg
})
```

## .env 文件说明

- **`.env`** - 本地环境配置文件（不提交到版本控制）
- **`env.example`** - 环境变量配置模板（提交到版本控制）

### .env 和 env.example 的区别

- `env.example` 是模板，包含所有可配置的环境变量和默认值
- `.env` 是你本地的实际配置，包含敏感信息（数据库密码等），不应该提交到版本控制
- `.gitignore` 已配置忽略 `.env` 文件

## 添加新的配置模块

如果需要添加新的配置模块（例如 `es.go` for Elasticsearch），按以下步骤：

1. **创建新配置文件** `pkg/config/es.go`:
```go
package config

type ElasticsearchConfig struct {
    Host string
    Port int
}

const DefaultESHost = "127.0.0.1"
const DefaultESPort = 9200

func LoadElasticsearchConfig() ElasticsearchConfig {
    return ElasticsearchConfig{
        Host: getEnvOrDefault("ES_HOST", DefaultESHost),
        Port: toInt(getEnvOrDefault("ES_PORT", strconv.Itoa(DefaultESPort))),
    }
}
```

2. **在 `loader.go` 中添加加载逻辑**:
```go
type AllConfig struct {
    // ... 其他配置 ...
    Elasticsearch ElasticsearchConfig
}

func LoadAllConfig() AllConfig {
    return AllConfig{
        // ... 其他配置 ...
        Elasticsearch: LoadElasticsearchConfig(),
    }
}
```

3. **在 `container.go` 中注册提供者**:
```go
c.Provide(func() ElasticsearchConfig { return allCfg.Elasticsearch })
```

4. **在 `env.example` 和 `.env` 中添加新的环境变量**:
```env
# Elasticsearch 配置
ES_HOST=127.0.0.1
ES_PORT=9200
```

## 常见问题

### Q: 为什么要分离配置到不同的文件？
A: 这样做有以下好处：
- 模块化清晰，易于维护
- 每个配置模块责任单一
- 便于添加新的配置模块
- 代码更易于扩展

### Q: 如何在本地开发中使用环境变量？
A: 有几种方式：
1. 使用 `.env` 文件配合 godotenv 库自动加载
2. 直接在系统环境中设置变量
3. 在 IDE 中配置环境变量

### Q: 默认值是什么？
A: 每个配置模块中都定义了默认值常量，可以查看相应的文件了解。
