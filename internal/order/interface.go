package order

import (
	"context"
	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/kitex_gen/order"
)

type OrderServiceImpl interface {
	Seckill(ctx context.Context, req *order.SeckillReq) (resp *order.SeckillResp, err error)
}

type OrderConsumer interface {
	StartConsumer()
}

type OrderDatabase interface {
	SeckillOrder(msg mq.SeckillMessage) error
}
