package order

import (
	"context"
	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/kitex_gen/order"
)

type OrderServiceImpl interface {
	Seckill(ctx context.Context, req *order.SeckillReq) (resp *order.SeckillResp, err error)
	GetSeckillPath(ctx context.Context, req *order.GetSeckillPathReq) (resp *order.GetSeckillPathResp, err error)
	GetSeckillResult_(ctx context.Context, req *order.GetSeckillResultReq) (resp *order.GetSeckillResultResp, err error)
}

type OrderConsumer interface {
	StartConsumer()
}

type OrderDatabase interface {
	SeckillOrder(msg mq.SeckillMessage) error
	GetProductEndTime(productID uint) (int64, error)
}
