#!/usr/bin/env bash
RUN_NAME="${RUN_NAME:-seckill.order}"
MAIN_PKG="${MAIN_PKG:-./cmd/order}"

mkdir -p output/bin
cp script/* output/
chmod +x output/bootstrap.sh

if [ "$IS_SYSTEM_TEST_ENV" != "1" ]; then
    go build -o output/bin/${RUN_NAME} "${MAIN_PKG}"
else
    go test -c -covermode=set -o output/bin/${RUN_NAME} -coverpkg=./... "${MAIN_PKG}"
fi
