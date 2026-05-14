package constant

// ============================================================
// 默认配置常量（被 .env 覆盖）
// ============================================================

// 服务端口
const (
	DefaultAPIGatewayPort = "8081"
	DefaultUserRPCPort   = "8889"
	DefaultProductRPCPort = "8890"
	DefaultOrderRPCPort  = "8891"
)

// 服务名称（用于 etcd 注册）
const (
	ServiceNameUser    = "user_service"
	ServiceNameProduct = "product_service"
	ServiceNameOrder   = "order_service"
)

// 数据库默认配置
const (
	DefaultMySQLDSN = "root:root@tcp(127.0.0.1:13306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local"

	DefaultRedisAddr     = "127.0.0.1:6379"
	DefaultRedisPassword = ""
	DefaultRedisDB       = 0

	DefaultRabbitMQURL = "amqp://guest:guest@127.0.0.1:5672/"

	DefaultEtcdEndpoint = "127.0.0.1:2379"
)
