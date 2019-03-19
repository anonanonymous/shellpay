# shellpay
[![Build Status](https://travis-ci.com/anonanonymous/shellpay.svg?branch=dev)](https://travis-ci.com/anonanonymous/shellpay)
[![Go Report Card](https://goreportcard.com/badge/github.com/anonanonymous/shellpay)](https://goreportcard.com/report/github.com/anonanonymous/shellpay)  
Turtlecoin payment gateway


# Merchant API

All `POST` requests are limited to 1KiB of data

## `POST` shellpay.ml/api/keys  
Generates a key pair for the merchant.  
All payment notifications sent to the merchant are signed with HMAC-SHA256 using the private key.
```
Params:
    @address: The Merchant's TurtleCoin address. It can be either a standard or integrated address

Result:
HTTP 200
{
    "status": "ok",
    "result":
    {
        "expiry": "key expiration date as unix timestamp
        "public_key": "64 character string",
        "private_key": "128 character string"
    }
}
HTTP 4xx
{
    "status": "error message"
}
```
Example
```
curl -d '{"address": "TRTLuySpDqd2fcvq5vx7Jiayw6yao7JHXFPuia5V83cVREtQSKyvWpxX9vamnUcG35BkQy6VfwUy5CsV9YNomioPGGyVhK3YXLq"}' shellpay.ml/api/keys
```

## `POST` shellpay.ml/api/invoice  
This creates an invoice. Sign the request body using HMAC-SHA256 with your private key so shellpay can verify request validity.  
Then place the hmac signature in the request header value "HMAC-SIGNATURE"  
```
Params:
    @ipn_uri: The URI where shellpay will send payment notifications
    @currency_amount: The amount to charge for the order
    @currency_code: The 3 letter code of the currency ie. "usd", "gbp", "btc", etc
    @order_id: Merchant specified id for this invoice
    @custom: Merchant specified custom data.
    @public_key: Merchant's public key

Result:
HTTP 200
{
    "status": "ok",
    "result":
    {
        "payment_id": "64 character string",
        "timestamp": "timestamp of the invoice creation"
    }
}
HTTP 4xx
{
    "status": "error message"
}
```
Python Example
```
import hashlib, hmac, json
from requests import post

def hmac_sign(message: str, secret: str):
    k = bytes(secret, 'utf8')
    m = bytes(message, 'utf8')
    return hmac.new(key=k, msg=m, digestmod=hashlib.sha256).hexdigest()

invoice = {
    'ipn_uri': 'https://starbucks.com/payments',
    'currency_amount': 10,  
    'currency_code': 'usd',
    'order_id': '123456',   
    'custom': "{'item_name': 'one tall coffee'}",
    'public_key': "PUBLIC KEY"
}
response = post('https://shellpay.ml/api/invoice', json=invoice,
                headers={'HMAC-SIGNATURE': hmac_sign(json.dumps(invoice), 'PRIVATE KEY')})
```

## `GET` shellpay.ml/api/invoice/:id  
Gets invoice status. `id` is the `payment_id` of the invoice
```
Params:
    None
Result:
HTTP 200
{
    "status": "ok",
    "result":
    {
        "status": "invoice status - unpaid/processing/paid"
        "amount_received": "the amount (in trtl) received in atomic units",
        "atomic_amount": "the amount (in trtl) due in atomic units",
        "fee": "shellpay fee amount (in trtl) in atomic units",
        "payment_address": "the address where funds should be sent",
        "payment_id": "64 character string",
        "currency_amount": "amount due in the original currency",  
        "currency_code": "the currency code",
        'order_id': '123456',   
        "expiration": "timestamp of the invoice expiration"
    }
}
HTTP 404
{
    "status": "not found"
}
```

# Payment Notifications
These notitifications are sent to the uri specified during invoice creation.
## Verifying Requests
Python Flask Snippet
```
@app.route('/ipn', methods=['POST'])
def ipn():
    # retrieve raw request data
    req = str(request.data, 'utf8')
    # compare hmac signatures
    if hmac_sign(req, keys['private_key']) == request.headers['HMAC-SIGNATURE']:
        process_ipn(req)

    return jsonify({})
```
Type - `invoice created`  
This is sent when an invoice is created
```
Payload:
{
    "type": 'invoice created',
    "currency_amount": "amount due in the original currency",  
    "currency_code": "the currency code",
    "custom": Merchant specified custom data.
    "merchant_address": 'the merchant's address',
    "order_id": "the order id",
    "public_key": "PUBLIC KEY",
}
```
Type - `paid`  
This is sent when an invoice is paid
```
Payload:
{
    "type": "paid",
    "currency_amount": "amount due in the original currency",  
    "currency_code": "the currency code",
    "custom": Merchant specified custom data.
    "merchant_address": 'the merchant's address',
    "order_id": "the order id",
    "public_key": "PUBLIC KEY",
}
```
Type - `expired`  
This is sent when an invoice is expires
```
Payload:
{
    "type": "expired",
    "currency_amount": "amount due in the original currency",  
    "currency_code": "the currency code",
    "custom": Merchant specified custom data.
    "merchant_address": 'the merchant's address',
    "order_id": "the order id",
    "public_key": "PUBLIC KEY",
}
```