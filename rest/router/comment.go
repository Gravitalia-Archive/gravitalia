package router

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
)

// getVanity permit to get vanity
func getVanity(req *http.Request) string {
	var vanity string
	if req.Header.Get("authorization") == "" {
		return ""
	} else {
		data, err := helpers.CheckToken(req.Header.Get("authorization"))
		if err != nil {
			return ""
		}
		vanity = data
	}

	return vanity
}

// doesCommentExists checks if a comment really exists
func doesCommentExists(id string) bool {
	res, err := database.MakeRequest("MATCH (c:Comment {id: $id}) RETURN c;", map[string]any{"id": id})
	if err != nil {
		return false
	}

	if res != nil {
		return true
	} else {
		return false
	}
}

// isAReply checks if the comment ID is a reply
// if yes, return the original comment
func isAReply(id string) string {
	res, err := database.MakeRequest("MATCH (:Comment {id: $id})-[:Reply]->(c:Comment) RETURN c.id;", map[string]any{"id": id})
	if err != nil {
		return ""
	}

	return res.(string)
}

func Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		get_comment(w, req)
	} else if req.Method == "POST" {
		add_comment(w, req)
	} else if req.Method == "DELETE" {
		delete_comment(w, req)
	}
}

func get_comment(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity := getVanity(req)

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

	var skip int
	if req.URL.Query().Has("skip") {
		intVar, _ := strconv.Atoi(req.URL.Query().Get("skip"))
		skip = intVar
	}

	var comments []any
	if req.URL.Query().Has("reply") {
		comments, err = database.GetReply(id, req.URL.Query().Get("reply"), skip*20, vanity)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Cannot get replies",
			})
			return
		}
	} else {
		comments, err = database.GetComments(id, skip*20, vanity)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Cannot get comments",
			})
			return
		}
	}

	if comments == nil {
		jsonEncoder.Encode(make([]any, 0))
	} else {
		jsonEncoder.Encode(comments)
	}
}

func add_comment(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity := getVanity(req)
	if vanity == "" {
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

	var comment_id string
	if getbody.ReplyTo == "" {
		comment_id, err = database.CommentPost(id, vanity, getbody.Content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Error while posting comment, double check body",
			})
			return
		}
	} else {
		if !doesCommentExists(getbody.ReplyTo) {
			w.WriteHeader(http.StatusBadRequest)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid 'reply' ID",
			})
			return
		}

		original_comment_id := isAReply(getbody.ReplyTo)
		if original_comment_id == "" {
			comment_id, err = database.CommentReply(getbody.ReplyTo, vanity, getbody.Content, getbody.ReplyTo)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				jsonEncoder.Encode(model.RequestError{
					Error:   true,
					Message: "Error while posting comment, double check body",
				})
				return
			}
		} else {
			comment_id, err = database.CommentReply(getbody.ReplyTo, vanity, getbody.Content, original_comment_id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				jsonEncoder.Encode(model.RequestError{
					Error:   true,
					Message: "Error while posting comment, double check body",
				})
				return
			}
		}
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: comment_id,
	})
}

func delete_comment(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity := getVanity(req)
	if vanity == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/comment/")

	database.DeleteComment(id, vanity)

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: "Deleted comment",
	})
}
