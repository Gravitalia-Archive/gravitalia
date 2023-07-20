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
func getVanity(token string) string {
	var vanity string
	if token == "" {
		return ""
	} else {
		data, err := helpers.CheckToken(token)
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

// Handler re-routes request to the right function
// based on its method
func Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		getComment(w, req)
	} else if req.Method == http.MethodPost {
		addComment(w, req)
	} else if req.Method == http.MethodDelete {
		deleteComment(w, req)
	}
}

// getComment returns the comment and replies
func getComment(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity := getVanity(req.Header.Get("authorization"))

	id := strings.TrimPrefix(req.URL.Path, "/comment/")
	post, err := database.GetPost(id, "")
	if err != nil || post.Id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidPost,
		})
		return
	}

	stats, err := database.GetProfile(post.Author)
	if err != nil || stats.Suspended {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidUser,
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
				Message: ErrorInvalidRelation,
			})
			return
		}

		allow_post_access = is
	}

	if !allow_post_access {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidPostAccess,
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
				Message: ErrorWithDatabase,
			})
			return
		}
	} else {
		comments, err = database.GetComments(id, skip*20, vanity)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorWithDatabase,
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

// addComment allows to create a new comment on a post
func addComment(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity := getVanity(req.Header.Get("authorization"))
	if vanity == "" {
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

	var getbody model.AddBody
	json.Unmarshal(body, &getbody)

	if strings.TrimSpace(getbody.Content) == "" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidBody,
		})
		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/comment/")
	post, err := database.GetPost(id, "")
	if err != nil || post.Id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidPost,
		})
		return
	}

	stats, err := database.GetProfile(post.Author)
	if err != nil || stats.Suspended {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidUser,
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
				Message: ErrorInvalidRelation,
			})
			return
		}

		allow_post_access = is
	}

	if !allow_post_access {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidPostAccess,
		})
		return
	}

	var comment_id string
	if getbody.ReplyTo == "" {
		// Notify post creator that a user posted a comment
		if vanity != post.Author {
			msg, _ := json.Marshal(
				model.Message{
					Type:      "post_comment",
					From:      vanity,
					To:        id,
					Important: true,
				},
			)
			helpers.Publish(post.Author, msg)
		}

		// Create comment on database
		comment_id, err = database.CommentPost(id, vanity, getbody.Content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidBody,
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
					Message: ErrorInvalidBody,
				})
				return
			}
		} else {
			comment_id, err = database.CommentReply(getbody.ReplyTo, vanity, getbody.Content, original_comment_id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				jsonEncoder.Encode(model.RequestError{
					Error:   true,
					Message: ErrorInvalidBody,
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

// deleteComment allows to remove a comment from database and
// all associated replies
func deleteComment(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity := getVanity(req.Header.Get("authorization"))
	if vanity == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/comment/")

	_, err := database.MakeRequest("MATCH (c:Comment {id: $to})<-[:WROTE]-(u:User {name: $id}) OPTIONAL MATCH (r:Comment)-[:REPLY]-(c) WITH c, r DETACH DELETE r, c;", map[string]any{"id": vanity, "to": id})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: Ok,
	})
}
