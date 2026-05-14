package config

type EtcdConfig struct {
	Endpoints []string `mapstructure:"endpoints"`
}
