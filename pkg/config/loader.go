package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// ============================================================
// 统一配置加载入口
// 从各自的配置模块加载所有配置
// ============================================================

// AllConfig 包含所有应用配置
type AllConfig struct {
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	Etcd     EtcdConfig     `mapstructure:"etcd"`
	Server   ServerConfig   `mapstructure:"server"`
}

type ServerConfig struct {
	ApiGatewayPort string `mapstructure:"api_gateway_port"`
	UserRpcPort    string `mapstructure:"user_rpc_port"`
	ProductRpcPort string `mapstructure:"product_rpc_port"`
	OrderRpcPort   string `mapstructure:"order_rpc_port"`
}

var AppConfig AllConfig

// LoadAllConfig 使用 Viper 从 yaml 文件及环境变量加载所有配置
// 推荐使用此方法处理所有配置
func LoadAllConfig() AllConfig {
	viper.SetConfigName("config") // 读取 config.yaml
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")     // 当前工作目录下的 config 目录
	viper.AddConfigPath("../config/")    // 支持上一级目录，方便在 cmd 目录下运行
	viper.AddConfigPath("../../config/") // 支持更深层级目录运行测试

	// 支持环境变量读取，如果存在环境变量的话，viper 会自动读取并覆盖 yaml 的配置
	viper.AutomaticEnv()
	// 使用 _ 分隔，例如 MYSQL_DSN
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Config file not found, will rely on ENV variables or defaults. Error: %v", err)
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	return AppConfig
}
