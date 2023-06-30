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

// UserHandler routes to the right function
func UserHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/users/")
	if req.Method == http.MethodOptions {
		Index(w, req)
	} else if id != "" && req.Method == http.MethodGet {
		GetUser(w, req)
	} else if req.Method == http.MethodDelete {
		Delete(w, req)
	} else if id != "" && id == ME && req.Method == http.MethodPatch {
		update(w, req)
	}
}

// GetUser allows getting user data such as posts
func GetUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	var me string
	id := strings.TrimPrefix(req.URL.Path, "/users/")
	username := id

	authHeader := req.Header.Get("Authorization")

	// Check actual user
	if username == ME && authHeader != "" {
		vanity, err := helpers.CheckToken(authHeader)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidToken,
			})
			return
		}
		username = vanity
		me = vanity
	}

	// Get user profile
	stats, err := database.GetProfile(username)
	if err != nil || stats.Suspended {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidUser,
		})
		return
	}

	// Check if viewer is following user
	var viewerFollows bool
	if authHeader != "" {
		viewerFollows, err = database.IsUserSubscrirerTo(me, username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidRelation,
			})
			return
		}
	}

	// Check if viewer have access to the user's post
	allowPostAccess := stats.Public || (authHeader != "" && id != ME) || viewerFollows || (authHeader != "" && id == me)

	posts := make([]model.Post, 0)
	if allowPostAccess {
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
		CanAccessPost:    allowPostAccess,
		FollowedByViewer: viewerFollows,
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
