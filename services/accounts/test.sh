#!/usr/bin/env bash
DB_USER=postgres \
DB_PWD=postgres \
DB_NAME=userdb_test \
MASTER_KEY=dsanon \
HOST_URI='http://localhost' \
HOST_PORT=':7071' \
go test -v -cover
