package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
		DB:   0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Printf("Redis not available: %v\n", err)
		os.Exit(1)
	}

	// 清理测试数据
	rdb.Del(ctx, "test:stock:1", "test:stock:1soldout", "test:order:1:1")

	// 场景1: 库存不存在（未预热）
	script := redis.NewScript(`
local stockKey = KEYS[1]
local soldoutKey = KEYS[2]
local orderKey = KEYS[3]

if redis.call("EXISTS", soldoutKey) == 1 then
    return -1
end

local stock = tonumber(redis.call("GET", stockKey))
if stock == nil then
    return -2
end

if stock <= 0 then
    redis.call("SET", soldoutKey, 1)
    return -1
end

if redis.call("EXISTS", orderKey) == 1 then
    return -4
end

redis.call("DECR", stockKey)
if redis.call("GET", stockKey) == "0" then
    redis.call("SET", soldoutKey, 1)
end

return 1
`)

	// 场景1: 未预热
	r1, _ := script.Run(ctx, rdb, []string{"test:stock:1", "test:stock:1soldout", "test:order:1:1"}).Int()
	fmt.Printf("场景1 (未预热): %d (期望 -2)\n", r1)

	// 场景2: 库存正常，首次下单
	rdb.Set(ctx, "test:stock:1", 10, 0)
	r2, _ := script.Run(ctx, rdb, []string{"test:stock:1", "test:stock:1soldout", "test:order:1:1"}).Int()
	stockAfterDecr, _ := rdb.Get(ctx, "test:stock:1").Int()
	fmt.Printf("场景2 (首次下单成功): %d (期望 1), 剩余库存: %d (期望 9)\n", r2, stockAfterDecr)

	// 场景3: 重复下单
	r3, _ := script.Run(ctx, rdb, []string{"test:stock:1", "test:stock:1soldout", "test:order:2:2"}).Int()
	// 这个场景没有 orderKey 时正常下单，再查 orderKey
	rdb.Set(ctx, "test:order:1:1", "SN123", 0)
	r3, _ = script.Run(ctx, rdb, []string{"test:stock:1", "test:stock:1soldout", "test:order:1:1"}).Int()
	// 应该返回 -4
	fmt.Printf("场景3 (重复下单): %d (期望 -4)\n", r3)

	// 场景4: 售罄（最后1件）
	rdb.Del(ctx, "test:stock:1soldout")
	rdb.Set(ctx, "test:stock:1", 1, 0)
	rdb.Del(ctx, "test:order:3:3")
	r4, _ := script.Run(ctx, rdb, []string{"test:stock:1", "test:stock:1soldout", "test:order:3:3"}).Int()
	isSoldout, _ := rdb.Exists(ctx, "test:stock:1soldout").Result()
	fmt.Printf("场景4 (最后1件): %d (期望 1), 售罄标记: %d (期望 1)\n", r4, isSoldout)

	// 场景5: 售罄后再次请求
	r5, _ := script.Run(ctx, rdb, []string{"test:stock:1", "test:stock:1soldout", "test:order:4:4"}).Int()
	fmt.Printf("场景5 (售罄后): %d (期望 -1)\n", r5)

	// 清理
	rdb.Del(ctx, "test:stock:1", "test:stock:1soldout", "test:order:1:1", "test:order:2:2", "test:order:3:3", "test:order:4:4")
	fmt.Println("\n所有场景测试完成 ✅")
}
