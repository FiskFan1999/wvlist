package main

import (
	"html/template"
	"net/http"
)

type HomePagePar struct {
}

func HomePage(w http.ResponseWriter, r *http.Request) {

	/*
		Check if the path is not equal to "/"
		(The request is for a page we don't have)
	*/
	if r.URL.Path != "/" {
		http.Error(w, "Page Not Found", 404)
		return
	}

	/*

		WVEntry, _ := ParseCurrentSingle("bach")
		for _, row := range WVEntry.WVList {
			fmt.Fprintln(w, row)
		}
	*/

	homeTemplatePath := "./template/homepage.html"
	tmp, err := template.ParseFiles(homeTemplatePath)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}
	tmp.Execute(w, nil)
}
