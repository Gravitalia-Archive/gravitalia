package router

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
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

	// Check valid relation
	relation := strings.ToUpper(strings.TrimPrefix(req.URL.Path, "/relation/"))
	if relation == "" || func() bool {
		for _, v := range []string{"LIKE", "SUBSCRIBER", "BLOCK", "LOVE", "VIEW"} {
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

	// Check token
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

	if getbody.Id == "" || vanity == getbody.Id {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidBody,
		})
		return
	}

	var content string
	switch relation {
	case "SUBSCRIBER", "BLOCK", "REQUEST":
		content = "User"
	case "LIKE", "VIEW":
		content = "Post"
	case "LOVE":
		content = "Comment"
	}

	var identifier string
	if content == "User" {
		identifier = "name"
	} else {
		identifier = "id"
	}

	// Remove subscription relations
	if relation == "BLOCK" {
		_, err = database.MakeRequest("MATCH (:User {name: $id})-[r:SUBSCRIBER]-(:User {name: $to}) DELETE r;",
			map[string]any{"id": vanity, "to": getbody.Id})
		if err != nil {
			log.Printf("(Relation) Cannot remove subscription: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorWithDatabase,
			})
			return
		}
	}

	if relation == "SUBSCRIBER" {
		// Check if account is blocked
		isBlocked, err := isAccountBlocked(vanity, getbody.Id)
		if err != nil {
			log.Printf("(Relation) Cannot know if users are blocked: %v", err)
			isBlocked = false
		}

		if isBlocked {
			w.WriteHeader(http.StatusConflict)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidUser,
			})
			return
		}

		// Check if account is private
		stats, err := database.GetBasicProfile(getbody.Id)
		if err != nil || stats.Suspended {
			log.Printf("(Relation) cannot get targeted user: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidUser,
			})
			return
		}

		if !stats.Public {
			// If sub relation exists, remove it
			res, err := database.MakeRequest("MATCH (a:User {name: $id}) MATCH (b:User {name: $to}) OPTIONAL MATCH (a)-[r:SUBSCRIBER]->(b) WITH r, a, b WHERE r IS NOT NULL DELETE r RETURN NOT(r IS NULL);",
				map[string]any{"id": vanity, "to": getbody.Id})
			if err != nil {
				log.Printf("(Relation) Cannot remove sub private: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonEncoder.Encode(model.RequestError{
					Error:   true,
					Message: ErrorWithDatabase,
				})
				return
			}

			if res != nil && res.(bool) {
				jsonEncoder.Encode(model.RequestError{
					Error:   false,
					Message: OkDeletedRelation,
				})
				return
			}

			// Remove or create sub request
			res, err = database.MakeRequest("MATCH (a:User {name: $id}) MATCH (b:User {name: $to}) OPTIONAL MATCH (a)-[r:REQUEST]->(b) DELETE r FOREACH (x IN CASE WHEN r IS NULL THEN [1] ELSE [] END |	CREATE (a)-[:REQUEST]->(b)	) RETURN NOT(r IS NULL);",
				map[string]any{"id": vanity, "to": getbody.Id})
			if err != nil {
				log.Printf("(Relation) Got an error : %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonEncoder.Encode(model.RequestError{
					Error:   true,
					Message: ErrorWithDatabase,
				})
				return
			}

			if res.(bool) {
				jsonEncoder.Encode(model.RequestError{
					Error:   false,
					Message: OkDeletedRelation,
				})
				return
			} else {
				// Notify target that requester wants to follow him
				msg, _ := json.Marshal(
					model.Message{
						Type:      "request_subscription",
						From:      vanity,
						To:        getbody.Id,
						Important: true,
					},
				)
				helpers.Nats.Publish(getbody.Id, msg)

				jsonEncoder.Encode(model.RequestError{
					Error:   false,
					Message: OkAddedRequest,
				})
				return
			}
		}
	}

	// Create or delete asked relation
	res, err := database.MakeRequest("MATCH (a:User {name: $id}) MATCH (b:"+content+"{"+identifier+": $to}) OPTIONAL MATCH (a)-[r:"+relation+"]->(b) DELETE r FOREACH (x IN CASE WHEN r IS NULL THEN [1] ELSE [] END |	CREATE (a)-[:"+relation+"]->(b)	) RETURN NOT(r IS NULL);",
		map[string]any{"id": vanity, "to": getbody.Id})
	if err != nil {
		log.Printf("(Relation) Got an error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	}

	if res.(bool) {
		jsonEncoder.Encode(model.RequestError{
			Error:   false,
			Message: OkDeletedRelation,
		})
	} else {
		// Notify post author if a new like appears
		if relation == "LIKE" {
			res, _ := database.MakeRequest("MATCH (u:User)-[:CREATE]-(:Post {id: $id}) RETURN u.name;",
				map[string]any{"id": getbody.Id})
			if err != nil {
				log.Printf("(Relation) Cannot get post creator: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonEncoder.Encode(model.RequestError{
					Error:   true,
					Message: ErrorWithDatabase,
				})
				return
			}

			if vanity != res.(string) {
				msg, _ := json.Marshal(
					model.Message{
						Type:      "post_like",
						From:      vanity,
						To:        getbody.Id,
						Important: true,
					},
				)
				helpers.Publish(res.(string), msg)
			}
		}

		jsonEncoder.Encode(model.RequestError{
			Error:   false,
			Message: OkCreatedRelation,
		})
	}
}

// Exists handles route to know if a relation
// exists between two nodes
func Exists(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	relation := strings.ToUpper(strings.TrimPrefix(req.URL.Path, "/relation/"))
	if relation == "" || func() bool {
		for _, v := range []string{"LIKE", "SUBSCRIBER", "BLOCK", "LOVE", "REQUEST"} {
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

	var content string
	switch relation {
	case "SUBSCRIBER", "BLOCK", "REQUEST":
		content = "User"
	case "LIKE", "VIEW":
		content = "Post"
	case "LOVE":
		content = "Comment"
	}

	var identifier string
	if content == "User" {
		identifier = "name"
	} else {
		identifier = "id"
	}

	var existence string
	res, err := database.MakeRequest("MATCH (a:User {name: $id})-[:"+relation+"]->(b:"+content+"{"+identifier+": $to}) RETURN a;",
		map[string]any{"id": vanity, "to": target})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	} else if res != nil {
		existence = "existent"
	} else if res == nil {
		existence = "non-existent"
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: existence,
	})
}
