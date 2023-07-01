package router

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/grpc"
	"github.com/Gravitalia/gravitalia/helpers"
	"github.com/Gravitalia/gravitalia/model"
)

const NEW = "new"

const (
	ErrorInvalidContent = "Content does not comply with our rules"
	ErrorWithDatabase   = "Couldn't get database reponse"
	ErrorUploading      = "Error occurs when uploading content"
)

// PostHandler re-routes to the requested handler
func PostHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/posts/")
	if r.Method == http.MethodGet && id != NEW {
		Get(w, r)
	} else if r.Method == http.MethodPost && id == NEW {
		New(w, r)
	} else if r.Method == http.MethodDelete {
		DeletePost(w, r)
	}
}

// Get routes to a post getter
func Get(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	var vanity string
	if req.Header.Get("authorization") != "" {
		data, err := helpers.CheckToken(req.Header.Get("authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: "Invalid token",
			})
			return
		}
		vanity = data
	}

	id := strings.TrimPrefix(req.URL.Path, "/posts/")
	post, err := database.GetPost(id, vanity)
	if err != nil || post.Id == "" {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid post",
		})
		return
	}

	jsonEncoder.Encode(post)
}

// New routes allows to create a new post
func New(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	// Checks authorization
	if req.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
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

	// Read body
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

	var getbody model.PostBody
	if err = json.Unmarshal(body, &getbody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidBody,
		})
		return
	}

	if len(getbody.Images) > 5 {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorExceededMaximumImages,
		})
		return
	}

	// Define channels
	tag := make(chan string)
	isNude := make([]chan bool, len(getbody.Images))

	go func() {
		res, _ := grpc.TagImage(0, getbody.Images[0])
		tag <- res
	}()

	for i, image := range getbody.Images {
		isNude[i] = make(chan bool)
		go func(i int, image []byte) {
			res, _ := grpc.TagImage(1, image)
			isNude[i] <- res == "nude"
		}(i, image)
	}

	// Checks if content is prohibited
	for _, isNudeChan := range isNude {
		if <-isNudeChan {
			w.WriteHeader(http.StatusBadRequest)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidContent,
			})
			return
		}
	}

	// Publish contents
	hashChans := make([]chan string, len(getbody.Images))
	for i, image := range getbody.Images {
		hashChans[i] = make(chan string)
		go func(i int, image []byte) {
			res, err := grpc.UploadImage(image)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				jsonEncoder.Encode(model.RequestError{
					Error:   true,
					Message: ErrorUploading,
				})
				return
			}

			hashChans[i] <- res
		}(i, image)
	}

	// Convert string channel to string
	hash := make([]string, len(getbody.Images))
	for i, hashChan := range hashChans {
		hash[i] = <-hashChan
	}

	id := helpers.Generate()
	_, err = database.CreatePost(id, vanity, <-tag, getbody.Description, hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	}

	// Success reponse with post ID
	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: id,
	})
}

// DeletePost delete wanted post if related to connected user
func DeletePost(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	vanity := getVanity(req.Header.Get("authorization"))
	if vanity == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	}

	id := strings.TrimPrefix(req.URL.Path, "/posts/")

	_, err := database.MakeRequest("MATCH (p:Post {id: $to})<-[:Create]-(:User {name: $id}) OPTIONAL MATCH (c:Comment)-[:Comment]-(p) WITH p, c DETACH DELETE r, c;", map[string]any{"id": vanity, "to": id})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorWithDatabase,
		})
		return
	}

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: Ok,
	})
}
