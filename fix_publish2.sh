sed -i '' 's/tttHeaders/Headers/g' internal/order/handler.go
sed -i '' 's/tttHeaders/Headers/g' internal/user/handler.go
sed -i '' 's/tttHeaders/Headers/g' internal/product/handler.go

sed -i '' 's/"golang.org\/x\/crypto\/bcrypt"/"golang.org\/x\/crypto\/bcrypt"\n\t"full_backend_practice\/infrastructure\/tacer"/g' internal/user/handler.go
