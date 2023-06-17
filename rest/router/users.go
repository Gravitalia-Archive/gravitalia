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

func UserHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/users/")
	if id != "" && req.Method == http.MethodGet {
		Users(w, req)
	} else if id != "" && id == ME && req.Method == http.MethodDelete {
		Delete(w, req)
	} else if id != "" && id == ME && req.Method == http.MethodPatch {
		update(w, req)
	}
}

func Users(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	username := strings.TrimPrefix(req.URL.Path, "/users/")
	if username == ME && req.Header.Get("Authorization") != "" {
		vanity, err := helpers.CheckToken(req.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidToken,
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
			Message: ErrorInvalidUser,
		})
		return
	}

	var viewer_follows bool
	if req.Header.Get("Authorization") != "" {
		vanity, err := helpers.CheckToken(req.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidToken,
			})
			return
		}

		viewer_follows, err = database.IsUserSubscrirerTo(vanity, username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidRelation,
			})
			return
		}
	}

	var allow_post_access bool
	if stats.Public {
		allow_post_access = true
	} else if !stats.Public && strings.TrimPrefix(req.URL.Path, "/users/") != ME && req.Header.Get("Authorization") != "" {
		allow_post_access = viewer_follows
	}

	posts := make([]model.Post, 0)
	if allow_post_access {
		posts, err = database.GetUserPost(username, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidUser,
			})
			return
		}
	}

	jsonEncoder.Encode(struct {
		Followers        int64        `json:"followers"`
		Following        int64        `json:"following"`
		Public           bool         `json:"public"`
		Suspended        bool         `json:"suspended"`
		CanAccessPost    bool         `json:"access_post"`
		FollowedByViewer bool         `json:"followed_by_viewer"`
		Posts            []model.Post `json:"posts"`
	}{
		Followers:        stats.Followers,
		Following:        stats.Following,
		Public:           stats.Public,
		Suspended:        stats.Suspended,
		CanAccessPost:    allow_post_access,
		FollowedByViewer: viewer_follows,
		Posts:            posts,
	})
}

// Delete allows users to delete their account
func Delete(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity := ""
	var err error

	if req.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	} else if req.Header.Get("Authorization") == os.Getenv("GLOBAL_AUTH") {
		vanity = req.URL.Query().Get("user")
	} else {
		vanity, err = helpers.CheckToken(req.Header.Get("Authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidToken,
			})
			return
		}
	}

	_, err = database.DeleteUser(vanity)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

	database.Set(vanity+"-gd", "ok", 3600)

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: Ok,
	})
}

// Handle patch method, allows to update user data
func update(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity, err := helpers.CheckToken(req.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

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

	var getbody model.UpdateBody
	json.Unmarshal(body, &getbody)

	if getbody.Public != nil {
		_, err := database.MakeRequest("MATCH (u:User {name: $id}) SET u.public = $public;", map[string]interface{}{"id": vanity, "public": *getbody.Public})
		if err != nil {
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: Ok,
			})
			return
		}
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: Ok,
	})
}
