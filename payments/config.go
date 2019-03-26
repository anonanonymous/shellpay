package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	// wallet address format
	addressFormat = "^TRTL([a-zA-Z0-9]{95}|[a-zA-Z0-9]{183})$"
	// smallest divisible unit
	atoms = 100
	// amount of time an invoice is active
	paymentExpire = time.Minute * 60
	// langth of api public key
	pubKeylen = 64
	// lenth of wallet address
	addrLen = 99
	// length of integrated wallet address
	intaddrLen = 187
	// how long in seconds invoice info is available after expiration
	expDelay = 60
	// dev fee percentage
	feeRate = 0.0
	// the number of confirms required for a payment to be verified
	minConfirms = 1
)

var (
	// MasterKey - the HMAC private key for request signing across services
	MasterKey, addr   string
	hostURI, hostPort string
	gatewayURI        string
	walletHost        string
	walletPort        string
	nodeFee           int64
	invoiceDB         *redis.Pool
	wallet            *Wallet
	logFile           *os.File
)

func init() {
	var err error
	MasterKey = os.Getenv("MASTER_KEY")
	redisPort := os.Getenv("REDIS_PORT")

	if redisPort == "" {
		redisPort = ":6379"
	}
	if hostURI = os.Getenv("HOST_URI"); hostURI == "" {
		panic("Set the HOST_URI variable.")
	}
	if hostPort = os.Getenv("HOST_PORT"); hostPort == "" {
		panic("Set the HOST_PORT variable")
	}
	if gatewayURI = os.Getenv("GATEWAY_URI"); gatewayURI == "" {
		panic("Set the GATEWAY_URI variable.")
	}
	if walletPort = os.Getenv("WALLET_PORT"); walletPort == "" {
		panic("Set the WALLET_PORT variable to the wallet port number")
	}
	if walletHost = os.Getenv("WALLET_HOST"); walletHost == "" {
		panic("Set the WALLET_HOST variable.")
	}

	logFile, err = os.OpenFile("payments.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	wallet, err = NewWallet(walletHost, walletPort)
	if err != nil {
		panic("Error communication with wallet: " + err.Error())
	}
	println("wallet address:", wallet.Address)
	println("node Fee:", wallet.NodeFee)

	hostURI += hostPort
	http.DefaultClient.Timeout = time.Second * 5
	invoiceDB = newPool(redisPort)
	cleanupHook()
}
