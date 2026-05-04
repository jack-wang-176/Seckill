package router

import (
	"full_backend_practice/internal/api/handler"
	"full_backend_practice/internal/api/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz, userH *handler.UserHandler, orderH *handler.OrderHandler) {
	v1 := h.Group("/api/v1")

	userGroup := v1.Group("/user")
	{
		userGroup.POST("/register", userH.Register)
		userGroup.GET("/login", userH.Login)
	}
	seckillGroup := v1.Group("/seckill", middleware.AuthMiddleWare())
	{
		seckillGroup.POST("/order", orderH.CreateOrder)
	}
}
