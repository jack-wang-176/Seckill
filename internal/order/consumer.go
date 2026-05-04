package order

import (
	"encoding/json"
	"full_backend_practice/pkg/mq"
	"log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderConsumer interface {
	StartConsumer()
}

type orderConsumer struct {
	mysql  OrderDatabase
	mq     *mq.RabbitClient
	logger *zap.Logger
}

func NewOrderConsumer(o OrderDatabase, mq *mq.RabbitClient, log *zap.Logger) OrderConsumer {
	return &orderConsumer{
		mysql:  o,
		mq:     mq,
		logger: log,
	}
}

func (o *orderConsumer) StartConsumer() {
	msgs, err := o.mq.Channel.Consume(
		"order_seckill", "", false, false, false, false, nil)
	if err != nil {
		o.logger.Fatal("MQ consume error", zap.Error(err))
	}

	go func() {
		for d := range msgs {
			var msg mq.SeckillMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("fail to parse message: %v\n", err)
				d.Nack(false, false)
				continue
			}

			err := o.mysql.SeckillOrder(msg)

			if err != nil {
				log.Printf("订单处理失败 [订单号:%s]: %v", msg.OrderNo, err)

				if err == gorm.ErrRecordNotFound {
					d.Ack(false)
				} else {
					d.Nack(false, true)
				}
			} else {
				d.Ack(false)
				log.Printf("订单成功落盘: %s", msg.OrderNo)
			}
		}
	}()
}
