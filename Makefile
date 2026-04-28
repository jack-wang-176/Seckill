.PHONY: build-api build-order run-api run-order tidy

build-api:
	go build -o bin/api-server backend/cmd/api/main.go

build-order:
	go build -o bin/order-server backend/cmd/order/main.go

run-api: build-api
	./bin/api-server

run-order: build-order
	./bin/order-server

tidy:
	go mod tidy
