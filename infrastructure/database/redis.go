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
	// 使用调用方传入的 context 进行检查通常更灵活，但这里仍然使用 background 做初始化探测
	_, err := client.Ping(context.Background()).Result()
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

func (rw *RedisWrapper) PreHeatStock(ctx context.Context, productID uint, stock int) error {
	key := fmt.Sprintf("seckill:stock:%d", productID)
	err := preHeatScript.Run(ctx, rw.Client, []string{key}, stock).Err()
	if err != nil {
		return fmt.Errorf("redis set err: %v", err)
	}
	return nil
}

func (rw *RedisWrapper) SimpleDecrStock(ctx context.Context, stockFirst []string) int {
	result, _ := rw.Client.Eval(ctx, decrStockLua, stockFirst).Int()
	return result
}

func (rw *RedisWrapper) SetPath(ctx context.Context, key, value string) error {
	return rw.Client.Set(ctx, key, value, 60*(time.Second)).Err()
}

func (rw *RedisWrapper) ConfirmPaht(ctx context.Context, path, key string) (bool, error) {
	val, err := rw.Client.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val == path, nil
}

// 这里临时的设置set的持续时间段，后续应该维护一个添加对应product的开始时间和结束时间，并根据当下的时间来动态设置对应字段的持续时间
func (rw *RedisWrapper) SendSeckillCre(ctx context.Context, productID, userID uint) error {
	key := fmt.Sprintf("seckill:order_create:%d:%d", productID, userID)
	return rw.Client.Set(ctx, key, "success_create_order", time.Hour*24).Err()
}

func (rw *RedisWrapper) GetSeckillCre(ctx context.Context, productID, userID uint) (bool, error) {
	key := fmt.Sprintf("seckill:order_create:%d:%d", productID, userID)
	res, err := rw.Client.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return res == "success_create_order", nil
}
