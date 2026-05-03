package user_service

import (
	"encoding/json"
	"fmt"
	"full_backend_practice/pkg/database/mysql"
	"full_backend_practice/pkg/mq"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserConsumer struct {
	MR     *mysql.MySqlWrapper
	MQ     *mq.RabbitClient
	Logger *zap.Logger
}

func NewConsumer(mr *mysql.MySqlWrapper, mq *mq.RabbitClient, log *zap.Logger) *UserConsumer {
	return &UserConsumer{
		MR:     mr,
		MQ:     mq,
		Logger: log,
	}
}

func (u *UserConsumer) StartRegisterConsumer() {
	msgs, err := u.MQ.Channel.Consume(
		"user_register", "", false, false, false, false, nil,
	)
	if err != nil {
		u.Logger.Fatal("MQ consume error", zap.Error(err))
	}
	go func() {
		for d := range msgs {
			var msg mq.UserMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				u.Logger.Error("fail to parse message", zap.Error(err))
				d.Nack(false, false)
				continue
			}
			err := u.MR.DB.Transaction(func(tx *gorm.DB) error {
				var user mysql.User
				err := tx.Model(&mysql.User{}).Where("username = ?", msg.Username).First(&user).Error
				if err == nil {
					return fmt.Errorf("username already exists")
				}
				if err != gorm.ErrRecordNotFound {
					return err
				}
				return tx.Create(&mysql.User{
					Username:     msg.Username,
					PasswordHash: msg.Password,
				}).Error
			})
			if err != nil {
				if err.Error() == "username already exists" {
					u.Logger.Warn("注册幂等，用户名已存在", zap.String("username", msg.Username))
					d.Ack(false)
				} else {
					u.Logger.Error("用户注册失败", zap.Error(err))
					d.Nack(false, true)
				}
			} else {
				u.Logger.Info("用户注册成功", zap.String("username", msg.Username))
				d.Ack(false)
			}
		}
	}()
}
