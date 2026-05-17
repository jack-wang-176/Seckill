sed -i '' '/container_name: seckill-jaeger/a\
\    environment:\
\      - MEMORY_MAX_TRACES=5000\
' config/docker-compose.yml
