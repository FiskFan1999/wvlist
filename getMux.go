package main

import (
	"net/http"
)

func GetMux(isTls bool) (m *http.ServeMux) {

	/*
		Creates the muxer that will be used in
		http.ListenAndServe. Note that some pages
		are intended to be different depending on if
		viewed over plaintext or over TLS.
	*/

	m = http.NewServeMux()

	m.HandleFunc("/", HomePage(isTls))
	m.HandleFunc("/view/", ViewPage)
	m.HandleFunc("/submit/", SubmitPage)
	m.HandleFunc("/edit/", GetEditPage)
	m.HandleFunc("/lilysand/", LilypondSandbox)
	m.HandleFunc("/incipit/", GetLilypond)
	m.HandleFunc("/api/v1/", APIv1Handler)

	var acHand func(http.ResponseWriter, *http.Request) = AdminConsolePlaintextHandler
	if isTls {
		acHand = AdminConsole
	}
	m.HandleFunc("/admin/", acHand)

	return
}
