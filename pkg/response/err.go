package response

import (
	"full_backend_practice/kitex_gen/base"
	"full_backend_practice/pkg/constant"
)

// ai改的 //提取变量的时候留下来这个问题
const (
	CodeOK            = constant.CodeOK
	CodeInvalidParams = constant.CodeInvalidParams
	CodeUnauthorized  = constant.CodeUnauthorized
	CodeInternal      = constant.CodeInternal
	CodeRPCError      = constant.CodeRPCError
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
