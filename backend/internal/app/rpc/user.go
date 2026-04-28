package rpc

import (
	"full_backend_practice/kitex_gen/user/userservice"
	"full_backend_practice/pkg/logger"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	etcd "github.com/kitex-contrib/registry-etcd"
)

var UserClient userservice.Client

func InitUserRpc() {
	r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
	if err != nil {
		logger.Log.Info("fail to create etcd resolver")
	}
	UserClient, err = userservice.NewClient(
		"user-service",
		client.WithResolver(r),
		client.WithLoadBalancer(loadbalance.NewWeightedRoundRobinBalancer()),
	)
	if err != nil {
		panic("初始化 RPC 客户端失败: " + err.Error())
	}

}
