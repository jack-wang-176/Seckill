package constant

// ============================================================
// 订单相关常量
// ============================================================

// OrderNoPrefix 订单号前缀
const OrderNoPrefix = "SN"

// OrderNoFormat 订单号生成格式（fmt.Sprintf）
// 参数: unixNano, userID
const OrderNoFormat = "SN%d%d"

// OrderStatus 订单状态
const (
	OrderStatusPending   int8 = 0 // 待处理
	OrderStatusSuccess   int8 = 1 // 秒杀成功
	OrderStatusFailed    int8 = 2 // 秒杀失败
	OrderStatusCancelled int8 = 3 // 已取消
)

// 队列名称
const (
	QueueOrderSeckill  = "order_seckill"
	QueueProductList   = "product_list"
	QueueProductSingle = "product_single"
)
