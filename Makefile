.PHONY: build-api build-order build-user run-api run-order run-user tidy

build-api:
	go build -o bin/api-server backend/cmd/api/main.go

build-order:
	go build -o bin/order-server backend/cmd/order/main.go

build-user:
	go build -o bin/user-server backend/cmd/user/main.go

run-api: build-api
	./bin/api-server

run-order: build-order
	./bin/order-server

run-user: build-user
	./bin/user-server

tidy:
	go mod tidy
