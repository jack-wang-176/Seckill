package order_service

import (
	"encoding/json"
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/mq"
	"log"

	"gorm.io/gorm"
)

func StartConsumer() {
	msgs, err := mq.Channel.Consume(
		"seckill_order_queue", "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 修复 1：goroutine 闭包在结尾必须加上 () 才能真正执行
	go func() {
		for d := range msgs {
			var msg mq.SeckillMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				log.Printf("fail to parse message: %v\n", err)
				// 消息格式错误，直接丢弃，不要重试
				d.Nack(false, false)
				continue
			}

			// 修复 2：去掉 Transaction 结尾多余的 ()
			err := database.DB.Transaction(func(tx *gorm.DB) error {
				res := tx.Model(&database.Product{}).
					Where("id = ? AND stock > 0", msg.ProductID).
					Update("stock", gorm.Expr("stock - 1"))

				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					return gorm.ErrRecordNotFound // 代表库存不足
				}

				// 修复 3：使用 fmt.Sprint 适配 int64 类型
				return tx.Create(&database.Order{
					OrderNo:   msg.OrderNo,
					UserID:    uint(msg.UserID),
					ProductID: uint(msg.ProductID),
					Status:    1, // 成功
				}).Error
			}) // <- 注意这里，不要加 ()

			// 修复 4：根据不同类型的错误，决定是 Ack 还是 Nack
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
	}() // <- 必须有这对括号，代表调用这个匿名函数
}
