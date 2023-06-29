package router

import (
	"fmt"
	"net/http"
)

const (
	ErrorGetLatestCommunity = "Cannot get last community posts"
	ErrorGetLatestFollowing = "Cannot get last following posts"
	ErrorGetLatestLikedPost = "Cannot get the latest posts from the most recently liked post tag"
	ErrorInvalidToken       = "Invalid token"
)

// Index is the main route,which is notably there
// for the healthcheck
func Index(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "OK")
}
