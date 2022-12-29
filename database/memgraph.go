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

// CREATE CONSTRAINT ON (u:User) ASSERT u.vanity IS UNIQUE; => don't allow 2 users with same vanity
// CREATE (:User {vanity: "realhinome"}); => Create user - Create account
// CREATE (:User {vanity: "arianagrande"});
// CREATE (:User {vanity: "abc"});
// MATCH (a:User), (b:User) WHERE a.vanity = 'realhinome' AND b.vanity = 'arianagrande' CREATE (a)-[r:Subscribers]->(b) RETURN type(r); - A will follow B
// MATCH (a:User), (b:User) WHERE a.vanity = 'abc' AND b.vanity = 'arianagrande' CREATE (a)-[r:Subscribers]->(b) RETURN type(r);
// MATCH (a:User), (b:User) WHERE a.vanity = 'abc' AND b.vanity = 'realhinome' CREATE (a)-[r:Subscribers]->(b) RETURN type(r);
// MATCH (:User) -[:Subscribers]->(d:User) WHERE d.vanity = 'arianagrande' RETURN count(*); => GET total followers
// MATCH (n:User) -[:Subscribers]->(:User) WHERE n.vanity = 'arianagrande' RETURN count(*); => GET total following
// CREATE (:Post {id: "12345678901234", tags: ["animals", "cat", "black"], text: "Look at my cat! Awe...", description: "A black cat on a chair"}); MATCH (a:User), (b:Post) WHERE a.vanity = 'realhinome' AND b.id = '12345678901234' CREATE (a)-[r:Create]->(b) RETURN type(r); - Create post
// MATCH (a:User), (b:Post) WHERE a.vanity = 'realhinome' AND b.id = '12345678901234' CREATE (a)-[r:View]->(b) RETURN type(r); - User a saw the post
// MATCH (a:User), (b:Post) WHERE a.vanity = 'arianagrande' AND b.id = '12345678901234' CREATE (a)-[r:View]->(b) RETURN type(r);
// MATCH (a:User), (b:Post) WHERE a.vanity = 'realhinome' AND b.id = '12345678901234' CREATE (a)-[r:Like]->(b) RETURN type(r); - User a likes b post
// MATCH (n:User) -[:Create]->(p:Post) WHERE n.vanity = 'realhinome' RETURN p; - get post
// MATCH (a:User), (b:User) WHERE a.vanity = 'realhinome' AND b.vanity = 'abc' CREATE (a)-[r:Block]->(b) RETURN type(r); - user a block b user

// CreateUser allows to create a new user into the graph database
func CreateUser(vanity string) (string, error) {
	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"CREATE (:User {vanity: $vanity});",
			map[string]any{"vanity": vanity})
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

// GetUserStats returns subscriptions and subscribers of the desired user
func GetUserStats(vanity string) (model.Stats, error) {
	followers, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (interface{}, error) {
		result, err := transaction.Run(ctx,
			"MATCH (:User) -[:Subscribers]->(d:User) WHERE d.vanity = $vanity RETURN d.vanity, count(*) QUERY MEMORY LIMIT 10 KB;",
			map[string]any{"vanity": vanity})
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
			"MATCH (n:User) -[:Subscribers]->(:User) WHERE n.vanity = $vanity RETURN count(*) QUERY MEMORY LIMIT 10 KB;",
			map[string]any{"vanity": vanity})
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

func GetUserPost(vanity string, skip uint8) ([]model.Post, error) {
	var list []model.Post

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (u:User) -[:Create]->(p:Post) WHERE u.vanity = $vanity RETURN p.id, p.description, p.text ORDER BY p.id SKIP $skip LIMIT 12 QUERY MEMORY LIMIT 5 KB;",
			map[string]any{"vanity": vanity, "skip": skip * 12})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			incr := 0
			pos := 0
			for i := 0; i < len(result.Record().Values); i++ {
				if i%3 == 0 && i != 0 {
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
