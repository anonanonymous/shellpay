package main

import (
	"bytes"
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
		`{"username":"anon2", "password":"", "email":"justbe@your.self"}`,
		`{"username":"", "password":"pass", "email":"justbe@your.self"}`,
		`{ame":"", "password":"pass", "email":"justbe@your.self"}`,
		`{"username":"anonymous", "password":"h4ck3r", "email":""}`, // dupliate db entry
	}
	badRes := []string{"Missing password", "Missing username", "Malformed data", "Username taken"}

	for _, v := range goodJSON {
		request, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(v))
		request.SetBasicAuth("", sign(v))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		assert.Equal(t, 201, resp.Code, "Created User")
	}

	for i, v := range badJSON {
		request, _ := http.NewRequest("POST", "/api/users", bytes.NewBufferString(v))
		request.SetBasicAuth("", sign(v))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		assert.Equal(t, 400, resp.Code, "Failed to Create User")
		assert.Equal(t, `{"status":"`+badRes[i]+`","result":null}`+"\n", resp.Body.String())
	}
	userDB.Exec("TRUNCATE users;")
}
