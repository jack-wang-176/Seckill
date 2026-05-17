sed -i '' -e 's|"go.uber.org/zap"|"go.uber.org/zap"\
\t"full_backend_practice/infrastructure/tacer"\
\t"go.opentelemetry.io/otel"|g' internal/product/consumer.go

sed -i '' -e 's/for d := range msgs {/for d := range msgs {\
\t\t\tfunc() {\
\t\t\t\tctx := tacer.ExtractAMQPHeaders(context.Background(), d.Headers)\
\t\t\t\tctx, span := otel.Tracer("product_service").Start(ctx, "ConsumeProductList")\
\t\t\t\tdefer span.End()/g' internal/product/consumer.go

sed -i '' -e 's/\td\.Ack/\t\t\td\.Ack/g' internal/product/consumer.go
sed -i '' -e 's/\td\.Nack/\t\t\td\.Nack/g' internal/product/consumer.go
sed -i '' -e 's/continue/return/g' internal/product/consumer.go

sed -i '' -e 's/\t\t}//g' internal/product/consumer.go

