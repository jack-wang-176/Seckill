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
	isValid, err := s.RedisWrapper.ConfirmPaht(ctx, req.Path, fmt.Sprintf("seckill:path:%d:%d", req.UserId, req.ProductId))
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

	soldOutKey := fmt.Sprintf("seckill:stock:%dsoldout", req.ProductId)
	//这里和预热中保持一致
	stockKey := fmt.Sprintf("seckill:stock:%d", req.ProductId)

	result := s.RedisWrapper.SimpleDecrStock(ctx, []string{stockKey, soldOutKey})
	if result == -1 {
		s.Logger.Warn("Stock sold out", zap.Int64("ProductID", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "sell out")
		return resp, nil
	} else if result == -2 {
		s.Logger.Error("Stock Decr inject error,nees further investigation", zap.Int64("ProductID", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to seckill order")
		return resp, nil
	} else if result == -3 {
		s.Logger.Error("Stock not find", zap.Int64("ProductID", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to seckill order")
		return resp, nil
	}
	orderNod := fmt.Sprintf("SN%d%d", time.Now().UnixNano(), req.UserId)
	//这里orderNO能不能继续用，在目前的代码中我好像直接忽略了这一点
	//从设计上有必要，其他路由需要对这个字段进行对接，目前在order里面可能只去用getseckillresult去对接
	//其他微服务在req的时候就要考虑这一点
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

	err = s.MQ.Channel.PublishWithContext(ctx, "", "order_seckill", false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})

	if err != nil {
		s.Logger.Error("RabbitMQ publish error", zap.Error(err))
		_ = s.RedisWrapper.Client.Incr(ctx, fmt.Sprintf("seckill:stock:%d", req.ProductId))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to seckill order")
		return resp, nil
	}

	s.Logger.Info("Order pushed to MQ successfully", zap.String("order_no", orderNod))

	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "seckill success")
	return resp, nil
}
func (s *orderServiceImpl) SeckillPath(ctx context.Context, req *order.GetSeckillPathReq) (resp *order.GetSeckillPathResp, err error) {
	resp = new(order.GetSeckillPathResp)
	resp.BaseResp = &base.BaseResp{}
	salt := fmt.Sprintf("May@)@#)(*&^$#@!%s", time.Now().String())
	path := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d:%d:%s", req.UserId, req.ProductId, salt))))
	//交给前端。生成对应路径，存入redis，后续校验路径和之前生成的是否一致
	key := fmt.Sprintf("seckill:path:%d:%d", req.UserId, req.ProductId)
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
	isValid, err := s.RedisWrapper.GetSeckillCre(ctx, uint(req.ProductId), uint(req.UserId))
	if err != nil {
		s.Logger.Error("Redis get seckill result error", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to get seckill result")
		return resp, nil
	}
	if !isValid {
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "invalid seckill result")
		return resp, nil
	}
	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "success")
	return resp, nil
}
