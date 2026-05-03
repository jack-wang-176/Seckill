package order_service

import (
	"context"
	"encoding/json"
	"fmt"
	"full_backend_practice/kitex_gen/base"
	"full_backend_practice/kitex_gen/order"
	"full_backend_practice/pkg/database/mysql"
	"full_backend_practice/pkg/database/redis"
	"full_backend_practice/pkg/logger"
	"full_backend_practice/pkg/mq"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type OrderServiceImpl struct {
	MySqlWrapper *mysql.MySqlWrapper
	RedisWrapper *redis.RedisWrapper
	MQ           *mq.RabbitClient
	Logger       *zap.Logger
}

func NewOrderServiceImpl(mr *mysql.MySqlWrapper, rw *redis.RedisWrapper, mq *mq.RabbitClient, logger *zap.Logger) *OrderServiceImpl {
	return &OrderServiceImpl{
		MySqlWrapper: mr,
		RedisWrapper: rw,
		MQ:           mq,
		Logger:       logger,
	}
}

func (s *OrderServiceImpl) Seckill(ctx context.Context, req *order.SeckillReq) (resp *order.SeckillResp, err error) {
	resp = new(order.SeckillResp)
	resp.BaseResp = &base.BaseResp{}
	success, err := s.RedisWrapper.SimpleDecrStock(uint(req.ProductId))
	if err != nil {
		s.Logger.Error("Redis Decr Stock Error", zap.Error(err))
		resp.BaseResp.Code = 500
		resp.BaseResp.Msg = err.Error()
		return resp, nil
	}
	if !success {
		s.Logger.Warn("Stock sold out", zap.Int64("ProductID", req.ProductId))
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
	err = s.MQ.Channel.PublishWithContext(ctx, "", "order_seckill", false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})
	if err != nil {
		logger.Log.Error("RabbitMQ publish error", zap.Error(err))
		_ = s.RedisWrapper.Client.Incr(ctx, fmt.Sprintf("seckill:stock:%d", req.ProductId))
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
