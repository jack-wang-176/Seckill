# Fix Tracer name inside the consumers to carry the correct kitex/opentelemetry globals
sed -i '' 's/tracer := otel.Tracer("consumer")/tracer := otel.Tracer("order_service")/g' internal/order/consumer.go
sed -i '' 's/tracer := otel.Tracer("consumer")/tracer := otel.Tracer("user_service")/g' internal/user/consumer.go
