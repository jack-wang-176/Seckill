package database

import (
	"context"
	_ "embed"
	"fmt"

	"full_backend_practice/pkg/config"
	"full_backend_practice/pkg/constant"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	preHeatScript   = redis.NewScript(`return redis.call('SET',KEYS[1],ARGV[1])`)
	decrStockScript *redis.Script
)

//go:embed scripts/decr_stock.lua
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
	key := fmt.Sprintf(constant.SeckillStockKeyFormat, productID)
	err := preHeatScript.Run(ctx, rw.Client, []string{key}, stock).Err()
	if err != nil {
		return fmt.Errorf("redis set err: %v", err)
	}
	return nil
}

func (rw *RedisWrapper) SimpleDecrStock(ctx context.Context, stockFirst []string, args ...string) int {
	result, _ := rw.Client.Eval(ctx, decrStockLua, stockFirst, args).Int()
	return result
}

func (rw *RedisWrapper) SetPath(ctx context.Context, key, value string) error {
	return rw.Client.Set(ctx, key, value, constant.SeckillPathTTL).Err()
}

func (rw *RedisWrapper) ConfirmPaht(ctx context.Context, path, key string) (bool, error) {
	val, err := rw.Client.Get(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val == path, nil
}

func (rw *RedisWrapper) SendSeckillCre(ctx context.Context, productID, userID uint, orderNo string) error {
	key := fmt.Sprintf(constant.SeckillOrderKeyFormat, productID, userID)
	return rw.Client.Set(ctx, key, orderNo, constant.SeckillOrderConfirmTTL).Err()
}

func (rw *RedisWrapper) GetSeckillCre(ctx context.Context, productID, userID uint) (string, error) {
	key := fmt.Sprintf(constant.SeckillOrderKeyFormat, productID, userID)
	return rw.Client.Get(ctx, key).Result()
}
func (rw *RedisWrapper) SetOrderSuccess(ctx context.Context, key string) error {
	return rw.Client.Set(ctx, key, "", constant.SeckillOrderConfirmTTL).Err()
}
