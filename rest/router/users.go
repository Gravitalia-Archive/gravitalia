package router

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
)

// UserHandler routes to the right function
func UserHandler(w http.ResponseWriter, req *http.Request) {
	id := strings.TrimPrefix(req.URL.Path, "/users/")
	if req.Method == http.MethodOptions {
		Index(w, req)
	} else if id != "" && req.Method == http.MethodGet {
		getUser(w, req)
	} else if req.Method == http.MethodDelete {
		deleteUser(w, req)
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
	if username == ME && authHeader != "" {
		vanity, err := helpers.CheckToken(authHeader)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidToken,
			})
			return
		}
		username = vanity
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
		viewerFollows, err = database.IsUserSubscrirerTo(me, username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidRelation,
			})
			return
		}
	}

	// Check if viewer have access to the user's post
	allowPostAccess := stats.Public || (authHeader != "" && id != ME) || viewerFollows || (authHeader != "" && id == me)

	posts := make([]model.Post, 0)
	if allowPostAccess {
		posts, err = database.GetUserPost(username, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidUser,
			})
			return
		}
	}

	jsonEncoder.Encode(struct {
		Followers        int64        `json:"followers"`
		Following        int64        `json:"following"`
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

// Delete allows users to delete their account
func deleteUser(w http.ResponseWriter, req *http.Request) {
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
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

	database.Set(vanity+"-gd", "ok", 3600)

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: Ok,
	})
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

// GetData returns a ZIP folder with two CSV files
// containing user and liked/created posts data
func GetData(w http.ResponseWriter, req *http.Request) {
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

	// Create CSV with user data
	userFilePath, err := database.MakeRequest("WITH \"MATCH (u:User {name: '"+vanity+"'}) RETURN u.name as vanity, u.community as community_id, u.rank as rank, u.public as is_public, u.suspended as is_suspended;\" as query CALL export_util.csv_query(query, \"/var/lib/memgraph/user.csv\", True) YIELD file_path RETURN file_path;",
		map[string]any{"id": vanity})
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	}

	// Create CSV with posts data
	postFilePath, err := database.MakeRequest("WITH \"OPTIONAL MATCH (u:User {name: '"+vanity+"'})-[r]-(p:Post)-[:Show]-(t:Tag) WHERE type(r) = 'Create' OR type(r) = 'Like' OPTIONAL MATCH (p)-[:Comment]-(c:Comment)-[:Wrote]-(u) OPTIONAL MATCH (u)-[l:Like]->(p) WITH DISTINCT p, r, t, count(DISTINCT l) as likes, collect({id: c.id, text: c.text, timestamp: c.timestamp }) as my_comment RETURN p.id as id, p.text as description, p.hash as images, p.description as automatic_legend, t.name as autmatic_tag, likes, type(r) as relation, my_comment\" as query CALL export_util.csv_query(query, \"/var/lib/memgraph/$posts.csv\", True) YIELD file_path RETURN file_path;",
		map[string]any{"id": vanity})
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	}

	fmt.Println(userFilePath.(string), postFilePath.(string))

	// Create a buffer to write the ZIP file
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	// Add user CSV file to the ZIP
	userCSVFile, err := os.Open(userFilePath.(string))
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}
	defer userCSVFile.Close()

	_, err = userCSVFile.Stat()
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

	userZipFile, err := zipWriter.Create("user.csv")
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

	_, err = io.Copy(userZipFile, userCSVFile)
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

	// Add post CSV file to the ZIP
	postCSVFile, err := os.Open(postFilePath.(string))
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}
	defer postCSVFile.Close()

	_, err = postCSVFile.Stat()
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

	postZipFile, err := zipWriter.Create("posts.csv")
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

	_, err = io.Copy(postZipFile, postCSVFile)
	if err != nil {
		log.Println("(getData)", err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInternalServerError,
		})
		return
	}

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
