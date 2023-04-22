package router

import (
	"fmt"
	"net/http"
)

// Index is the main route,which is notably there
// for the healthcheck
func Index(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "OK")
}
