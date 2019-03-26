package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	//"./models"
	//"github.com/anonanonymous/shellpay/services/payments/models"

	"github.com/julienschmidt/httprouter"
)

var timers = map[string]*time.Timer{}

func main() {
	defer logFile.Close()
	log.SetOutput(logFile)

	router := httprouter.New()

	router.POST("/api/transaction/received", handleTXReceived)
	router.POST("/api/invoice", handleCreateInvoice)
	router.GET("/api/invoice/:id", handleGetInvoice)
	router.GET("/api/status", handleStatus)

	srv := &http.Server{
		Addr:         hostPort,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
		Handler:      router,
	}

	go paymentsProcessor()
	log.Println("Info: Starting Service on:", hostURI)
	log.Fatal(srv.ListenAndServe())
}

// handleCreateInvoice - make a payment invoice
func handleCreateInvoice(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var inv invoice
	var devFee int64

	data, code, err := validateRequest(req)
	if err != nil {
		handleError(res, err.Error(), code)
		return
	}
	if err = json.Unmarshal(data, &inv); err != nil {
		handleError(res, "Error Parsing JSON", http.StatusBadRequest)
		return
	}
	if !checkInvoice(&inv) {
		handleError(res, "Invoice Error: invalid/missing data", http.StatusBadRequest)
		return
	}

	rate := exchangeRates[inv.CurCode]
	if rate <= 0.0 {
		handleError(res, "Error calculating cost in "+inv.CurCode, http.StatusBadRequest)
		return
	}

	paymentID, err := randHexStr(32)
	if err != nil {
		handleError(res, "Error creating paymentID", http.StatusInternalServerError)
		return
	}

	paymentAddr, err := wallet.newIntegratedAddress(paymentID)
	if err != nil {
		handleError(res, "Error creating payment address", http.StatusInternalServerError)
		return
	}

	// calculate cost and (((devFee))). node fee is the minimum fee
	inv.AtomicAmount = int64((inv.CurAmount / rate) * atoms)
	if devFee = int64((float64(inv.AtomicAmount) * feeRate)); devFee < wallet.NodeFee {
		devFee = wallet.NodeFee
	}
	inv.AtomicAmount += devFee
	// dump invoice details into redis
	now := time.Now()
	inv.Timestamp = now.Unix()
	inv.Expiry = now.Add(paymentExpire).Unix()
	err = redisHMSet(paymentID, map[string]interface{}{
		"status":           "unpaid",
		"order_id":         inv.OrderID,
		"ipn_uri":          inv.IpnURI,
		"currency_code":    inv.CurCode,
		"currency_amount":  inv.CurAmount,
		"payment_id":       paymentID,
		"merchant_address": inv.MerchantAddress,
		"payment_address":  paymentAddr,
		"return_address":   inv.ReturnAddress,
		"custom":           inv.Custom,
		"fee":              devFee,
		"atomic_amount":    inv.AtomicAmount,
		"amount_received":  0,
		"timestamp":        inv.Timestamp,
		"expiration":       inv.Expiry,
		"public_key":       inv.PublicKey,
	})
	if err != nil {
		handleError(res, "Redis Error", http.StatusInternalServerError)
		return
	}

	err = redisExpire(paymentID, int64(paymentExpire.Seconds())+expDelay)
	if err != nil {
		handleError(res, "Redis Error", http.StatusInternalServerError)
		return
	}
	// function to run if the invoice expires
	timers[paymentID] = time.AfterFunc(paymentExpire, func() {
		delete(timers, paymentID)
		redisHMSet(paymentID, map[string]interface{}{"status": "expired"})
		redisSAdd("expired_invoices", []string{paymentID})
	})

	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: map[string]string{
			"payment_id": paymentID,
			"timestamp":  strconv.FormatInt(inv.Timestamp, 10),
		},
	})
}

// handleGetInvoice - gets info about an invoice from redis
func handleGetInvoice(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	_, code, err := validateRequest(req)
	if err != nil {
		handleError(res, err.Error(), code)
		return
	}

	fields := []string{
		"status",
		"order_id",
		"atomic_amount",
		"amount_received",
		"expiration",
		"payment_address",
		"currency_code",
		"currency_amount",
		"fee",
	}
	data, err := redisHMGet(p.ByName("id"), fields)
	if err != nil || data[0] == "" {
		handleError(res, "Not found", http.StatusNotFound)
		return
	}
	// build payload
	payload := make(map[string]string, len(fields))
	for i, v := range data {
		payload[fields[i]] = v
	}

	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: payload,
	})
}

// handleTXReceived - handles payments
func handleTXReceived(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var jresp map[string]interface{}
	// TODO: hmac sign request in wallet service
	data, err := getBody(req)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	if err = json.Unmarshal(data, &jresp); err != nil {
		log.Println("Error Parsing Json:", err)
		return
	}

	pID, amount := jresp["payment_id"].(string), int64(jresp["amount"].(float64))
	block := jresp["block"].(string)
	if err := redisHIncrBy(pID, "amount_received", amount); err != nil {
		log.Println("Redis Error:", err)
		return
	}

	inv, err := redisHMGet(pID, []string{"atomic_amount", "amount_received"})
	if err != nil {
		log.Println("Redis Error:", err)
		return
	}
	// we can trust that these are valid ints
	cost, _ := strconv.ParseInt(inv[0], 10, 64)
	received, _ := strconv.ParseInt(inv[1], 10, 64)
	// if invoice is paid, add it to the `processing_invoices` map
	if received >= cost {
		timers[pID].Stop()
		delete(timers, pID)
		redisHMSet(pID, map[string]interface{}{"status": "processing", "block": block})
		redisSAdd("processing_invoices", []string{pID})
	}
}

// handleStatus - service status
func handleStatus(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	_, code, err := validateRequest(req)
	if err != nil {
		handleError(res, err.Error(), code)
		return
	}
	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: map[string]string{
			"timestamp": strconv.FormatInt(time.Now().Unix(), 10),
		},
	})
}
