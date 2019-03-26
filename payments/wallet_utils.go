package main

import (
	"errors"
	"fmt"
)

// Wallet - communicates with the wallet service
type Wallet struct {
	Host, Port, Address string
	NodeFee             int64
}

// Transaction - hold transaction amount and address
type Transaction struct {
	Amount    int64
	Address   string
	PaymentID string
}

// NewWallet - creates a Wallet instance to communicate with the wallet service api
func NewWallet(host, port string) (*Wallet, error) {
	resp, err := sendRequest("GET", host+":"+port+"/wallet", "", MasterKey)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Host:    host,
		Port:    port,
		Address: (*resp)["address"].(string),
		NodeFee: int64((*resp)["node_fee"].(float64)),
	}, nil
}

// newIntegratedAddress - creates an integrated address
func (w *Wallet) newIntegratedAddress(paymentID string) (string, error) {
	resp, err := sendRequest(
		"POST", w.Host+":"+w.Port+"/wallet/integrated_address",
		"{\"payment_id\": \""+paymentID+"\"}", // manually creating a json payload.
		MasterKey,
	)
	if err != nil {
		return "", err
	}

	return (*resp)["integrated_address"].(string), nil
}

// sendTransaction - sends funds to the specified list of address, amounts pairs
func (w *Wallet) sendTransaction(transactions []Transaction) error {
	if len(transactions) == 0 {
		return nil
	}

	// building payload.. this doesn't feel right, should change
	payload := "{\"transfers\":["
	for i, tx := range transactions {
		payload += fmt.Sprintf("[\"%s\", %d]", tx.Address, tx.Amount)
		if i+1 < len(transactions) {
			payload += ", "
		}
	}
	payload += "]}"

	resp, err := sendRequest("POST", w.Host+":"+w.Port+"/wallet/send_transaction", payload, MasterKey)
	if err != nil {
		return err
	}
	if (*resp)["status"].(string) != "ok" {
		return errors.New("Failed to send transaction: " + (*resp)["status"].(string))
	}

	return nil
}

// getStatus - gets the wallet status. (block height, balance, is_synced)
func (w *Wallet) getStatus() (int64, int64, bool, error) {
	resp, err := sendRequest("GET", w.Host+":"+w.Port+"/wallet/status", "", MasterKey)
	if err != nil {
		return 0, 0, false, err
	}
	return int64((*resp)["block"].(float64)),
		int64((*resp)["balance"].(float64)),
		(*resp)["is_synced"].(bool), nil
}
