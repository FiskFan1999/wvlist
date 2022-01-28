/*

This page serves as the sandbox (/lilysand)
Both as the initial page, and for the
output of the user-submitted lilypond input.

*/
package main

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"os"
	"strings"
	ttemplate "text/template"
	"time"
)

const (
	InputMaxLength = 256
)

type LilySandInput struct {
	IsPost           bool
	Command          string
	CommandMaxLength uint
	CommandOutput    string
	ImageHref        string
}

func LilypondSandbox(w http.ResponseWriter, r *http.Request) {
	tmp, err := template.ParseFiles("./template/lilypondSandbox.html")
	if err != nil {
		http.Error(w, "500 Internal Server Error", 500)
		return
	}

	var lsi LilySandInput
	query_lilypond := r.URL.Query()["lilypond"]
	lsi.Command = ""
	if len(query_lilypond) > 0 {
		lsi.Command = html.UnescapeString(strings.TrimSpace(query_lilypond[0]))
	}
	lsi.IsPost = len(lsi.Command) != 0 // Note: r.PostFormValue silently returns null string if is GET
	lsi.CommandMaxLength = InputMaxLength

	if lsi.IsPost {
		if len(lsi.Command) > InputMaxLength {
			http.Error(w, "413 Payload Too Large", 413)
			return
		}
		// Execute lilypond

		// Load the lilypond template
		lilypondTmp, err := ttemplate.ParseFiles(LILYPOND_TEMPLATE_FILE)
		if err != nil {
			fmt.Println("CreateLilypondIncipit error: ", err)
			http.Error(w, "500 Server Internal Error", 500)
			return
		}

		var lti LilypondTemplateInput
		lti.LilypondVersion = FullConfig.LilypondVersion
		lti.Score = lsi.Command

		var buf bytes.Buffer

		lilypondTmp.Execute(&buf, lti)

		//fmt.Println(buf.String())

		// save this to a temporary file

		deleteFileFunc := func(f *os.File) {
			time.Sleep(30 * time.Second)
			f.Close()
			os.Remove(f.Name())

		}

		tmpFileIn, err := os.CreateTemp("./rootstatic/", "lilyfile_in.*.ly")
		if err != nil {
			fmt.Println("os.CreateTemp error", err)
			return
		}

		_, err = tmpFileIn.WriteString(buf.String())
		if err != nil {
			fmt.Println("os.WriteString error", err)
			return
		}

		go deleteFileFunc(tmpFileIn)

		tmpFileOut, err := os.CreateTemp("./rootstatic/", "lilyfile_out.*")
		if err != nil {
			fmt.Println("os.CreateTemp error", err)
			return
		}

		go deleteFileFunc(tmpFileOut)

		tmpFileOutWD := tmpFileOut.Name()[len("./rootstatic/"):]

		go func(filename string) {
			time.Sleep(30 * time.Second)
			//Delete all files
			allFiles, err := os.ReadDir("./rootstatic")
			if err != nil {
				fmt.Println(err)
				return
			}
			for _, entry := range allFiles {
				name := entry.Name()
				if strings.HasPrefix(name, filename) {
					os.Remove("./rootstatic/" + name)
				}
				fmt.Println(name)
			}
		}(tmpFileOutWD)

		lilypondExec := GetLilypondExec(tmpFileIn.Name()[len("./rootstatic/"):], tmpFileOut.Name()[len("./rootstatic/"):], "./rootstatic")

		combinedOutput, err := lilypondExec.CombinedOutput()

		if err != nil {
		}

		lsi.CommandOutput = string(combinedOutput)
		lsi.ImageHref = tmpFileOut.Name()[len("./rootstatic"):] + ".png"

		// Execute lilypond command
	}

	tmp.Execute(w, lsi)
}
