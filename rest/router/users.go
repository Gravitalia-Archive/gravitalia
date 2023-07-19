package router

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
	zipkinhttp "github.com/openzipkin/zipkin-go/middleware/http"
)

// isAccountBlocked returns a boolean. If one of the both
// account blocked the other, returns true.
func isAccountBlocked(id string, to string) (bool, error) {
	if id == "" || to == "" {
		return false, nil
	}

	res, err := database.MakeRequest("MATCH (a:User {name: $id}) MATCH (b:User {name: $to}) OPTIONAL MATCH (a)-[r:Block]-(b) RETURN NOT(r IS NULL);",
		map[string]any{"id": id, "to": to})
	if err != nil {
		log.Printf("(isAccountBlocked) %v", err)
		return false, err
	} else if res == nil {
		return false, nil
	}

	return res.(bool), nil
}

// UserHandler routes to the right function
func UserHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/users/")
	if req.Method == http.MethodOptions {
		Index(w, req)
	} else if id != "" && req.Method == http.MethodGet {
		getUser(w, req)
	} else if id != "" && id == ME && req.Method == http.MethodPatch {
		update(w, req)
	}
}

// GetUser allows getting user data such as posts
func getUser(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	var me string
	id := strings.TrimPrefix(req.URL.Path, "/users/")
	username := id

	authHeader := req.Header.Get("Authorization")

	// Check actual user
	if authHeader != "" {
		vanity, err := helpers.CheckToken(authHeader)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidToken,
			})
			return
		}
		if username == ME {
			username = vanity
		}
		me = vanity
	}

	// Get user profile
	stats, err := database.GetProfile(username)
	if err != nil || stats.Suspended {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidUser,
		})
		return
	}

	// Check if viewer is following user
	var viewerFollows bool
	if authHeader != "" {
		res, err := database.MakeRequest("MATCH (:User {name: $id})-[r:Subscriber]->(:User {name: $to}) RETURN NOT(r IS NULL);",
			map[string]any{"id": me, "to": username})
		if err != nil {
			log.Printf("(getUser) cannot know if user follows: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidRelation,
			})
			return
		}

		if res == nil {
			viewerFollows = false
		} else {
			viewerFollows = res.(bool)
		}
	}

	// Check if account is blocked
	isBlocked, err := isAccountBlocked(me, username)
	if err != nil {
		isBlocked = false
	}

	// Check if viewer have access to the user's post
	allowPostAccess := stats.Public || viewerFollows || (authHeader != "" && id == me)
	if isBlocked {
		allowPostAccess = false
	}

	posts := make([]model.Post, 0)
	if allowPostAccess {
		posts, err = database.GetUserPost(username, 0)
		if err != nil {
			log.Printf("(getUser) cannot get posts: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidUser,
			})
			return
		}
	}

	jsonEncoder.Encode(struct {
		Followers        uint32       `json:"followers"`
		Following        uint32       `json:"following"`
		Public           bool         `json:"public"`
		Suspended        bool         `json:"suspended"`
		CanAccessPost    bool         `json:"access_post"`
		FollowedByViewer bool         `json:"followed_by_viewer"`
		Posts            []model.Post `json:"posts"`
	}{
		Followers:        stats.Followers,
		Following:        stats.Following,
		Public:           stats.Public,
		Suspended:        stats.Suspended,
		CanAccessPost:    allowPostAccess,
		FollowedByViewer: viewerFollows,
		Posts:            posts,
	})
}

// DeleteUser allows users to delete their account
func DeleteUser(zipkinClient *zipkinhttp.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// If method is OPTIONS send OK
		if req.Method == http.MethodOptions {
			Index(w, req)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		jsonEncoder := json.NewEncoder(w)

		vanity := ""
		var err error

		if req.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusBadRequest)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidToken,
			})
			return
		} else if req.Header.Get("Authorization") == os.Getenv("GLOBAL_AUTH") {
			vanity = req.URL.Query().Get("user")
		} else {
			vanity, err = helpers.CheckToken(req.Header.Get("Authorization"))
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				jsonEncoder.Encode(model.RequestError{
					Error:   true,
					Message: ErrorInvalidToken,
				})
				return
			}
		}

		_, err = database.MakeRequest("MATCH (u:User {name: $id}) OPTIONAL MATCH (u)-[:Wrote]->(p:Post) OPTIONAL MATCH (u)-[:Wrote]->(c:Comment) OPTIONAL MATCH (u)-[r]-() DETACH DELETE p, c, r, u;",
			map[string]interface{}{"id": vanity})
		if err != nil {
			log.Printf("(DeleteUser) cannot delete user: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInternalServerError,
			})
			return
		}

		database.Set(vanity+"-gd", "ok", 3600)

		// Add user into document in case of search
		documentUser, _ := json.Marshal(struct {
			Vanity   string `json:"vanity"`
			Username string `json:"username"`
			Flags    int    `json:"flags"`
		}{
			Vanity:   vanity,
			Username: "",
			Flags:    0,
		})
		go makeRequest(zipkinClient, os.Getenv("SEARCH_API")+"/search/delete", "DELETE", bytes.NewBuffer(documentUser), os.Getenv("GLOBAL_AUTH"))

		jsonEncoder.Encode(model.RequestError{
			Error:   false,
			Message: Ok,
		})
	}
}

