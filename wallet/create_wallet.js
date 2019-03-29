#!/usr/bin/env node
const WB = require('turtlecoin-wallet-backend');

if (process.argv.length !== 4) {
    console.log("Usage: ./create_wallet <wallet file> <wallet password>")
    return
}

(async () => {
    const daemon = new WB.BlockchainCacheApi('blockapi.turtlepay.io', true);
    const wallet = WB.WalletBackend.createWallet(daemon)
    wallet.saveWalletToFile(process.argv[2], process.argv[3])

})().catch(err => {
    console.log('Caught promise rejection: ' + err);
});
