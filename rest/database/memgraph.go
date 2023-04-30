package database

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
	"github.com/bradfitz/gomemcache/memcache"
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
	Mem = memcache.New(os.Getenv("MEM_URL"))
}

// MakeRequest is a simple way to send a query
func MakeRequest(query string, params map[string]any) (any, error) {
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
func CreateUser(id string) (bool, error) {
	_, err := MakeRequest("CREATE (:User {name: $id, public: true, suspended: false});", map[string]any{"id": id})
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetProfile returns followers, following and other account data of the desired user
func GetProfile(id string) (model.Profile, error) {
	var profile model.Profile

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (n:User {name: $id}) OPTIONAL MATCH (n)-[:Subscriber]->(d:User) WITH n, count(d) as following OPTIONAL MATCH (u:User)-[:Subscriber]->(n) WITH n, following, count(u) as followers RETURN followers, following, n.public, n.suspended;",
			map[string]any{"id": id})
		if err != nil {
			return nil, err
		}

		for result.Next(ctx) {
			if result.Record().Values[2] == nil {
				return nil, errors.New("invalid user")
			}

			profile.Followers = result.Record().Values[0].(int64)
			profile.Following = result.Record().Values[1].(int64)
			profile.Public = result.Record().Values[2].(bool)
			profile.Suspended = result.Record().Values[3].(bool)
		}

		return profile, nil
	})
	if err != nil {
		return model.Profile{Followers: -1, Following: -1}, err
	}

	return profile, nil
}

// GetUserPost is a function for getting every posts of a user
// and see their likes
func GetUserPost(id string, skip uint8) ([]model.Post, error) {
	list := make([]model.Post, 0)

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (u:User {name: $id})-[:Create]->(p:Post) OPTIONAL MATCH (p)<-[l:Like]-(liker:User) RETURN p.id as id, p.description, p.text, count(DISTINCT l) ORDER BY id SKIP 0 LIMIT 12;",
			map[string]any{"id": id, "skip": skip * 12})
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

// UserRelation create a new relation (edge) between two nodes
func UserRelation(id string, to string, relationType string) (bool, error) {
	var content string
	switch relationType {
	case "Subscriber", "Block":
		content = "User"
	case "Like", "View":
		content = "Post"
	case "Love":
		content = "Comment"
	}

	var identifier string
	if content == "User" {
		identifier = "name"
	} else {
		identifier = "id"
	}

	res, err := MakeRequest("MATCH (a:User {name: $id})-[:"+relationType+"]->(b:"+content+"{"+identifier+": $to}) RETURN a;",
		map[string]any{"id": id, "to": to})
	if err != nil {
		return false, err
	} else if res != nil {
		return false, errors.New("already " + relationType + "ed")
	}

	res, err = MakeRequest("MATCH (a:User {name: $id}), (b:"+content+" {"+identifier+": $to}) CREATE (a)-[r:"+relationType+"]->(b) RETURN type(r) QUERY MEMORY LIMIT 1 KB;",
		map[string]any{"id": id, "to": to})
	if err != nil {
		return false, err
	} else if res == nil {
		return false, errors.New("invalid " + content)
	}

	return true, nil
}

// UserUnRelation delete a relation (edge) between two nodes
func UserUnRelation(id string, to string, relationType string) (bool, error) {
	var content string
	switch relationType {
	case "Subscriber", "Block":
		content = "User"
	case "Like", "View":
		content = "Post"
	case "Love":
		content = "Comment"
	}

	var identifier string
	if content == "User" {
		identifier = "name"
	} else {
		identifier = "id"
	}

	_, err := MakeRequest("MATCH (a:User {name: $id})-[r:"+relationType+"]->(b:"+content+" {"+identifier+": $to}) DELETE r QUERY MEMORY LIMIT 1 KB;",
		map[string]any{"id": id, "to": to})
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetPost allows to get data of a post
func GetPost(id string, user string) (model.Post, error) {
	var post model.Post

	_, err := Session.ExecuteWrite(ctx, func(transaction neo4j.ManagedTransaction) (any, error) {
		result, err := transaction.Run(ctx,
			"MATCH (author:User)-[:Create]->(p:Post {id: $id}) MATCH (:User)-[l:Like]->(p) WITH author, p, count(DISTINCT l) as numLikes OPTIONAL MATCH (p)<-[:Comment]-(c:Comment)<-[:Wrote]-(u:User) OPTIONAL MATCH (c:Comment)-[love:Love]-(lover:User) WITH author, u, lover, p, numLikes, c, count(DISTINCT love) as loveComment WITH author, p, numLikes, collect({id: c.id, text: c.text, timestamp: c.timestamp, user: u.name, love: loveComment, me_loved: lover.name = $user })[..20] as comments RETURN p.id, p.description, p.text, numLikes, author.name, comments;",
			map[string]any{"id": id, "user": user})
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
			post.Author = record.Values[4].(string)
			post.Comments = record.Values[5].([]any)

			return post, nil
		}

		return nil, result.Err()
	})
	if err != nil {
		return model.Post{}, err
	}

	return post, nil
}

