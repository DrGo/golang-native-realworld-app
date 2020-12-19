package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func check(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func regexMatch(regex, text string) bool {
	matched, err := regexp.MatchString(regex, text)
	if err != nil {
		panic(err)
	}
	return matched
}

var unauthRequestTests = []struct {
	init           func(*http.Request)
	url            string
	method         string
	bodyData       string
	expectedCode   int
	responseRegexg string
	msg            string
}{
	//Testing will run one by one, so you can combine it to a user story till another init().
	//And you can modified the header or body in the func(req *http.Request) {}

	//---------------------   Testing for user register   ---------------------
	{
		func(req *http.Request) {},
		"/api/users/",
		"POST",
		`{"user":{"username": "wangzitian0","email": "wzt@gg.cn","password": "jakejxke"}}`,
		http.StatusCreated,
		`{"user":{"username":"wangzitian0","email":"wzt@gg.cn","bio":"","image":null,"token":"([a-zA-Z0-9-_.]{115})"}}`,
		"valid data and should return StatusCreated",
	},
}

func TestWithoutAuth(t *testing.T) {
	s := &server{}
	for _, td := range unauthRequestTests {
		bodyData := td.bodyData
		req, err := http.NewRequest(td.method, td.url, bytes.NewBufferString(bodyData))
		check(t, err)
		req.Header.Set("Content-Type", "application/json")
		td.init(req)
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, req)
		if rec.Code != td.expectedCode {
			t.Errorf("wrong status code, got %v wanted %v", rec.Code, td.expectedCode)
		}
		body := rec.Body.String()
		if !regexMatch(td.responseRegexg, body) {
			t.Errorf("wrong content, got %v wanted %v", body, td.responseRegexg)
		}
	}
}
