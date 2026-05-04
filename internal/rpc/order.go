package rpc

import (
	"full_backend_practice/kitex_gen/order/orderservice"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/zap"
)

var OrderClient orderservice.Client

func InitOrderRpc(l *zap.Logger) {
	r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
	if err != nil {
		if l != nil {
			l.Error("fail to create etcd resolver", zap.Error(err))
		}
	}

	OrderClient, err = orderservice.NewClient(
		"order-service",
		client.WithResolver(r),
		client.WithLoadBalancer(loadbalance.NewWeightedRoundRobinBalancer()),
	)

	if err != nil {
		if l != nil {
			l.Fatal("Init Order RPC Client failed", zap.Error(err))
		}
	}
}
