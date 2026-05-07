package response

import (
	"full_backend_practice/kitex_gen/base"
)

const (
	CodeOK int32 = 200
	//for product service 3xx
	//for user service 4xx
	CodeInvalidParams int32 = 400
	CodeUnauthorized  int32 = 401
	//for order service 5xx
	CodeInternal int32 = 600

	CodeRPCError int32 = 1001
)

func BuildBaseResp(code int32, msg string) *base.BaseResp {
	return &base.BaseResp{
		Code: code,
		Msg:  msg,
	}
}

func BuildOK() *base.BaseResp {
	return BuildBaseResp(CodeOK, "success")
}
