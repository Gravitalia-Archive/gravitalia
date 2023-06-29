package router

import (
	"fmt"
	"net/http"
)

const ME = "@me"

// Every possible error list
const (
	ErrorInvalidPost         = "Invalid post"
	ErrorInvalidPostAccess   = "No access to this post"
	ErrorInternalServerError = "Internal server error"
	ErrorInvalidToken        = "Invalid token"
	ErrorUnableReadBody      = "Unable to read body"
	ErrorInvalidBody         = "Invalid body"
	ErrorInvalidRelation     = "Invalid relation"
	ErrorInvalidQuery        = "Invalid query"
	ErrorInvalidUser         = "Invalid user"
)

// Every OK message reponse
const (
	Ok                = "OK"
	OkCreatedRelation = "Created relation"
	OkDeletedRelation = "Deleted relation"
)

func Index(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "OK")
}
