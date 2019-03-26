package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// invoice
type invoice struct {
	ReturnAddress   string  `json:"return_address"`
	IpnURI          string  `json:"ipn_uri"`
	CurCode         string  `json:"currency_code"`
	CurAmount       float64 `json:"currency_amount"`
	AtomicAmount    int64   `json:"atomic_amount"`
	OrderID         string  `json:"order_id"`
	Custom          string  `json:"custom"`
	PublicKey       string  `json:"public_key"`
	MerchantAddress string  `json:"merchant_address"`
	Timestamp       int64
	Expiry          int64
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

// sign - returns a HMAC signature for a message
func sign(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// validateAddress - check payment address format
func validateAddress(addr string) bool {
	addr = strings.TrimSpace(addr)
	matched, err := regexp.MatchString(addressFormat, addr)
	return err == nil && matched
}

// randHexStr - creates a string of random hex values
func randHexStr(len uint) (string, error) {
	buf := make([]byte, len)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// checkInvoice - verifies that required fields are provided and valid in an invoice struct
func checkInvoice(inv *invoice) bool {
	inv.CurCode = strings.ToLower(inv.CurCode)

	if _, ok := exchangeRates[inv.CurCode]; !ok {
		return false
	}
	if inv.CurAmount <= 0 {
		return false
	}
	if !validateAddress(inv.MerchantAddress) {
		return false
	}
	if inv.IpnURI == "" || inv.OrderID == "" || len(inv.PublicKey) != pubKeylen {
		return false
	}

	// redis can't use an empty string as a value, "-" is a placeholder
	if len(inv.ReturnAddress) == 0 {
		inv.ReturnAddress = "-"
	} else {
		return validateAddress(inv.ReturnAddress)
	}

	return true
}
