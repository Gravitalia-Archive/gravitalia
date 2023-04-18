package router

import (
	"fmt"
	"net/http"
)

func New(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Created")
}
