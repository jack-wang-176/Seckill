package order_service

import (
	"context"
	"encoding/json"
	"fmt"
	"full_backend_practice/kitex_gen/base"
	"full_backend_practice/kitex_gen/order"
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/mq"
	"full_backend_practice/pkg/logger"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type OrderServiceImpl struct{}

func (s *OrderServiceImpl) Seckill(ctx context.Context, req *order.SeckillReq) (resp *order.SeckillResp, err error) {
	resp = new(order.SeckillResp)
	resp.BaseResp = &base.BaseResp{}
	success, err := database.SimpleDecrStock(uint(req.ProductId))
	if err != nil {
		logger.Log.Error("Redis Decr Stock Error", zap.Error(err))
		resp.BaseResp.Code = 500
		resp.BaseResp.Msg = err.Error()
		return resp, nil
	}
	if !success {
		logger.Log.Warn("Stock sold out", zap.Int64("ProductID", req.ProductId))
		resp.BaseResp.Code = 500
		resp.BaseResp.Msg = "sell out"
		return resp, nil
	}
	orderNod := fmt.Sprintf("SN%d%d", time.Now().UnixNano(), req.UserId)
	msgStuct := mq.SeckillMessage{
		UserID:    uint64(req.UserId),
		ProductID: uint64(req.ProductId),
		OrderNo:   orderNod,
	}
	body, err := json.Marshal(msgStuct)
	if err != nil {
		logger.Log.Error("JSON marshal error", zap.Error(err))
		resp.BaseResp.Code = 500
		resp.BaseResp.Msg = "JSON marshal error"
		return resp, nil
	}
	err = mq.Channel.PublishWithContext(ctx, "", "seckill_order_queue", false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})
	if err != nil {
		logger.Log.Error("RabbitMQ publish error", zap.Error(err))
		_ = database.Client.Incr(ctx, fmt.Sprintf("seckill:stock:%d", req.ProductId))
		resp.BaseResp.Code = 500
		resp.BaseResp.Msg = "fail to seckill order"
		return resp, nil
	}

	logger.Log.Info("Order pushed to MQ successfully", zap.String("order_no", orderNod))
	
	resp.BaseResp.Code = 200
	resp.BaseResp.Msg = "seckill success"
	resp.OrderNo = &orderNod
	return resp, nil
}
