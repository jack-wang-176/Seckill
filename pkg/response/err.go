package response

import (
	"full_backend_practice/kitex_gen/base"
)

const (
	CodeOK            int32 = 200
	CodeInvalidParams int32 = 400
	CodeUnauthorized  int32 = 401
	CodeInternal      int32 = 500
	// CodeRPCError can be used when gateway detects downstream microservice error
	CodeRPCError int32 = 1001
)

func BuildBaseResp(code int32, msg string) *base.BaseResp {
	return &base.BaseResp{
		Code: code,
		Msg:  msg,
	}
}

// Helpers for common responses
func BuildOK() *base.BaseResp {
	return BuildBaseResp(CodeOK, "success")
}
