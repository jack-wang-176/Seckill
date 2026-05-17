sed -i '' '/provider.WithInsecure(),/a\
\t\tprovider.WithEnableMetrics(false),\
' infrastructure/tacer/tracer.go

