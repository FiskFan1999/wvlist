package main

import (
	"html/template"
	"net/http"
)

type SubmitPageTemplateInput struct {
	Config            ConfigStr
	SubmissionMessage string
}

func SubmitPage(w http.ResponseWriter, r *http.Request) {
	/*
		My plan is that this page only handles
		serving the page to submit.

		The json should be sent to an API endpoint because of
		javascript sending json.
	*/

	tmp, err := template.ParseFiles("./template/submitPage.html")
	if err != nil {
		http.Error(w, "Internal server error "+err.Error(), 500)
		return
	}

	var inp SubmitPageTemplateInput
	inp.Config = *FullConfig
	inp.SubmissionMessage = "sub message here"

	tmp.Execute(w, inp)
}
