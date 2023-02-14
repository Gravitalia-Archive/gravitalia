package router

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Gravitalia/gravitalia/database"
	"github.com/Gravitalia/gravitalia/model"
)

func Post(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	jsonEncoder := json.NewEncoder(w)

	id := strings.TrimPrefix(req.URL.Path, "/posts/")
	post, err := database.GetPost(id)
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
