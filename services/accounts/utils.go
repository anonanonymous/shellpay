package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/anonanonymous/shellpay/services/accounts/models"
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

// updateUser - updates user properties
func updateUser(usr *user.User, id string) error {
	var totpKey *string
	if usr.TwoFactorEnabled() {
		totpKey = &usr.TOTPKey
	}
	_, err := userDB.Exec(`
					UPDATE users
					SET
					email = $1,
					verifier = $2,
					totpKey = $3,
					privateKey = $4
					WHERE id = $5;`,
		usr.Email, usr.Verifier, totpKey, hex.EncodeToString(usr.PrivateKey), id,
	)
	return err
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
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
	return rawData, err
}

// sign - returns a HMAC signature for a message
func sign(message string) string {
	mac := hmac.New(sha256.New, []byte(MasterKey))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// getJSON - gets the json from a request
func getJSON(req *http.Request) ([]byte, error) {
	rawData, err := getBody(req)
	if err != nil {
		return nil, errors.New("Internal error")
	}
	return rawData, nil
}

// getUser - retrieves a user from the database
func getUser(userid string) (*user.User, error) {
	var email, totp []byte

	row := userDB.QueryRow(`
			SELECT ih, verifier, username, email, privateKey, totpKey
			FROM users
			WHERE id = $1;`, userid,
	)
	usr := user.User{}
	err := row.Scan(
		&usr.IH, &usr.Verifier, &usr.Username,
		&email, &usr.PrivateKey, &totp,
	)
	if err != nil {
		return nil, err
	}

	usr.PrivateKey, err = hex.DecodeString(string(usr.PrivateKey))
	if err != nil {
		return nil, err
	}
	if email != nil {
		usr.Email = string(email)
	}
	if totp != nil {
		usr.TOTPKey = string(totp)
	}

	return &usr, nil
}

// getUserID - retrieves a user id from the database
func getUserID(username string) (string, error) {
	var uid string

	row := userDB.QueryRow(`
			SELECT id
			FROM users
			WHERE username = $1;`, username,
	)
	err := row.Scan(&uid)
	return uid, err
}

// deleteUser - deletes a user from the database
func deleteUser(userid string) error {
	_, err := userDB.Exec(`
			DELETE FROM
			users
			WHERE id = $1;`, userid,
	)
	return err
}
