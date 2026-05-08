package handler

import (
	"context"
	product "full_backend_practice/kitex_gen/product"
	"full_backend_practice/kitex_gen/product/productservice"
	"full_backend_practice/pkg/response"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"go.uber.org/zap"
)

type ProductServiceHandler struct {
	Client productservice.Client
	logger *zap.Logger
}

func NewProductServiceHandler(client productservice.Client, logger *zap.Logger) *ProductServiceHandler {
	return &ProductServiceHandler{Client: client, logger: logger}
}

type CreateProductRequest struct {
	Name         string  `json:"name" vd:"len($)>0"`
	Price        float64 `json:"price" vd:"$>0"`
	SeckillPrict float64 `json:"seckill_prict" vd:"$>0"`
	Stock        int32   `json:"stock" vd:"$>=0"`
	Version      int32   `json:"version"`
	StartTime    int64   `json:"start_time" vd:"$>0"`
	EndTime      int64   `json:"end_time" vd:"$>0"`
}

func (s ProductServiceHandler) GetProductList(c context.Context, ctx *app.RequestContext) {
	rsp, err := s.Client.GetProductList(c, &product.GetProductListReq{})
	if err != nil {
		response.ServerError(ctx, err)
		return
	}
	if rsp.BaseResp == nil || rsp.BaseResp.Code != response.CodeOK {
		code := response.CodeInternal
		msg := "downstream service error"
		if rsp.BaseResp != nil {
			code = rsp.BaseResp.Code
			msg = rsp.BaseResp.Msg
		}
		response.Error(ctx, int(code), msg)
		return
	}
	response.Success(ctx, utils.H{"code": rsp.BaseResp.Code, "products": rsp.Products})
}

func (s ProductServiceHandler) GetProduct(c context.Context, ctx *app.RequestContext) {
	idStr := ctx.Param("id")
	productID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, 400, "invalid product id")
		return
	}
	rsp, err := s.Client.GetProduct(c, &product.GetProductReq{ProductId: productID})
	if err != nil {
		response.ServerError(ctx, err)
		return
	}
	if rsp.BaseResp == nil || rsp.BaseResp.Code != response.CodeOK {
		code := response.CodeInternal
		msg := "downstream service error"
		if rsp.BaseResp != nil {
			code = rsp.BaseResp.Code
			msg = rsp.BaseResp.Msg
		}
		response.Error(ctx, int(code), msg)
		return
	}
	response.Success(ctx, utils.H{"code": rsp.BaseResp.Code, "product": rsp.Product})
}

func (s *ProductServiceHandler) CreateProduct(c context.Context, ctx *app.RequestContext) {
	var req CreateProductRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		response.Error(ctx, 400, "Invalid Request Parameters: "+err.Error())
		return
	}
	rsp, err := s.Client.CreateProduct(c, &product.CreateProductReq{
		Name:         req.Name,
		Price:        req.Price,
		SeckillPrict: req.SeckillPrict,
		Stock:        req.Stock,
		Version:      req.Version,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})
	if err != nil {
		response.ServerError(ctx, err)
		return
	}
	if rsp.BaseResp == nil || rsp.BaseResp.Code != response.CodeOK {
		code := response.CodeInternal
		msg := "downstream service error"
		if rsp.BaseResp != nil {
			code = rsp.BaseResp.Code
			msg = rsp.BaseResp.Msg
		}
		response.Error(ctx, int(code), msg)
		return
	}
	response.Success(ctx, utils.H{"code": rsp.BaseResp.Code, "product": rsp.Product})
}

// 针对这个接口目前暂时不做role的区分，如果有时间的话鉴权可以放在router或者这里也行吧
func (s *ProductServiceHandler) HeatProduct(c context.Context, ctx *app.RequestContext) {
	rsp, err := s.Client.HeatProduct(c, &product.HeatProductReq{})
	if err != nil {
		response.ServerError(ctx, err)
		return
	}
	if rsp.BaseResp == nil || rsp.BaseResp.Code != response.CodeOK {
		code := response.CodeInternal
		msg := "downstream service error"
		if rsp.BaseResp != nil {
			code = rsp.BaseResp.Code
			msg = rsp.BaseResp.Msg
		}
		response.Error(ctx, int(code), msg)
		return
	}
	response.Success(ctx, utils.H{"code": rsp.BaseResp.Code})

}
