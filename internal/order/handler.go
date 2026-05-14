package order

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"full_backend_practice/infrastructure/database"
	"full_backend_practice/kitex_gen/base"
	order "full_backend_practice/kitex_gen/order"

	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/pkg/constant"
	"full_backend_practice/pkg/response"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type orderServiceImpl struct {
	MySqlWrapper OrderDatabase
	RedisWrapper *database.RedisWrapper
	MQ           *mq.RabbitClient
	Logger       *zap.Logger
}

func NewOrderServiceImpl(mr OrderDatabase, rw *database.RedisWrapper, mq *mq.RabbitClient, logger *zap.Logger) OrderServiceImpl {
	return &orderServiceImpl{
		MySqlWrapper: mr,
		RedisWrapper: rw,
		MQ:           mq,
		Logger:       logger,
	}
}
func (s *orderServiceImpl) Seckill(ctx context.Context, req *order.SeckillReq) (resp *order.SeckillResp, err error) {
	resp = new(order.SeckillResp)
	resp.BaseResp = &base.BaseResp{}

	//添加校验path
	isValid, err := s.RedisWrapper.ConfirmPaht(ctx, req.Path, fmt.Sprintf(constant.SeckillPathKeyFormat, req.UserId, req.ProductId))
	if err != nil {
		s.Logger.Error("Redis confirm path error", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to confirm path")
		return resp, nil
	}
	if !isValid {
		s.Logger.Warn("Invalid seckill path", zap.Int64("UserID", req.UserId), zap.Int64("ProductID", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "invalid path")
		return resp, nil
	}

	soldOutKey := fmt.Sprintf(constant.SeckillSoldoutKeyFormat, req.ProductId)
	//这里和预热中保持一致
	stockKey := fmt.Sprintf(constant.SeckillStockKeyFormat, req.ProductId)
	successKey := fmt.Sprintf(constant.SeckillOrderKeyFormat, req.ProductId, req.UserId)

	// 先生成 orderNo，然后传给 Lua 脚本进行 Redis 预占位
	// 这样 GetSeckillResult_ 可以立即从 Redis 读到同一个 orderNo，无需等待 consumer
	orderNod := fmt.Sprintf(constant.OrderNoFormat, time.Now().UnixNano(), req.UserId)

	result := s.RedisWrapper.SimpleDecrStock(ctx, []string{stockKey, soldOutKey, successKey}, orderNod)
	if result == constant.SeckillLuaSoldOut {
		s.Logger.Warn("Stock sold out", zap.Int64("ProductID", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "sell out")
		return resp, nil
	} else if result == constant.SeckillLuaNoStock {
		s.Logger.Error("Stock Decr inject error,nees further investigation", zap.Int64("ProductID", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to seckill order")
		return resp, nil
	} else if result == constant.SeckillLuaStockAnomaly {
		s.Logger.Error("Stock anomaly (<=0 without soldout flag)", zap.Int64("ProductID", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "stock anomaly")
		return resp, nil
	} else if result == constant.SeckillLuaAlreadyOrdered {
		s.Logger.Warn("already have order", zap.Int64("UserID", req.UserId), zap.Int64("ProductID", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "already have order")
		return resp, nil
	}
	msgStuct := mq.SeckillMessage{
		UserID:    uint64(req.UserId),
		ProductID: uint64(req.ProductId),
		OrderNo:   orderNod,
	}
	body, err := json.Marshal(msgStuct)
	if err != nil {
		s.Logger.Error("JSON marshal error", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "JSON marshal error")
		return resp, nil
	}

	err = s.MQ.Channel.PublishWithContext(ctx, "", constant.QueueOrderSeckill, false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})

	if err != nil {
		s.Logger.Error("RabbitMQ publish error", zap.Error(err))
		_ = s.RedisWrapper.Client.Incr(ctx, fmt.Sprintf(constant.SeckillStockKeyFormat, req.ProductId))
		// MQ 发布失败，删除预占位的 orderKey，让用户可重试
		_ = s.RedisWrapper.Client.Del(ctx, successKey)
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to seckill order")
		return resp, nil
	}

	s.Logger.Info("Order pushed to MQ successfully", zap.String("order_no", orderNod))

	resp.OrderNo = orderNod
	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "seckill success")
	return resp, nil
}
func (s *orderServiceImpl) GetSeckillPath(ctx context.Context, req *order.GetSeckillPathReq) (resp *order.GetSeckillPathResp, err error) {
	resp = new(order.GetSeckillPathResp)
	resp.BaseResp = &base.BaseResp{}
	salt := fmt.Sprintf("May@)@#)(*&^$#@!%s", time.Now().String())
	path := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d:%d:%s", req.UserId, req.ProductId, salt))))
	//交给前端。生成对应路径，存入redis，后续校验路径和之前生成的是否一致
	key := fmt.Sprintf(constant.SeckillPathKeyFormat, req.UserId, req.ProductId)
	err = s.RedisWrapper.SetPath(ctx, key, path)
	if err != nil {
		s.Logger.Error("Redis Set Path error", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to set path")
		return resp, nil
	}
	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "success")
	//这里的path有必要在resp返回吗。我后续再看看，毕竟已经写入了redis
	resp.Path = path
	return resp, nil
}
func (s *orderServiceImpl) GetSeckillResult_(ctx context.Context, req *order.GetSeckillResultReq) (resp *order.GetSeckillResultResp, err error) {
	resp = new(order.GetSeckillResultResp)
	resp.BaseResp = &base.BaseResp{}
	orderNo, err := s.RedisWrapper.GetSeckillCre(ctx, uint(req.ProductId), uint(req.UserId))
	if err != nil {
		s.Logger.Error("Redis get seckill result error", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to get seckill result")
		return resp, nil
	}
	if orderNo == "" {
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "invalid seckill result")
		return resp, nil
	}
	resp.OrderNo = orderNo
	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "success")
	return resp, nil
}
