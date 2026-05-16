package main

import (
	"full_backend_practice/infrastructure/logger"
	"full_backend_practice/infrastructure/tacer"
	"full_backend_practice/internal/api/handler"
	"full_backend_practice/internal/api/router"
	"full_backend_practice/internal/rpc"
	"full_backend_practice/pkg/config"

	"github.com/cloudwego/hertz/pkg/app/server"
	hertztracing "github.com/hertz-contrib/obs-opentelemetry/tracing"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func buildContainer() *dig.Container {
	c := dig.New()

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
		traceCfg *config.TracerConfig,
	) {
		shutdown := tacer.InitTracer(traceCfg, "api_gateway")
		defer shutdown()
		serverTracer, tracerCfg := hertztracing.NewServerTracer()
		port := cfg.ApiGatewayPort
		if port == "" {
			port = "8081" // fallback
		}
		hostPort := "0.0.0.0:" + port
		h := server.Default(server.WithHostPorts(hostPort), serverTracer)
		h.Use(hertztracing.ServerMiddleware(tracerCfg))
		log.Info("Starting API Gateway...")
		router.Register(h, userH, orderH, productH)
		log.Info("API Gateway is running on " + hostPort)
		h.Spin()
	})
	if err != nil {
		panic(err)
	}
}
