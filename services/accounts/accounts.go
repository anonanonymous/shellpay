package main

import (
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
	router.GET("/api/user/id/:username", GetUserID)
	router.PUT("/api/users/:user_id/:setting", UpdateUser)
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
		handleError(res, "Username taken", http.StatusMethodNotAllowed)
		return
	}

	pwd, ok := data["password"]
	if !ok || pwd == "" {
		handleError(res, "Missing password", http.StatusBadRequest)
		return
	}

	email, ok := data["email"]
	if !ok {
		handleError(res, "Missing email", http.StatusBadRequest)
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
			"username": usr.Username,
			"verifier": usr.Verifier,
			"email":    usr.Email,
			"identity": usr.IH,
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
		Result: usr.Jsonify(),
	})
}

// GetUserID - retrieves the id from of a user
func GetUserID(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	if auth := isAuthorized(req, nil); !auth {
		handleError(res, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userid, err := getUserID(params.ByName("username"))
	if err != nil {
		handleError(res, "Not found", http.StatusNotFound)
		return
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: map[string]string{
			"id": userid,
		},
	})
}

// UpdateUser - changes the properties an existing user
func UpdateUser(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	var body request
	var uid string

	uid = params.ByName("user_id")
	usr, err := getUser(uid)
	if err != nil {
		handleError(res, "Not found", http.StatusNotFound)
		return
	}

	rawData, err := getJSON(req)
	if err != nil {
		handleError(res, err.Error(), http.StatusInternalServerError)
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

	switch params.ByName("setting") {
	case "password":
		pwd, ok := body["password"]
		if auth, err := usr.Verify(pwd); !ok || !auth || err != nil {
			handleError(res, "Unauthorized", http.StatusUnauthorized)
			return
		}
		pwd, ok = body["new_password"]
		if !ok {
			handleError(res, "Invalid new password", http.StatusInternalServerError)
			return
		}
		usr, err = user.NewUser(usr.Username, pwd, usr.Email)
		if err != nil {
			handleError(res, err.Error(), http.StatusInternalServerError)
			return
		}
	case "email":
		if !usr.SetEmail(body["email"]) {
			handleError(res, "Invalid email", http.StatusBadRequest)
			return
		} else if _, ok := body["email"]; !ok {
			handleError(res, "No email provided", http.StatusBadRequest)
			return
		}
	case "two_factor":
		val, ok := body["active"]
		if !ok {
			handleError(res, "Missing setting", http.StatusBadRequest)
			return
		}
		code, ok := body["code"]
		if !ok {
			handleError(res, "Missing code", http.StatusBadRequest)
			return
		}
		if secret, ok := body["secret"]; val == "true" && ok {
			if err = usr.EnableTwoFactor(secret, code); err != nil {
				handleError(res, err.Error(), http.StatusForbidden)
				return
			}
		} else if val == "false" {
			if err := usr.DisableTwoFactor(code); err != nil {
				handleError(res, err.Error(), http.StatusForbidden)
				return
			}
		} else {
			handleError(res, "Invalid parameter", http.StatusBadRequest)
			return
		}
	default:
		handleError(res, "Invalid setting", http.StatusBadRequest)
		return
	}

	if err := updateUser(usr, uid); err != nil {
		handleError(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: usr.Jsonify(),
	})
}

// DeleteUser - removes an existing user
func DeleteUser(res http.ResponseWriter, req *http.Request, params httprouter.Params) {
	uid := params.ByName("user_id")
	if uid == "" {
		handleError(res, "Missing user id", http.StatusBadRequest)
		return
	}
	if auth := isAuthorized(req, nil); !auth {
		handleError(res, "Unauthorized", http.StatusUnauthorized)
		return
	}

	usr, err := getUser(uid)
	if err != nil {
		handleError(res, "Not found", http.StatusNotFound)
		return
	}
	if err = deleteUser(uid); err != nil {
		handleError(res, "Internal error", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	json.NewEncoder(res).Encode(response{
		Status: "ok",
		Result: usr.Jsonify(),
	})
}
