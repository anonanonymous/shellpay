package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/julienschmidt/httprouter"
)

var txs []string

func Router() *httprouter.Router {
	router := httprouter.New()
	router.POST("/api/transaction/received", handleTXReceived)
	router.POST("/api/invoice", handleCreateInvoice)
	router.GET("/api/invoice/:id", handleGetInvoice)
	router.GET("/api/status", handleStatus)
	return router
}

func TestGetStatus(t *testing.T) {
	var result response

	req, _ := http.NewRequest("GET", "/api/status", bytes.NewBufferString(""))
	req.Header.Add("HMAC-SIGNATURE", sign("", MasterKey))
	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)
	json.Unmarshal(resp.Body.Bytes(), &result)
	assert.Equal(t, "ok", result.Status)
}

func TestCreateInvoice(t *testing.T) {
	var result response
	validInv := []invoice{
		invoice{
			IpnURI:          "http://localhost:3333",
			CurCode:         "usd",
			CurAmount:       1.0,
			OrderID:         "wewve3",
			Custom:          "{\"item_name\": \"거북이\"}",
			PublicKey:       "ad221b7f753f55f95eb4ab1de082388a3dde35c822f458b2658fa2b5eb476cdb",
			MerchantAddress: "TRTLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
		},
		invoice{
			IpnURI:          "http://localhost:3333",
			CurCode:         "btc",
			CurAmount:       0.0001,
			OrderID:         "w\"e'wve3",
			Custom:          "{item_na'me\": \"one\" coffee\"}",
			PublicKey:       "ad221b7f753f55f95eb4ab1de082388a3dde35c822f458b2658fa2b5eb476cdb",
			MerchantAddress: "TRTLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
		},
		invoice{
			IpnURI:          "http://localhost:3333",
			CurCode:         "gbp",
			CurAmount:       10.0,
			OrderID:         ":) :DDD\"",
			Custom:          "10",
			PublicKey:       "ad221b7f753f55f95eb4ab1de082388a3dde35c822f458b2658fa2b5eb476cdb",
			MerchantAddress: "TRTLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
		},
		invoice{
			IpnURI:          "http://localhost:3333",
			CurCode:         "gbp",
			CurAmount:       10.0,
			OrderID:         ":) :DDD\"",
			Custom:          "10",
			PublicKey:       "ad221b7f753f55f95eb4ab1de082388a3dde35c822f458b2658fa2b5eb476cdb",
			MerchantAddress: "TRTLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
			ReturnAddress:   "TRTLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
		},
	}

	invalidInv := []invoice{
		// invalid currency code
		invoice{
			IpnURI:          "http://localhost:3333",
			CurCode:         "gb111p",
			CurAmount:       10.0,
			OrderID:         ":) :DDD\"",
			Custom:          "10",
			PublicKey:       "ad221b7f753f55f95eb4ab1de082388a3dde35c822f458b2658fa2b5eb476cdb",
			MerchantAddress: "TRTLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
		},
		// invalid public key
		invoice{
			IpnURI:          "http://localhost:3333",
			CurCode:         "gbp",
			CurAmount:       -10.0,
			OrderID:         ":) :DDD\"",
			Custom:          "10",
			PublicKey:       "ad221b7f753f55f95eb4ab1de082388a3dde35c822f458b2658fa2b5eb476cdb",
			MerchantAddress: "TRTLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
		},
		// invalid currency amount
		invoice{
			IpnURI:          "http://localhost:3333",
			CurCode:         "gbp",
			CurAmount:       -10.0,
			OrderID:         ":) :DDD\"",
			Custom:          "10",
			PublicKey:       "ad753f55f95eb4ab1de082388a3dde35c822f458b2658fa2b5eb476cdb",
			MerchantAddress: "TRTLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
		},
		// invalid merchant address
		invoice{
			IpnURI:          "http://localhost:3333",
			CurCode:         "gbp",
			CurAmount:       1.0,
			OrderID:         ":^)",
			Custom:          "10",
			PublicKey:       "ad753f55f95eb4ab1de082388a3dde35c822f458b2658fa2b5eb476cdb",
			MerchantAddress: "uxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
		},
		// invalid return address
		invoice{
			IpnURI:          "http://localhost:3333",
			CurCode:         "gbp",
			CurAmount:       10.0,
			OrderID:         "",
			Custom:          "10",
			PublicKey:       "ad753f55f95eb4ab1de082388a3dde35c822f458b2658fa2b5eb476cdb",
			MerchantAddress: "TRTLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
			ReturnAddress:   "TRLuxsu1u9HLuYVS4XhkDRxiHEuBJTQ17Y2TRFSQaLtbkgJWYjNGUwMeWSnW4sy3Q9r9yPw8jEsdS9GLysqy4zigZBhCkysg3T",
		},
	}

	for _, inv := range validInv {
		raw, _ := json.Marshal(inv)
		req, _ := http.NewRequest("POST", "/api/invoice", bytes.NewBuffer(raw))
		req.Header.Add("HMAC-SIGNATURE", sign(string(raw), MasterKey))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, req)
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Equal(t, "ok", result.Status)
		txs = append(txs, result.Result["payment_id"])
	}

	for _, inv := range invalidInv {
		raw, _ := json.Marshal(inv)
		req, _ := http.NewRequest("POST", "/api/invoice", bytes.NewBuffer(raw))
		req.Header.Add("HMAC-SIGNATURE", sign(string(raw), MasterKey))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, req)
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.NotEqual(t, "ok", result.Status)
	}
}

func TestGetInvoice(t *testing.T) {
	var result response

	for _, tx := range txs {
		req, _ := http.NewRequest("GET", "/api/invoice/"+tx, bytes.NewBufferString(""))
		req.Header.Add("HMAC-SIGNATURE", sign("", MasterKey))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, req)
		json.Unmarshal(resp.Body.Bytes(), &result)
		assert.Equal(t, "ok", result.Status)
	}

	req, _ := http.NewRequest("GET", "/api/invoice/"+txs[0], bytes.NewBufferString(""))
	req.Header.Add("HMAC-SIGNATURE", sign(":^)", MasterKey))
	resp := httptest.NewRecorder()
	Router().ServeHTTP(resp, req)
	json.Unmarshal(resp.Body.Bytes(), &result)
	assert.NotEqual(t, "ok", result.Status)

}

func TestTXReceived(t *testing.T) {
	for _, tx := range txs {
		mock, _ := json.Marshal(map[string]interface{}{
			"amount":     100000000000,
			"payment_id": tx,
			"block":      "1",
		})
		req, _ := http.NewRequest("POST", "/api/transaction/received", bytes.NewBuffer(mock))
		req.Header.Add("HMAC-SIGNATURE", sign(string(mock), MasterKey))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, req)
		inv, _ := redisHMGet(tx, []string{"status"})
		assert.Equal(t, "processing", inv[0])
		redisDel(tx)
	}
	redisDel("processing_invoices")
}
