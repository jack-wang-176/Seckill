package handler

import (
	"context"
	"full_backend_practice/kitex_gen/order"
	"full_backend_practice/kitex_gen/order/orderservice"
	"full_backend_practice/pkg/response"
	"strconv"

	"go.uber.org/zap"

	"github.com/cloudwego/hertz/pkg/app"
)

type OrderHandler struct {
	Client orderservice.Client
	logger *zap.Logger
}

func NewOrderHandler(client orderservice.Client, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{Client: client, logger: logger}
}

type SeckillReq struct {
	ProductId string `json:"product_id" vd:"$>0"`
}

func (h *OrderHandler) CreateOrder(c context.Context, ctx *app.RequestContext) {
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
	resp, err := h.Client.Seckill(c, &order.SeckillReq{
		UserId:    userID,
		ProductId: productID,
	})
	if err != nil {
		if h.logger != nil {
			h.logger.Error("RPC OrderClient error", zap.Error(err))
		}
		response.ServerError(ctx, err)
		return
	}

	if h.logger != nil {
		h.logger.Info("网关层接收到下单请求", zap.Int64("UserID", userID), zap.String("ProductID", req.ProductId))
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

	response.Success(ctx, map[string]interface{}{
		"order_no": resp.OrderNo,
	})
}
