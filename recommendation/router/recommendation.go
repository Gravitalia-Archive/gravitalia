package router

import (
	"encoding/json"
	"net/http"

	"github.com/Gravitalia/recommendation/database"
	"github.com/Gravitalia/recommendation/helpers"
	"github.com/Gravitalia/recommendation/model"
)

// Get handle the route for /for_you_feed and return
// the posts that the user may like
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

	followingPost, err := database.LastFollowingPost(vanity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Cannot get last following posts",
		})
		return
	}

	communityPost, err := database.LastCommunityPost(vanity)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Cannot get last community posts",
		})
		return
	}

	var posts []model.Post
	posts = append(posts, tagPost...)
	posts = append(posts, followingPost...)
	posts = append(posts, communityPost...)

	posts = helpers.RemoveDuplicates(posts)

	ids := []string{}

	for _, post := range posts {
		ids = append(ids, post.Id)
	}

	posts, _ = database.JaccardRank(vanity, ids)

	jsonEncoder.Encode(posts)
}
