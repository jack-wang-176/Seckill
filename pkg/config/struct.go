package config

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type RabbitMQConfig struct {
	URL    string   `mapstructure:"url"`
	Queues []string `mapstructure:"queues"`
}

type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
}
