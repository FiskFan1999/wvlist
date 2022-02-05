package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	LilypondDirPath = "./lilypond/"
)

type LilypondFileToMakeStr struct {
	Command  string
	Filename string
}

var LilypondFilesToMake chan LilypondFileToMakeStr

func GetModifiedFilename(s string) string {
	/*
		TODO: do this with a for loop instead
	*/
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(s, "/", ""), "\\", ""), "'", ""), "\"", "")
}

func GetLilypond(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	if len(query["id"]) != 1 || len(query["no"]) != 1 {
		http.Error(w, "400 Bad Request", 400)
		return
	}

	/*
		Is a valid request.
	*/

	id := query["id"][0]
	no := query["no"][0]

	/*
		Remove those keys and see if there are other queries
		(bad request)
	*/
	delete(query, "id")
	delete(query, "no")
	if len(query) != 0 {
		http.Error(w, "400 Bad Request", 400)
		return
	}

	/*
		Safely continue with id no
	*/

	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "%s%s.%s.png", LilypondDirPath, id, no)
	LilypondImageFilename := buf.String()

	if _, err := os.ReadFile(buf.String()); err != nil && os.IsNotExist(err) {
		// File doesn't exist.
		go func() {
			var newFile LilypondFileToMakeStr

			/*
				Read the incipit command from the file
			*/

			path := "./current/" + id + ".csv"

			contents, err := os.ReadFile(path)
			if err != nil {
				fmt.Println("getting lilypond file error", err.Error())
				return
			}

			r := csv.NewReader(bytes.NewReader(contents))
			fileContents, err := r.ReadAll()
			fmt.Println(fileContents)

			/*
				Check that the row is not out of bounds
			*/

			noInt, err := strconv.Atoi(no)
			if err != nil {
				http.Error(w, "400 bad request", 400)
				return
			}

			if noInt >= len(fileContents) {
				http.Error(w, "400 bad request", 400)
				return
			}
			incipit := fileContents[noInt][3]
			fmt.Println(incipit)

			newFile.Command = incipit
			newFile.Filename = LilypondImageFilename
			LilypondFilesToMake <- newFile
		}()

	} else {
		http.ServeFile(w, r, LilypondImageFilename)
	}
}

func LilypondWriteIncipitsFromChannel() {
	/*
		This function is run in a subroutine.
		For each entry recieved, write the lilypond
		incipit to the file.
	*/
	for true {
		var newFile LilypondFileToMakeStr
		newFile = <-LilypondFilesToMake

		/*
			First check if this file may have already been made
		*/

		if _, err := os.ReadFile(newFile.Filename); !(err != nil && os.IsNotExist(err)) {
			continue
		}

		filename := strings.TrimLeft(strings.TrimRight(newFile.Filename, ".png"), LilypondDirPath) // .png is appended by lilypond
		CreateLilypondIncipit(newFile.Command, filename)
	}

}

func OldGetLilypond(w http.ResponseWriter, r *http.Request) {
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
