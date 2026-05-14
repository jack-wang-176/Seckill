package main

import (
	"net"

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

func buildContainer() *dig.Container {
	c := dig.New()

	if err := config.RegisterConfigProviders(c); err != nil {
		panic(err)
	}

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
		serverCfg *config.ServerConfig,
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

		port := serverCfg.ProductRpcPort
		if port == "" {
			port = "8890"
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
