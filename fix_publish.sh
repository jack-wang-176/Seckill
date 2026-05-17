set -e

# Order handler
sed -i '' 's/body, err := json.Marshal(msgStuct)/headers := amqp091.Table{}\n\t\ttacer.InjectAMQPHeaders(ctx, headers)\n\n\t\tbody, err := json.Marshal(msgStuct)/g' internal/order/handler.go
sed -i '' '/DeliveryMode: amqp091.Persistent,/a\
\t\t\tHeaders:      headers,
' internal/order/handler.go
sed -i '' 's/"github.com\/rabbitmq\/amqp091-go"/"github.com\/rabbitmq\/amqp091-go"\n\t"full_backend_practice\/infrastructure\/tacer"/g' internal/order/handler.go

# User handler
sed -i '' 's/body, err := json.Marshal(mqmsg)/headers := amqp091.Table{}\n\t\ttacer.InjectAMQPHeaders(ctx, headers)\n\n\t\tbody, err := json.Marshal(mqmsg)/g' internal/user/handler.go
sed -i '' '/DeliveryMode: amqp091.Persistent,/a\
\t\t\tHeaders:      headers,
' internal/user/handler.go

# Product handler
sed -i '' 's/body, err := json.Marshal(msg)/headers := amqp091.Table{}\n\t\ttacer.InjectAMQPHeaders(ctx, headers)\n\n\t\tbody, err := json.Marshal(msg)/g' internal/product/handler.go
sed -i '' '/DeliveryMode: amqp091.Persistent,/a\
\t\t\tHeaders:      headers,
' internal/product/handler.go
sed -i '' 's/"github.com\/rabbitmq\/amqp091-go"/"github.com\/rabbitmq\/amqp091-go"\n\t"full_backend_practice\/infrastructure\/tacer"/g' internal/product/handler.go

