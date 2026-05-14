package config

import (
	"strings"
)

// ============================================================
// etcd 配置
// ============================================================

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
}

// etcd 默认配置常量
const DefaultEtcdEndpoint = "127.0.0.1:2379"

// LoadEtcdConfig 从环境变量加载 etcd 配置
func LoadEtcdConfig() EtcdConfig {
	endpointsStr := getEnvOrDefault("ETCD_ENDPOINTS", DefaultEtcdEndpoint)

	return EtcdConfig{
		Endpoints: strings.Split(endpointsStr, ","),
	}
}
