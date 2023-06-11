package router

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
)

// contains checks if a string is present in a slice of strings
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// RelationHandler re-route to the requested handler
func RelationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		Exists(w, r)
	} else if r.Method == "POST" {
		Relation(w, r)
	}
}

// Relation is a route for allowing users to subscribe to each other
// or like posts, depending on the chosen route
func Relation(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	relation := strings.TrimPrefix(req.URL.Path, "/relation/")
	if relation == "" || !contains([]string{"like", "subscribe", "block", "love"}, relation) {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid relation type",
		})
		return
	}

	switch relation {
	case "like":
		relation = "Like"
	case "love":
		relation = "Love"
	case "subscribe":
		relation = "Subscriber"
	case "block":
		relation = "Block"
	}

	var vanity string
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	}

	data, err := helpers.CheckToken(authHeader)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	}
	vanity = data

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Unable to read body",
		})
		return
	}

	var getbody model.SetBody
	err = json.Unmarshal(body, &getbody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid body",
		})
		return
	}

	if getbody.Id == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid body: missing ID",
		})
		return
	}

	isValid, err := database.UserRelation(vanity, getbody.Id, relation)

	if err != nil && strings.Contains(err.Error(), "already") {
		database.UserUnRelation(vanity, getbody.Id, relation)
		jsonEncoder.Encode(model.RequestError{
			Error:   false,
			Message: "OK: Deleted relation",
		})
	} else if err != nil || !isValid {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: err.Error(),
		})
		return
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: "OK: Created relation",
	})
}

// Exists handle route to know if a
func Exists(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	relation := strings.TrimPrefix(req.URL.Path, "/relation/")
	if relation == "" || !contains([]string{"like", "subscribe", "block", "love"}, relation) {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid relation type",
		})
		return
	}

	switch relation {
	case "like":
		relation = "Like"
	case "love":
		relation = "Love"
	case "subscribe":
		relation = "Subscriber"
	case "block":
		relation = "Block"
	}

	if req.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	}

	vanity, err := helpers.CheckToken(req.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	}

	target := req.URL.Query().Get("target")
	if target == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid query: missing target",
		})
		return
	}

	isValid, err := database.UserRelation(vanity, target, relation)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: err.Error(),
		})
		return
	}

	existence := "non-existent"
	if isValid {
		existence = "existent"
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: existence,
	})
}
