package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type V1UploadUglySanitizedInput struct {
	ComposerFirst   string
	ComposerLast    string
	ComposerBirth   int
	ComposerDeath   int
	Notes           string
	SubmitName      string
	SubmitEmail     string
	CompositionList []WVEntry
}

func APIv1Handler(w http.ResponseWriter, r *http.Request) {
	var argvraw []string
	for _, arg := range strings.Split(r.URL.Path, "/")[3:] {
		if len(strings.TrimSpace(arg)) != 0 {
			argvraw = append(argvraw, strings.ToLower(arg))
		}
	}

	command := argvraw[0]
	argv := argvraw[1:]

	switch command {
	case "uploadugly":
		if r.Method != "POST" {
			http.Error(w, "405 Only POST method is allowed", 405)
			return
		}
		if len(argv) != 0 {
			http.Error(w, "400 Bad Request (no path after /api/v1/uploadugly)", 400)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "500 Internal Server Error "+err.Error(), 500)
			return
		}

		var bodySliceMaps []map[string]string

		err = json.Unmarshal(body, &bodySliceMaps)
		if err != nil {
			http.Error(w, "500 Internal Server Error "+err.Error(), 500)
			return
		}

		if !UploadUglyCheckForCorrectPostBody(bodySliceMaps) {
			http.Error(w, "400 Bad Request", 400)
			return
		}

		var sanitizedInputStr V1UploadUglySanitizedInput

		for _, row := range bodySliceMaps {
			key, ok1 := row["name"]
			value, ok2 := row["value"]
			if !ok1 || !ok2 {
				http.Error(w, "400 Unknown key "+row["name"], 400)
				return
			}

			switch key {
			case "composerFirst":
				sanitizedInputStr.ComposerFirst = value
			case "composerLast":
				sanitizedInputStr.ComposerLast = value
			case "birth":
				sanitizedInputStr.ComposerBirth, err = strconv.Atoi(value)
				if err != nil {
					sanitizedInputStr.ComposerBirth = 0
				}
			case "death":
				sanitizedInputStr.ComposerDeath, err = strconv.Atoi(value)
				if err != nil {
					sanitizedInputStr.ComposerDeath = 0
				}
			case "notes":
				sanitizedInputStr.Notes = value
			case "submitname":
				sanitizedInputStr.SubmitName = value
			case "email":
				sanitizedInputStr.SubmitEmail = value
			case "classification":
				var newWVEntry WVEntry
				sanitizedInputStr.CompositionList = append(sanitizedInputStr.CompositionList, newWVEntry)
				index := len(sanitizedInputStr.CompositionList) - 1
				sanitizedInputStr.CompositionList[index].Classifier = value
			case "number":
				index := len(sanitizedInputStr.CompositionList) - 1
				sanitizedInputStr.CompositionList[index].Number, err = strconv.Atoi(value)
				if err != nil {
					fmt.Println("BWV number error")
					sanitizedInputStr.CompositionList[index].Number = -1
				}
			case "other":
				index := len(sanitizedInputStr.CompositionList) - 1
				sanitizedInputStr.CompositionList[index].Extra = value
			case "title":
				index := len(sanitizedInputStr.CompositionList) - 1
				sanitizedInputStr.CompositionList[index].Title = value
			case "incipit":
				index := len(sanitizedInputStr.CompositionList) - 1
				sanitizedInputStr.CompositionList[index].Incipit = value
			default:
				fmt.Println("unknown key", row["name"])
				if false {
					http.Error(w, "400 Unknown key "+row["name"], 400)
					return
				}
			}
		}
		fmt.Printf("%+v\n", sanitizedInputStr)

		/*
			TODO: send SMTP email to sanitizedInputStr.SubmitEmail
		*/

		/*
			Marshall sanitizedInputStr to a new tempfile in /submissions/
		*/

		file, err := os.CreateTemp("./submissions", "submission.*.unverified")
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "500 Internal server error "+err.Error(), 500)
			return
		}
		defer file.Close()

		marshaled, err := json.Marshal(sanitizedInputStr)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "500 Internal server error "+err.Error(), 500)
			return
		}

		if _, err = file.Write(marshaled); err != nil {
			fmt.Println(err.Error())
			http.Error(w, "500 Internal server error "+err.Error(), 500)
			return
		}

		w.WriteHeader(201)

	default:
		http.Error(w, "404 Not Found", 404)
		return
	}
}

func UploadUglyCheckForCorrectPostBody(body []map[string]string) bool {
	/*
		TODO: check for correctly formatted input.
		Check that all fields have been sent even
		if some of them are empty, and check for
		correct order of
	*/
	return true
}