// DeleteUser allows to remove every relations, posts, comments and user
func DeleteUser(vanity string) (bool, error) {
	_, err := MakeRequest("MATCH (u:User {name: $id})-[:Create]->(p:Post) DETACH DELETE p WITH u MATCH (u)-[:Commented]->(c:Comment) DETACH DELETE c WITH u MATCH (u)-[r]->() DELETE r WITH u DETACH DELETE u;",
		map[string]any{"id": vanity})
	if err != nil {
		return false, err
	}

	return true, nil
}

// IsUserSubscrirerTo check if a user (id) is subscrired to another one (user)
// and respond with true if a relation (edge) exists
// or with false if no relation exists
func IsUserSubscrirerTo(id string, user string) (bool, error) {
	res, err := MakeRequest("MATCH (a:User {name: $id})-[:Subscriber]->(b:User {name: $to}) RETURN a;",
		map[string]any{"id": id, "to": user})
	if err != nil {
		return false, err
	}

	if res != nil {
		return true, nil
	} else {
		return false, nil
	}
}

// CommentPost allows to post a comment on a post
func CommentPost(id string, user string, content string) (string, error) {
	comment_id := helpers.Generate()

	_, err := MakeRequest("CREATE (c:Comment {id: $comment_id, text: $content, timestamp: "+strconv.FormatInt(time.Now().Unix(), 10)+"}) WITH c MATCH (p:Post {id: $to}) MATCH (u:User {name: $id}) CREATE (c)-[:Comment]->(p) CREATE (u)-[:Wrote]->(c);", map[string]any{"id": user, "to": id, "comment_id": comment_id, "content": content})
	if err != nil {
		return "", err
	}

	return comment_id, nil
}

// CommentReply allows to post a comment on another comment
func CommentReply(id string, user string, content string) (string, error) {
	comment_id := helpers.Generate()

	_, err := MakeRequest("CREATE (new_comment:Comment {id: $comment_id, text: $content, timestamp: "+strconv.FormatInt(time.Now().Unix(), 10)+"}) WITH new_comment MATCH (ref:Comment {id: $to})<-[:Wrote]-(u:User) SET new_comment.replied_to = u.name WITH ref, new_comment MATCH (u:User {name: $id}) CREATE (new_comment)-[:Reply]->(ref) CREATE (u)-[:Wrote]->(new_comment);", map[string]any{"id": user, "to": id, "comment_id": comment_id, "content": content})
	if err != nil {
		return "", err
	}

	return comment_id, nil
}

// DeleteComment allows to remove a comment on a post
func DeleteComment(id string, user string) (bool, error) {
	_, err := MakeRequest("MATCH (r:Comment)-[:Reply]-(c:Comment {id: $to})<-[:Wrote]-(u:User {name: $id}) DETACH DELETE r, c;", map[string]any{"id": user, "to": id})
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetComments sends 20 comments of a post
func GetComments(id string, skip int, user string) ([]any, error) {
	res, err := MakeRequest("MATCH (:Post {id: $id})<-[:Comment]-(c:Comment)<-[:Wrote]-(u:User) OPTIONAL MATCH (c:Comment)-[love:Love]-(lover:User) WITH  u, lover, c, count(DISTINCT love) as loveComment WITH collect({id: c.id, text: c.text, timestamp: c.timestamp, user: u.name, love: loveComment, me_loved: lover.name = $user }) as comments SKIP $skip LIMIT 20 RETURN comments;",
		map[string]any{"id": id, "skip": skip, "user": user})
	if err != nil {
		return nil, err
	}

	if res != nil {
		return res.([]any), nil
	} else {
		return nil, nil
	}
}

// GetReply sends 20 replies of a comment
func GetReply(post_id string, id string, skip int, user string) ([]any, error) {
	res, err := MakeRequest("MATCH (:Post {id: $post_id})<-[:Comment]-(:Comment {id: $id})<-[:Reply]-(c:Comment)<-[:Wrote]-(u:User) OPTIONAL MATCH (c:Comment)-[love:Love]-(lover:User) WITH  u, lover, c, count(DISTINCT love) as loveComment WITH collect({id: c.id, text: c.text, timestamp: c.timestamp, user: u.name, love: loveComment, me_loved: lover.name = $user }) as comments SKIP $skip LIMIT 20 RETURN comments;",
		map[string]any{"post_id": post_id, "id": id, "skip": skip, "user": user})
	if err != nil {
		return nil, err
	}

	if res != nil {
		return res.([]any), nil
	} else {
		return nil, nil
	}
}
