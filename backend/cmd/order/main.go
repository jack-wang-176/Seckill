package main

import (
	"net"

	"go.uber.org/zap"

	order_service "full_backend_practice/backend/internal/app/order-service"
	"full_backend_practice/kitex_gen/order/orderservice"
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/logger"
	"full_backend_practice/pkg/mq"

	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	// 0. 初始化高性能日志组件
	logger.InitLogger()
	logger.Log.Info("Starting Order Service...")

	// 1. 初始化数据库
	// 根据 docker-compose.yml 的配置，数据库为 seckill_db，密码为 root
	database.InitMYSQL("root:root@tcp(127.0.0.1:3306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local")
	logger.Log.Info("MySQL initialized successfully")

	// 2. 初始化 Redis
	database.InitRedis("127.0.0.1:6379", "", 0)
	logger.Log.Info("Redis initialized successfully")

	etcdEndPoints := []string{"127.0.0.1:2379"}
	mqUrl, _ := mq.GetConfigFromEtcd(etcdEndPoints, "/config/rbbitmq/order/url")
	mq.InitRabbitMQ(mqUrl, []string{"order_seckill"})
	logger.Log.Info("RabbitMQ initialized successfully")

	// 4. 启动 MQ 消费者
	order_service.StartConsumer()
	logger.Log.Info("Order Consumer started, listening for messages...")

	// 5. 初始化 ETCD 服务注册
	r, err := etcd.NewEtcdRegistry([]string{"127.0.0.1:2379"})
	if err != nil {
		logger.Log.Fatal("fail to create etcd registry", zap.Error(err))
	}

	addr, _ := net.ResolveTCPAddr("tcp", "0.0.0.0:8888")

	// 6. 初始化 Kitex Server 服务，注意注入注册中心和服务的信息
	svr := orderservice.NewServer(
		new(order_service.OrderServiceImpl),
		server.WithServiceAddr(addr), // 指定监听端口
		server.WithRegistry(r),       // ETCD 注册中心
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
			ServiceName: "order-service", // 和 Gateway API 调用的服务名一致
		}),
	)

	logger.Log.Info("Order Service RPC Server is running on 0.0.0.0:8888...")

	if err := svr.Run(); err != nil {
		logger.Log.Error("Order Service exit with error", zap.Error(err))
	}
}
