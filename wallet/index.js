#!/usr/bin/env node
const fastify = require('fastify')({logger: false});
const fs = require('fs');
const WB = require('turtlecoin-wallet-backend');
const fetch = require('node-fetch');
const crypto = require('crypto');
const conf = require('./config.json');

const [wallet_file, wallet_pass] = [conf['wallet_file'], conf['wallet_pass']];
const payments_uri = conf['payments_uri'];

function hmac_sign(message) {
    let hmac = crypto.createHmac('sha256', conf['master_key']);
    hmac.update(message);
    return hmac.digest('hex');
}

(async () => {
    const daemon = new WB.BlockchainCacheApi('blockapi.turtlepay.io', true);

    const [wallet, error] = WB.WalletBackend.openWalletFromFile(daemon,wallet_file, wallet_pass) ;
	if (error !== undefined) { return; }
    console.log('Opened wallet', wallet.getPrimaryAddress());

    await wallet.start();

    console.log('Started wallet');
    console.log(wallet.getBalance());
 
    wallet.on('sync', (walletHeight, networkHeight) => {
        wallet.saveWalletToFile(wallet_file, wallet_pass);
        console.log(`Wallet synced! Wallet height: ${walletHeight}, Network height: ${networkHeight}`);
    })

    // send payment details to payment service
    wallet.on('incomingtx', (transaction) => {
        let payload = {
            'amount': transaction.totalAmount(),
            'payment_id': transaction.paymentID,
            'block': transaction.blockHeight.toString()
        };
        fetch(payments_uri + '/api/transaction/received', {
            method: 'POST',
            body: JSON.stringify(payload),
            headers: {
                "HMAC-SIGNATURE": hmac_sign(JSON.stringify(payload))
            }
        });
        console.log('Incoming transaction', transaction);
    })

    // send payment details to payment service
    wallet.on('outgoingtx', (transaction) => {
        fetch(payments_uri+'/api/transaction/sent', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify(transaction)
        })
        console.log(`Outgoing transaction of ${transaction.totalAmount()} sent!`);
    })

    // send wallet infomation to payments service
    fastify.get('/wallet', (request, reply) => {
        reply.send({
            address: wallet.getPrimaryAddress(),
            node_fee: wallet.getNodeFee()[1]
        })
    })

    // send wallet status to payments service
    fastify.get('/wallet/status', (request, reply) => {
        console.log(wallet.getSyncStatus());
        reply.send({
            block: wallet.getSyncStatus()[1],
            balance: wallet.getBalance()[0],
            is_synced: ((wallet.getSyncStatus()[1] - wallet.getSyncStatus()[0]) < 2)
        });
    })

    // saves the wallet
    fastify.post('/wallet/save', (request, reply) => {
        wallet.saveWalletToFile(wallet_file, wallet_pass);
        reply.send({})
    })

    // create integrated address
    fastify.post('/wallet/integrated_address', (request, reply) => {
        let address = WB.createIntegratedAddress(wallet.getPrimaryAddress(), request.body["payment_id"]);
        reply.send({ integrated_address: address });
    })

    // sends a transaction
    fastify.post('/wallet/send_transaction', (request, reply) => {
        console.log(request.body["transfers"]);
        wallet.sendTransactionAdvanced(request.body["transfers"])
        .then((response) => {
            if (response[1] !== undefined) {
                reply.send({ status: 'error' });
            } else {
                reply.send({ status: 'ok', block: wallet.getSyncStatus()[1]});
            }
            wallet.saveWalletToFile(wallet_file, wallet_pass);
        });
    })

    // run the server
    fastify.listen(process.env.NODE_PORT || 8070, (err, host) => {
        if (err) throw err
        fastify.log.info(`server listening on ${host}`)
    })

})().catch(err => {
    console.log('Caught promise rejection: ' + err);
});
