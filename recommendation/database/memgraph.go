package database

import (
	"context"
	"os"

	"github.com/Gravitalia/recommendation/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var (
	ctx     = context.Background()
	Session neo4j.SessionWithContext
)

// Init create the main variable for neo4j connection
func Init() {
	driver, _ := neo4j.NewDriverWithContext(os.Getenv("GRAPH_URL"), neo4j.BasicAuth(os.Getenv("GRAPH_USERNAME"), os.Getenv("GRAPH_PASSWORD"), ""))
	Session = driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

// loopResults allows to assort posts in an array
func loopResults(result neo4j.ResultWithContext) []model.Post {
	list := make([]model.Post, 0)

	pos := 0
	for result.Next(ctx) {
		if result.Record().Values[0] == nil {
			return list
		}

		record := result.Record()
		list = append(list, model.Post{})

		list[pos].Id = record.Values[0].(string)
		list[pos].Text = record.Values[1].(string)
		list[pos].Description = record.Values[2].(string)
		list[pos].Tag = record.Values[3].(string)

		pos++
	}

	return list
}

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
	var list []model.Post

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (n:User {name: $id})-[:Subscriber]->(u:User) MATCH (u)-[:Create]->(p:Post)-[:Show]->(t:Tag) WHERE NOT EXISTS((n)-[:View]->(p)) WITH p, t ORDER BY p.id DESC LIMIT 20 RETURN p.id, p.text, p.description, t.name;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		list = loopResults(result)

		return true, nil
	})
	if err != nil {
		return list, err
	}

	return list, nil
}

// LastCommunityPost allows to find the last n publications
// posted by the same community as the user
func LastCommunityPost(id string) ([]model.Post, error) {
	var list []model.Post

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (u:User {name: $id}) WITH u.community AS community MATCH (a:User {community: community})-[r]->(p:Post)-[:Show]->(t:Tag) WHERE NOT EXISTS((u)-[:View]->(p)) WITH p, t, count(r) AS connections ORDER BY connections DESC LIMIT 100 WITH p, t ORDER BY p.id DESC LIMIT 30 RETURN p.id, p.text, p.description, t.name;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		list = loopResults(result)

		return true, nil
	})
	if err != nil {
		return list, err
	}

	return list, nil
}

// LastLikedPost allows access to the last posts
// made with the same tag as the last liked post
func LastLikedPost(id string) ([]model.Post, error) {
	var list []model.Post

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx, "MATCH (u:User {name: $id})-[:Like]->(p:Post)-[:Show]->(t:Tag) WHERE NOT EXISTS((u)-[:View]->(p)) WITH p, t ORDER BY p.id DESC LIMIT 1 WITH t MATCH (p:Post)-[:Show]->(t:Tag) WHERE NOT EXISTS((u)-[:View]->(p)) WITH p, t ORDER BY p.id DESC LIMIT 10 RETURN p.id, p.text, p.description, t.name;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		list = loopResults(result)

		return true, nil
	})
	if err != nil {
		return list, err
	}

	return list, nil
}

// JaccardRank ranks every id in idList with the
// Jaccard similarity algorithm
func JaccardRank(id string, idList []string) ([]model.Post, error) {
	var list []model.Post

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (u:User {name: $id})-[:Like]->(p:Post) WITH u, p LIMIT 10 MATCH (l:Post) WHERE l.id IN $list AND NOT EXISTS((u)-[:View]->(l)) WITH l, p ORDER BY p.id DESC WITH collect(l) as posts, collect(p) as likedPosts CALL node_similarity.jaccard_pairwise(posts, likedPosts) YIELD node1, node2, similarity WITH node1, similarity ORDER BY similarity DESC LIMIT 15 OPTIONAL MATCH (a:User)-[:Like]->(node1) WITH node1, count(DISTINCT a) as numLikes MATCH (creator:User)-[:Create]-(node1) WITH node1, numLikes, creator OPTIONAL MATCH (:User {name: $id})-[r:Like]-(node1) RETURN node1.id, node1.text, node1.description, numLikes, node1.hash, creator.name, CASE WHEN r IS NULL THEN false ELSE true END;",
			map[string]any{"id": id, "list": idList})
		if err != nil {
			return nil, err
		}

		pos := 0
		for result.Next(ctx) {
			if result.Record().Values[0] == nil {
				return false, nil
			}

			record := result.Record()
			list = append(list, model.Post{})

			list[pos].Id = record.Values[0].(string)
			list[pos].Text = record.Values[1].(string)
			list[pos].Description = record.Values[2].(string)
			list[pos].Like = record.Values[3].(int64)
			list[pos].Hash = record.Values[4].([]any)
			list[pos].Author = record.Values[5].(string)

			pos++
		}

		return true, nil
	})
	if err != nil {
		return list, err
	}

	return list, nil
}
