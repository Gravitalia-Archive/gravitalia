package database

import (
	"context"
	"errors"
	"os"

	"github.com/Gravitalia/gravitalia/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	ctx       = context.Background()
	driver, _ = neo4j.NewDriverWithContext("bolt://localhost:7687", neo4j.BasicAuth(os.Getenv("GRAPH_USERNAME"), os.Getenv("GRAPH_PASSWORD"), ""))
	Session   = driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
)

func makeRequest(query string, params map[string]any) (any, error) {
	data, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			query,
			params)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}

// CreateUser allows to create a new user into the graph database
func CreateUser(id string) (string, error) {
	_, err := makeRequest("CREATE (:User {id: $id});", map[string]any{"id": id})
	if err != nil {
		return "", err
	}

	return "test", nil
}

// GetUserStats returns subscriptions and Subscriber of the desired user
func GetUserStats(id string) (model.Stats, error) {
	followers, err := makeRequest("MATCH (:User) -[:Subscriber]->(d:User) WHERE d.id = $id RETURN d.id, count(*) QUERY MEMORY LIMIT 10 KB;",
		map[string]any{"id": id})
	if err != nil {
		return model.Stats{Followers: -1, Following: -1}, err
	}

	follwing, err := makeRequest("MATMATCH (n:User) -[:Subscriber]->(:User) WHERE n.id = $id RETURN count(*) QUERY MEMORY LIMIT 10 KB;",
		map[string]any{"id": id})
	if err != nil {
		return model.Stats{Followers: -1, Following: -1}, err
	}

	return model.Stats{
		Followers: followers.(int64),
		Following: follwing.(int64),
	}, nil
}

// GetUserPost is a function for getting every posts of a user
// and see their likes
func GetUserPost(id string, skip uint8) ([]model.Post, error) {
	var list []model.Post

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (u:User) -[:Create]->(p:Post)<-[:Like]-(l:User) WHERE u.id = $id WITH p, count(l) as numLikes RETURN p.id, p.description, p.text, numLikes ORDER BY p.id SKIP 0 LIMIT 12 QUERY MEMORY LIMIT 5 KB;",
			map[string]any{"id": id, "skip": skip * 12})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			incr := 0
			pos := 0
			for i := 0; i < len(result.Record().Values); i++ {
				if i%4 == 0 && i != 0 {
					incr++
					pos = 0
				}
				if pos == 0 {
					list = append(list, model.Post{})
					list[incr].Id = result.Record().Values[i].(string)
				}
				if pos == 1 {
					list[incr].Description = result.Record().Values[i].(string)
				}
				if pos == 2 {
					list[incr].Text = result.Record().Values[i].(string)
				}
				if pos == 3 {
					list[incr].Like = result.Record().Values[i].(int64)
				}
				pos++
			}
			return list, nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return list, err
	}

	return list, nil
}

// UserSub allows a user to subscriber to another one
func UserRelation(id string, to_user string, relation_type string) (bool, error) {
	var content string
	switch relation_type {
	case "Subscriber", "Block":
		content = "User"
	case "Like", "View":
		content = "Post"
	}

	res, err := makeRequest("MATCH (a:User)-[:"+relation_type+"]->(b:"+content+") WHERE a.id = $id AND b.id = $to RETURN a QUERY MEMORY LIMIT 1 KB;",
		map[string]any{"id": id, "to": to_user})
	if err != nil {
		return false, err
	} else if res != nil {
		return false, errors.New("already " + relation_type + "ed")
	}

	res, err = makeRequest("MATCH (a:User), (b:"+content+") WHERE a.id = $id AND b.id = $to CREATE (a)-[r:"+relation_type+"]->(b) RETURN type(r) QUERY MEMORY LIMIT 1 KB;",
		map[string]any{"id": id, "to": to_user})
	if err != nil {
		return false, err
	} else if res == nil {
		return false, errors.New("invalid " + content)
	}

	return true, nil
}

// UserUnSub allows a user to unsubscriber to another one
func UserUnRelation(id string, to_user string, relation_type string) (bool, error) {
	var content string
	switch relation_type {
	case "Subscriber", "Block":
		content = "User"
	case "Like", "View":
		content = "Post"
	}

	_, err := makeRequest("MATCH (a:User)-[r:"+relation_type+"]->(b:"+content+") WHERE a.id = $id AND b.id = $to DELETE r QUERY MEMORY LIMIT 1 KB;",
		map[string]any{"id": id, "to": to_user})
	if err != nil {
		return false, err
	}

	return true, nil
}
