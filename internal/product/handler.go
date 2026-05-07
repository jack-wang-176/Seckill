package product

import (
	"context"
	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/mq"
	product "full_backend_practice/kitex_gen/product"

	"go.uber.org/zap"
)

// ProductServiceImpl implements the last service interface defined in the IDL.
type productServiceImpl struct {
	MySqlWrapper ProductDatabase
	RedisWrapper *database.RedisWrapper
	MQ           *mq.RabbitClient
	Logger       *zap.Logger
}

func NewProductServiceImpl(mr ProductDatabase, rw *database.RedisWrapper, mq *mq.RabbitClient, logger *zap.Logger) *productServiceImpl {
	return &productServiceImpl{
		MySqlWrapper: mr,
		RedisWrapper: rw,
		MQ:           mq,
		Logger:       logger,
	}
}

// GetProductList implements the ProductServiceImpl interface.
func (s *productServiceImpl) GetProductList(ctx context.Context, req *product.GetProductListReq) (resp *product.GetProductListResp, err error) {
	// TODO: Your code here...
	return
}

// GetProduct implements the ProductServiceImpl interface.
func (s *productServiceImpl) GetProduct(ctx context.Context, req *product.GetProductReq) (resp *product.GetProductResp, err error) {
	// TODO: Your code here...
	return
}

// HeatProduct implements the ProductServiceImpl interface.
func (s *productServiceImpl) HeatProduct(ctx context.Context, req *product.HeatProductReq) (resp *product.HeatProductResp, err error) {
	// TODO: Your code here...
	return
}
