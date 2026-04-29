package mq

import (
	"log"
	"time"

	"context"

	"github.com/rabbitmq/amqp091-go"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var Client *RabbitClient

type SeckillMessage struct {
	UserID    uint64 `json:"user_id"`
	ProductID uint64 `json:"product_id"`
	OrderNo   string `json:"order_no"`
}
type UserMessage struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RabbitClient struct {
	Conn    *amqp091.Connection
	Channel *amqp091.Channel
}

func GetConfigFromEtcd(endpoins []string, key string) (string, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoins,
		DialTimeout: 5 * time.Second,
	})
	defer cli.Close()
	if err != nil {
		return "", err
	}
	resp, err := cli.Get(context.Background(), key)
	if err != nil || len(resp.Kvs) == 0 {
		return "amqp://guest:guest@localhost:5672/", nil
	}
	return string(resp.Kvs[0].Value), nil
}

func InitRabbitMQ(url string, queues []string) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		log.Fatal(err)
	}
	ch, err := conn.Channel()

	if err != nil {
		log.Fatal(err)
	}
	for _, queue := range queues {
		_, err = ch.QueueDeclare(
			queue, // name
			true,  // durable
			false, // delete when unused
			false, // exclusive
			false, // no-wait
			nil,   // arguments
		)
	}
	Client = &RabbitClient{
		Conn:    conn,
		Channel: ch,
	}
	log.Println("rabbitmq channel created")
}
func CloseRabbitMQ() {
	if Client != nil {
		if Client.Channel != nil {
			Client.Channel.Close()
		}
		if Client.Conn != nil {
			Client.Conn.Close()
		}
	}
}
