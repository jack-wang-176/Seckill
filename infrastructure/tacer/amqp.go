package tacer

import (
	"context"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
)

type AmqpHeadersCarrier map[string]interface{}

func (c AmqpHeadersCarrier) Get(key string) string {
	if v, ok := c[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (c AmqpHeadersCarrier) Set(key string, value string) {
	c[key] = value
}

func (c AmqpHeadersCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return keys
}

func InjectAMQPHeaders(ctx context.Context, headers amqp091.Table) {
	propagation.TraceContext{}.Inject(ctx, AmqpHeadersCarrier(headers))
}

func ExtractAMQPHeaders(ctx context.Context, headers amqp091.Table) context.Context {
	return propagation.TraceContext{}.Extract(ctx, AmqpHeadersCarrier(headers))
}
