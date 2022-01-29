package main

import (
	"fmt"
	"net/http"
	"strings"
)

func APIv1Handler(w http.ResponseWriter, r *http.Request) {
	var argvraw []string
	for _, arg := range strings.Split(r.URL.Path, "/")[3:] {
		if len(strings.TrimSpace(arg)) != 0 {
			argvraw = append(argvraw, strings.ToLower(arg))
		}
	}

	command := argvraw[0]
	argv := argvraw[1:]
	fmt.Println(argv)

	switch command {
	case "uploadugly":
		/*
			Placeholder
		*/
		http.Error(w, "501 Not Implemented", 501)
		return
		if len(argv) != 0 {
			http.Error(w, "400 Bad Request (no path after /api/v1/uploadugly)", 400)
		}
	default:
		http.Error(w, "404 Not Found", 404)
		return
	}
}
