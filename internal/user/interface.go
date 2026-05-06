package user

import (
	"context"
	"full_backend_practice/kitex_gen/user"

	"full_backend_practice/infrastructure/mq"
)

type UserServiceImpl interface {
	Register(ctx context.Context, req *user.RegisterReq) (resp *user.RegisterResp, err error)
	Login(ctx context.Context, req *user.LoginReq) (resp *user.LoginResp, err error)
}

type UserConsumer interface {
	StartRegisterConsumer()
}

type UserDatabase interface {
	RegisterUser(msg mq.UserMessage) error
	FindUser(msg mq.UserMessage) error
	CreateUser(msg mq.UserMessage) error
	LoginUser(username string) (*User, error)
}
