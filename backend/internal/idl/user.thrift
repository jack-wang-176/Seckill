namespace go user
include "base.thrift"

struct UserReq {
    1: required string username
    2: required string password
}

struct LoginResp {
    1: required base.BaseResp base_resp
    2: optional string token 
}
stuct RegisterResp{
    1: required bool success
    2: optional string token
}

service UserService {
    LoginResp Login(1: LoginReq req)
}