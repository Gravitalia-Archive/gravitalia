package router

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/model"
)

func Users(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	username := strings.TrimPrefix(req.URL.Path, "/users/")
	stats, err := database.GetUserStats(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid user",
		})
		return
	}

	posts, err := database.GetUserPost(username, 0)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonEncoder.Encode(model.RequestError{
			Error:   true,
			Message: "Invalid user",
		})
		return
	}

	jsonEncoder.Encode(struct {
		Followers int64        `json:"followers"`
		Following int64        `json:"following"`
		Posts     []model.Post `json:"posts"`
	}{
		Followers: stats.Followers,
		Following: stats.Following,
		Posts:     posts,
	})
}
