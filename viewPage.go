/*
handlerFunc for viewing a page.
such as wvlist.net/view/bach
*/
package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
)

func ViewPage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimRight(r.URL.Path[6:], "/") //sanitize input

	var WV *CurrentSingle
	var err error
	WV, err = ParseCurrentSingle(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "404 Page not found.", 404)
			fmt.Fprintln(w, err)
			return
		}
	}

	tmp, err := template.ParseFiles("./template/view.html")
	if err != nil {
		http.Error(w, "500 internal server error", 500)
		fmt.Fprintln(w, err)
		return
	}
	tmp.Execute(w, *WV)
}
