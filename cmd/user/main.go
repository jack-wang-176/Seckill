package main

import (
	"net"
	"os"

	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/internal/user"
	"full_backend_practice/kitex_gen/user/userservice"
	"full_backend_practice/pkg/config"
	"full_backend_practice/pkg/constant"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func provideConfigs(c *dig.Container) {
	mysqlCfg, redisCfg, mqCfg, etcdCfg := config.LoadConfigFromEnv()

	// 覆盖队列：用户服务关心 order_seckill
	mqCfg.Queues = []string{constant.QueueOrderSeckill}

	c.Provide(func() *config.MySQLConfig { return &mysqlCfg })
	c.Provide(func() *config.RedisConfig { return &redisCfg })
	c.Provide(func() *config.RabbitMQConfig { return &mqCfg })
	c.Provide(func() *config.EtcdConfig { return &etcdCfg })
}

func buildContainer() *dig.Container {
	c := dig.New()
	provideConfigs(c)

	c.Provide(logger.InitLogger)
	c.Provide(database.InitMYSQL)
	c.Provide(user.NewUserMysql)
	c.Provide(database.InitRedis)
	c.Provide(database.NewRedisWrapper)
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
		log.Info("Starting Application of user service...")
		consumer.StartRegisterConsumer()
		log.Info("User Consumer Started")

		r, err := etcd.NewEtcdRegistry(etcdCfg.Endpoints)
		if err != nil {
			log.Error("Failed to create etcd registry", zap.Error(err))
			return err
		}

		port := os.Getenv("USER_RPC_PORT")
		if port == "" {
			port = constant.DefaultUserRPCPort
		}
		addrStr := "0.0.0.0:" + port
		addr, _ := net.ResolveTCPAddr("tcp", addrStr)

		svr := userservice.NewServer(
			impl,
			server.WithServiceAddr(addr),
			server.WithRegistry(r),
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: constant.ServiceNameUser}),
		)
		log.Info("User Service RPC Server is running on " + addrStr)
		return svr.Run()
	})
	if err != nil {
		panic(err)
	}
}
