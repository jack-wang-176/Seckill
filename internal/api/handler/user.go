package handler

import (
	"context"
	"full_backend_practice/internal/rpc"
	"full_backend_practice/kitex_gen/user"
	"full_backend_practice/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

type UserRequest struct {
	Username string `json:"username" vd:"len($)>0"`
	Password string `json:"password" vd:"len($)>0"`
}

func Register(c context.Context, ctx *app.RequestContext) {
	var req UserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 400, "Invalid Request Parameters"+err.Error())
		return
	}
	rpcReq := &user.RegisterReq{
		Username: req.Username,
		Password: req.Password,
	}
	resp, err := rpc.UserClient.Register(c, rpcReq)
	if err != nil {
		response.ServerError(ctx, err)
		return
	}

	if resp.BaseResp == nil || resp.BaseResp.Code != response.CodeOK {
		var msg string
		if resp.BaseResp != nil {
			msg = resp.BaseResp.Msg
		} else {
			msg = "downstream service error"
		}
		response.Error(ctx, int(resp.BaseResp.Code), msg)
		return
	}

	response.Success(ctx, utils.H{
		"code": resp.BaseResp.Code,
	})
}
func Login(c context.Context, ctx *app.RequestContext) {
	var req UserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 400, "Invalid Request Parameters: "+err.Error())
		return
	}
	rpcReq := &user.LoginReq{
		Username: req.Username,
		Password: req.Password,
	}
	resp, err := rpc.UserClient.Login(c, rpcReq)
	if err != nil {
		response.ServerError(ctx, err)
		return
	}

	if resp.BaseResp == nil || resp.BaseResp.Code != response.CodeOK {
		var msg string
		if resp.BaseResp != nil {
			msg = resp.BaseResp.Msg
		} else {
			msg = "downstream service error"
		}
		response.Error(ctx, int(resp.BaseResp.Code), msg)
		return
	}

	var tokenStr string
	if resp.Token != nil {
		tokenStr = *resp.Token
	}

	response.Success(ctx, utils.H{
		"code":  resp.BaseResp.Code,
		"msg":   resp.BaseResp.Msg,
		"token": tokenStr,
	})
}
