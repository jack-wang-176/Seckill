package product

import (
	"context"
	"encoding/json"
	"fmt"
	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/infrastructure/tacer"
	"full_backend_practice/pkg/constant"

	"go.opentelemetry.io/otel"
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
	msgs, err := p.MQ.Channel.Consume(constant.QueueProductList, "", false, false, false, false, nil)
	if err != nil {
		p.Logger.Fatal("MQ consume error", zap.Error(err))
	}

	go func() {
		tracer := otel.Tracer("product_service")

		for d := range msgs {
			func() {
				ctx := tacer.ExtractAMQPHeaders(context.Background(), d.Headers)
				ctx, span := tracer.Start(ctx, "ConsumeProductList")
				defer span.End()

				var msg mq.ProductMessage
				if err := json.Unmarshal(d.Body, &msg); err != nil {
					p.Logger.Error("fail to parse message", zap.Error(err))
					d.Nack(false, false)
					return
				}

				products, err := p.MR.GetProductList(ctx, msg)
				if err != nil {
					p.Logger.Error("Failed to get product list", zap.Error(err))
					d.Nack(false, true)
					return
				}

				// 缓存结果到 Redis
				key := constant.ProductListKey
				b, err := json.Marshal(products)
				if err == nil {
					_ = p.RedisWorker.Client.Set(ctx, key, b, constant.ProductCacheTTL).Err()
				}

				d.Ack(false)
			}()
		}
	}()
}

func (p *productConsumer) StartSingleConsumer() {
	msgs, err := p.MQ.Channel.Consume(constant.QueueProductSingle, "", false, false, false, false, nil)
	if err != nil {
		p.Logger.Fatal("MQ consume error", zap.Error(err))
	}

	go func() {
		tracer := otel.Tracer("product_service")

		for d := range msgs {
			func() {
				ctx := tacer.ExtractAMQPHeaders(context.Background(), d.Headers)
				ctx, span := tracer.Start(ctx, "ConsumeProductSingle")
				defer span.End()

				var msg mq.ProductMessage
				if err := json.Unmarshal(d.Body, &msg); err != nil {
					p.Logger.Error("fail to parse message", zap.Error(err))
					d.Nack(false, false)
					return
				}

				prod, err := p.MR.GetProduct(ctx, msg)
				if err != nil {
					p.Logger.Error("Failed to get product", zap.Error(err))
					d.Nack(false, true)
					return
				}

				key := fmt.Sprintf(constant.ProductKeyFormat, msg.ProductID)
				b, err := json.Marshal(prod)
				if err == nil {
					_ = p.RedisWorker.Client.Set(ctx, key, b, constant.ProductCacheTTL).Err()
				}

				d.Ack(false)
			}()
		}
	}()
}
