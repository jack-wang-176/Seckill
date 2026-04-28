package router

import (
	"full_backend_practice/backend/internal/api-gateway/biz/handler"
	"full_backend_practice/backend/internal/api-gateway/biz/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz) {
	v1 := h.Group("/api/v1")

	userGroup := v1.Group("/user")
	{
		userGroup.POST("/register", handler.Register)
	}
	seckillGroup := v1.Group("/seckill", middleware.AuthMiddleWare())
	{
		seckillGroup.POST("/order", handler.CreateOrder)
	}
}
