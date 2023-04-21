package database

import (
	"context"
	"os"

	"github.com/Gravitalia/recommendation/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	ctx       = context.Background()
	driver, _ = neo4j.NewDriverWithContext("bolt://localhost:7687", neo4j.BasicAuth(os.Getenv("GRAPH_USERNAME"), os.Getenv("GRAPH_PASSWORD"), ""))
	Session   = driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
)

// PageRank starts a new calculation of PageRank
// in the database
func PageRank() (bool, error) {
	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		_, err := transaction.Run(ctx,
			"MATCH p=(n:User)-[r]->(m:User) WHERE type(r) <> 'Block' WITH project(p) as graph CALL pagerank_online.update(graph) YIELD node, rank SET node.rank = rank;",
			map[string]any{})
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

// CommunityDetection starts a new calculation of community
// detection in the database
func CommunityDetection() (bool, error) {
	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		_, err := transaction.Run(ctx,
			"MATCH p=(n:User)-[r]->(m) WHERE type(r) <> 'Block' AND type(r) <> 'View' WITH project(p) as graph CALL community_detection_online.update(graph) YIELD node, community_id WITH node, community_id WHERE labels(node) = ['User'] SET node.community = community_id;",
			map[string]any{})
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

// LastFollowingPost allows to find the last n publications
// posted by followings account
func LastFollowingPost(id string) ([]model.Post, error) {
	list := make([]model.Post, 0)

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (n:User {name: 'realhinome'})-[:Subscriber]->(u:User) MATCH (u)-[:Create]->(p:Post)-[:Show]->(t:Tag) WHERE NOT EXISTS((n)-[:View]->(p)) WITH p, t ORDER BY p.id DESC LIMIT 20 RETURN p.id, p.description, p.text, t.name;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		pos := 0
		for result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return list, nil
			}

			record := result.Record()
			list = append(list, model.Post{})

			list[pos].Id = record.Values[0].(string)
			list[pos].Description = record.Values[1].(string)
			list[pos].Text = record.Values[2].(string)
			list[pos].Tag = record.Values[3].(string)

			pos++
		}

		return list, nil
	})
	if err != nil {
		return list, err
	}

	return list, nil
}

// LastCommunityPost allows to find the last n publications
// posted by the same community as the user
func LastCommunityPost(id string) ([]model.Post, error) {
	list := make([]model.Post, 0)

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (u:User {name: $id}) WITH u.community as community MATCH (a)-[r]->(p:Post)-[:Show]->(t:Tag) WHERE a.community = community WITH p, t, count(r) as connections ORDER BY connections DESC LIMIT 100 WITH p, t ORDER BY p.id DESC LIMIT 30 RETURN p.id, p.text, p.description, t.name;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		pos := 0
		for result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return list, nil
			}

			record := result.Record()
			list = append(list, model.Post{})

			list[pos].Id = record.Values[0].(string)
			list[pos].Text = record.Values[1].(string)
			list[pos].Description = record.Values[2].(string)
			list[pos].Tag = record.Values[3].(string)

			pos++
		}

		return list, nil
	})
	if err != nil {
		return list, err
	}

	return list, nil
}

// LastLikedPost allows access to the last posts
// made with the same tag as the last liked post
func LastLikedPost(id string) ([]model.Post, error) {
	list := make([]model.Post, 0)

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (:User {name: 'realhinome'})-[:Like]->(p:Post)-[:Show]->(t:Tag) WITH p, t ORDER BY p.id DESC LIMIT 1 WITH t MATCH (p:Post)-[:Show]->(t:Tag) WITH p, t ORDER BY p.id DESC LIMIT 10 RETURN p.id, p.text, p.description, t.name;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		pos := 0
		for result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return list, nil
			}

			record := result.Record()
			list = append(list, model.Post{})

			list[pos].Id = record.Values[0].(string)
			list[pos].Text = record.Values[1].(string)
			list[pos].Description = record.Values[2].(string)
			list[pos].Tag = record.Values[3].(string)

			pos++
		}

		return list, nil
	})
	if err != nil {
		return list, err
	}

	return list, nil
}
