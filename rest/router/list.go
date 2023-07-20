package router

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// ListHandler routes to the right function
func ListHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodOptions {
		Index(w, req)
	} else if req.Method == http.MethodGet {
		getList(w, req)
	}
}

// getList allows to return a user or post list
// based on the wanted list
func getList(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	authToken := req.Header.Get("authorization")

	// Check token
	if authToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	vanity, err := helpers.CheckToken(authToken)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	id := strings.ToUpper(strings.TrimPrefix(req.URL.Path, "/list/"))
	if id == "" || func() bool {
		for _, v := range []string{"SUBSCRIBER", "SUBSCRIPTION", "BLOCK", "REQUEST"} {
			if v == id {
				return false
			}
		}
		return true
	}() {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidList,
		})
		return
	}

	list := make([]any, 0)
	ctx := context.Background()
	if id == "Subscription" {
		_, err := database.Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (:User {name: $id})-[:SUBSCRIBER]->(u:User) RETURN u.name;",
				map[string]any{"id": vanity})
			if err != nil {
				return nil, err
			}

			for result.Next(ctx) {
				if result.Record().Values[0] == nil {
					return list, nil
				}

				list = append(list, result.Record().Values[0].(string))
			}

			return nil, result.Err()
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorWithDatabase,
			})
			return
		}
	} else {
		_, err := database.Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (u:User)-[:"+id+"]->(:User {name: $id}) RETURN u.name;",
				map[string]any{"id": vanity})
			if err != nil {
				return nil, err
			}

			for result.Next(ctx) {
				if result.Record().Values[0] == nil {
					return list, nil
				}

				list = append(list, result.Record().Values[0].(string))
			}

			return nil, result.Err()
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorWithDatabase,
			})
			return
		}
	}

	jsonEncoder.Encode(list)
}
