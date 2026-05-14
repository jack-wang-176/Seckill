package main

import (
	"net"
	"os"

	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/internal/product"
	"full_backend_practice/kitex_gen/product/productservice"

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

	// 覆盖队列：商品服务关心 product_list 和 product_single
	mqCfg.Queues = []string{constant.QueueProductList, constant.QueueProductSingle}

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

		port := os.Getenv("PRODUCT_RPC_PORT")
		if port == "" {
			port = constant.DefaultProductRPCPort
		}
		addrStr := "0.0.0.0:" + port
		addr, _ := net.ResolveTCPAddr("tcp", addrStr)

		svr := productservice.NewServer(
			impl,
			server.WithServiceAddr(addr),
			server.WithRegistry(r),
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: constant.ServiceNameProduct}),
		)
		log.Info("Product Service RPC Server is running on " + addrStr)
		return svr.Run()
	})
	if err != nil {
		panic(err)
	}
}
