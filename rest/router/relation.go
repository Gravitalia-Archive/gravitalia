package router

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// RelationHandler re-routes to the requested handler
func RelationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		Exists(w, r)
	} else if r.Method == http.MethodPost {
		Relation(w, r)
	}
}

// Relation is a route for allowing users to subscribe to each other
// or like posts, depending on the chosen route
func Relation(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	relation := cases.Title(language.English, cases.Compact).String(strings.TrimPrefix(req.URL.Path, "/relation/"))
	if relation == "" || func() bool {
		for _, v := range []string{"Like", "Subscriber", "Block", "Love"} {
			if v == relation {
				return false
			}
		}
		return true
	}() {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidRelation,
		})
		return
	}

	var vanity string
	if req.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	data, err := helpers.CheckToken(req.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
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
			Message: ErrorUnableReadBody,
		})
		return
	}

	var getbody model.SetBody
	err = json.Unmarshal(body, &getbody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidBody,
		})
		return
	}

	if getbody.Id == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidBody,
		})
		return
	}

	isValid, err := database.UserRelation(vanity, getbody.Id, relation)

	if err != nil && strings.Contains(err.Error(), "already") {
		database.UserUnRelation(vanity, getbody.Id, relation)
		jsonEncoder.Encode(model.RequestError{
			Error:   false,
			Message: OkDeletedRelation,
		})
		return
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
		Message: OkCreatedRelation,
	})
}

// Exists handles route to know if a relation
// exists between two nodes
func Exists(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	relation := cases.Title(language.English, cases.Compact).String(strings.TrimPrefix(req.URL.Path, "/relation/"))
	if relation == "" || func() bool {
		for _, v := range []string{"Like", "Subscriber", "Block", "Love"} {
			if v == relation {
				return false
			}
		}
		return true
	}() {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidRelation,
		})
		return
	}

	if req.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	vanity, err := helpers.CheckToken(req.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	target := req.URL.Query().Get("target")
	if target == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidQuery,
		})
		return
	}

	isValid, _ := database.UserRelation(vanity, target, relation)

	existence := "non-existent"
	if isValid {
		existence = "existent"
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: existence,
	})
}
