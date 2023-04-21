package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Gravitalia/recommendation/database"
	"github.com/Gravitalia/recommendation/helpers"
	"github.com/Gravitalia/recommendation/model"
)

func Get(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	var vanity string
	if req.Header.Get("authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid token",
		})
		return
	} else {
		data, err := helpers.Check(req.Header.Get("authorization"))
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

	tagPost, err := database.LastLikedPost(vanity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Cannot get the latest posts from the most recently liked post tag",
		})
		return
	}
	fmt.Println("tag liked post: ", tagPost)

	followingPost, err := database.LastFollowingPost(vanity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Cannot get last following posts",
		})
		return
	}
	fmt.Println("Following post: ", followingPost)

	communityPost, err := database.LastCommunityPost(vanity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Cannot get last community posts",
		})
		return
	}
	fmt.Println("community post: ", communityPost)

	jsonEncoder.Encode(model.RequestError{
		Error:   false,
		Message: vanity,
	})
}
