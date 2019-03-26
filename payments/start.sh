#!/usr/bin/env bash
# payments service config
export WALLET_HOST='http://localhost'
export WALLET_PORT='8070'
export MASTER_KEY=
export HOST_PORT=':7073'
export REDIS_PORT=':6379'
export HOST_URI='http://localhost'
export GATEWAY_URI='http://localhost:7072'

if [[ $1 == 'test' ]]; then
	go test -v -cover
elif [[ $1 == 'build' ]]; then
	go build utils.go tx_processor.go http_utils.go redis_utils.go config.go exchange_rates.go  payments.go wallet_utils.go ; mv utils payments
else
	go run payments.go utils.go tx_processor.go http_utils.go redis_utils.go config.go exchange_rates.go wallet_utils.go
fi
