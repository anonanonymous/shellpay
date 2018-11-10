package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"./models"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.POST("/api/users", createUser)
	log.Fatal(http.ListenAndServe(hostPort, router))
}

// createUser - creates a new user, stores it in the userDB
func createUser(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var body request
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
		handleError(res, "Error processing request", http.StatusInternalServerError)
		return
	}

	uname, ok := body["username"]
	if !ok {
		handleError(res, "Error processing request: Missing Username", http.StatusBadRequest)
		return
	}
	pwd, ok := body["password"]
	if !ok {
		handleError(res, "Error processing request: Missing Password", http.StatusBadRequest)
		return
	}
	email, ok := body["email"]
	if !ok {
		handleError(res, "Error processing request: Missing Email", http.StatusBadRequest)
		return
	}

	if isRegistered(uname) {
		handleError(res, "Error: Username Taken", http.StatusBadRequest)
		return
	}

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
