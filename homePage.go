package main

import (
	"html/template"
	"net/http"
)

type HomePagePar struct {
}

func HomePage(w http.ResponseWriter, r *http.Request) {

	/*
		If path is not equal to /, treat it as calling a root static file
	*/
	if r.URL.Path != "/" {
		GetRootStaticFile(w, r)
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
