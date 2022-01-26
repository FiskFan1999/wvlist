package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"os"
	"strconv"
)

type FullCurrentList []CurrentSingle

const (
	WVEntryRowLength = 5
)

type WVEntry struct {
	Classifier  string
	Number      int
	Extra       string
	Title       string
	Incipit     string
	IncipitHTML string
}

type Note struct {
	Message string
	Author  string
	DateSTR string
}

type CurrentSingle struct {
	ID            string
	ComposerFirst string `json:"first"`
	ComposerLast  string `json:"last"`
	ComposerBirth int    `json:"birth"`
	ComposerDeath int    `json:"death"`
	Notes         []Note
	WVList        []WVEntry
}

func GetAllCurrent() FullCurrentList {
	return FullCurrentList{}
}

func ParseCurrentSingle(id string) (*CurrentSingle, error) {
	// "id" is the path to the collection of files in current/
	// (for example, current/bach.json current/bach.notes current/bach.csv -> "bach"

	info := "./current/" + id + ".json"
	notes := "./current/" + id + ".notes"
	list := "./current/" + id + ".csv"

	// Check if any of these three to not exist.
	fileInfo, errInfo := os.ReadFile(info)
	fileNotes, errNotes := os.ReadFile(notes)
	fileList, errList := os.ReadFile(list)

	if errInfo != nil {
		return nil, errInfo
	}
	if errNotes != nil {
		return nil, errNotes
	}
	if errList != nil {
		return nil, errList
	}

	var ComposerInfo CurrentSingle

	err := json.Unmarshal(fileInfo, &ComposerInfo)
	if err != nil {
		return nil, err
	}

	var CSVList [][]string
	CSVList, err = csv.NewReader(bytes.NewReader(fileList)).ReadAll()
	if err != nil {
		return nil, err
	}
	fmt.Println(CSVList)

	/*
		Parse CSV list into []WVEntry
	*/

	var AllWVEntries []WVEntry

	for _, row := range CSVList {
		if len(row) != WVEntryRowLength {
			return nil, errors.New("CSV: row with invalid length found")
		}

		fmt.Println(row)
		var newEntry WVEntry
		newEntry.Classifier = row[0]
		newEntry.Number, err = strconv.Atoi(row[1])
		newEntry.Extra = row[2]
		newEntry.Title = row[3]
		newEntry.Incipit = row[4]
		newEntry.IncipitHTML = html.EscapeString(row[4])

		if err != nil {
			return nil, errors.New("CSV: non-number found in column 1")
		}

		AllWVEntries = append(AllWVEntries, newEntry)
	}

	ComposerInfo.WVList = AllWVEntries

	fmt.Println(fileNotes, ComposerInfo)

	return &ComposerInfo, nil
}
