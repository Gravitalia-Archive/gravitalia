package router

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"

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
	if r.Method == http.MethodPost && id != NEW {
		Get(w, r)
	} else if r.Method == http.MethodPost && id == NEW {
		New(w, r)
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

	// Define channels
	tag := make(chan string)
	is_nude := make(chan bool)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()

		res, _ := grpc.TagImage(0, getbody.Images[0])
		tag <- res
	}()

	go func() {
		defer wg.Done()

		res, _ := grpc.TagImage(1, getbody.Images[0])
		is_nude <- res == "nude"
	}()

	if <-is_nude {
		w.WriteHeader(http.StatusBadRequest)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidContent,
		})
		return
	}

	// Wait until gRPC requests finished
	wg.Wait()

	// Publish content
	hash, err := grpc.UploadImage(getbody.Images[0])
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorUploading,
		})
		return
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
