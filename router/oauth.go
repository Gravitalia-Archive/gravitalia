package router

import (
	"fmt"
	"net/http"
)

func OAuth(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL.Query())

	fmt.Fprintf(w, "OK")
}
