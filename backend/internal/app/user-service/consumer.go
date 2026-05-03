package user_service

import (
	"encoding/json"
	"fmt"
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/logger"
	"full_backend_practice/pkg/mq"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func StartRegisterConsumer() {
	msgs, err := mq.Client.Channel.Consume(
		"user_register", "", false, false, false, false, nil,
	)
	if err != nil {
		logger.Log.Fatal("MQ consume error", zap.Error(err))
	}
	go func() {
		for d := range msgs {
			var msg mq.UserMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				logger.Log.Error("fail to parse message", zap.Error(err))
				d.Nack(false, false)
				continue
			}
			err := database.DB.Transaction(func(tx *gorm.DB) error {
				var user database.User
				err := tx.Model(&database.User{}).Where("username = ?", msg.Username).First(&user).Error
				if err == nil {
					return fmt.Errorf("username already exists")
				}
				if err != gorm.ErrRecordNotFound {
					return err // 系统错误
				}
				return tx.Create(&database.User{
					Username:     msg.Username,
					PasswordHash: msg.Password,
				}).Error
			})
			if err != nil {
				if err.Error() == "username already exists" {
					logger.Log.Warn("注册幂等，用户名已存在", zap.String("username", msg.Username))
					d.Ack(false) // 幂等，直接确认
				} else {
					logger.Log.Error("用户注册失败", zap.Error(err))
					d.Nack(false, true) // 系统错误，重试
				}
			} else {
				logger.Log.Info("用户注册成功", zap.String("username", msg.Username))
				d.Ack(false)
			}
		}
	}()
}
