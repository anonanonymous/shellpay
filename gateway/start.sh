#!/usr/bin/env bash
export MASTER_KEY=
export DB_USER=
export DB_PWD=
export USER_URI='http://localhost:7071'
export PAYMENT_URI='http://localhost:7073'
export HOST_URI='http://localhost'
export HOST_PORT=':7072'

if [[ $1 == 'test' ]]; then
	go test -v -cover
elif [[ $1 == 'start' ]]; then
	./gateway
elif [[ $1 == 'build' ]]; then
	go build gateway.go http_utils.go utils.go db_utils.go config.go payments_handlers.go web_handlers.go
else
	go run gateway.go http_utils.go utils.go db_utils.go config.go payments_handlers.go web_handlers.go
fi
