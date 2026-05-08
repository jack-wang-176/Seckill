package main

import (
	"net"

	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/internal/product"
	"full_backend_practice/kitex_gen/product/productservice"

	"full_backend_practice/pkg/config"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func provideConfigs(c *dig.Container) {
	c.Provide(func() *config.MySQLConfig {
		return &config.MySQLConfig{DSN: "root:root@tcp(127.0.0.1:13306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local"}
	})
	c.Provide(func() *config.RedisConfig {
		return &config.RedisConfig{Addr: "127.0.0.1:6379", Password: "", DB: 0}
	})
	c.Provide(func() *config.RabbitMQConfig {
		return &config.RabbitMQConfig{URL: "amqp://guest:guest@127.0.0.1:5672/", Queues: []string{"product_list", "product_single"}}
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
	c.Provide(product.NewProductMysql)
	c.Provide(database.InitRedis)
	c.Provide(database.NewRedisWrapper)
	c.Provide(mq.InitRabbitMQ)

	c.Provide(product.NewProductServiceImpl)
	c.Provide(product.NewProductConsumer)
	return c
}

func main() {
	c := buildContainer()
	err := c.Invoke(func(impl product.ProductServiceImpl,
		consumer product.ProductConsumer,
		etcdCfg *config.EtcdConfig,
		log *zap.Logger,
	) error {
		log.Info("Starting Application of product service...")
		consumer.StartConsumer()
		log.Info("Product Consumer Started")
		r, err := etcd.NewEtcdRegistry(etcdCfg.Endpoints)
		if err != nil {
			log.Error("Failed to create etcd registry", zap.Error(err))
			return err
		}
		addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:8890")
		svr := productservice.NewServer(
			impl,
			server.WithServiceAddr(addr),
			server.WithRegistry(r),
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "product_service"}),
		)
		return svr.Run()
	})
	if err != nil {
		panic(err)
	}
}
