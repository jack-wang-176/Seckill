package router

import (
	"full_backend_practice/internal/api/handler"
	"full_backend_practice/internal/api/middleware"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func Register(h *server.Hertz, userH *handler.UserHandler, orderH *handler.OrderHandler, productH *handler.ProductServiceHandler) {
	v1 := h.Group("/api/v1")

	userGroup := v1.Group("/user")
	{
		userGroup.POST("/register", userH.Register)
		userGroup.POST("/login", userH.Login)
	}
	// 产品相关路由
	productGroup := v1.Group("/product")
	{
		productGroup.GET("/list", productH.GetProductList)
		productGroup.GET("/:id", productH.GetProduct)
		productGroup.POST("/create", productH.CreateProduct)
		productGroup.POST("/heat", productH.HeatProduct)
	}
	seckillGroup := v1.Group("/seckill", middleware.AuthMiddleWare())
	{
		seckillGroup.POST("/path", orderH.OrderPath)
		seckillGroup.POST("/result", orderH.OrderResult)
		seckillGroup.POST("/order/:path", orderH.CreateOrder)
	}
}
