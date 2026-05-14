package config

import "go.uber.org/dig"

// ============================================================
// 依赖注入配置提供者
// 为 dig 容器提供各配置对象
// ============================================================

// ProvideConfigs 为 dig 容器提供所有配置 (兼容旧代码)
// 使用示例:
//
//	c := dig.New()
//	ProvideConfigs(c)
//	c.Provide(logger.InitLogger)
func ProvideConfigs(c *dig.Container) {
	mysqlCfg, redisCfg, mqCfg, etcdCfg := LoadConfigFromEnv()

	c.Provide(func() MySQLConfig { return mysqlCfg })
	c.Provide(func() RedisConfig { return redisCfg })
	c.Provide(func() RabbitMQConfig { return mqCfg })
	c.Provide(func() EtcdConfig { return etcdCfg })
}

// RegisterConfigProviders 注册所有配置提供者到 dig 容器 (新方式)
// 使用示例:
//
//	c := dig.New()
//	RegisterConfigProviders(c)
//	c.Provide(logger.InitLogger)
func RegisterConfigProviders(c *dig.Container) error {
	allCfg := LoadAllConfig()

	if err := c.Provide(func() MySQLConfig { return allCfg.MySQL }); err != nil {
		return err
	}
	if err := c.Provide(func() RedisConfig { return allCfg.Redis }); err != nil {
		return err
	}
	if err := c.Provide(func() RabbitMQConfig { return allCfg.RabbitMQ }); err != nil {
		return err
	}
	if err := c.Provide(func() EtcdConfig { return allCfg.Etcd }); err != nil {
		return err
	}

	return nil
}
