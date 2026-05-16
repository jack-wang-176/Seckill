package user

import (
	"encoding/json"
	"full_backend_practice/infrastructure/mq"

	"go.uber.org/zap"
	"context"
	"full_backend_practice/infrastructure/tacer"
	"go.opentelemetry.io/otel"
)

type userConsumer struct {
	MR     UserDatabase
	MQ     *mq.RabbitClient
	Logger *zap.Logger
}

func NewConsumer(mr UserDatabase, mq *mq.RabbitClient, log *zap.Logger) UserConsumer {
	return &userConsumer{
		MR:     mr,
		MQ:     mq,
		Logger: log,
	}
}

func (u *userConsumer) StartRegisterConsumer() {
	_, qerr := u.MQ.Channel.QueueDeclare(
		"user_register",
		true,
		false,
		false,
		false,
		nil,
	)
	if qerr != nil {
		u.Logger.Fatal("MQ queue declare error", zap.Error(qerr))
	}

	msgs, err := u.MQ.Channel.Consume(
		"user_register", "", false, false, false, false, nil,
	)
	if err != nil {
		u.Logger.Fatal("MQ consume error", zap.Error(err))
	}
	go func() {
			tracer := otel.Tracer("consumer")

		for d := range msgs {
			ctx := tacer.ExtractAMQPHeaders(context.Background(), d.Headers)
						ctx, span := tracer.Start(ctx, "ConsumeUserRegister")
						defer span.End()
						var msg mq.UserMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				u.Logger.Error("fail to parse message", zap.Error(err))
				d.Nack(false, false)
				continue
			}
			err := u.MR.RegisterUser(msg)
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
