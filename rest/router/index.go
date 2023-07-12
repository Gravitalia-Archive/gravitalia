package router

import (
	"fmt"
	"net/http"
)

const ME = "@me"

// Every possible error list
const (
	ErrorDataRequested         = "Data requested less than 24 hours ago"
	ErrorExceededMaximumImages = "Maximum images exceeded"
	ErrorInvalidList           = "Invalid list"
	ErrorInvalidPost           = "Invalid post"
	ErrorInvalidPostAccess     = "No access to this post"
	ErrorInternalServerError   = "Internal server error"
	ErrorInvalidToken          = "Invalid token"
	ErrorInvalidBody           = "Invalid body"
	ErrorInvalidRelation       = "Invalid relation"
	ErrorInvalidQuery          = "Invalid query"
	ErrorInvalidUser           = "Invalid user"
	ErrorMethodNotAllowed      = "Method not allowed"
	ErrorUnableReadBody        = "Unable to read body"
)

// Every OK message reponse
const (
	Ok                = "OK"
	OkAddedRequest    = "Request added"
	OkCreatedRelation = "Created relation"
	OkDeletedRelation = "Deleted relation"
)

func Index(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "OK")
}
