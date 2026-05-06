package order

import (
	"context"
	"encoding/json"
	"fmt"
	"full_backend_practice/kitex_gen/base"
	"full_backend_practice/kitex_gen/order"
	"full_backend_practice/pkg/database"

	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/pkg/response"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type orderServiceImpl struct {
	MySqlWrapper OrderDatabase
	RedisWrapper *database.RedisWrapper
	MQ           *mq.RabbitClient
	Logger       *zap.Logger
}

func NewOrderServiceImpl(mr OrderDatabase, rw *database.RedisWrapper, mq *mq.RabbitClient, logger *zap.Logger) OrderServiceImpl {
	return &orderServiceImpl{
		MySqlWrapper: mr,
		RedisWrapper: rw,
		MQ:           mq,
		Logger:       logger,
	}
}

func (s *orderServiceImpl) Seckill(ctx context.Context, req *order.SeckillReq) (resp *order.SeckillResp, err error) {
	resp = new(order.SeckillResp)
	resp.BaseResp = &base.BaseResp{}
	success, err := s.RedisWrapper.SimpleDecrStock(uint(req.ProductId))
	if err != nil {
		s.Logger.Error("Redis Decr Stock Error", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, err.Error())
		return resp, nil
	}
	if !success {
		s.Logger.Warn("Stock sold out", zap.Int64("ProductID", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "sell out")
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
		s.Logger.Error("JSON marshal error", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "JSON marshal error")
		return resp, nil
	}
	err = s.MQ.Channel.PublishWithContext(ctx, "", "order_seckill", false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})
	if err != nil {
		s.Logger.Error("RabbitMQ publish error", zap.Error(err))
		_ = s.RedisWrapper.Client.Incr(ctx, fmt.Sprintf("seckill:stock:%d", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to seckill order")
		return resp, nil
	}

	s.Logger.Info("Order pushed to MQ successfully", zap.String("order_no", orderNod))

	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "seckill success")
	resp.OrderNo = &orderNod
	return resp, nil
}
