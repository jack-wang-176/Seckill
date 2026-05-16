package rpc

import (
	"full_backend_practice/kitex_gen/product/productservice"
	"full_backend_practice/pkg/config"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/loadbalance"
	kitextracing "github.com/kitex-contrib/obs-opentelemetry/tracing"
	etcd "github.com/kitex-contrib/registry-etcd"
	"go.uber.org/zap"
)

func InitProductRpc(cfg *config.EtcdConfig, l *zap.Logger) (productservice.Client, error) {
	r, err := etcd.NewEtcdResolver(cfg.Endpoints)
	if err != nil {
		if l != nil {
			l.Info("fail to create etcd resolver")
		}
		return nil, err
	}
	clientImpl, err := productservice.NewClient(
		"product_service",
		client.WithResolver(r),
		client.WithLoadBalancer(loadbalance.NewWeightedRoundRobinBalancer()),
		client.WithSuite(kitextracing.NewClientSuite()),
	)
	if err != nil {
		if l != nil {
			l.Info("fail to create product service client")
		}
		return nil, err
	}
	return clientImpl, nil
}
