namespace go user
include "base.thrift"

struct RegisterReq{
    1: required string username
    2: required string password
}
struct RegisterResp{
    1: required base.BaseResp base_resp
    2: optional i64 user_id
}
struct LoginReq{
    1: required string username
    2: required string password
}

struct LoginResp {
    1: required base.BaseResp base_resp
    2: optional string token 
}


service UserService {
    RegisterResp Register(1: RegisterReq req)
    LoginResp Login(1: LoginReq req)
}