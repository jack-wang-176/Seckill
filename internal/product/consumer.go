package product

import (
	"encoding/json"
	"full_backend_practice/infrastructure/mq"

	"go.uber.org/zap"
)

type productConsumer struct {
	MR     ProductDatabase
	MQ     *mq.RabbitClient
	Logger *zap.Logger
}

func NewProductConsumer(mr ProductDatabase, mq *mq.RabbitClient, log *zap.Logger) ProductConsumer {
	return &productConsumer{
		MR:     mr,
		MQ:     mq,
		Logger: log,
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

			_, err := p.MR.GetProductList(msg)

			if err != nil {
				p.Logger.Error("Failed to get product list", zap.Error(err))
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
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

			_, err := p.MR.GetProduct(msg)

			if err != nil {
				p.Logger.Error("Failed to get product", zap.Error(err))
				d.Nack(false, true)
			} else {
				d.Ack(false)
			}
		}
	}()
}
