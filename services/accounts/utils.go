package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"./models"
)

// response - outgoing data format
type response struct {
	Status string            `json:"status"`
	Result map[string]string `json:"result"`
}

// request - incoming data format
type request map[string]string

// isAuthorized - check HMAC signature
func isAuthorized(req *http.Request, message []byte) bool {
	_, signature, ok := req.BasicAuth()
	if !ok {
		return false
	}

	mac, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	return checkMAC(message, mac, []byte(MasterKey))
}

// isRegistered - check if username is already present in the database
func isRegistered(username string) bool {
	row := userDB.QueryRow("SELECT username FROM users WHERE username = $1;", username)
	err := row.Scan()
	return err != sql.ErrNoRows
}

// insertUser - inserts a user into the database
func insertUser(usr *user.User) error {
	_, err := userDB.Exec(`
					INSERT INTO
					users (username, verifier, ih, email, privateKey)
					VALUES ($1, $2, $3, $4, $5);`,
		usr.Username, usr.Verifier, usr.IH, usr.Email, hex.EncodeToString(usr.PrivateKey),
	)
	return err
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha512.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

// handleError - writes json response and sets header
func handleError(res http.ResponseWriter, text string, code int) {
	res.WriteHeader(code)
	json.NewEncoder(res).Encode(response{
		Status: text,
	})
}

// getBody - retrieves the raw data recieved in a request
func getBody(req *http.Request) ([]byte, error) {
	rawData, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return rawData, nil
}
