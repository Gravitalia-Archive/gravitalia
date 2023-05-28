package router

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
)

// UserHandler route /users/* route into the well path
func UserHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/users/")
	if id != "" && req.Method == "GET" {
		Users(w, req)
	} else if id != "" && id == "@me" && req.Method == "DELETE" {
		Delete(w, req)
	} else if id != "" && id == "@me" && req.Method == "PATCH" {
		update(w, req)
	}
}

// isBlockedAccount check if a user (id) is blocked to another one (user)
// and respond with true if a relation (edge) exists
// or with false if no relation exists
func isBlockedAccount(id string, user string) (bool, error) {
	res, err := database.MakeRequest("MATCH (a:User {name: $id})-[:Block]->(b:User {name: $to}) RETURN a;",
		map[string]any{"id": id, "to": user})
	if err != nil {
		return false, err
	}

	return res != nil, nil
}

// Users is the GET route
func Users(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	username := strings.TrimPrefix(req.URL.Path, "/users/")
	if username == "@me" && req.Header.Get("authorization") != "" {
		vanity, err := helpers.CheckToken(req.Header.Get("authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid token",
			})
			return
		}
		username = vanity
	}

	stats, err := database.GetProfile(username)
	if err != nil || stats.Suspended {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid user",
		})
		return
	}

	var allow_post_access bool
	var vanity string
	if stats.Public {
		allow_post_access = true
	} else if !stats.Public && strings.TrimPrefix(req.URL.Path, "/users/") != "@me" && req.Header.Get("authorization") != "" {
		vanity, err = helpers.CheckToken(req.Header.Get("authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid token",
			})
			return
		}

		is, err := database.IsUserSubscrirerTo(vanity, username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid relation",
			})
			return
		}

		allow_post_access = is
	}

	if strings.TrimPrefix(req.URL.Path, "/users/") != "@me" && req.Header.Get("authorization") != "" {
		is, err := isBlockedAccount(vanity, username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid relation Block",
			})
			return
		}

		if is {
			allow_post_access = true
		}
	}

	posts := make([]model.Post, 0)
	if allow_post_access {
		posts, err = database.GetUserPost(username, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid user",
			})
			return
		}
	}

	jsonEncoder.Encode(struct {
		Followers     int64        `json:"followers"`
		Following     int64        `json:"following"`
		Public        bool         `json:"public"`
		Suspended     bool         `json:"suspended"`
		CanAccessPost bool         `json:"access_post"`
		Posts         []model.Post `json:"posts"`
	}{
		Followers:     stats.Followers,
		Following:     stats.Following,
		Public:        stats.Public,
		Suspended:     stats.Suspended,
		CanAccessPost: allow_post_access,
		Posts:         posts,
	})
}

// Delete allows users to delete their account
func Delete(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity := ""
	var err error

	if req.Header.Get("authorization") == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	} else if req.Header.Get("authorization") == os.Getenv("GLOBAL_AUTH") {
		vanity = req.URL.Query().Get("user")
	} else {
		vanity, err = helpers.CheckToken(req.Header.Get("authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid token",
			})
			return
		}
	}

	_, err = database.DeleteUser(vanity)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Internal server error",
		})
		return
	}

	database.Set(vanity+"-gd", "ok", 3600)

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: "OK",
	})
}

// Handle patch method, allows to update user data
func update(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity, err := helpers.CheckToken(req.Header.Get("authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	}

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Unable to get body",
		})
		return
	}

	var getbody model.UpdateBody
	json.Unmarshal(body, &getbody)

	if getbody.Public != nil {
		_, err := database.MakeRequest("MATCH (u:User {name: $id}) SET u.public = $public;", map[string]interface{}{"id": vanity, "public": *getbody.Public})
		if err != nil {
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "OK",
			})
			return
		}
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: "OK",
	})
}
