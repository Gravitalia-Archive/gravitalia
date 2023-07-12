package router

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/model"
)

// Suspend allows to suspend a user if have
// a valid authorization
func Suspend(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorMethodNotAllowed,
		})
		return
	}

	if req.Header.Get("authorization") != os.Getenv("GLOBAL_AUTH") {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	if !req.URL.Query().Has("vanity") {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidUser,
		})
		return
	}

	is_suspend := true
	if req.URL.Query().Has("suspend") {
		d, err := strconv.ParseBool(req.URL.Query().Get("suspend"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidQuery,
			})
			return
		}
		is_suspend = d
	}

	_, err := database.MakeRequest("MATCH (u:User {name: $id}) SET u.suspended = $suspended;", map[string]interface{}{"id": req.URL.Query().Get("vanity"), "suspended": is_suspend})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: Ok,
	})
}
