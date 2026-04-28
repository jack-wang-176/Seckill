package mq

import (
	"full_backend_practice/pkg/logger"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var Channel *amqp091.Channel

type SeckillMessage struct {
	UserID    uint64 `json:"user_id"`
	ProductID uint64 `json:"product_id"`
	OrderNo   string `json:"order_no"`
}

type UserMessage struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func InitRabbitMQ() {
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.Log.Fatal("rabbitmq dial err", zap.Error(err))
	}
	ch, err := conn.Channel()
	Channel = ch
	if err != nil {
		logger.Log.Fatal("rabbitmq channel err", zap.Error(err))
	}

	_, _ = Channel.QueueDelete("seckill_order_queue", true, true, false)
	_, _ = Channel.QueueDelete("user_register_queue", true, true, false)

	_, err = Channel.QueueDeclare("seckill_order_queue", true, false, false, false, nil)
	if err != nil {
		logger.Log.Fatal("seckill queue declare err", zap.Error(err))
	}
	_, err = Channel.QueueDeclare("user_register_queue", true, false, false, false, nil)
	if err != nil {
		logger.Log.Fatal("user queue declare err", zap.Error(err))
	}

	logger.Log.Info("rabbitmq channel created and queues declared")
}
