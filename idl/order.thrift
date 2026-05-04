namespace go order

include "base.thrift"


struct GetSeckillPath{
    1: required i64 user_id
    2: required i64 product_id
}

struct GetSeckillPathResp{
    1: required base.BaseResp base_resp
    2: required string path
}

struct SeckillReq{
    1: required i64 user_id
    2: required i64 product_id
    3: required string path
}
struct SeckillResp{
    1: required base.BaseResp base_resp
}
struct GetSeckillResultReq{
    1: required i64 user_id
    2: required i64 product_id
}
struct GetSeckillResultResp{
    1: required base.BaseResp base_resp
    2: i8 status
    3: string order_no
}

service OrderService{
    SeckillResp Seckill(1: SeckillReq req)
    GetSeckillPathResp GetSeckillPath(1: GetSeckillPath req)
    GetSeckillResultResp GetSeckillResult(1: GetSeckillResultReq req)
}