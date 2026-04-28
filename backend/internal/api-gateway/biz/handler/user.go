package handler

import (
	"context"
	"full_backend_practice/backend/internal/app/rpc"
	"full_backend_practice/kitex_gen/user"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type UserRequest struct {
	Username string `json:"username" vd:"len($)>0"`
	Password string `json:"password" vd:"len($)>0"`
}

func Register(c context.Context, ctx *app.RequestContext) {
	var req UserRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(consts.StatusBadRequest, utils.H{
			"msg": err.Error(),
		})
		return
	}
	rpcReq := &user.RegisterReq{
		Username: req.Username,
		Password: req.Password,
	}
	resp, err := rpc.UserClient.Register(c, rpcReq)
	if err != nil {
		ctx.JSON(consts.StatusInternalServerError, utils.H{
			"msg": err.Error(),
		})
		return
	}
	ctx.JSON(consts.StatusOK, utils.H{
		"code": resp.BaseResp.Code,
		"msg":  resp.BaseResp.Msg,
	})
}
