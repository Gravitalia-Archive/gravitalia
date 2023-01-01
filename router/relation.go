package router

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
)

// contains allows to check if a map of string contains
// a particular string
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// Relation is a route for allow users to subscribe to each other
// or like posts, depending on route chosen
func Relation(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json_encoder := json.NewEncoder(w)

	relation := strings.TrimPrefix(req.URL.Path, "/relation/")
	fmt.Println(relation)
	if relation == "" || !contains([]string{"like", "subscribe", "block"}, relation) {
		w.WriteHeader(http.StatusBadRequest)
		json_encoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid relation type",
		})
		return
	}

	switch relation {
	case "like":
		relation = "Like"
	case "subscribe":
		relation = "Subscriber"
	case "block":
		relation = "Block"
	}
	fmt.Println(relation)

	var vanity string
	if req.Header.Get("authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json_encoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	} else {
		data, err := helpers.CheckToken(req.Header.Get("authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json_encoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid token",
			})
			return
		}
		vanity = data
	}

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json_encoder.Encode(model.RequestError{
			Error:   true,
			Message: "Unable to get body",
		})
		return
	}

	var getbody model.SetBody
	json.Unmarshal(body, &getbody)

	if getbody.Id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json_encoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid body",
		})
		return
	}

	is_valid, err := database.UserRelation(vanity, getbody.Id, relation)

	if err != nil && strings.Contains(err.Error(), "already") {
		database.UserUnRelation(vanity, getbody.Id, relation)
		json_encoder.Encode(model.RequestError{
			Error:   false,
			Message: "OK: Deleted relation",
		})
	} else if err != nil || !is_valid {
		w.WriteHeader(http.StatusBadRequest)
		json_encoder.Encode(model.RequestError{
			Error:   true,
			Message: err.Error(),
		})
		return
	} else {
		json_encoder.Encode(model.RequestError{
			Error:   false,
			Message: "OK: Create relation",
		})
	}
}
