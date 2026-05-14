package constant

import "time"

// ============================================================
// Redis Key 命名空间 & 格式常量
// ============================================================

// --- 商品缓存 ---
const (
	// ProductListKey Redis key for product list cache
	ProductListKey = "product:list"

	// ProductKeyFormat Redis key format for single product (with productID)
	// Use fmt.Sprintf(ProductKeyFormat, productID) to build
	ProductKeyFormat = "product:%d"

	// ProductCacheTTL 商品列表/详情缓存 TTL
	ProductCacheTTL = 5 * time.Minute
)

// --- 秒杀核心 ---
const (
	// SeckillStockKeyFormat 秒杀库存 key 格式: seckill:stock:{productID}
	SeckillStockKeyFormat = "seckill:stock:%d"
	// SeckillSoldoutKeyFormat 秒杀售罄标记 key 格式: seckill:stock:{productID}soldout
	SeckillSoldoutKeyFormat = "seckill:stock:%dsoldout"
	// SeckillOrderKeyFormat 秒杀订单预占位 key 格式: seckill:order_create:{productID}:{userID}
	SeckillOrderKeyFormat = "seckill:order_create:%d:%d"
	// SeckillPathKeyFormat 秒杀路径校验 key 格式: seckill:path:{userID}:{productID}
	SeckillPathKeyFormat = "seckill:path:%d:%d"
)

// --- 秒杀相关 TTL ---
const (
	// SeckillPathTTL 秒杀路径有效期（用户拿到 path 后必须在 TTL 内下单）
	SeckillPathTTL = 60 * time.Second

	// SeckillOrderPreholdTTL 秒杀订单预占位 TTL（Lua 脚本中设置，给 consumer 的缓冲时间）
	// 预占位时间: consumer 成功后会覆盖为 24h；consumer 失败则 TTL 过期自动释放
	SeckillOrderPreholdTTL = 30 * time.Second

	// SeckillOrderConfirmTTL 秒杀订单确认后 TTL（consumer 写入后的持久化时间）
	SeckillOrderConfirmTTL = 24 * time.Hour
)

// --- Lua 脚本返回码 ---
const (
	SeckillLuaSuccess        = 1  // 成功（库存扣减 + orderNo 预占位）
	SeckillLuaSoldOut        = -1 // 售罄
	SeckillLuaNoStock        = -2 // 库存不存在（未预热）
	SeckillLuaStockAnomaly   = -3 // 库存异常（<=0 但未标记售罄）
	SeckillLuaAlreadyOrdered = -4 // 该用户已下单（防重复秒杀）
)
