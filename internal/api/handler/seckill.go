package handler

import (
	"context"
	"full_backend_practice/internal/rpc"
	"full_backend_practice/kitex_gen/order"
	"full_backend_practice/pkg/logger"
	"strconv"

	"go.uber.org/zap"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

type SeckillReq struct {
	ProductId string `json:"product_id" vd:"$>0"`
}

func CreateOrder(c context.Context, ctx *app.RequestContext) {
	userIDVal, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(consts.StatusUnauthorized, utils.H{
			"msg": "user_id not exists",
		})
		return
	}
	userID := userIDVal.(int64)
	var req SeckillReq
	if err := ctx.BindAndValidate(&req); err != nil {
		ctx.JSON(consts.StatusBadRequest, utils.H{
			"msg": err.Error(),
		})
		return
	}
	productID, _ := strconv.ParseInt(req.ProductId, 10, 64)
	resp, err := rpc.OrderClient.Seckill(c, &order.SeckillReq{
		UserId:    userID,
		ProductId: productID,
	})
	if err != nil {
		logger.Log.Error("RPC OrderClient error", zap.Error(err))
		ctx.JSON(consts.StatusInternalServerError, utils.H{
			"msg":  err.Error(),
			"code": consts.StatusInternalServerError,
		})
		return
	}

	logger.Log.Info("网关层接收到下单请求", zap.Int64("UserID", userID), zap.String("ProductID", req.ProductId))

	ctx.JSON(consts.StatusOK, map[string]interface{}{
		"code": resp.BaseResp.Code,
		"msg":  resp.BaseResp.Msg,
		"data": map[string]interface{}{
			"order_no": resp.OrderNo,
		},
	})
}
