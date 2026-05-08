package main

import (
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/internal/api/handler"
	"full_backend_practice/internal/api/router"
	"full_backend_practice/internal/rpc"

	"github.com/cloudwego/hertz/pkg/app/server"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

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
		h := server.Default(server.WithHostPorts("0.0.0.0:8080"))
		log.Info("Starting API Gateway...")
		router.Register(h, userH, orderH, productH)
		log.Info("API Gateway is running on 0.0.0.0:8080")
		h.Spin()
	})
	if err != nil {
		panic(err)
	}
}
