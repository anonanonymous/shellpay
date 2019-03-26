package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

// the amount of time between polling invoices
const scanInterval = time.Second * 15

// number of times to attempt forwarding funds
const maxRetries = 1
const retryDelay = time.Second * 30

// paymentsProcessor - does the payment processing
func paymentsProcessor() {
	var transactions []Transaction
	var wg sync.WaitGroup

	for ; ; time.Sleep(scanInterval) {
		currentBlock, _, _, err := wallet.getStatus()
		if err != nil {
			log.Println("Wallet Error:", err)
		}
		expiredInv, _ := redisSMembers("expired_invoices")
		processingInv, _ := redisSMembers("processing_invoices")
		if len(processingInv) == 0 && len(expiredInv) == 0 {
			continue
		}
		for key := range processingInv {
			data, _ := redisHMGet(key, []string{
				"block",
				"merchant_address",
				"amount_received",
				"fee",
			})
			invBlock, _ := strconv.ParseInt(data[0], 10, 64)
			invRec, _ := strconv.ParseInt(data[2], 10, 64)
			invFee, _ := strconv.ParseInt(data[3], 10, 64)
			// mark confirmed invoices as paid
			if currentBlock-invBlock >= minConfirms {
				redisHMSet(key, map[string]interface{}{"status": "paid"})
				redisSRem("processing_invoices", key)
				redisExpire(key, int64(expDelay))
				transactions = append(transactions, Transaction{
					Amount:    invRec - invFee,
					Address:   data[1],
					PaymentID: key,
				})
			}
		}
		for key := range expiredInv {
			var addr string
			data, _ := redisHMGet(key, []string{
				"merchant_address",
				"return_address",
				"amount_received",
				"fee",
			})
			if data[1] != "-" {
				addr = data[1]
			} else {
				addr = data[0]
			}
			invRec, _ := strconv.ParseInt(data[2], 10, 64)
			invFee, _ := strconv.ParseInt(data[3], 10, 64)
			redisSRem("expired_invoices", key)
			if invRec > invFee {
				transactions = append(transactions, Transaction{
					Amount:    invRec - invFee,
					Address:   addr,
					PaymentID: key,
				})
			}
		}
		fmt.Println("transactions:", transactions)
		for i := 0; i < maxRetries; i++ {
			if err = wallet.sendTransaction(transactions); err != nil {
				log.Println("Error sending transaction:", err, transactions)
				time.Sleep(retryDelay)
			} else {
				break
			}
		}
		// send payment notifications
		wg.Add(len(transactions))
		for _, key := range transactions {
			go func(paymentID string) {
				fields := []string{
					"status",
					"order_id",
					"custom",
					"atomic_amount",
					"amount_received",
					"currency_code",
					"currency_amount",
					"ipn_uri",
					"public_key",
				}
				data, _ := redisHMGet(paymentID, fields)
				result := make(map[string]string, len(fields))
				for i, v := range data {
					if fields[i] == "status" {
						result["type"] = v
					} else {
						result[fields[i]] = v
					}
				}
				payload, _ := json.Marshal(result)
				_, err := sendRequest("POST", gatewayURI+"/api/ipn", string(payload), MasterKey)
				if err != nil {
					log.Println("Error sending IPN:", string(payload))
				}
				wg.Done()
			}(key.PaymentID)
		}
		wg.Wait()
		transactions = nil
	}
}
