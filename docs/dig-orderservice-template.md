# orderservice 的 Dig 构造函数模版

这份补充文件只演示如何按你当前的 `orderservice` 结构来写 `dig` 的依赖关系，不改现有业务代码。

## 1. 当前 orderservice 的依赖边界

从你现有代码看，`orderservice` 的入口在 `backend/cmd/order/main.go`，核心实现是：

- `backend/internal/app/order-service/handler.go`：`OrderServiceImpl.Seckill`
- `backend/internal/app/order-service/consumer.go`：MQ 消费者
- `pkg/database`：MySQL、Redis、订单和商品表
- `pkg/mq`：RabbitMQ 连接
- `pkg/logger`：Zap 日志

因此，`dig` 最适合注入的对象是这些长生命周期、可复用的组件：

- `*zap.Logger`
- `*gorm.DB`
- Redis 客户端或封装函数
- RabbitMQ 连接/Channel 封装
- `OrderRepo`
- `OrderService`
- `OrderServiceImpl`

请求上下文、单次 RPC 请求参数、事务里的短生命周期 `tx`，不适合做全局注入。

## 2. 推荐的构造函数分层

### 基础设施层

```go
func NewLogger() (*zap.Logger, error)
func NewDB(dsn string) (*gorm.DB, error)
func NewRedis(addr, password string, db int) (*redis.Client, error)
func NewMQ(url string, queues []string) (*mq.Client, error)
func NewEtcdRegistry(endpoints []string) (*etcd.Registry, error)
```

### 业务数据层

```go
type OrderRepo struct {
    DB *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo
```

### 业务服务层

```go
type OrderService struct {
    repo   *OrderRepo
    logger *zap.Logger
    mq     *mq.Client
    redis  *redis.Client
}

func NewOrderService(repo *OrderRepo, logger *zap.Logger, mq *mq.Client, redis *redis.Client) *OrderService
```

### Kitex 实现层

```go
type OrderServiceImpl struct {
    svc *OrderService
}

func NewOrderServiceImpl(svc *OrderService) *OrderServiceImpl
```

如果你暂时不拆 `OrderServiceImpl`，也可以先让它保持现状，只把基础设施和业务服务注入进来，再逐步拆分。

## 3. 一份可直接套用的 Dig 模版

```go
package main

import (
    "net"

    order_service "full_backend_practice/backend/internal/app/order-service"
    "full_backend_practice/kitex_gen/order/orderservice"
    "full_backend_practice/pkg/database"
    "full_backend_practice/pkg/logger"
    "full_backend_practice/pkg/mq"

    "github.com/cloudwego/kitex/pkg/rpcinfo"
    "github.com/cloudwego/kitex/server"
    "github.com/uber-go/dig"
    etcd "github.com/kitex-contrib/registry-etcd"
)

func buildContainer() *dig.Container {
    c := dig.New()

    c.Provide(func() string {
        return "root:root@tcp(127.0.0.1:3306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local"
    })
    c.Provide(func() []string {
        return []string{"127.0.0.1:2379"}
    })

    c.Provide(func() *logger.Logger {
        return logger.Log
    })
    c.Provide(database.InitMYSQL)
    c.Provide(database.InitRedis)
    c.Provide(mq.InitRabbitMQ)

    c.Provide(order_service.NewOrderRepo)
    c.Provide(order_service.NewOrderService)
    c.Provide(order_service.NewOrderServiceImpl)

    return c
}

func main() {
    c := buildContainer()

    _ = c.Invoke(func(
        impl *order_service.OrderServiceImpl,
        etcdEndpoints []string,
    ) error {
        r, err := etcd.NewEtcdRegistry(etcdEndpoints)
        if err != nil {
            return err
        }

        addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:8888")
        svr := orderservice.NewServer(
            impl,
            server.WithServiceAddr(addr),
            server.WithRegistry(r),
            server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "order-service"}),
        )

        return svr.Run()
    })
}
```

这份模版的关键点是：

- `Provide` 只负责注册构造函数
- `Invoke` 才负责拿最终对象并启动 Kitex
- `main.go` 不再直接手写一串初始化顺序，而是交给容器解析依赖图

## 4. 更贴近你当前项目的过渡方案

如果你暂时不想重构所有包，可以先只把这几项抽成构造函数：

```go
func NewOrderRepo() *OrderRepo
func NewOrderService() *OrderService
func NewOrderServiceImpl() *OrderServiceImpl
```

然后保留你原先的全局初始化，只把 `main.go` 改成让 `dig` 负责组装对象：

```go
c := dig.New()
c.Provide(func() *zap.Logger { return logger.Log })
c.Provide(func() *gorm.DB { return database.DB })
c.Provide(func() *mq.Client { return mq.Client })
c.Provide(order_service.NewOrderServiceImpl)

c.Invoke(func(impl *order_service.OrderServiceImpl) {
    // 这里只负责起服务
})
```

这是“先注入、后拆全局变量”的过渡方案，风险最低。

## 5. `orderservice` 里最容易写错的点

- `dig` 不能自动猜你想注入的是值还是指针，构造函数签名要统一。
- `OrderServiceImpl` 如果还是空结构体，那 `dig` 注入意义不大，最好让它持有业务服务对象。
- `consumer.go` 这种后台协程不建议自己创建一堆隐式依赖，最好在启动时把需要的 `DB`、`MQ`、`Logger` 明确注入进来。
- `main.go` 里不要既用 `dig` 又手动初始化同一份对象，否则会出现双实例问题。

## 6. 你现在这套结构，最合理的注入顺序

```text
logger / db / redis / mq / etcd
        ↓
     order repo
        ↓
    order service
        ↓
  order service impl
        ↓
   kitex server start
```

## 7. 结论

你这个 `orderservice` 最适合的 Dig 思路不是把所有函数都塞进容器，而是：

- 先把基础设施注册成单例
- 再把 repo/service/impl 一层层构造成有显式依赖的对象
- 最后在 `Invoke` 里只做启动，不做业务初始化

如果你下一步要，我可以继续按这个模板，给你补一版 `api gateway` 侧的 Dig 依赖图。