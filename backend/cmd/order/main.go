package main

import (
	"net"

	order_service "full_backend_practice/backend/internal/app/order-service"
	"full_backend_practice/kitex_gen/order/orderservice"
	"full_backend_practice/pkg/config"
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/logger"
	"full_backend_practice/pkg/mq"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

// 注入配置文件
func provideConfigs(c *dig.Container) {
	c.Provide(func() *config.MySQLConfig {
		return &config.MySQLConfig{DSN: "root:root@tcp(127.0.0.1:3306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local"}
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

func buildContainer() *dig.Container {
	c := dig.New()

	// 1. 提供配置
	provideConfigs(c)

	// 2. 提供基础组件 (注意：不加括号，传入函数名！)
	c.Provide(logger.InitLogger)
	c.Provide(database.InitMYSQL)
	c.Provide(database.InitRedis)
	c.Provide(mq.InitRabbitMQ)

	// 3. 提供业务逻辑组件
	c.Provide(order_service.NewOrderServiceImpl)
	c.Provide(order_service.NewOrderConsumer)

	return c
}

func main() {
	c := buildContainer()
	err := c.Invoke(func(impl *order_service.OrderServiceImpl,
		consumer *order_service.OrderConsumer,
		etcdCfg *config.EtcdConfig,
		log *zap.Logger,
	) error {

		log.Info("Starting Application...")

		// 1. 启动消费者协程
		consumer.StartConsumer()
		log.Info("Order Consumer started, listening for messages...")

		// 2. 初始化服务注册
		r, err := etcd.NewEtcdRegistry(etcdCfg.Endpoints)
		if err != nil {
			return err
		}

		addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:8888")

		svr := orderservice.NewServer(
			impl,
			server.WithServiceAddr(addr),
			server.WithRegistry(r),
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
				ServiceName: "order-service",
			}),
		)

		log.Info("Order Service RPC Server is running on 0.0.0.0:8888...")
		return svr.Run()
	})

	if err != nil {
		panic(err)
	}
}
