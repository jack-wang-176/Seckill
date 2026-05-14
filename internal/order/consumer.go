package order

import (
	"context"
	"encoding/json"
	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/pkg/constant"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type orderConsumer struct {
	mysql        OrderDatabase
	RedisWrapper *database.RedisWrapper
	mq           *mq.RabbitClient
	logger       *zap.Logger
}

func NewOrderConsumer(o OrderDatabase, mq *mq.RabbitClient, redisWrapper *database.RedisWrapper, log *zap.Logger) OrderConsumer {
	return &orderConsumer{
		mysql:        o,
		RedisWrapper: redisWrapper,
		mq:           mq,
		logger:       log,
	}
}

func (o *orderConsumer) StartConsumer() {
	msgs, err := o.mq.Channel.Consume(
		constant.QueueOrderSeckill, "", false, false, false, false, nil)
	if err != nil {
		o.logger.Fatal("MQ consume error", zap.Error(err))
	}

	go func() {
		for d := range msgs {
			var msg mq.SeckillMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				o.logger.Error("fail to parse message", zap.Error(err))
				d.Nack(false, false)
				continue
			}

			err := o.mysql.SeckillOrder(msg)

			if err != nil {
				o.logger.Error("订单处理失败", zap.String("order_no", msg.OrderNo), zap.Error(err))

				if err == gorm.ErrRecordNotFound {
					d.Ack(false)
				} else {
					d.Nack(false, true)
				}
			} else {
				//调用sql查询product时间。
				endTime, err := o.mysql.GetProductEndTime(uint(msg.ProductID))
				if err != nil {
					o.logger.Error("Failed to detect product end time", zap.String("order_no", msg.OrderNo), zap.Error(err))
					d.Nack(false, true)
					continue
				}
				if endTime < time.Now().Unix() {
					o.logger.Warn("Product already ended", zap.String("order_no", msg.OrderNo), zap.Int64("end_time", endTime))
					d.Ack(false)
					continue
				}

				d.Ack(false)
				//调用redis
				err = o.RedisWrapper.SendSeckillCre(context.Background(), uint(msg.ProductID), uint(msg.UserID), msg.OrderNo)
				if err != nil {
					o.logger.Error("Redis send seckill result error", zap.String("order_no", msg.OrderNo), zap.Error(err))
				}
				o.logger.Info("订单处理成功", zap.String("order_no", msg.OrderNo))
			}
		}
	}()
}
