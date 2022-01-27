package main

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"strings"
)

func GetModifiedFilename(s string) string {
	/*
		TODO: do this with a for loop instead
	*/
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "/", ""), "\\", ""), "'", ""), "\"", "")
}

func GetLilypond(w http.ResponseWriter, r *http.Request) {
	// serve lilypond png file
	pathSan := (r.URL.Query()["f"])
	if len(pathSan) != 1 {
		http.Error(w, "400 Bad request", 400)
		return
	}
	pathSanEscaped := html.UnescapeString(pathSan[0])
	filepathWD := GetModifiedFilename(pathSanEscaped)
	filepath := "./lilypond/" + filepathWD
	file, err := os.Open(filepath)
	file.Close()
	if errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(w, pathSanEscaped, filepathWD)
		go CreateLilypondIncipit(strings.TrimRight(pathSanEscaped, ".png"), strings.TrimRight(filepathWD, ".png"))
		http.Error(w, "404 not found", 404)
		return

	}
	http.ServeFile(w, r, filepath)
}
