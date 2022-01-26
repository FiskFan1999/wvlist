package main

import (
	"net/http"
	"strings"
)

func GetRootStaticFile(w http.ResponseWriter, r *http.Request) {
	sanitisedPath := strings.Trim(r.URL.Path, "/") // sanitize input
	http.ServeFile(w, r, "./rootstatic/"+sanitisedPath)
}
