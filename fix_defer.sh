sed -i '' -e 's/for d := range msgs {/for d := range msgs {\
\t\t\tfunc(d struct{amqp091.Delivery}) {\
/' internal/order/consumer.go

# But d's type is amqp091.Delivery. Since we are inside the same package block we can just use func() and capture d? 
# Wait, d is re-declared in ranging from Go 1.22+ safe, but passing it as an argument is better.
