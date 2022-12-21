package router

import (
	"fmt"
	"net/http"
)

func New(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Created")
}
