package user_service

import (
	"context"
	"encoding/json"
	"fmt"
	"full_backend_practice/kitex_gen/base"
	"full_backend_practice/kitex_gen/user"
	"full_backend_practice/pkg/database"
	"full_backend_practice/pkg/logger"
	"full_backend_practice/pkg/mq"
	"full_backend_practice/pkg/token"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserServiceImpl struct{}

// Register 用户注册接口，异步写入 MQ，密码加密
func (s *UserServiceImpl) Register(ctx context.Context, req *user.RegisterReq) (resp *user.RegisterResp, err error) {
	resp = new(user.RegisterResp)
	resp.BaseResp = &base.BaseResp{}

	// 1. 参数校验
	if req.Username == "" || req.Password == "" {
		resp.BaseResp.Code = 400
		resp.BaseResp.Msg = "username and password required"
		return resp, nil
	}

	// 2. 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		resp.BaseResp.Code = 500
		resp.BaseResp.Msg = "password hash error"
		return resp, nil
	}

	// 3. 写入 MQ
	mqmsg := mq.UserMessage{
		Username: req.Username,
		Password: string(hash),
	}
	body, err := json.Marshal(mqmsg)
	if err != nil {
		resp.BaseResp.Code = 500
		resp.BaseResp.Msg = fmt.Sprintf("marshal error: %v", err)
		return resp, nil
	}
	err = mq.Client.Channel.PublishWithContext(ctx, "", "user_register", false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         body,
	})
	if err != nil {
		resp.BaseResp.Code = 500
		resp.BaseResp.Msg = "fail to register user"
		return resp, nil
	}
	resp.BaseResp.Code = 200
	resp.BaseResp.Msg = "success to register user"
	return resp, nil
}

// Login 用户登录接口
func (s *UserServiceImpl) Login(ctx context.Context, req *user.LoginReq) (resp *user.LoginResp, err error) {
	resp = new(user.LoginResp)
	resp.BaseResp = &base.BaseResp{}

	if req.Username == "" || req.Password == "" {
		resp.BaseResp.Code = 400
		resp.BaseResp.Msg = "username and password required"
		return resp, nil
	}

	// 1. 同步查询数据库校验用户是否存在
	var u database.User
	err = database.DB.Where("username = ?", req.Username).First(&u).Error
	if err != nil {
		logger.Log.Warn("login failed: user not found", zap.String("username", req.Username))
		resp.BaseResp.Code = 401
		resp.BaseResp.Msg = "invalid username or password"
		return resp, nil
	}

	// 2. 校验密码哈希是否匹配
	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password))
	if err != nil {
		logger.Log.Warn("login failed: incorrect password", zap.String("username", req.Username))
		resp.BaseResp.Code = 401
		resp.BaseResp.Msg = "invalid username or password"
		return resp, nil
	}

	// 3. 生成 JWT Token
	accessToken, _, err := token.TokenCreate(&u)
	if err != nil {
		logger.Log.Error("login failed: fail to create token", zap.Error(err))
		resp.BaseResp.Code = 500
		resp.BaseResp.Msg = "fail to generate token"
		return resp, nil
	}

	resp.BaseResp.Code = 200
	resp.BaseResp.Msg = "success to login"
	resp.Token = &accessToken // 携带生成的 Token 返回给调用方
	return resp, nil
}
