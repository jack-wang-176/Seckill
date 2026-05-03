package order_service

import (
	"encoding/json"
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/mq"
	"log"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderConsumer struct {
	DB     *gorm.DB
	MQ     *mq.RabbitClient
	Logger *zap.Logger
}

func NewOrderConsumer(db *gorm.DB, mq *mq.RabbitClient, log *zap.Logger) *OrderConsumer {
	return &OrderConsumer{
		DB:     db,
		MQ:     mq,
		Logger: log,
	}
}

func (o *OrderConsumer) StartConsumer() {
	msgs, err := o.MQ.Channel.Consume(
		"order_seckill", "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for d := range msgs {
			var msg mq.SeckillMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("fail to parse message: %v\n", err)
				d.Nack(false, false)
				continue
			}

			err := o.DB.Transaction(func(tx *gorm.DB) error {
				res := tx.Model(&database.Product{}).
					Where("id = ? AND stock > 0", msg.ProductID).
					Update("stock", gorm.Expr("stock - 1"))

				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					return gorm.ErrRecordNotFound
				}

				return tx.Create(&database.Order{
					OrderNo:   msg.OrderNo,
					UserID:    uint(msg.UserID),
					ProductID: uint(msg.ProductID),
					Status:    1, // 成功
				}).Error
			})

			if err != nil {
				log.Printf("订单处理失败 [订单号:%s]: %v", msg.OrderNo, err)

				if err == gorm.ErrRecordNotFound {
					// 业务失败：库存不足。这种错误重试也没用，直接 Ack 确认掉（或丢入死信队列）
					d.Ack(false)
				} else {
					// 系统失败：比如 MySQL 宕机、网络波动。Nack 并重回队列 (requeue=true) 等待重试
					d.Nack(false, true)
				}
			} else {
				// 成功落盘
				d.Ack(false)
				log.Printf("订单成功落盘: %s", msg.OrderNo)
			}
		}
	}()
}
