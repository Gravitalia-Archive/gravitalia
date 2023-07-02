package router

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/model"

	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// randomString generates a random character string with a predefined number
func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// makeRequest allows to make requests and return the body
func makeRequest(zipkinClient *zipkinhttp.Client, url string, method string, reqBody io.Reader, authHeader string) ([]byte, error) {
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, errors.New("unable to create request")
	}

	if authHeader != "" {
		req.Header.Add("Authorization", authHeader)
	}

	response, err := zipkinClient.Do(req)
	if err != nil {
		return nil, errors.New("unable to make request")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("unable to read request")
	}

	return body, nil
}

// OAuth handles requests for connections, and will grant a Json Web Token
// or redirect the user to the public data sharing acceptance page.
func OAuth(zipkinClient *zipkinhttp.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if req.URL.Query().Has("state") && req.URL.Query().Has("code") {
			jsonEncoder := json.NewEncoder(w)

			val, err := database.Mem.Get(req.URL.Query().Get("state"))
			if err != nil || string(val.Value) != "ok" {
				state := randomString(24)
				database.Set(state, "ok", 500)
				http.Redirect(w, req, os.Getenv("OAUTH_HOST")+"/oauth2/authorize?client_id=suba&scope=identity&redirect_uri=https://api.gravitalia.com/callback&response_type=code&state="+state, http.StatusTemporaryRedirect)
			} else {
				postBody, _ := json.Marshal(struct {
					ClientId     string `json:"client_id"`
					ClientSecret string `json:"client_secret"`
					Code         string `json:"code"`
					RedirectUri  string `json:"redirect_uri"`
				}{
					ClientId:     "suba",
					ClientSecret: os.Getenv("SECRET"),
					Code:         req.URL.Query().Get("code"),
					RedirectUri:  os.Getenv("REDIRECT_URL"),
				})

				body, err := makeRequest(zipkinClient, os.Getenv("OAUTH_API")+"/oauth2/token", "POST", bytes.NewBuffer(postBody), "")
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					jsonEncoder.Encode(model.RequestError{
						Error:   true,
						Message: "Internal error:" + err.Error(),
					})
					return
				}
				var data model.RequestError
				json.Unmarshal(body, &data)
				if data.Error {
					w.WriteHeader(http.StatusBadRequest)
					jsonEncoder.Encode(model.RequestError{
						Error:   true,
						Message: "Invalid code",
					})
					return
				}

				body, err = makeRequest(zipkinClient, os.Getenv("OAUTH_API")+"/users/@me", "GET", nil, data.Message)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					jsonEncoder.Encode(model.RequestError{
						Error:   true,
						Message: "Internal error:" + err.Error(),
					})
					return
				}
				var user model.AuthaUser
				json.Unmarshal(body, &user)
				if user.Vanity == "" {
					w.WriteHeader(http.StatusBadRequest)
					jsonEncoder.Encode(model.RequestError{
						Error:   true,
						Message: "Invalid code",
					})
					return
				}

				// Check if account has been deleted 1 hour ago
				val, _ = database.Mem.Get(user.Vanity + "-gd")
				if val != nil && string(val.Value) == "ok" {
					w.WriteHeader(http.StatusBadRequest)
					jsonEncoder.Encode(model.RequestError{
						Error:   true,
						Message: "Account deleted too soon",
					})
					return
				}

				database.CreateUser(user.Vanity)

				// Add user into document in case of search
				documentUser, _ := json.Marshal(struct {
					Vanity   string `json:"vanity"`
					Username string `json:"username"`
					Flags    int    `json:"flags"`
				}{
					Vanity:   user.Vanity,
					Username: user.Username,
					Flags:    user.Flags,
				})
				go makeRequest(zipkinClient, os.Getenv("SEARCH_API")+"/search/add", "POST", bytes.NewBuffer(documentUser), os.Getenv("GLOBAL_AUTH"))

				http.Redirect(w, req, "https://www.gravitalia.com/callback?token="+data.Message, http.StatusTemporaryRedirect)
			}
		} else {
			state := randomString(24)
			database.Set(state, "ok", 500)
			http.Redirect(w, req, os.Getenv("OAUTH_HOST")+"/oauth2/authorize?client_id=suba&scope=identity&redirect_uri=https://api.gravitalia.com/callback&response_type=code&state="+state, http.StatusTemporaryRedirect)
		}
	}
}
