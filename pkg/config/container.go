package config

import "go.uber.org/dig"

// RegisterConfigProviders 注册所有配置提供者到 dig 容器 (新方式)
func RegisterConfigProviders(c *dig.Container) error {
	allCfg := LoadAllConfig()

	if err := c.Provide(func() *MySQLConfig { return &allCfg.MySQL }); err != nil {
		return err
	}
	if err := c.Provide(func() *RedisConfig { return &allCfg.Redis }); err != nil {
		return err
	}
	if err := c.Provide(func() *RabbitMQConfig { return &allCfg.RabbitMQ }); err != nil {
		return err
	}
	if err := c.Provide(func() *EtcdConfig { return &allCfg.Etcd }); err != nil {
		return err
	}
	if err := c.Provide(func() *ServerConfig { return &allCfg.Server }); err != nil {
		return err
	}
	if err := c.Provide(func() *TracerConfig { return &allCfg.Tracer }); err != nil {
		return err
	}

	return nil
}
