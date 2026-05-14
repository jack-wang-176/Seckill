package rpc

import (
	"full_backend_practice/kitex_gen/order/orderservice"
	"full_backend_practice/pkg/config"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/zap"
)

func InitOrderRpc(cfg *config.EtcdConfig, l *zap.Logger) (orderservice.Client, error) {
	r, err := etcd.NewEtcdResolver(cfg.Endpoints)
	if err != nil {
		if l != nil {
			l.Error("fail to create etcd resolver", zap.Error(err))
		}
		return nil, err
	}

	clientImpl, err := orderservice.NewClient(
		"order_service",
		client.WithResolver(r),
		client.WithLoadBalancer(loadbalance.NewWeightedRoundRobinBalancer()),
	)

	if err != nil {
		if l != nil {
			l.Fatal("Init Order RPC Client failed", zap.Error(err))
		}
		return nil, err
	}

	return clientImpl, nil
}
