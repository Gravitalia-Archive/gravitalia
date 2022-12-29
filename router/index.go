package router

import (
	"fmt"
	"net/http"
)

func Index(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "OK")
}
