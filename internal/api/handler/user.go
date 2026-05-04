package handler

import (
	"context"
	"full_backend_practice/kitex_gen/user"
	"full_backend_practice/kitex_gen/user/userservice"
	"full_backend_practice/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"go.uber.org/zap"
)

type UserHandler struct {
	Client userservice.Client
	Log    *zap.Logger
}

func NewUserHandler(client userservice.Client, log *zap.Logger) *UserHandler {
	return &UserHandler{Client: client, Log: log}
}

type UserRequest struct {
	Username string `json:"username" vd:"len($)>0"`
	Password string `json:"password" vd:"len($)>0"`
}

func (h *UserHandler) Register(c context.Context, ctx *app.RequestContext) {
	var req UserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 400, "Invalid Request Parameters"+err.Error())
		return
	}
	rpcReq := &user.RegisterReq{
		Username: req.Username,
		Password: req.Password,
	}
	resp, err := h.Client.Register(c, rpcReq)
	if err != nil {
		response.ServerError(ctx, err)
		return
	}

	if resp.BaseResp == nil || resp.BaseResp.Code != response.CodeOK {
		code := response.CodeInternal
		var msg string
		if resp.BaseResp != nil {
			code = resp.BaseResp.Code
			msg = resp.BaseResp.Msg
		}
		if msg == "" {
			msg = "downstream service error"
		}
		response.Error(ctx, int(code), msg)
		return
	}

	response.Success(ctx, utils.H{
		"code": resp.BaseResp.Code,
	})
}

func (h *UserHandler) Login(c context.Context, ctx *app.RequestContext) {
	var req UserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 400, "Invalid Request Parameters: "+err.Error())
		return
	}
	rpcReq := &user.LoginReq{
		Username: req.Username,
		Password: req.Password,
	}
	resp, err := h.Client.Login(c, rpcReq)
	if err != nil {
		response.ServerError(ctx, err)
		return
	}

	if resp.BaseResp == nil || resp.BaseResp.Code != response.CodeOK {
		code := response.CodeInternal
		var msg string
		if resp.BaseResp != nil {
			code = resp.BaseResp.Code
			msg = resp.BaseResp.Msg
		}
		if msg == "" {
			msg = "downstream service error"
		}
		response.Error(ctx, int(code), msg)
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
