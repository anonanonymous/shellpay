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
	router.GET("/api/users/:user_id", GetUser)
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
}

func TestGetUser(t *testing.T) {
	userDB.Exec("UPDATE users SET id = 1 WHERE username ='anon';")
	userDB.Exec("UPDATE users SET id = 2 WHERE username ='an on';")
	userDB.Exec("UPDATE users SET id = 3 WHERE username ='username';")
	userDB.Exec("UPDATE users SET id = 4 WHERE username ='anonymous';")
	goodUIDS := []string{"1", "2", "3", "4"}
	badUIDS := []string{"-1", "io2-", "; TRUNCATE users;", "hi"}
	//badUIDS := []string{"-1", "12", "31", "pp"}

	for _, v := range goodUIDS {
		request, _ := http.NewRequest("GET", "/api/users/"+v, nil)
		request.SetBasicAuth("", sign(""))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		assert.Equal(t, 200, resp.Code, "Retrieved User")
	}
	for _, v := range badUIDS {
		request, _ := http.NewRequest("GET", "/api/users/"+v, nil)
		request.SetBasicAuth("", sign(""))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		assert.Equal(t, 404, resp.Code, "No User")
	}
	userDB.Exec("TRUNCATE users;")
}
