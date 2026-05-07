package product

import (
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
func (p *productConsumer) StartConsumer() {}
