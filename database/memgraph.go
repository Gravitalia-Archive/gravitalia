package database

import (
	"context"
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
// CREATE (:User {vanity: "arianagrande"}); => Create user
// CREATE (:User {vanity: "abc"});
// MATCH (a:User), (b:User) WHERE a.vanity = 'realhinome' AND b.vanity = 'arianagrande' CREATE (a)-[r:Subscribers]->(b) RETURN type(r); - A will follow B
// MATCH (a:User), (b:User) WHERE a.vanity = 'abc' AND b.vanity = 'arianagrande' CREATE (a)-[r:Subscribers]->(b) RETURN type(r);
// MATCH (a:User), (b:User) WHERE a.vanity = 'abc' AND b.vanity = 'realhinome' CREATE (a)-[r:Subscribers]->(b) RETURN type(r);
// MATCH (:User) -[:Subscribers]->(d:User) WHERE d.vanity = 'arianagrande' RETURN count(*); => GET total followers
// MATCH (n:User) -[:Subscribers]->(:User) WHERE n.vanity = 'arianagrande' RETURN count(*); => GET total following
// CREATE (:Post {id: "12345678901234", tags: ["animals", "cat", "black"], text: "Look at my cat! Awe...", description: "A black cat on a chair"}); - Create post
// MATCH (a:User), (b:Post) WHERE a.vanity = 'realhinome' AND b.id = '12345678901234' CREATE (a)-[r:Likes]->(b) RETURN type(r); - User a likes b post

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
			"MATCH (:User) -[:Subscribers]->(d:User) WHERE d.vanity = $vanity RETURN count(*);",
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
		return model.Stats{}, err
	}

	follwing, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (interface{}, error) {
		result, err := transaction.Run(ctx,
			"MATCH (n:User) -[:Subscribers]->(:User) WHERE n.vanity = $vanity RETURN count(*);",
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
		return model.Stats{}, err
	}

	return model.Stats{
		Followers: followers.(int64),
		Following: follwing.(int64),
	}, nil
}
