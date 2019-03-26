package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

// validateAddress - check payment address format
func validateAddress(addr string) bool {
	addr = strings.TrimSpace(addr)
	matched, err := regexp.MatchString(addressFormat, addr)
	return err == nil && matched
}

// sign - returns a HMAC signature for a message
func sign(message, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
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
