package router

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
)

// Subscribers is a route for allow users to subscribe to each other
func Subscribers(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json_encoder := json.NewEncoder(w)

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

	is_valid, err := database.UserRelation(vanity, getbody.Id, "Subscriber")

	if err != nil && err.Error() == "already Subscribered" {
		database.UserUnRelation(vanity, getbody.Id, "Subscriber")
		json_encoder.Encode(model.RequestError{
			Error:   false,
			Message: vanity + " stopped to follow " + getbody.Id,
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
			Message: vanity + " now follow " + getbody.Id,
		})
	}
}
