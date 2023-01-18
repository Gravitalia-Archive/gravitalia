package router

import (
	"fmt"
	"net/http"
)

func GetPost(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "OK")
}
