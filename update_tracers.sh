#!/bin/bash
# Modify cmd/user/main.go
sed -i '' 's/etcdCfg \*config.EtcdConfig,/etcdCfg \*config.EtcdConfig,\n\t\ttraceCfg \*config.TracerConfig,/g' cmd/user/main.go
sed -i '' 's/log.Info("Starting Application/shutdown \:= tacer.InitTracer(traceCfg, constant.ServiceNameUser)\n\t\tdefer shutdown()\n\t\tlog.Info("Starting Application/g' cmd/user/main.go
sed -i '' 's/server.WithServerBasicInfo/server.WithSuite(kitextracing.NewServerSuite()),\n\t\t\tserver.WithServerBasicInfo/g' cmd/user/main.go
sed -i '' 's/"go.uber.org\/zap"/"go.uber.org\/zap"\n\t"full_backend_practice\/infrastructure\/tacer"\n\tkitextracing "github.com\/kitex-contrib\/obs-opentelemetry\/tracing"/g' cmd/user/main.go

# Modify cmd/order/main.go
sed -i '' 's/etcdCfg \*config.EtcdConfig,/etcdCfg \*config.EtcdConfig,\n\t\ttraceCfg \*config.TracerConfig,/g' cmd/order/main.go
sed -i '' 's/log.Info("Starting Application/shutdown \:= tacer.InitTracer(traceCfg, constant.ServiceNameOrder)\n\t\tdefer shutdown()\n\t\tlog.Info("Starting Application/g' cmd/order/main.go
sed -i '' 's/server.WithServerBasicInfo/server.WithSuite(kitextracing.NewServerSuite()),\n\t\t\tserver.WithServerBasicInfo/g' cmd/order/main.go
sed -i '' 's/"go.uber.org\/zap"/"go.uber.org\/zap"\n\t"full_backend_practice\/infrastructure\/tacer"\n\tkitextracing "github.com\/kitex-contrib\/obs-opentelemetry\/tracing"/g' cmd/order/main.go

# Modify cmd/product/main.go
sed -i '' 's/etcdCfg \*config.EtcdConfig,/etcdCfg \*config.EtcdConfig,\n\t\ttraceCfg \*config.TracerConfig,/g' cmd/product/main.go
sed -i '' 's/log.Info("Starting Application/shutdown \:= tacer.InitTracer(traceCfg, constant.ServiceNameProduct)\n\t\tdefer shutdown()\n\t\tlog.Info("Starting Application/g' cmd/product/main.go
sed -i '' 's/server.WithServerBasicInfo/server.WithSuite(kitextracing.NewServerSuite()),\n\t\t\tserver.WithServerBasicInfo/g' cmd/product/main.go
sed -i '' 's/"go.uber.org\/zap"/"go.uber.org\/zap"\n\t"full_backend_practice\/infrastructure\/tacer"\n\tkitextracing "github.com\/kitex-contrib\/obs-opentelemetry\/tracing"/g' cmd/product/main.go

