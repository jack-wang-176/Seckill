package mq

import (
	"full_backend_practice/pkg/config"

	"log"
	"time"

	"context"

	"github.com/rabbitmq/amqp091-go"
	clientv3 "go.etcd.io/etcd/client/v3"
)

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

func InitRabbitMQ(cfg *config.RabbitMQConfig) *RabbitClient {
	conn, err := amqp091.Dial(cfg.URL)
	if err != nil {
		log.Fatal(err)
	}
	ch, err := conn.Channel()

	if err != nil {
		log.Fatal(err)
	}
	for _, queue := range cfg.Queues {
		_, err = ch.QueueDeclare(
			queue,
			true,
			false,
			false,
			false,
			nil,
		)
	}
	client := &RabbitClient{
		Conn:    conn,
		Channel: ch,
	}
	log.Println("rabbitmq channel created")
	return client
}
func CloseRabbitMQ(client *RabbitClient) {
	if client != nil {
		if client.Channel != nil {
			client.Channel.Close()
		}
		if client.Conn != nil {
			client.Conn.Close()
		}
	}
}
