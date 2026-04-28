package main

import (
	"full_backend_practice/backend/internal/api-gateway/biz/router"
	"full_backend_practice/backend/internal/app/rpc"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	// 1. 初始化 RPC 客户端，连接到后面要启动的 Order 微服务
	rpc.InitOrderRpc()

	// 2. 初始化 Hertz 网关 Server，监听 8080 端口
	h := server.Default(server.WithHostPorts("0.0.0.0:8080"))

	// 3. 注册路由
	router.Register(h)

	// 4. 启动服务
	h.Spin()
}
