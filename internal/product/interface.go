package product

import (
	"context"
	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/kitex_gen/product"
)

type ProductDatabase interface {
	GetProductList(msg mq.ProductMessage) ([]Product, error)
	GetProduct(msg mq.ProductMessage) (Product, error)
}

type ProductServiceImpl interface {
	GetProductList(c context.Context, req *product.GetProductListReq) (resp *product.GetProductListResp, err error)
	GetProduct(c context.Context, req *product.GetProductReq) (resp *product.GetProductResp, err error)
	HeatProduct(c context.Context, req *product.HeatProductReq) (resp *product.HeatProductResp, err error)
}
type ProductConsumer interface {
	StartConsumer()
}
