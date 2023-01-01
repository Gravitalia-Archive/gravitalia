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

// CREATE CONSTRAINT ON (u:User) ASSERT u.id IS UNIQUE; => don't allow 2 users with same id
// CREATE (:User {id: "realhinome"}); => Create user - Create account
// CREATE (:User {id: "arianagrande"});
// CREATE (:User {id: "abc"});
// MATCH (a:User), (b:User) WHERE a.id = 'realhinome' AND b.id = 'arianagrande' CREATE (a)-[r:Subscriber]->(b) RETURN type(r); - A will follow B
// MATCH (a:User), (b:User) WHERE a.id = 'abc' AND b.id = 'arianagrande' CREATE (a)-[r:Subscriber]->(b) RETURN type(r);
// MATCH (a:User), (b:User) WHERE a.id = 'abc' AND b.id = 'realhinome' CREATE (a)-[r:Subscriber]->(b) RETURN type(r);
// MATCH (:User) -[:Subscriber]->(d:User) WHERE d.id = 'arianagrande' RETURN count(*); => GET total followers
// MATCH (n:User) -[:Subscriber]->(:User) WHERE n.id = 'arianagrande' RETURN count(*); => GET total following
// CREATE (:Post {id: "12345678901234", tags: ["animals", "cat", "black"], text: "Look at my cat! Awe...", description: "A black cat on a chair"}); MATCH (a:User), (b:Post) WHERE a.id = 'realhinome' AND b.id = '12345678901234' CREATE (a)-[r:Create]->(b) RETURN type(r); - Create post
// MATCH (a:User), (b:Post) WHERE a.id = 'realhinome' AND b.id = '12345678901234' CREATE (a)-[r:View]->(b) RETURN type(r); - User a saw the post
// MATCH (a:User), (b:Post) WHERE a.id = 'arianagrande' AND b.id = '12345678901234' CREATE (a)-[r:View]->(b) RETURN type(r);
// MATCH (a:User), (b:Post) WHERE a.id = 'realhinome' AND b.id = '12345678901234' CREATE (a)-[r:Like]->(b) RETURN type(r); - User a likes b post
// MATCH (n:User) -[:Create]->(p:Post) WHERE n.id = 'realhinome' RETURN p; - get post
// MATCH (a:User), (b:User) WHERE a.id = 'realhinome' AND b.id = 'abc' CREATE (a)-[r:Block]->(b) RETURN type(r); - user a block b user

// CreateUser allows to create a new user into the graph database
func CreateUser(id string) (string, error) {
	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"CREATE (:User {id: $id});",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return "", err
	}

	return "test", nil
}

// GetUserStats returns subscriptions and Subscriber of the desired user
func GetUserStats(id string) (model.Stats, error) {
	followers, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (interface{}, error) {
		result, err := transaction.Run(ctx,
			"MATCH (:User) -[:Subscriber]->(d:User) WHERE d.id = $id RETURN d.id, count(*) QUERY MEMORY LIMIT 10 KB;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return nil, errors.New("invalid user")
			} else {
				return result.Record().Values[1], nil
			}
		}

		return nil, result.Err()
	})
	if err != nil {
		return model.Stats{Followers: -1, Following: -1}, err
	}

	follwing, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (interface{}, error) {
		result, err := transaction.Run(ctx,
			"MATCH (n:User) -[:Subscriber]->(:User) WHERE n.id = $id RETURN count(*) QUERY MEMORY LIMIT 10 KB;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})
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

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, "MATCH (a:User)-[:"+relation_type+"]->(b:"+content+") WHERE a.id = $id AND b.id = $to RETURN a QUERY MEMORY LIMIT 1 KB;",
			map[string]any{"id": id, "to": to_user})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return true, nil
			} else {
				return false, errors.New("already " + relation_type + "ed")
			}
		} else {
			return true, nil
		}
	})
	if err != nil {
		return false, err
	}

	_, err = Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (a:User), (b:"+content+") WHERE a.id = $id AND b.id = $to CREATE (a)-[r:"+relation_type+"]->(b) RETURN type(r) QUERY MEMORY LIMIT 1 KB;",
			map[string]any{"id": id, "to": to_user})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return nil, errors.New("invalid " + content)
			} else {
				return true, nil
			}
		} else {
			return nil, errors.New("invalid " + content)
		}
	})
	if err != nil {
		return false, err
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

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, "MATCH (a:User)-[r:"+relation_type+"]->(b:"+content+") WHERE a.id = $id AND b.id = $to DELETE r QUERY MEMORY LIMIT 1 KB;",
			map[string]any{"id": id, "to": to_user})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return true, nil
			} else {
				return false, errors.New("already " + relation_type + "ed")
			}
		} else {
			return true, nil
		}
	})
	if err != nil {
		return false, err
	}

	return true, nil
}
