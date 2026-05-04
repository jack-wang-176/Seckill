package main

import (
	"full_backend_practice/internal/api/router"
	"full_backend_practice/internal/rpc"
	"full_backend_practice/pkg/logger"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {

	lg := logger.InitLogger()

	logger.SetLogger(lg)
	lg.Info("Starting API Gateway...")

	rpc.InitOrderRpc(lg)
	rpc.InitUserRpc(lg)

	h := server.Default(server.WithHostPorts("0.0.0.0:8080"))

	router.Register(h)

	lg.Info("API Gateway is running on 0.0.0.0:8080...")
	h.Spin()
}
