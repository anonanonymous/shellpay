package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/julienschmidt/httprouter"
)

func Router() *httprouter.Router {
	router := httprouter.New()
	router.POST("/api/users", CreateUser)
	return router
}

func TestCreateUser(t *testing.T) {
	goodJSON := []string{
		`{"username":"anon", "password":"pass", "email":"justbe@your.self"}`,
		`{"username":"an on", "password":"pass", "email":"justbe@your.self"}`,
		`{"username":"username", "password":"pass", "email":"turtle@me.io"}`,
		`{"username":"anonymous", "password":"h4ck3r", "email":""}`,
	}
	badJSON := []string{
		`{"username":"anon", "password":"", "email":"justbe@your.self"}`,
		`{"username":"", "password":"pass", "email":"justbe@your.self"}`,
		`{"username":"anonymous", "password":"h4ck3r", "email":""}`, // dupliate db entry
	}
	badResp := []int{400, 400, 400}

	for _, v := range goodJSON {
		request, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(v))
		request.SetBasicAuth("", sign(v))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		assert.Equal(t, 201, resp.Code, "OK Created User")
	}

	for i, v := range badJSON {
		request, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(v))
		request.SetBasicAuth("", sign(v))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		fmt.Println(resp)
		assert.Equal(t, badResp[i], resp.Code, "Failed to Create User")
	}
	userDB.Exec("TRUNCATE users;")
}

// sign - returns a HMAC signature for a message
func sign(message string) string {
	mac := hmac.New(sha512.New, []byte(MasterKey))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
