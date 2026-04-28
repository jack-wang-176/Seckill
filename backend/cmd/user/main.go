package main

import (
	"net"

	user_service "full_backend_practice/backend/internal/app/user-service"
	"full_backend_practice/kitex_gen/user/userservice"
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/logger"
	"full_backend_practice/pkg/mq"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	logger.InitLogger()
	logger.Log.Info("Starting User Service...")

	database.InitMYSQL("root:root@tcp(127.0.0.1:3306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local")
	logger.Log.Info("MySQL initialized successfully")

	database.InitRedis("127.0.0.1:6379", "", 0)
	logger.Log.Info("Redis initialized successfully")

	mq.InitRabbitMQ()
	logger.Log.Info("RabbitMQ initialized successfully")

	user_service.StartConsumer()
	logger.Log.Info("User Consumer started, listening for messages...")

	r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"})
	if err != nil {
		logger.Log.Fatal("fail to create etcd registry: " + err.Error())
	}

	addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:8889")

	svr := userservice.NewServer(
		new(user_service.UserServiceImpl),
		server.WithServiceAddr(addr),                 
		server.WithRegistry(r),                       
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "user-service",         
		}),
	)

	logger.Log.Info("User Service RPC Server is running on 0.0.0.0:8889...")

	if err := svr.Run(); err != nil {
		logger.Log.Error("User Service exit with error: " + err.Error())
	}
}
