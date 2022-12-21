package database

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var ctx = context.Background()
var driver, _ = neo4j.NewDriverWithContext("bolt://localhost:7687", neo4j.BasicAuth("", "", ""))
var Session = driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

// CREATE (:User {vanity: "arianagrande"})-[:Subscribers]->(:User {vanity: "realhinome"})-[:Likes]->(:Post {id: "1055103553418567740"}); => Create user - ArianaGrande will be subscribed to RealHinome
// MATCH (:User) -[:Subscribers]->(d:User) WHERE d.vanity = 'realhinome' RETURN count(*); => GET total followers
// MATCH (n:User) -[:Subscribers]-(d:User) WHERE n.vanity = 'arianagrande' RETURN count(*); => GET total following

func CreateUser() (string, error) {
	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"CREATE (:User {vanity: $vanity})-[:Subscribers]->(:User {vanity: 'realhinome'})-[:Likes]->(:Post {id: '1055103553418567740'});",
			map[string]any{"vanity": "arianagrande"})
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
