namespace go product
include "base.thrift"

struct ProductInfo {
    1: i64 id
    2: string name
    3: double price
    5: double seckill_prict
    6: i32 stock
    7: i32 version
    8: string start_time
    9: string end_time
}

struct GetProductListReq{
}
struct GetProductListResp{
    1: required base.BaseResp base_resp
    2: list<ProductInfo> products
}

struct GetProductReq {
    1: required i64 product_id
}

struct GetProductResp {
    1: required base.BaseResp base_resp
    2: optional ProductInfo product
}

struct HeatProductReq{
}
struct HeatProductResp{
    1: required base.BaseResp base_resp
}


service ProductService {
    GetProductListResp GetProductList(1: GetProductListReq req)
    GetProductResp GetProduct(1: GetProductReq req)
    HeatProductResp HeatProduct(1: HeatProductReq req)
}