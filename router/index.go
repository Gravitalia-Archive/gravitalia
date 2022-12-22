package router

import (
	"fmt"
	"net/http"
)

func Index(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "OK")
}
