# Payments Service

This is the payments service. It handles invoices and payments.  

## Getting Started
**Install the Dependencies**  
- Arch Linux  - `pacman -Sy go redis`  
- Debian/Ubuntu 
    - `apt install redis-server`
    - https://tecadmin.net/install-go-on-debian
- Go packages
    ```
    go get github.com/gomodule/redigo/redis \
           github.com/julienschmidt/httprouter
    ```

Now edit the following variables in run.sh:
```
HOST_URI        // uri of payments service
HOST_PORT       // host port of payments service
MASTER_KEY      // Key for HMAC signing http requests between services. Use the same value on all services
WALLET_HOST     // hostname of the wallet service
WALLET_PORT     // port of the wallet service
REDIS_PORT      // redis port number, default is :6379
GATEWAY_URI     // uri of the gateway service.
```
More config options are available in `config.go`.  
Make sure the wallet service is running before running the tests.  
Run the tests: `./run.sh test`  
If all tests pass you can now run this service with `./run.sh`. Logs will be written to `payments.log`  
## Running in production
**Make sure this service is not publicly available**  
[pm2](https://pm2.keymetrics.io) or service file  
[Redis replication](https://redis.io/topics/replication)  
[Use a Load Balancer with multiple instances of this service](https://www.digitalocean.com/community/tutorials/an-introduction-to-haproxy-and-load-balancing-concepts)