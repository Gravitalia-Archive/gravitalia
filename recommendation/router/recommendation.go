package router

import (
	"encoding/json"
	"net/http"

	"github.com/Gravitalia/recommendation/database"
	"github.com/Gravitalia/recommendation/helpers"
	"github.com/Gravitalia/recommendation/model"
)

// Handler allows to choose the best route based on the method
func Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "OPTIONS" {
		Index(w, req)
	} else if req.Method == "GET" {
		recommendationGet(w, req)
	}
}

// Get handle the route for /for_you_feed and return
// the posts that the user may like
func recommendationGet(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	var vanity string
	if req.Header.Get("authorization") == "" {
		w.WriteHeader(http.StatusUnauthorized)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorInvalidToken,
		})
		return
	} else {
		data, err := helpers.Check(req.Header.Get("authorization"))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			jsonEncoder.Encode(model.RequestError{
				Error:   true,
				Message: ErrorInvalidToken,
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
			Message: ErrorGetLatestLikedPost,
		})
		return
	}

	followingPost, err := database.LastFollowingPost(vanity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorGetLatestFollowing,
		})
		return
	}

	communityPost, err := database.LastCommunityPost(vanity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: ErrorGetLatestCommunity,
		})
		return
	}

	var posts []model.Post
	posts = append(posts, tagPost...)
	posts = append(posts, followingPost...)
	posts = append(posts, communityPost...)

	// Remove duplicates
	posts = helpers.RemoveDuplicates(posts)

	var ids []string

	for _, post := range posts {
		ids = append(ids, post.Id)
	}

	posts, _ = database.JaccardRank(vanity, ids)
	// Remove duplicates
	posts = helpers.RemoveDuplicates(posts)

	jsonEncoder.Encode(posts)
}
