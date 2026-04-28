package main

import (
	"full_backend_practice/backend/internal/api-gateway/biz/router"
	"full_backend_practice/backend/internal/app/rpc"
	"full_backend_practice/pkg/logger"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	// 0. 初始化日志
	logger.InitLogger()
	logger.Log.Info("Starting API Gateway...")

	// 1. 初始化 RPC 客户端
	rpc.InitOrderRpc()
	rpc.InitUserRpc()

	// 2. 初始化 Hertz 网关 Server，监听 8080 端口
	h := server.Default(server.WithHostPorts("0.0.0.0:8080"))

	// 3. 注册路由
	router.Register(h)

	// 4. 启动服务
	logger.Log.Info("API Gateway is running on 0.0.0.0:8080...")
	h.Spin()
}
