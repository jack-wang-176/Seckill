.PHONY: build-api build-order build-product build-user run-api run-order run-product run-user tidy

build-api:
	go build -o bin/api-server backend/cmd/api/main.go

build-order:
	go build -o bin/order-server backend/cmd/order/main.go

build-product:
	go build -o bin/product-server backend/cmd/product/main.go

build-user:
	go build -o bin/user-server backend/cmd/user/main.go

run-api: build-api
	./bin/api-server

run-order: build-order
	./bin/order-server

run-product: build-product
	./bin/product-server

run-user: build-user
	./bin/user-server

tidy:
	go mod tidy
