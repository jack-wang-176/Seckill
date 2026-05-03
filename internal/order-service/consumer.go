package order_service

import (
	"encoding/json"
	"full_backend_practice/pkg/database/mysql"
	"full_backend_practice/pkg/mq"
	"log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderConsumer struct {
	MR     *mysql.MySqlWrapper
	MQ     *mq.RabbitClient
	Logger *zap.Logger
}

func NewOrderConsumer(mr *mysql.MySqlWrapper, mq *mq.RabbitClient, log *zap.Logger) *OrderConsumer {
	return &OrderConsumer{
		MR:     mr,
		MQ:     mq,
		Logger: log,
	}
}

func (o *OrderConsumer) StartConsumer() {
	msgs, err := o.MQ.Channel.Consume(
		"order_seckill", "", false, false, false, false, nil)
	if err != nil {
		o.Logger.Fatal("MQ consume error", zap.Error(err))
	}

	go func() {
		for d := range msgs {
			var msg mq.SeckillMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("fail to parse message: %v\n", err)
				d.Nack(false, false)
				continue
			}

			err := o.MR.SeckillOrder(msg)

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
