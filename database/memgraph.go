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
	followers, err := makeRequest("MATCH (:User) -[:Subscriber]->(d:User) WHERE d.id = $id RETURN count(*), d.id QUERY MEMORY LIMIT 10 KB;",
		map[string]any{"id": id})
	if err != nil {
		return model.Stats{Followers: -1, Following: -1}, err
	}

	follwing, err := makeRequest("MATCH (n:User) -[:Subscriber]->(:User) WHERE n.id = $id RETURN count(*) QUERY MEMORY LIMIT 10 KB;",
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
			"MATCH (u:User {id: $id})-[:Create]->(p:Post) OPTIONAL MATCH (p)<-[:Like]-(liker:User) RETURN p.id as id, p.description as description, p.text as text, count(liker) AS likes ORDER BY id SKIP 0 LIMIT 12;",
			map[string]any{"id": id, "skip": skip * 12})
		if err != nil {
			return nil, err
		}

		pos := 0
		for result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return nil, errors.New("invalid user")
			}

			record := result.Record()
			list = append(list, model.Post{})

			list[pos].Id = record.Values[0].(string)
			list[pos].Description = record.Values[1].(string)
			list[pos].Text = record.Values[2].(string)
			list[pos].Like = record.Values[3].(int64)

			pos++
		}

		return list, nil
	})
	if err != nil {
		return list, err
	}

	return list, nil
}

// UserSub allows a user to subscriber to another one
func UserRelation(id string, toUser string, relationType string) (bool, error) {
	var content string
	switch relationType {
	case "Subscriber", "Block":
		content = "User"
	case "Like", "View":
		content = "Post"
	}

	res, err := makeRequest("MATCH (a:User)-[:"+relationType+"]->(b:"+content+") WHERE a.id = $id AND b.id = $to RETURN a QUERY MEMORY LIMIT 1 KB;",
		map[string]any{"id": id, "to": toUser})
	if err != nil {
		return false, err
	} else if res != nil {
		return false, errors.New("already " + relationType + "ed")
	}

	res, err = makeRequest("MATCH (a:User), (b:"+content+") WHERE a.id = $id AND b.id = $to CREATE (a)-[r:"+relationType+"]->(b) RETURN type(r) QUERY MEMORY LIMIT 1 KB;",
		map[string]any{"id": id, "to": toUser})
	if err != nil {
		return false, err
	} else if res == nil {
		return false, errors.New("invalid " + content)
	}

	return true, nil
}

// UserUnSub allows a user to unsubscriber to another one
func UserUnRelation(id string, toUser string, relationType string) (bool, error) {
	var content string
	switch relationType {
	case "Subscriber", "Block":
		content = "User"
	case "Like", "View":
		content = "Post"
	}

	_, err := makeRequest("MATCH (a:User)-[r:"+relationType+"]->(b:"+content+") WHERE a.id = $id AND b.id = $to DELETE r QUERY MEMORY LIMIT 1 KB;",
		map[string]any{"id": id, "to": toUser})
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetPost allows to get data of a post
func GetPost(id string) (model.Post, error) {
	var post model.Post

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (:User)-[:Create]->(p:Post {id: $id}) MATCH (:User)-[:Like]->(p) WITH p, count(*) as numLikes OPTIONAL MATCH (p)<-[r:Comment]-(c:Comment) WITH p, numLikes, collect({id: c.id, text: c.text, user: c.user})[..20] as comments RETURN p.id, p.description, p.text, numLikes, comments ORDER BY p.id DESC;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return nil, errors.New("invalid post")
			}
			record := result.Record()

			post.Id = record.Values[0].(string)
			post.Description = record.Values[1].(string)
			post.Text = record.Values[2].(string)
			post.Like = record.Values[3].(int64)
			post.Comments = record.Values[4].([]any)

			return post, nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return model.Post{}, err
	}

	return post, nil
}

func Deleteuser(vanity string) (bool, error) {
	_, err := makeRequest("MATCH (u:User {id: 'realhinome'})-[:Create]->(p:Post) DETACH DELETE p WITH u MATCH (u)-[:Commented]->(c:Comment) DETACH DELETE c WITH u MATCH (u)-[r]->() DELETE r WITH u DETACH DELETE u;",
		map[string]any{"id": vanity})
	if err != nil {
		return false, err
	}

	return true, nil
}
