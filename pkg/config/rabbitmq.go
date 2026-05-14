package config

type RabbitMQConfig struct {
	URL    string   `mapstructure:"url"`
	Queues []string `mapstructure:"queues"`
}
