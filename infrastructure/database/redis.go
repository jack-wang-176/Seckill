package database

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"full_backend_practice/pkg/config"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// var Client *redis.Client
// 这里记得去把这个context通过注入和前面路由触发的context关联起来
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
	key := fmt.Sprintf("seckill:stock:%d", productID)
	err := preHeatScript.Run(ctx, rw.Client, []string{key}, stock).Err()
	if err != nil {
		return fmt.Errorf("redis set err: %v", err)
	}
	return nil
}

func (rw *RedisWrapper) SimpleDecrStock(stockFirst []string) int {
	result, _ := rw.Client.Eval(ctx, decrStockLua, stockFirst).Int()
	return result
}
func (rw *RedisWrapper) SetPath(path, key string) error {
	return rw.Client.Set(ctx, path, key, 60*(time.Second)).Err()
}
func (rw *RedisWrapper) ConfirmPaht(path, key string) (bool, error) {
	val, err := rw.Client.Get(ctx, path).Result()
	if err != nil {
		return false, err
	}
	return val == key, nil
}

// 这里临时的设置set的持续时间段，后续应该维护一个添加对应product的开始时间和结束时间，并根据当下的时间来动态设置对应字段的持续时间
func (rw *RedisWrapper) SendSeckillCre(productID, userID uint) error {
	key := fmt.Sprintf("seckill:order_create:%d:%d", productID, userID)
	return rw.Client.Set(ctx, key, "success_create_order", time.Hour*24).Err()
}
func (rw *RedisWrapper) GetSeckillCre(productID, userID uint) (bool, error) {
	key := fmt.Sprintf("seckill:order_create:%d:%d", productID, userID)
	res, err := rw.Client.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res == "success_create_order", nil
}
