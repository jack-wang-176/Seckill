package config

type MySQLConfig struct {
	DSN string
}
type EtcdConfig struct {
	Endpoints []string
}
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}
type RabbitMQConfig struct {
	URL    string
	Queues []string
}
