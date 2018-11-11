package main

import (
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

var (
	// MasterKey - the HMAC private key required for request signing
	MasterKey             string
	dbUser, dbPwd, dbName string
	hostURI, hostPort     string
	userDB                *sql.DB
)

func init() {
	var err error
	MasterKey = os.Getenv("MASTER_KEY")

	if dbUser = os.Getenv("DB_USER"); dbUser == "" {
		panic("Set the DB_USER env variable")
	}
	if dbPwd = os.Getenv("DB_PWD"); dbPwd == "" {
		panic("Set the DB_PWD env variable")
	}
	if dbName = os.Getenv("DB_NAME"); dbName == "" {
		panic("Set the DB_NAME env variable")
	}

	userDB, err = sql.Open("postgres", "postgres://"+dbUser+":"+dbPwd+"@localhost/"+dbName+"?sslmode=disable")
	if err != nil {
		panic(err)
	}
	if err = userDB.Ping(); err != nil {
		panic(err)
	}

	if hostURI = os.Getenv("HOST_URI"); hostURI == "" {
		panic("Set the HOST_URI env variable")
	}
	if hostPort = os.Getenv("HOST_PORT"); hostPort == "" {
		panic("Set the HOST_PORT variable")
	}

	hostURI += hostPort
}
