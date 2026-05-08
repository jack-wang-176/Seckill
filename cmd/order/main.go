package main

import (
	"net"

	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/infrastructure/mq"
	order "full_backend_practice/internal/order"

	"full_backend_practice/kitex_gen/order/orderservice"
	"full_backend_practice/pkg/config"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

// 注入配置文件
func provideConfigs(c *dig.Container) {
	c.Provide(func() *config.MySQLConfig {
		return &config.MySQLConfig{DSN: "root:root@tcp(127.0.0.1:13306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local"}
	})
	c.Provide(func() *config.RedisConfig {
		return &config.RedisConfig{Addr: "127.0.0.1:6379", Password: "", DB: 0}
	})
	c.Provide(func() *config.RabbitMQConfig {
		return &config.RabbitMQConfig{URL: "amqp://guest:guest@127.0.0.1:5672/", Queues: []string{"order_seckill"}}
	})
	c.Provide(func() *config.EtcdConfig {
		return &config.EtcdConfig{Endpoints: []string{"127.0.0.1:2379"}}
	})
}

package main

import (
	"os" // 新增
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/internal/api/handler"
	"full_backend_practice/internal/api/router"
	"full_backend_practice/internal/rpc"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

// ... buildContainer 保持不变 ...
func buildContainer() *dig.Container {
	c := dig.New()
	c.Provide(logger.InitLogger)
	c.Provide(rpc.InitOrderRpc)
	c.Provide(rpc.InitUserRpc)
	c.Provide(rpc.InitProductRpc)
	c.Provide(handler.NewUserHandler)
	c.Provide(handler.NewOrderHandler)
	c.Provide(handler.NewProductServiceHandler)
	return c
}

func main() {
	container := buildContainer()
	err := container.Invoke(func(
		userH *handler.UserHandler,
		orderH *handler.OrderHandler,
		productH *handler.ProductServiceHandler,
		log *zap.Logger,
	) {
		// 动态获取端口，默认为 8080
		port := os.Getenv("API_GATEWAY_PORT")
		if port == "" {
			port = "8080"
		}
		addr := "0.0.0.0:" + port

		h := server.Default(server.WithHostPorts(addr))
		log.Info("Starting API Gateway on " + addr + "...")
		router.Register(h, userH, orderH, productH)
		h.Spin()
	})
	if err != nil {
		panic(err)
	}
}