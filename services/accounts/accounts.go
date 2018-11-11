package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/anonanonymous/shellpay/services/accounts/models"
	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.POST("/api/users", CreateUser)
	log.Fatal(http.ListenAndServe(hostPort, router))
}

// CreateUser - creates a new user, stores it in the userDB
func CreateUser(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var body request
	ctx := req.Context()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	rawData, err := getBody(req)
	if err != nil {
		handleError(res, "Internal error", http.StatusInternalServerError)
		return
	}

	if auth := isAuthorized(req, rawData); !auth {
		handleError(res, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err = json.Unmarshal(rawData, &body); err != nil {
		handleError(res, "Malformed data", http.StatusBadRequest)
		return
	}

	uname, ok := body["username"]
	if !ok || uname == "" {
		handleError(res, "Missing username", http.StatusBadRequest)
		return
	}
	if isRegistered(uname) {
		handleError(res, "Username taken", http.StatusBadRequest)
		return
	}
	pwd, ok := body["password"]
	if !ok || pwd == "" {
		handleError(res, "Missing password", http.StatusBadRequest)
		return
	}

	email, ok := body["email"]
	usr, err := user.NewUser(uname, pwd, email)
	if err != nil {
		handleError(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := insertUser(usr); err != nil {
		handleError(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: map[string]string{
			"username":   usr.Username,
			"verifier":   usr.Verifier,
			"email":      usr.Email,
			"identity":   usr.IH,
			"privateKey": hex.EncodeToString(usr.PrivateKey),
		},
	})
}
