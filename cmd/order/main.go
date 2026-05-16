package main

import (
	"net"

	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/infrastructure/mq"
	order "full_backend_practice/internal/order"

	"full_backend_practice/kitex_gen/order/orderservice"
	"full_backend_practice/pkg/config"
	"full_backend_practice/pkg/constant"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/dig"
	"go.uber.org/zap"
	"full_backend_practice/infrastructure/tacer"
	kitextracing "github.com/kitex-contrib/obs-opentelemetry/tracing"
)

func buildContainer() *dig.Container {
	c := dig.New()

	if err := config.RegisterConfigProviders(c); err != nil {
		panic(err)
	}

	c.Provide(logger.InitLogger)
	c.Provide(database.InitMYSQL)
	c.Provide(order.NewOrderMysql)
	c.Provide(database.InitRedis)
	c.Provide(database.NewRedisWrapper)
	c.Provide(mq.InitRabbitMQ)

	c.Provide(order.NewOrderServiceImpl)
	c.Provide(order.NewOrderConsumer)
	return c
}

func main() {
	c := buildContainer()
	err := c.Invoke(func(impl order.OrderServiceImpl,
		consumer order.OrderConsumer,
		serverCfg *config.ServerConfig,
		etcdCfg *config.EtcdConfig,
		traceCfg *config.TracerConfig,
		log *zap.Logger,
	) error {
		shutdown := tacer.InitTracer(traceCfg, constant.ServiceNameOrder)
		defer shutdown()
		log.Info("Starting Application of order service...")
		consumer.StartConsumer()
		log.Info("Order Consumer Started")

		r, err := etcd.NewEtcdRegistry(etcdCfg.Endpoints)
		if err != nil {
			log.Error("Failed to create etcd registry", zap.Error(err))
			return err
		}

		port := serverCfg.OrderRpcPort
		if port == "" {
			port = "8891"
		}
		addrStr := "0.0.0.0:" + port
		addr, _ := net.ResolveTCPAddr("tcp", addrStr)

		svr := orderservice.NewServer(
			impl,
			server.WithServiceAddr(addr),
			server.WithRegistry(r),
			server.WithSuite(kitextracing.NewServerSuite()),
			server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: constant.ServiceNameOrder}),
		)
		log.Info("Order Service RPC Server is running on " + addrStr)
		return svr.Run()
	})
	if err != nil {
		panic(err)
	}
}
