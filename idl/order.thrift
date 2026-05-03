namespace go order

include "base.thrift"

struct SeckillReq{
    1: required i64 user_id
    2: required i64 product_id
}
struct SeckillResp{
    1: required base.BaseResp base_resp
    2: optional string order_no
}
service OrderService{
    SeckillResp Seckill(1: SeckillReq req)
}