package mq

import (
	"log"

	"github.com/rabbitmq/amqp091-go"
)

var Channel *amqp091.Channel

type SeckillMessage struct {
	UserID    uint64 `json:"user_id"`
	ProductID uint64 `json:"product_id"`
	OrderNo   string `json:"order_no"`
}

func InitRabbitMQ() {
	conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	ch, err := conn.Channel()
	Channel = ch
	if err != nil {
		log.Fatal(err)
	}
	_, err = Channel.QueueDelete(
		"seckill_order_queue", true, true, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("rabbitmq channel created")
}
