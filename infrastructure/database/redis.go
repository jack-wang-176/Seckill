package database

import (
	"context"
	_ "embed"
	"fmt"

	"full_backend_practice/pkg/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// var Client *redis.Client
var ctx = context.Background()

var (
	preHeatScript   = redis.NewScript(`return redis.call('SET',KEYS[1],ARGV[1])`)
	decrStockScript *redis.Script
)

//go:embd scripts/decr_stock.lua
var decrStockLua string

func init() {
	decrStockScript = redis.NewScript(decrStockLua)
}

func InitRedis(cfg *config.RedisConfig, log *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: 10,
	})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		if log != nil {
			log.Error("redis init err", zap.Error(err))
		}
		return nil, err
	}
	if log != nil {
		log.Info("redis init success")
	}
	return client, nil
}

type RedisWrapper struct {
	Client *redis.Client
}

func NewRedisWrapper(cli *redis.Client) *RedisWrapper {
	return &RedisWrapper{Client: cli}
}

func (rw *RedisWrapper) PreHeatStock(productID uint, stock int) error {
	key := fmt.Sprintf("skill:stock:product:%d", productID)
	err := preHeatScript.Run(ctx, rw.Client, []string{key}, stock).Err()
	if err != nil {
		return fmt.Errorf("redis set err: %v", err)
	}
	return nil
}

func (rw *RedisWrapper) SimpleDecrStock(productID uint) (bool, error) {
	key := fmt.Sprintf("seckill:stock:%d", productID)
	result, err := decrStockScript.Run(ctx, rw.Client, []string{key}).Result()
	if err != nil {
		return false, fmt.Errorf("lua script execution error: %w", err)
	}
	if resultInt, ok := result.(int64); ok {
		if resultInt == 1 {
			return true, nil
		}
		return false, nil
	}
	return false, fmt.Errorf("unexpected running line")
}
