package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"os"
)

type HomepageMenuContentSingle struct {
	Name string
	Href string
}

type FullHomepageMenuContents []HomepageMenuContentSingle

type HomePageTemplateInput struct {
	List  []FullListIndex
	Name  string
	Table []HomepageMenuContentSingle
}

func HomePage(w http.ResponseWriter, r *http.Request) {

	/*
		If path is not equal to /, treat it as calling a root static file
	*/
	if r.URL.Path != "/" {
		GetRootStaticFile(w, r)
		return
	}

	fullList := GetAllLists()

	var inp HomePageTemplateInput
	inp.Name = FullConfig.Name
	inp.List = fullList
	inp.Table = GetHomePageMenuContents()

	homeTemplatePath := "./template/homepage.html"
	tmp, err := template.ParseFiles(homeTemplatePath)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}
	tmp.Execute(w, inp)
}

func GetHomePageMenuContents() (menu FullHomepageMenuContents) {
	filename := "homepageMenuContents.csv"
	file, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading"+filename, err)
		return
	}

	breader := bytes.NewReader(file)

	contentsCSV, err := csv.NewReader(breader).ReadAll()

	if err != nil {
		fmt.Println("Error reading csv from ", filename)
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
