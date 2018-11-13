package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func Router() *httprouter.Router {
	router := httprouter.New()
	router.POST("/api/users", CreateUser)
	router.GET("/api/users/:user_id", GetUser)
	router.GET("/api/user/id/:username", GetUserID)
	router.PUT("/api/users/:user_id/:setting", UpdateUser)
	return router
}

func TestCreateUser(t *testing.T) {
	userDB.Exec("TRUNCATE users;")
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
	badCode := []int{400, 400, 400, 405}

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
		assert.Equal(t, badCode[i], resp.Code, "Failed to Create User")
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
}

func TestGetUserID(t *testing.T) {
	goodUIDS := []string{"anon", "an on", "username", "anonymous"}
	badUIDS := []string{"-1", "io2-", "; TRUNCATE users;", "0"}

	for i, v := range goodUIDS {
		request, _ := http.NewRequest("GET", "/api/user/id/"+v, nil)
		request.SetBasicAuth("", sign(""))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		assert.Equal(t,
			`{"status":"ok","result":{"id":"`+strconv.Itoa(i+1)+"\"}}\n",
			resp.Body.String(), "Failed to query user",
		)
		assert.Equal(t, 200, resp.Code, "Retrieved user id")
	}

	for _, v := range badUIDS {
		request, _ := http.NewRequest("GET", "/api/user/id/"+v, nil)
		request.SetBasicAuth("", sign(""))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		assert.Equal(t, 404, resp.Code, "No user with this id")
	}
}

func TestUpdateUser(t *testing.T) {
	var body response
	goodSettings := [][]string{
		{"1", "password", `{"password":"pass", "new_password":"pass101"}`},
		{"2", "email", `{"password":"pass", "email":"new@email.com"}`},
		{"3", "two_factor", `{"password":"pass", "two_factor":"true"}`},
		{"3", "two_factor", `{"password":"pass", "two_factor":"false"}`},
		{"4", "email", `{"password":"h4ck3r", "email":"pass@101.co"}`},
	}

	badSettings := [][]string{
		{"1", "password", `{"password":"pass1", "new_password":"pass101"}`},
		{"1", "password", `{"password":"pass1", "new_password":""}`},
		{"2", "email", `{"password":"pass", "email":"newemailcom"}`},
		{"2", "email", `{"password":"pass", "":"newemailcom"}`},
		{"3", "two_factor", `{"password":"pass", "two_factor":"false"}`},
		{"3", "two_factor", `{"password":"pass", "two_factor":"fa"}`},
		{"3", "two_factor", `{"password":"pass", "tw_factor":"fa"}`},
		{"4", "email", `assword":"h4ck3r", "email":"pass@101.co"}`},
		{"4", "eail", `assword":"h4ck3r", "email":"pass@101.co"}`},
		{"4", "email", `password":"h4ck3r", "emal":"pass@101.co"}`},
	}
	badCode := []int{401, 401, 400, 400, 409, 400, 400, 400, 400, 400}

	for _, v := range goodSettings {
		request, _ := http.NewRequest("PUT", "/api/users/"+v[0]+"/"+v[1], bytes.NewBufferString(v[2]))
		request.SetBasicAuth("", sign(v[2]))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		json.Unmarshal([]byte(resp.Body.String()), &body)
		assert.Len(t, body.Result["privateKey"], 64, "Valid length")
		assert.Equal(t, 200, resp.Code, "User setting changed")
	}

	for i, v := range badSettings {
		request, _ := http.NewRequest("PUT", "/api/users/"+v[0]+"/"+v[1], bytes.NewBufferString(v[2]))
		request.SetBasicAuth("", sign(v[2]))
		resp := httptest.NewRecorder()
		Router().ServeHTTP(resp, request)
		json.Unmarshal([]byte(resp.Body.String()), &body)
		assert.Equal(t, badCode[i], resp.Code, "User setting changed")
	}
	userDB.Exec("TRUNCATE users;")
}
