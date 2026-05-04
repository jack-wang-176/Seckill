package main

import (
	"context"
	order "full_backend_practice/kitex_gen/order"
	product "full_backend_practice/kitex_gen/product"
)

// OrderServiceImpl implements the last service interface defined in the IDL.
type OrderServiceImpl struct{}

// Seckill implements the OrderServiceImpl interface.
func (s *OrderServiceImpl) Seckill(ctx context.Context, req *order.SeckillReq) (resp *order.SeckillResp, err error) {
	// TODO: Your code here...
	return
}

// GetProduct implements the ProductServiceImpl interface.
func (s *ProductServiceImpl) GetProduct(ctx context.Context, req *product.GetProductReq) (resp *product.GetProductResp, err error) {
	// TODO: Your code here...
	return
}
