package tacer

import (
	"context"
	"full_backend_practice/pkg/config"

	"github.com/kitex-contrib/obs-opentelemetry/provider"
)

func InitTracer(cfg *config.TracerConfig, serviceName string) func() {
	if cfg == nil || !cfg.Enabled {
		return func() {}
	}
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(serviceName),
		provider.WithExportEndpoint(cfg.Endpoint),
		provider.WithInsecure(),
	)
	return func() {
		_ = p.Shutdown(context.Background())
	}
}
