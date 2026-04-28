package user_service

import (
	"context"
	"encoding/json"
	"fmt"
	"full_backend_practice/kitex_gen/base"
	"full_backend_practice/kitex_gen/user"
	"full_backend_practice/pkg/mq"

	"github.com/rabbitmq/amqp091-go"
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
	err = mq.Channel.PublishWithContext(ctx, "", "user_register_queue", false, false, amqp091.Publishing{
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

	// TODO: 在这里编写你的登录业务逻辑（例如查询数据库、生成和验证分发 JWT Token）

	return resp, nil
}
