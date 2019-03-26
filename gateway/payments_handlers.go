package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

// InitPaymentsHandlers - sets up the http handlers
func InitPaymentsHandlers(r *httprouter.Router) {
	r.POST("/api/ipn", limit(handleIPN, ratelimiter))
	r.POST("/api/invoice", limit(handleNewInvoice, ratelimiter))
	r.GET("/api/invoice/:id", limit(handleGetInvoice, ratelimiter))
	r.GET("/api/status", limit(handleStatus, ratelimiter))
	r.POST("/api/keys", limit(handleNewKey, strictRL))
}

// handleNewInvoice - handles invoice requests
func handleNewInvoice(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var data []byte
	var jresp map[string]interface{}
	var keys *Key
	var err error

	data, err = getBody(req)
	if err != nil {
		handleError(res, "Server Error", http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(data, &jresp); err != nil {
		handleError(res, "Error Parsing JSON", http.StatusBadRequest)
		return
	}

	// get key from signatures db
	keys, err = getKey(jresp["public_key"].(string))
	if err != nil {
		handleError(res, "Invalid Public Key", http.StatusBadRequest)
		return
	}
	if sign(string(data), keys.Priv) != req.Header.Get("HMAC-SIGNATURE") {
		handleError(res, "Not Authorized", http.StatusForbidden)
		log.Println("Forbidden Merchant:", req.RemoteAddr, keys.Pub)
		return
	}

	jresp["merchant_address"] = keys.Address
	if data, err = json.Marshal(jresp); err != nil {
		handleError(res, "Server Error", http.StatusInternalServerError)
		return
	}
	// send invoice info to payments service
	resp, err := sendRequest("POST", paymentURI+"/api/invoice", string(data), MasterKey)
	if err != nil {
		handleError(res, "Failed to ping payment api", http.StatusInternalServerError)
		return
	}
	if (*resp)["status"].(string) != "ok" {
		handleError(res, "Failed to create invoice", http.StatusInternalServerError)
		log.Println("Failed to create invoice:", (*resp)["status"])
		return
	}
	json.NewEncoder(res).Encode(*resp)

	// send "invoice created" payment notification to merchant
	jresp["type"] = "invoice created"
	ipn := jresp["ipn_uri"].(string)
	delete(jresp, "ipn_uri")
	if data, err = json.Marshal(jresp); err != nil {
		log.Println("Error Creating IPN Payload:", err)
		return
	}
	go func() {
		_, err = sendRequest("POST", ipn, string(data), keys.Priv)
		if err != nil {
			log.Println("Error Sending IPN:", err)
		}
	}()
}

// handleGetInvoice - gets invoice details from payments service
func handleGetInvoice(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
	resp, err := sendRequest("GET", paymentURI+"/api/invoice/"+p.ByName("id"), "", MasterKey)
	if err != nil {
		handleError(res, "Server Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(*resp)
}

// handleStatus - service status
func handleStatus(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	payments, err := sendRequest("GET", paymentURI+"/api/status", "", MasterKey)
	if err != nil {
		payments = &map[string]interface{}{"status": "offline"}
	}
	/*
		accounts, err := sendRequest("GET", accountsURI+"/api/status", "", MasterKey)
		if err != nil {
			acounts = &map[string]interface{}{"status": "offline"}
		}
	*/
	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: map[string]string{
			"payments": (*payments)["status"].(string),
			// "accounts": (*accounts)["status"].(string),
		},
	})
}

// handleNewKey - generates a pub/priv key for merchant api requests
func handleNewKey(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var jresp request
	/* maybe add request challenge here */

	data, err := getBody(req)
	if err != nil {
		handleError(res, "Server Error", http.StatusInternalServerError)
		return
	}
	if err = json.Unmarshal(data, &jresp); err != nil {
		handleError(res, "Error Parsing JSON", http.StatusBadRequest)
		return
	}
	if !validateAddress(jresp["address"]) {
		handleError(res, "Invalid Address", http.StatusBadRequest)
		return
	}

	random, err := randHexStr(32)
	if err != nil {
		handleError(res, "Server Error", http.StatusInternalServerError)
		return
	}
	// this is an example pub/priv you should use something more secure
	pub := sign(string(data), random)
	priv := sign(pub, pub+random) + random
	exp := time.Now().Round(time.Hour * 24).Add(time.Hour * 24 * keyValidDays).Unix()
	if !setKey(pub, priv, jresp["address"], exp) {
		handleError(res, "Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: map[string]string{
			"public_key":  pub,
			"private_key": priv,
			"expiry":      strconv.FormatInt(exp, 10),
		},
	})
}

// handleIPN - sends payment notification to merchant uri
func handleIPN(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var jresp map[string]string
	data, code, err := validateRequest(req)
	if err != nil {
		handleError(res, err.Error(), code)
		return
	}
	// remove the `ipn_uri` field from the payload
	json.Unmarshal(data, &jresp)
	ipn := jresp["ipn_uri"]
	delete(jresp, "ipn_uri")

	data, _ = json.Marshal(jresp)
	merchant, err := getKey(jresp["public_key"])
	if err != nil {
		log.Println("Error in handleIPN:", err.Error())
	}
	_, err = sendRequest("POST", ipn, string(data), merchant.Priv)
	if err != nil {
		log.Println("Error Sending IPN:", err)
	}
}
