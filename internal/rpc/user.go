package rpc

import (
	"full_backend_practice/kitex_gen/user/userservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/zap"
)

func InitUserRpc(l *zap.Logger) (userservice.Client, error) {
	r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
	if err != nil {
		if l != nil {
			l.Info("fail to create etcd resolver")
		}
		return nil, err
	}
	clientImpl, err := userservice.NewClient(
		"user-service",
		client.WithResolver(r),
		client.WithLoadBalancer(loadbalance.NewWeightedRoundRobinBalancer()),
	)
	if err != nil {
		if l != nil {
			l.Fatal("Init User RPC Client failed", zap.Error(err))
		}
		return nil, err
	}

	return clientImpl, nil
}
