package config

import "os"

// ============================================================
// 环境变量通用工具函数
// ============================================================

// getEnvOrDefault 从环境变量获取值，如果未设置则返回默认值
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
