package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/ulule/limiter/drivers/middleware/stdlib"

	"github.com/julienschmidt/httprouter"
)

// response - outgoing data format
type response struct {
	Status string            `json:"status"`
	Result map[string]string `json:"result"`
}

// request - incoming data format
type request map[string]string

// httpsRedirect - redirects http to https
func httpsRedirect(res http.ResponseWriter, req *http.Request) {
	newURI := "https://" + req.Host + req.RequestURI
	http.Redirect(res, req, newURI, http.StatusPermanentRedirect)
}

// limit - rate limiter middleware
func limit(h httprouter.Handle, rl *stdlib.Middleware) httprouter.Handle {
	return func(res http.ResponseWriter, req *http.Request, p httprouter.Params) {
		context, err := rl.Limiter.Get(req.Context(), rl.Limiter.GetIPKey(req))
		if err != nil {
			rl.OnError(res, req, err)
			return
		}

		res.Header().Add("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		res.Header().Add("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		res.Header().Add("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			rl.OnLimitReached(res, req)
			return
		}
		res.Header().Set("Access-Control-Allow-Origin", "*")
		req.Body = http.MaxBytesReader(res, req.Body, maxSize) // limit post size
		h(res, req, p)
	}
}

// handleError - writes json response and sets header
func handleError(res http.ResponseWriter, text string, code int) {
	res.WriteHeader(code)
	json.NewEncoder(res).Encode(response{
		Status: text,
	})
}

// isAuthorized - check HMAC signature agints MasterKey
func isAuthorized(req *http.Request, message []byte) bool {
	signature := req.Header.Get("HMAC-SIGNATURE")
	mac, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	return checkMAC(message, mac, []byte(MasterKey))
}

// getBody - retrieves the raw data received in a request
func getBody(req *http.Request) ([]byte, error) {
	rawData, err := ioutil.ReadAll(req.Body)
	return rawData, err
}

// validateRequest - check request hmac signature and body
func validateRequest(req *http.Request) ([]byte, int, error) {
	data, err := getBody(req)
	if err != nil {
		return nil, http.StatusInternalServerError, errors.New("Server Error")
	}
	if ok := isAuthorized(req, data); !ok {
		return nil, http.StatusUnauthorized, errors.New("Unauthorized")
	}
	return data, 200, nil
}

// sendRequest - helper function for sending data with hmac signature
func sendRequest(method, uri, data, key string) (*map[string]interface{}, error) {
	var rawData []byte
	var body map[string]interface{}

	req, err := http.NewRequest(method, uri, bytes.NewBufferString(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("HMAC-SIGNATURE", sign(data, key))
	req.Header.Add("Content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	req.Close = true
	defer resp.Body.Close()

	rawData, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(rawData, &body); err != nil {
		return nil, err
	}

	return &body, nil
}
