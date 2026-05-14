package main

import (
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/internal/api/handler"
	"full_backend_practice/internal/api/router"
	"full_backend_practice/internal/rpc"
	"full_backend_practice/pkg/config"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func buildContainer() *dig.Container {
	c := dig.New()
	// 加载所有外部配置并注册到容器
	if err := config.RegisterConfigProviders(c); err != nil {
		panic(err)
	}
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
		cfg *config.ServerConfig,
		log *zap.Logger,
	) {
		port := cfg.ApiGatewayPort
		if port == "" {
			port = "8081" // fallback
		}
		hostPort := "0.0.0.0:" + port
		h := server.Default(server.WithHostPorts(hostPort))
		log.Info("Starting API Gateway...")
		router.Register(h, userH, orderH, productH)
		log.Info("API Gateway is running on " + hostPort)
		h.Spin()
	})
	if err != nil {
		panic(err)
	}
}
