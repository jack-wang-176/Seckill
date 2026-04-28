package rpc

import (
	"full_backend_practice/kitex_gen/order/orderservice"
	"log"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	etcd "github.com/kitex-contrib/registry-etcd"
)

var OrderClient orderservice.Client

func InitOrderRpc() {
	r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
	if err != nil {
		log.Printf("fail to create etcd resolver: %v\n", err)
	}

	OrderClient, err = orderservice.NewClient(
		"order-service",
		client.WithResolver(r),
		client.WithLoadBalancer(loadbalance.NewWeightedRoundRobinBalancer()),
	)

	if err != nil {
		panic("初始化 RPC 客户端失败: " + err.Error())
	}
}
