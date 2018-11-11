package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/anonanonymous/shellpay/services/accounts/models"
	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()

	router.POST("/api/users", CreateUser)
	router.GET("/api/users/:user_id", GetUser)
	router.PUT("/api/users/:user_id", UpdateUser)
	router.DELETE("/api/users/:user_id", DeleteUser)
	log.Fatal(http.ListenAndServe(hostPort, router))
}

// CreateUser - creates a new user, stores it in the userDB
func CreateUser(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var data request
	rawData, err := getJSON(req)
	if err != nil {
		handleError(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if auth := isAuthorized(req, rawData); !auth {
		handleError(res, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err = json.Unmarshal(rawData, &data); err != nil {
		handleError(res, "Malformed data", http.StatusBadRequest)
		return
	}

	uname, ok := data["username"]
	if !ok || uname == "" {
		handleError(res, "Missing username", http.StatusBadRequest)
		return
	}
	if isRegistered(uname) {
		handleError(res, "Username taken", http.StatusBadRequest)
		return
	}

	pwd, ok := data["password"]
	if !ok || pwd == "" {
		handleError(res, "Missing password", http.StatusBadRequest)
		return
	}

	email, ok := data["email"]
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

// GetUser - retrieves a user by their id
func GetUser(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	userid := params.ByName("user_id")
	if userid == "" {
		handleError(res, "Missing user id", http.StatusBadRequest)
		return
	}
	if auth := isAuthorized(req, nil); !auth {
		handleError(res, "Unauthorized", http.StatusUnauthorized)
		return
	}
	usr, err := getUser(userid)
	if err != nil {
		handleError(res, "Not found", http.StatusNotFound)
		return
	}
	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: map[string]string{
			"id":         userid,
			"username":   usr.Username,
			"verifier":   usr.Verifier,
			"email":      usr.Email,
			"identity":   usr.IH,
			"privateKey": hex.EncodeToString(usr.PrivateKey),
			"totpKey":    usr.TOTPKey,
		},
	})
}

// UpdateUser - changes the properties an existing user
func UpdateUser(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	userid := params.ByName("user_id")
	if userid == "" {
		handleError(res, "Missing user id", http.StatusBadRequest)
		return
	}
	if auth := isAuthorized(req, nil); !auth {
		handleError(res, "Unauthorized", http.StatusUnauthorized)
		return
	}
}

// DeleteUser - removes an existing user
func DeleteUser(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	uid := params.ByName("user_id")
	if uid == "" {
		handleError(res, "Missing user id", http.StatusBadRequest)
		return
	}

}
