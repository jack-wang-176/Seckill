package database

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var ctx = context.Background()

func InitRedis(addr string, password string, db int) {
	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 10,
	})
	_, err := Client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("redis init err: %v", err)
	}
	fmt.Println("redis init success")
}
func PreHeatStock(productID uint, stock int) error {
	key := fmt.Sprintf("skill:stock:product:%d", productID)
	err := Client.Set(ctx, key, stock, 0).Err()
	if err != nil {
		return fmt.Errorf("redis set err: %v", err)
	}
	return nil
}
func SimpleDecrStock(productID uint) (bool, error) {
	key := fmt.Sprintf("seckill:stock:%d", productID)
	remain, err := Client.Decr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if remain < 0 {
		return false, nil
	}
	return true, nil
}
