#!/usr/bin/env bash
# This program is used to run multiple instances of the wallet service
# Replace <placeholder> with the appropiate values for your configuration

if [[ $# -lt 3 ]]; then
	echo 'Usage: ./wallet_start.sh <containter file path> <bind-port>';
	exit;
fi

while true; do
    ./turtle-service \
    -w $1 \
    -p <placeholder> --rpc-password <placeholder> \
    --daemon-address <placeholder> \
    --log-file service-$2 \
    --bind-port $2 &
    pid=$(($(echo $$)+1));
    trap ' echo killing service on port $2; kill -9 $pid; exit ' SIGTERM SIGINT SIGQUIT
    
	sleep 30 &
	wait $!
    while true; do
    	curl -m 5 -d '{"jsonrpc":"2.0","id":1,"password":"<placeholder>","method":"getStatus","params":{}}' http://localhost:$2/json_rpc;
    	res=$(echo $?);
    	if [[ $res -ne 0 ]]; then
    		break;
    	fi
    
    	sleep 60 &
    	wait $!
    	curl -m 60 -d '{"jsonrpc":"2.0","id":1,"password":"<placeholder>","method":"save","params":{}}' http://localhost:$2/json_rpc;
    	res=$(echo $?);
    	if [[ $res -ne 0 ]]; then
    		break;
    	fi
    done
    kill -9 $pid;
    echo service on port $2 died;
	sleep 5;
done