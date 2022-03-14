package main

import (
	"bytes"
	"encoding/csv"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gosimple/unidecode"
)

type HomepageMenuContentSingle struct {
	Name string
	Href string
}

//type FullHomepageMenuContents []HomepageMenuContentSingle

type HomePageTemplateInput struct {
	List          []FullListIndex
	Config        ConfigStr
	Name          string
	Table         []HomepageMenuContentSingle
	SearchTerm    string
	CommitHTML    string
	CommitSnippet string
}

func HomePage(isTLS bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		/*
			If path is not equal to /, treat it as calling a root static file
		*/
		if r.URL.Path != "/" {
			GetRootStaticFile(w, r)
			return
		}

		if FullConfig.TorAddress != "" && isTLS {
			w.Header().Add("Onion-Location", FullConfig.TorAddress)
		}

		fullList := GetAllLists()

		var inp HomePageTemplateInput
		inp.Config = *FullConfig
		inp.List = fullList
		inp.Table = GetHomePageMenuContents()
		inp.CommitHTML, inp.CommitSnippet = GetLinkToCommitInRepositry(FullConfig.Commit)

		/*
			Parse search and if there is a searched query,
			search for composers by that name
		*/

		searchQueryTermList := r.URL.Query()["search"]
		if len(searchQueryTermList) != 0 && len(searchQueryTermList[0]) != 0 {
			searchQueryTerm := searchQueryTermList[0]
			inp.SearchTerm = searchQueryTerm
			log.Println("searching", searchQueryTerm)
			inp.List = GetResultsSearchComposerIndex(inp.List[:], searchQueryTerm)
		}

		homeTemplatePath := "./template/homepage.html"
		tmp, err := template.ParseFiles(homeTemplatePath)
		if err != nil {
			http.Error(w, "Internal server error", 500)
			return
		}
		tmp.Execute(w, inp)
	}
}

func GetHomePageMenuContents() (menu []HomepageMenuContentSingle) {
	filename := "homepageMenuContents.csv"
	file, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Error reading"+filename, err)
		return
	}

	breader := bytes.NewReader(file)

	contentsCSV, err := csv.NewReader(breader).ReadAll()

	if err != nil {
		log.Println("Error reading csv from ", filename)
		return
	}

	for _, row := range contentsCSV {
		/*
			[0] = Name
			[1] = Href
		*/
		var nextItem HomepageMenuContentSingle
		nextItem.Name = row[0]
		nextItem.Href = row[1]
		menu = append(menu, nextItem)
	}

	return
}

func GetResultsSearchComposerIndex(origContents []FullListIndex, query string) (finalContents []FullListIndex) {
	for _, row := range origContents {
		nameToSearch := row.Name
		nameToSearchUni := unidecode.Unidecode(nameToSearch)

		if strings.Contains(strings.ToLower(nameToSearch), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(nameToSearchUni), strings.ToLower(query)) {
			// Found a match
			finalContents = append(finalContents, row)
		}
	}
	return
}
