namespace go product
include "base.thrift"

struct ProductInfo {
    1: i64 id
    2: string name
    3: double price
    4: i32 stock
}

struct GetProductReq {
    1: required i64 product_id
}

struct GetProductResp {
    1: required base.BaseResp base_resp
    2: optional ProductInfo product
}

service ProductService {
    GetProductResp GetProduct(1: GetProductReq req)
}