package handler

import (
	"context"
	"full_backend_practice/internal/rpc"
	"full_backend_practice/kitex_gen/order"
	"full_backend_practice/pkg/logger"
	"full_backend_practice/pkg/response"
	"strconv"

	"go.uber.org/zap"

	"github.com/cloudwego/hertz/pkg/app"
)

type SeckillReq struct {
	ProductId string `json:"product_id" vd:"$>0"`
}

func CreateOrder(c context.Context, ctx *app.RequestContext) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		response.Error(ctx, 401, "user_id not exists")
		return
	}
	userID := userIDVal.(int64)
	var req SeckillReq
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 400, err.Error())
		return
	}
	productID, _ := strconv.ParseInt(req.ProductId, 10, 64)
	resp, err := rpc.OrderClient.Seckill(c, &order.SeckillReq{
		UserId:    userID,
		ProductId: productID,
	})
	if err != nil {
		if lg := logger.GetLogger(); lg != nil {
			lg.Error("RPC OrderClient error", zap.Error(err))
		}
		response.ServerError(ctx, err)
		return
	}

	if lg := logger.GetLogger(); lg != nil {
		lg.Info("网关层接收到下单请求", zap.Int64("UserID", userID), zap.String("ProductID", req.ProductId))
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

	response.Success(ctx, map[string]interface{}{
		"order_no": resp.OrderNo,
	})
}
