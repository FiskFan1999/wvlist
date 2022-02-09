package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
)

func GetEditPage(w http.ResponseWriter, r *http.Request) {
	argv := strings.Split(r.URL.Path, "/")
	/*
		[0] = ""
		[1] = "edit"
		[2] = ID
	*/

	if len(argv) < 3 || strings.TrimSpace(argv[2]) == "" {
		http.Error(w, "400 syntax /edit/<id>", 400)
		return
	}

	id := argv[2]

	var WV *CurrentSingle
	var err error
	WV, err = ParseCurrentSingle(id)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "404 Page not found.", 404)
			fmt.Fprintln(w, err)
			return
		} else {
			http.Error(w, err.Error(), 500)
			return
		}

	}
	tmp, err := template.ParseFiles("./template/editPage.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	tmp.Execute(w, WV)
	/*

		dirName := "./current/"
		mainFileName := dirName + id

		infoFileName := mainFileName + ".json" // just for some information to put on the edit page
		listFileName := mainFileName + ".csv"
		// don't care about notes file

		infoContents, err := os.ReadFile(infoFileName)
		if err != nil {
			http.Error(w, "404 File not Found", 404)
			return
		}

		listContents, err := os.ReadFile(listFileName)
		if err != nil {
			http.Error(w, "404 File not Found", 404)
			return
		}

		fmt.Fprintln(w, string(infoContents))
		fmt.Fprintln(w, string(listContents))
	*/

	/*
		Load the information into the same struct as the view page.
	*/

}
