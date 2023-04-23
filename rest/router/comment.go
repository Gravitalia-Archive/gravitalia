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

func Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		add(w, req)
	} else if req.Method == "DELETE" {
		delete(w, req)
	}
}

func add(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	var vanity string
	if req.Header.Get("authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	} else {
		data, err := helpers.CheckToken(req.Header.Get("authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
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
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Unable to get body",
		})
		return
	}

	var getbody model.AddBody
	json.Unmarshal(body, &getbody)

	if getbody.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid body: missing content field",
		})
		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/comment/")
	post, err := database.GetPost(id, "")
	if err != nil || post.Id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid post",
		})
		return
	}

	stats, err := database.GetProfile(post.Author)
	if err != nil || stats.Suspended {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid user",
		})
		return
	}

	var allow_post_access bool
	if stats.Public {
		allow_post_access = true
	} else if !stats.Public && req.Header.Get("authorization") != "" {
		is, err := database.IsUserSubscrirerTo(post.Author, vanity)
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

	if !allow_post_access {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "You don't have access to this post",
		})
		return
	}

	comment_id, err := database.CommentPost(id, vanity, getbody.Content)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Error while posting comment, double check body",
		})
		return
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: comment_id,
	})
}

func delete(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	var vanity string
	if req.Header.Get("authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	} else {
		data, err := helpers.CheckToken(req.Header.Get("authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid token",
			})
			return
		}
		vanity = data
	}

	id := strings.TrimPrefix(req.URL.Path, "/comment/")

	database.DeleteComment(id, vanity)

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: "Deleted comment",
	})
}
