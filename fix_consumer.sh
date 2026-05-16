sed -i '' 's/go func() {/go func() {\n\t\t\ttracer := otel.Tracer("consumer")\n/g' internal/order/consumer.go
sed -i '' 's/var msg mq.SeckillMessage/ctx := tacer.ExtractAMQPHeaders(context.Background(), d.Headers)\n\t\t\t\t\t\tctx, span := tracer.Start(ctx, "ConsumeOrderSeckill")\n\t\t\t\t\t\tdefer span.End()\n\t\t\t\t\t\tvar msg mq.SeckillMessage/g' internal/order/consumer.go
sed -i '' 's/"go.uber.org\/zap"/"go.uber.org\/zap"\n\t"full_backend_practice\/infrastructure\/tacer"\n\t"go.opentelemetry.io\/otel"/g' internal/order/consumer.go

sed -i '' 's/go func() {/go func() {\n\t\t\ttracer := otel.Tracer("consumer")\n/g' internal/user/consumer.go
sed -i '' 's/var msg mq.UserMessage/ctx := tacer.ExtractAMQPHeaders(context.Background(), d.Headers)\n\t\t\t\t\t\tctx, span := tracer.Start(ctx, "ConsumeUserRegister")\n\t\t\t\t\t\tdefer span.End()\n\t\t\t\t\t\tvar msg mq.UserMessage/g' internal/user/consumer.go
sed -i '' 's/"go.uber.org\/zap"/"go.uber.org\/zap"\n\t"context"\n\t"full_backend_practice\/infrastructure\/tacer"\n\t"go.opentelemetry.io\/otel"/g' internal/user/consumer.go
