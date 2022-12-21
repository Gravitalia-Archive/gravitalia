package router

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/model"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func OAuth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if req.URL.Query().Has("state") && req.URL.Query().Has("code") {
		val, err := database.Mem.Get(req.URL.Query().Get("state"))
		if err != nil || string(val.Value) != "ok" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(model.Error{
				Error:   true,
				Message: "Invalid state",
			})
		} else {
			postBody, _ := json.Marshal(map[string]string{
				"client_id":     "suba",
				"client_secret": os.Getenv("secret"),
				"code":          req.URL.Query().Get("code"),
				"redirect_uri":  "https://www.gravitalia.com/callback",
			})
			resp, err := http.Post("http://localhost:1111/oauth2/token", "application/json", bytes.NewBuffer(postBody))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(model.Error{
					Error:   true,
					Message: "Internal error: unable to make request",
				})
			}
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(model.Error{
					Error:   true,
					Message: "Internal error: unable to read request",
				})
			}
			data := model.Request{}
			json.Unmarshal(body, &data)

			client := &http.Client{}
			req, err := http.NewRequest("GET", "http://localhost:1111/users/@me", nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(model.Error{
					Error:   true,
					Message: "Internal error: unable to make request",
				})
			}

			req.Header.Add("Authorization", data.Message)
			response, err := client.Do(req)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(model.Error{
					Error:   true,
					Message: "Internal error: unable to make request",
				})
			}
			defer response.Body.Close()
			body, err = ioutil.ReadAll(response.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(model.Error{
					Error:   true,
					Message: "Internal error: unable to read request",
				})
			}
			user := model.AuthaUser{}
			json.Unmarshal(body, &user)

			// Create a JWT token with user.Vanity
		}
	} else {
		state := randomString(24)
		database.Set(state, "ok")
		http.Redirect(w, req, "https://account.gravitalia.com/oauth2/authorize?response_type=code&client_id=suba&scope=user&state="+state, http.StatusTemporaryRedirect)
	}
}
