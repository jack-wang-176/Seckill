package handler

import (
	"context"
	"full_backend_practice/kitex_gen/product/productservice"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

type ProductServiceHandler struct {
	Client productservice.Client
	logger *zap.Logger
}

func NewProductServiceHandler(client productservice.Client, logger *zap.Logger) *ProductServiceHandler {
	return &ProductServiceHandler{Client: client, logger: logger}
}

func (s ProductServiceHandler) GetProductList(c context.Context, ctx *app.RequestContext) {
	//感觉直接调用就好，这几个接口handler级别都没有什么前置操作
}
func (s ProductServiceHandler) GetProduct(c context.Context, ctx *app.RequestContext) {

}

// 针对这个接口目前暂时不做role的区分，如果有时间的话鉴权可以放在router或者这里也行吧
func (s *ProductServiceHandler) HeatProduct(c context.Context, ctx *app.RequestContext) {

}
