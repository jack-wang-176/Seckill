package product

import (
	"context"
	"encoding/json"
	"fmt"
	"full_backend_practice/infrastructure/database"
	"full_backend_practice/infrastructure/mq"
	product "full_backend_practice/kitex_gen/product"
	"full_backend_practice/pkg/constant"
	"full_backend_practice/pkg/response"

	"full_backend_practice/infrastructure/tacer"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// ProductServiceImpl implements the last service interface defined in the IDL.
type productServiceImpl struct {
	MySqlWrapper ProductDatabase
	RedisWrapper *database.RedisWrapper
	MQ           *mq.RabbitClient
	Logger       *zap.Logger
}

func NewProductServiceImpl(mr ProductDatabase, rw *database.RedisWrapper, mq *mq.RabbitClient, logger *zap.Logger) ProductServiceImpl {
	return &productServiceImpl{
		MySqlWrapper: mr,
		RedisWrapper: rw,
		MQ:           mq,
		Logger:       logger,
	}
}

// GetProductList implements the ProductServiceImpl interface.
func (s *productServiceImpl) GetProductList(ctx context.Context, req *product.GetProductListReq) (resp *product.GetProductListResp, err error) {
	resp = new(product.GetProductListResp)
	// 先尝试从 Redis 缓存读取
	cacheKey := constant.ProductListKey
	if s.RedisWrapper != nil {
		if val, e := s.RedisWrapper.Client.Get(ctx, cacheKey).Result(); e == nil {
			var products []Product
			if jerr := json.Unmarshal([]byte(val), &products); jerr == nil {
				resp.Products = make([]*product.ProductInfo, 0, len(products))
				for _, p := range products {
					resp.Products = append(resp.Products, &product.ProductInfo{
						Id:           int64(p.ID),
						Name:         p.Name,
						Price:        p.Price,
						SeckillPrict: p.SeckillPrice,
						Stock:        int32(p.Stock),
						Version:      int32(p.Version),
						StartTime:    fmt.Sprintf("%d", p.StartTime),
						EndTime:      fmt.Sprintf("%d", p.EndTime),
					})
				}
				resp.BaseResp = response.BuildBaseResp(response.CodeOK, "success")
				return resp, nil
			}
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		s.Logger.Error("Failed to marshal GetProductListReq", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "json marshal error")
		return resp, err
	}
	headers := amqp091.Table{}
	tacer.InjectAMQPHeaders(ctx, headers)
	err = s.MQ.Channel.PublishWithContext(ctx, "", constant.QueueProductList, false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Headers:      headers,
		Body:         body,
	})

	if err != nil {
		s.Logger.Error("RabbitMQ publish error", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to get product list")
		return resp, nil
	}

	// 缓存未命中：通过 MQ 异步触发查询，返回处理中状态，消费者会把结果写入缓存
	resp.BaseResp = response.BuildBaseResp(response.CodeRPCError, "processing")
	return resp, nil
}

// GetProduct implements the ProductServiceImpl interface.
func (s *productServiceImpl) GetProduct(ctx context.Context, req *product.GetProductReq) (resp *product.GetProductResp, err error) {
	resp = new(product.GetProductResp)
	// 先尝试从缓存读取单条
	cacheKey := fmt.Sprintf(constant.ProductKeyFormat, req.ProductId)
	if s.RedisWrapper != nil {
		if val, e := s.RedisWrapper.Client.Get(ctx, cacheKey).Result(); e == nil {
			var p Product
			if jerr := json.Unmarshal([]byte(val), &p); jerr == nil {
				resp.Product = &product.ProductInfo{
					Id:           int64(p.ID),
					Name:         p.Name,
					Price:        p.Price,
					SeckillPrict: p.SeckillPrice,
					Stock:        int32(p.Stock),
					Version:      int32(p.Version),
					StartTime:    fmt.Sprintf("%d", p.StartTime),
					EndTime:      fmt.Sprintf("%d", p.EndTime),
				}
				resp.BaseResp = response.BuildBaseResp(response.CodeOK, "success")
				return resp, nil
			}
		}
	}

	body, err := json.Marshal(req)
	if err != nil {
		s.Logger.Error("Failed to marshal GetProductReq", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "json marshal error")
		return resp, err
	}
	headers := amqp091.Table{}
	tacer.InjectAMQPHeaders(ctx, headers)
	err = s.MQ.Channel.PublishWithContext(ctx, "", constant.QueueProductSingle, false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Headers:      headers,
		Body:         body,
	})

	if err != nil {
		s.Logger.Error("RabbitMQ publish error", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to get product")
		return resp, nil
	}

	// 缓存未命中，异步触发消费者查询并写缓存，返回处理中
	resp.BaseResp = response.BuildBaseResp(response.CodeRPCError, "processing")
	return resp, nil
}

// CreateProduct 直接写入 MySQL，并在成功后清理相关缓存
func (s *productServiceImpl) CreateProduct(ctx context.Context, req *product.CreateProductReq) (resp *product.CreateProductResp, err error) {
	resp = new(product.CreateProductResp)
	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "success")

	if req == nil || req.Name == "" {
		resp.BaseResp = response.BuildBaseResp(response.CodeInvalidParams, "name is required")
		return resp, nil
	}
	if req.StartTime <= 0 || req.EndTime <= 0 || req.StartTime >= req.EndTime {
		resp.BaseResp = response.BuildBaseResp(response.CodeInvalidParams, "invalid activity time")
		return resp, nil
	}

	prod := &Product{
		Name:         req.Name,
		Price:        req.Price,
		SeckillPrice: req.SeckillPrict,
		Stock:        int(req.Stock),
		Version:      int(req.Version),
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	}

	if err := s.MySqlWrapper.CreateProduct(ctx, prod); err != nil {
		s.Logger.Error("Failed to create product", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to create product")
		return resp, nil
	}

	//这里还是设计的问题，这两个路由应该穿起来调用
	if s.RedisWrapper != nil {
		_, _ = s.RedisWrapper.Client.Del(ctx, constant.ProductListKey, fmt.Sprintf(constant.ProductKeyFormat, prod.ID)).Result()
	}

	resp.Product = &product.ProductInfo{
		Id:           int64(prod.ID),
		Name:         prod.Name,
		Price:        prod.Price,
		SeckillPrict: prod.SeckillPrice,
		Stock:        int32(prod.Stock),
		Version:      int32(prod.Version),
		StartTime:    fmt.Sprintf("%d", prod.StartTime),
		EndTime:      fmt.Sprintf("%d", prod.EndTime),
	}
	return resp, nil
}

// HeatProduct implements the ProductServiceImpl interface.
func (s *productServiceImpl) HeatProduct(ctx context.Context, req *product.HeatProductReq) (resp *product.HeatProductResp, err error) {
	resp = new(product.HeatProductResp)
	s.Logger.Info("begin to heat product")
	msg := mq.ProductMessage{}
	products, err := s.MySqlWrapper.GetProductList(ctx, msg)
	if err != nil {
		s.Logger.Error("Failed to load products for preheat", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to heat product")
		return resp, nil
	}
	for _, p := range products {
		if err := s.RedisWrapper.PreHeatStock(ctx, uint(p.ID), p.Stock); err != nil {
			s.Logger.Error("Failed to pre heat stock", zap.Uint("product_id", p.ID), zap.Error(err))
			resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to heat product")
			return resp, nil
		}
	}

	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "success")
	return
}
