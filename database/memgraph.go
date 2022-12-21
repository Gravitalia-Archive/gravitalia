package database

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Item struct {
	Message string
}

//var driver, _ = neo4j.NewDriver("bolt://localhost:7687", neo4j.BasicAuth("", "", ""))
var driver, err = neo4j.NewDriverWithContext("bolt://localhost:7687", neo4j.BasicAuth("", "", ""))
var ctx = context.Background()
var Session = driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})

//var Session = driver.NewSession(neo4j.SessionConfig{})

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

/*func Createuser() (*Item, error) {
	result, err := Session.WriteTransaction(createUser)
	if err != nil {
		return nil, err
	}

	return result.(*Item), nil
}

func createUser(tx neo4j.Transaction) (interface{}, error) {
	records, err := tx.Run("CREATE (:User {vanity: 'arianagrande'})-[:Subscribers]->(:User {vanity: 'realhinome'})-[:Likes]->(:Post {id: '1055103553418567740'});", nil)
	if err != nil {
		return nil, err
	}
	record, err := records.Single()
	if err != nil {
		return nil, err
	}
	// You can also retrieve values by name, with e.g. `id, found := record.Get("n.id")`
	return &Item{
		Message: record.Values[0].(string),
	}, nil
}

func GetUserFollowers() (*Item, error) {
	result, err := Session.WriteTransaction(getFollower)
	if err != nil {
		return nil, err
	}

	return result.(*Item), nil
}

func getFollower(tx neo4j.Transaction) (interface{}, error) {
	records, err := tx.Run("MATCH (:User) -[:Likes]->(:Post) -[:Subs]->(d:User) WHERE d.name = 'Alice' RETURN count(*);", nil)
	if err != nil {
		return nil, err
	}
	record, err := records.Single()
	if err != nil {
		return nil, err
	}
	// You can also retrieve values by name, with e.g. `id, found := record.Get("n.id")`
	return &Item{
		Message: record.Values[0].(string),
	}, nil
}
*/
