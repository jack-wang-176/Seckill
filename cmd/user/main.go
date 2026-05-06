package main

import (
	"net"

	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/internal/user"
	"full_backend_practice/kitex_gen/user/userservice"
	"full_backend_practice/pkg/config"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

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
	provideConfigs(c)

	c.Provide(logger.InitLogger)
	c.Provide(database.InitMYSQL)
	c.Provide(database.InitRedis)
	c.Provide(mq.InitRabbitMQ)

	c.Provide(user.NewUserServiceImpl)
	c.Provide(user.NewConsumer)
	return c
}

func main() {
	c := buildContainer()
	err := c.Invoke(func(impl user.UserServiceImpl,
		consumer user.UserConsumer,
		etcdCfg *config.EtcdConfig,
		log *zap.Logger,
	) error {
		log.Info("Starting Application...")
		consumer.StartRegisterConsumer()
		log.Info("Order Consumer Started")

		r, err := etcd.NewEtcdRegistry(etcdCfg.Endpoints)
		if err != nil {
			return err
		}

		addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:8889")

		svr := userservice.NewServer(
			impl,
			server.WithServiceAddr(addr),
			server.WithRegistry(r),
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "user_service"}),
		)
		log.Info("User Service RPC Server is running on 0.0.0.0:8889")
		return svr.Run()
	})
	if err != nil {
		panic(err)
	}
}
