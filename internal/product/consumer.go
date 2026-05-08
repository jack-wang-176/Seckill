package product

import (
	"context"
	"encoding/json"
	"fmt"
	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/mq"
	"time"

	"go.uber.org/zap"
)

type productConsumer struct {
	MR          ProductDatabase
	MQ          *mq.RabbitClient
	RedisWorker *database.RedisWrapper
	Logger      *zap.Logger
}

func NewProductConsumer(mr ProductDatabase, mq *mq.RabbitClient, rw *database.RedisWrapper, log *zap.Logger) ProductConsumer {
	return &productConsumer{
		MR:          mr,
		MQ:          mq,
		RedisWorker: rw,
		Logger:      log,
	}
}
func (p *productConsumer) StartConsumer() {
	p.StartListConsumer()
	p.StartSingleConsumer()
}

func (p *productConsumer) StartListConsumer() {
	msgs, err := p.MQ.Channel.Consume("product_list", "", false, false, false, false, nil)
	if err != nil {
		p.Logger.Fatal("MQ consume error", zap.Error(err))
	}

	go func() {
		for d := range msgs {
			var msg mq.ProductMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				p.Logger.Error("fail to parse message", zap.Error(err))
				d.Nack(false, false)
				continue
			}

			products, err := p.MR.GetProductList(msg)
			if err != nil {
				p.Logger.Error("Failed to get product list", zap.Error(err))
				d.Nack(false, true)
				continue
			}

			// 缓存结果到 Redis，供 API 层读取
			key := "product:list"
			b, err := json.Marshal(products)
			if err == nil {
				_ = p.RedisWorker.Client.Set(context.Background(), key, b, 60*5*time.Second).Err()
			}

			d.Ack(false)
		}
	}()
}

func (p *productConsumer) StartSingleConsumer() {
	msgs, err := p.MQ.Channel.Consume("product_single", "", false, false, false, false, nil)
	if err != nil {
		p.Logger.Fatal("MQ consume error", zap.Error(err))
	}

	go func() {
		for d := range msgs {
			var msg mq.ProductMessage
			if err := json.Unmarshal(d.Body, &msg); err != nil {
				p.Logger.Error("fail to parse message", zap.Error(err))
				d.Nack(false, false)
				continue
			}

			prod, err := p.MR.GetProduct(msg)
			if err != nil {
				p.Logger.Error("Failed to get product", zap.Error(err))
				d.Nack(false, true)
				continue
			}

			key := fmt.Sprintf("product:%d", msg.ProductID)
			b, err := json.Marshal(prod)
			if err == nil {
				_ = p.RedisWorker.Client.Set(context.Background(), key, b, 60*5*time.Second).Err()
			}

			d.Ack(false)
		}
	}()
}
