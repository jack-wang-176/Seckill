package config

// ============================================================
// MySQL 配置
// ============================================================

type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

// DefaultMySQLDSN MySQL 默认连接字符串
const DefaultMySQLDSN = "root:root@tcp(127.0.0.1:13306)/seckill_db?charset=utf8mb4&parseTime=True&loc=Local"

// LoadMySQLConfig 从环境变量加载 MySQL 配置
func LoadMySQLConfig() MySQLConfig {
	return MySQLConfig{
		DSN: getEnvOrDefault("MYSQL_DSN", DefaultMySQLDSN),
	}
}
