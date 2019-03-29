# Wallet Service
This is the wallet service. It handles receiving/sending transactions  

## Getting Started
**Install the Dependencies**  
- Arch Linux  - `pacman -Sy nodejs yarn`  
- Debian/Ubuntu - https://tecadmin.net/install-latest-nodejs-npm-on-debian/  

Then run the `yarn` command in this directory  

Next create a wallet file. This wallet will be receiving and forwarding funds from customers to merchants.  
```./create_wallet.js <name of wallet file> <wallet password>```  

Edit the config.json to meet your requirements
```
wallet_file     // name of the wallet file
wallet_pass     // wallet file passsword
master_key      // Key for HMAC signing http requests between services
payments_uri    // uri of the payments service, default is localhost:7073
```
Now start the service with `./index.js`. It will run on port 8070  
## Running in production
**Make sure this service is not publicly available**  
[pm2](https://pm2.keymetrics.io) or [service file](https://nodesource.com/blog/running-your-node-js-app-with-systemd-part-1) See `payments.service`  
[Use a Load Balancer with multiple instances of this service](https://www.digitalocean.com/community/tutorials/an-introduction-to-haproxy-and-load-balancing-concepts)