package user

import (
	"context"
	"encoding/json"
	"fmt"
	"full_backend_practice/infrastructure/database"
	"full_backend_practice/kitex_gen/base"
	"full_backend_practice/kitex_gen/user"

	"full_backend_practice/infrastructure/mq"
	"full_backend_practice/pkg/response"
	"full_backend_practice/pkg/token"

	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"full_backend_practice/infrastructure/tacer"
)

type userServiceImpl struct {
	MySqlWrapper UserDatabase
	RedisWrapper *database.RedisWrapper
	MQ           *mq.RabbitClient
	Logger       *zap.Logger
}

func NewUserServiceImpl(mr UserDatabase, rw *database.RedisWrapper, mq *mq.RabbitClient, logger *zap.Logger) UserServiceImpl {
	return &userServiceImpl{
		MySqlWrapper: mr,
		RedisWrapper: rw,
		MQ:           mq,
		Logger:       logger,
	}
}

func (s *userServiceImpl) Register(ctx context.Context, req *user.RegisterReq) (resp *user.RegisterResp, err error) {
	resp = new(user.RegisterResp)
	resp.BaseResp = &base.BaseResp{}

	if req.Username == "" || req.Password == "" {
		resp.BaseResp = response.BuildBaseResp(response.CodeInvalidParams, "username and password required")
		return resp, nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "password hash error")
		return resp, nil
	}

	mqmsg := mq.UserMessage{
		Username: req.Username,
		Password: string(hash),
	}
	headers := amqp091.Table{}
		tacer.InjectAMQPHeaders(ctx, headers)

		body, err := json.Marshal(mqmsg)
	if err != nil {
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, fmt.Sprintf("marshal error: %v", err))
		return resp, nil
	}
	err = s.MQ.Channel.PublishWithContext(ctx, "", "user_register", false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
Headers:      headers,
		Body:         body,
	})
	if err != nil {
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to register user")
		return resp, nil
	}
	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "success to register user")
	return resp, nil
}

func (s *userServiceImpl) Login(ctx context.Context, req *user.LoginReq) (resp *user.LoginResp, err error) {
	resp = new(user.LoginResp)
	resp.BaseResp = &base.BaseResp{}

	if req.Username == "" || req.Password == "" {
		resp.BaseResp = response.BuildBaseResp(response.CodeInvalidParams, "username and password required")
		return resp, nil
	}

	u, err := s.MySqlWrapper.LoginUser(req.Username)
	if err != nil {
		s.Logger.Warn("login failed: user not found", zap.String("username", req.Username))
		resp.BaseResp = response.BuildBaseResp(response.CodeUnauthorized, "invalid username or password")
		return resp, nil
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password))
	if err != nil {
		s.Logger.Warn("login failed: incorrect password", zap.String("username", req.Username))
		resp.BaseResp = response.BuildBaseResp(response.CodeUnauthorized, "invalid username or password")
		return resp, nil
	}

	accessToken, _, err := token.TokenCreate(&token.Payload{
		UserID:   u.ID,
		Username: u.Username,
	})
	if err != nil {
		s.Logger.Error("login failed: fail to create token", zap.Error(err))
		resp.BaseResp = response.BuildBaseResp(response.CodeInternal, "fail to generate token")
		return resp, nil
	}

	resp.BaseResp = response.BuildBaseResp(response.CodeOK, "success to login")
	resp.Token = &accessToken
	return resp, nil
}
