#!/usr/bin/env bash
DB_USER=postgres \
DB_PWD=postgres \
MASTER_KEY=dsanon \
HOST_URI='http://localhost' \
HOST_PORT=':7071' \
go run *.go
