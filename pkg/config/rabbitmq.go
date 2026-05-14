package config

import (
	"strings"
)

// ============================================================
// RabbitMQ 配置
// ============================================================

type RabbitMQConfig struct {
	URL    string   `mapstructure:"url"`
	Queues []string `mapstructure:"queues"`
}

// RabbitMQ 默认配置常量
const DefaultRabbitMQURL = "amqp://guest:guest@127.0.0.1:5672/"

// LoadRabbitMQConfig 从环境变量加载 RabbitMQ 配置
func LoadRabbitMQConfig() RabbitMQConfig {
	url := getEnvOrDefault("RABBITMQ_URL", DefaultRabbitMQURL)
	queues := loadQueuesFromEnv()

	return RabbitMQConfig{
		URL:    url,
		Queues: queues,
	}
}

// loadQueuesFromEnv 从环境变量读取队列配置
// 格式: RABBITMQ_QUEUES=order_seckill,product_list,product_single
func loadQueuesFromEnv() []string {
	queuesStr := getEnvOrDefault("RABBITMQ_QUEUES", "")
	if queuesStr == "" {
		return []string{}
	}
	return strings.Split(queuesStr, ",")
}
