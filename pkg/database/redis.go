package database

import (
	"context"
	"fmt"

	"full_backend_practice/pkg/config"
	"full_backend_practice/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// var Client *redis.Client
var ctx = context.Background()

func InitRedis(cfg *config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 10,
	})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Log.Fatal("redis init err", zap.Error(err))
	}
	logger.Log.Info("redis init success")
	return client
}
type RedisWrapper struct{
	Client *redis.Client
}
func NewRedisWrapper(cli *redis.Client)*RedisWrapper{
	return &RedisWrapper{Client: cli}
}

func (rw *RedisWrapper) PreHeatStock(productID uint, stock int) error {
	key := fmt.Sprintf("skill:stock:product:%d", productID)
	err := rw.Client.Set(ctx, key, stock, 0).Err()
	if err != nil {
		return fmt.Errorf("redis set err: %v", err)
	}
	return nil
}

func (rw *RedisWrapper) SimpleDecrStock(productID uint) (bool, error) {
	key := fmt.Sprintf("seckill:stock:%d", productID)
	remain, err := rw.Client.Decr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if remain < 0 {
		return false, nil
	}
	return true, nil
}