// Handle patch method, allows to update user data
func update(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity, err := helpers.CheckToken(req.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorUnableReadBody,
		})
		return
	}

	var getbody model.UpdateBody
	json.Unmarshal(body, &getbody)

	if getbody.Public != nil {
		_, err := database.MakeRequest("MATCH (u:User {name: $id}) SET u.public = $public;", map[string]any{"id": vanity, "public": *getbody.Public})
		if err != nil {
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: Ok,
			})
			return
		}
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: Ok,
	})
}

// AcceptOrDecline permits to accept or decline the following request
func AcceptOrDecline(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	choice := strings.TrimPrefix(req.URL.Path, "/request/")
	if choice != "accept" && choice != "decline" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidRelation,
		})
		return
	}

	vanity, err := helpers.CheckToken(req.Header.Get("Authorization"))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	if !req.URL.Query().Has("target") {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidUser,
		})
		return
	}

	// Check if relation exists
	res, err := database.MakeRequest("MATCH (:User {name: $id})-[r:Request]->(:User {name: $to}) RETURN r;",
		map[string]any{"id": req.URL.Query().Get("target"), "to": vanity})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	}

	if res == nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidRelation,
		})
		return
	}

	if choice == "accept" {
		// Delete old relation, and create new one
		_, err = database.MakeRequest("MATCH (a:User {name: $id})-[r:Request]->(b:User {name: $to}) DELETE r CREATE (a)-[:Subscriber]->(b);",
			map[string]any{"id": req.URL.Query().Get("target"), "to": vanity})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorWithDatabase,
			})
			return
		}

		// Notify requester of the acceptance
		msg, _ := json.Marshal(
			model.Message{
				Type:      "subscription_accepted",
				From:      vanity,
				To:        req.URL.Query().Get("target"),
				Important: false,
			},
		)
		helpers.Publish(req.URL.Query().Get("target"), msg)
	} else {
		// Delete old relation
		_, err = database.MakeRequest("MATCH (a:User {name: $id})-[r:Request]->(b:User {name: $to}) DELETE r;",
			map[string]any{"id": req.URL.Query().Get("target"), "to": vanity})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorWithDatabase,
			})
			return
		}
	}
}

// GetData returns a ZIP folder with two CSV files
// containing user and liked/created posts data
func GetData(w http.ResponseWriter, req *http.Request) {
	// If method is OPTIONS send OK
	if req.Method == http.MethodOptions {
		Index(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	// Check authorization header
	authToken := req.Header.Get("authorization")

	var vanity string
	if authToken == os.Getenv("GLOBAL_AUTH") && req.URL.Query().Has("vanity") {
		vanity = req.URL.Query().Get("vanity")
	} else if authToken != "" {
		user, err := helpers.CheckToken(authToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidToken,
			})
			return
		}
		vanity = user
	}

	if vanity == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	// Check if data has been recuperated 24 hours ago
	if val, _ := database.Mem.Get(vanity + "-data"); val != nil && string(val.Value) == "ok" {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorDataRequested,
		})
		return
	}

	// Create CSV with user data
	userFilePath, err := database.MakeRequest("WITH \"MATCH (u:User {name: '"+vanity+"'}) RETURN u.name as vanity, u.community as community_id, u.rank as rank, u.public as is_public, u.suspended as is_suspended;\" as query CALL export_util.csv_query(query, \"/var/lib/memgraph/user.csv\", True) YIELD file_path RETURN file_path;",
		map[string]interface{}{"id": vanity})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	}

	// Create CSV with posts data
	postFilePath, err := database.MakeRequest("WITH \"OPTIONAL MATCH (u:User {name: '"+vanity+"'})-[r]-(p:Post)-[:Show]-(t:Tag) WHERE type(r) = 'Create' OR type(r) = 'Like' OPTIONAL MATCH (p)-[:Comment]-(c:Comment)-[:Wrote]-(u) OPTIONAL MATCH (u)-[l:Like]->(p) WITH DISTINCT p, r, t, count(DISTINCT l) as likes, collect({id: c.id, text: c.text, timestamp: c.timestamp }) as my_comment RETURN p.id as id, p.text as description, p.hash as images, p.description as automatic_legend, t.name as autmatic_tag, likes, type(r) as relation, my_comment\" as query CALL export_util.csv_query(query, \"/var/lib/memgraph/posts.csv\", True) YIELD file_path RETURN file_path;",
		map[string]interface{}{"id": vanity})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	}

	// Create a buffer to write the ZIP file
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	// Create a WaitGroup to synchronize goroutines
	var wg sync.WaitGroup

	// Parallelize CSV file operations
	wg.Add(2)

	go func() {
		defer wg.Done()

		// Add user CSV file to the ZIP
		if err := addFileToZip(zipWriter, userFilePath.(string), "user.csv"); err != nil {
			log.Println("(getData)", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInternalServerError,
			})
			return
		}
	}()

	go func() {
		defer wg.Done()

		// Add post CSV file to the ZIP
		if err := addFileToZip(zipWriter, postFilePath.(string), "posts.csv"); err != nil {
			log.Println("(getData)", err)
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInternalServerError,
			})
			return
		}
	}()

	// Wait for goroutines to finish
	wg.Wait()

	// Close the ZIP writer
	err = zipWriter.Close()
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

	// Add 24h cooldown
	database.Set(vanity+"-data", "ok", 86400)

	// Set the appropriate headers for the ZIP file
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=data.zip")

	// Write the ZIP buffer to the response
	_, err = zipBuffer.WriteTo(w)
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}
}

func addFileToZip(zipWriter *zip.Writer, filePath, fileName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Stat()
	if err != nil {
		return err
	}

	zipFile, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}

	_, err = io.Copy(zipFile, file)
	if err != nil {
		return err
	}

	return nil
}
