package router

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ListHandler routes to the right function
func ListHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		Index(w, req)
	} else if req.Method == http.MethodGet {
		getList(w, req)
	}
}

// getList allows to return a user or post list
// based on the wanted list
func getList(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	authToken := req.Header.Get("authorization")

	// Check token
	if authToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	vanity, err := helpers.CheckToken(authToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	id := cases.Title(language.English, cases.Compact).String(strings.TrimPrefix(req.URL.Path, "/list/"))
	if id == "" || func() bool {
		for _, v := range []string{"Subscriber", "Subscription", "Block"} {
			if v == id {
				return false
			}
		}
		return true
	}() {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidList,
		})
		return
	}

	var list []any
	if id == "Subscription" {
		users, err := database.MakeRequest("MATCH (:User {name: $id})-[:Subscriber]->(u:User) RETURN u.name;",
			map[string]any{"id": vanity})
		list = users.([]any)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorWithDatabase,
			})
			return
		}
	} else {
		users, err := database.MakeRequest("MATCH (u:User)-[:"+id+"]->(:User {name: $id}) RETURN u.name;",
			map[string]any{"id": vanity})
		list = users.([]any)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorWithDatabase,
			})
			return
		}
	}

	jsonEncoder.Encode(list)
}
