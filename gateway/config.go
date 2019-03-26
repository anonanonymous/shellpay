package main

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/ulule/limiter"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
	"github.com/ulule/limiter/drivers/store/memory"

	_ "github.com/lib/pq"
)

const (
	// Forking config.
	addressFormat          = "^TRTL([a-zA-Z0-9]{95}|[a-zA-Z0-9]{183})$"
	divisor        float64 = 100 // This is 100 for TRTL
	transactionFee         = 10  // This is 10 for TRTL

	maxSize      = 1024 // maximum api request size in bytes
	keyValidDays = 1    // the number of days an api key is valid
)

var (
	// MasterKey - the HMAC private key required for request signing
	MasterKey             string
	dbUser, dbPwd         string
	hostURI, hostPort     string
	paymentURI, userURI   string
	templates             *template.Template
	ratelimiter, strictRL *stdlib.Middleware
	gatewayDB             *sql.DB
	logFile               *os.File
)

func init() {
	var err error
	// MasterKey should be the same for all services
	MasterKey = os.Getenv("MASTER_KEY")

	// databse config
	if dbUser = os.Getenv("DB_USER"); dbUser == "" {
		panic("Set the DB_USER env variable")
	}
	if dbPwd = os.Getenv("DB_PWD"); dbPwd == "" {
		panic("Set the DB_PWD env variable")
	}

	gatewayDB, err = sql.Open(
		"postgres",
		"postgres://"+dbUser+":"+dbPwd+"@localhost/signatures?sslmode=disable",
	)
	if err != nil {
		panic(err)
	}
	if err = gatewayDB.Ping(); err != nil {
		panic(err)
	}
	// logging setup
	logFile, err = os.OpenFile("gateway.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	// services config
	if hostURI = os.Getenv("HOST_URI"); hostURI == "" {
		panic("Set the HOST_URI env variable")
	}
	if hostPort = os.Getenv("HOST_PORT"); hostPort == "" {
		panic("Set the HOST_PORT variable")
	}
	if userURI = os.Getenv("USER_URI"); userURI == "" {
		panic("Set the USER_URI env variable")
	}
	if paymentURI = os.Getenv("PAYMENT_URI"); paymentURI == "" {
		panic("Set the PAYMENT_URI env variable")
	}

	// ratelimiter config
	ratelimiter = stdlib.NewMiddleware(limiter.New(memory.NewStore(), limiter.Rate{
		Period: time.Minute * 1,
		Limit:  100,
	}))
	strictRL = stdlib.NewMiddleware(limiter.New(memory.NewStore(), limiter.Rate{
		Period: time.Hour * 24,
		Limit:  10,
	}))

	templates = template.Must(template.ParseGlob("templates/*.html"))

	hostURI += hostPort
	http.DefaultClient.Timeout = time.Second * 5
	go cleanupDB()
}
