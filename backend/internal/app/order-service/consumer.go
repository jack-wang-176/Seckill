package order_service

import (
	"encoding/json"
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/logger"
	"full_backend_practice/pkg/mq"
	"go.uber.org/zap"

	"gorm.io/gorm"
)

func StartConsumer() {
	msgs, err := mq.Channel.Consume(
		"seckill_order_queue", "", false, false, false, false, nil)
	if err != nil {
		logger.Log.Fatal("seckill consumer start err", zap.Error(err))
	}

	go func() {
		for d := range msgs {
			var msg mq.SeckillMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				logger.Log.Error("fail to parse seckill message", zap.Error(err))
				d.Nack(false, false)
				continue
			}

			err := database.DB.Transaction(func(tx *gorm.DB) error {
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
					Status:    1,
				}).Error
			})

			if err != nil {
				logger.Log.Error("seckill order process err", zap.String("order_no", msg.OrderNo), zap.Error(err))

				if err == gorm.ErrRecordNotFound {
					d.Ack(false)
				} else {
					d.Nack(false, true)
				}
			} else {
				d.Ack(false)
				logger.Log.Info("seckill order success", zap.String("order_no", msg.OrderNo))
			}
		}
	}()
}
