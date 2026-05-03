package main

import (
	"full_backend_practice/internal/api-gateway/router"
	"full_backend_practice/internal/rpc"
	"full_backend_practice/pkg/logger"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {

	logger.InitLogger()
	logger.Log.Info("Starting API Gateway...")

	rpc.InitOrderRpc()
	rpc.InitUserRpc()

	h := server.Default(server.WithHostPorts("0.0.0.0:8080"))

	router.Register(h)

	logger.Log.Info("API Gateway is running on 0.0.0.0:8080...")
	h.Spin()
}
