package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"log"
	"os"
)

type FullCurrentList []CurrentSingle

const (
	WVEntryRowLength = 4
)

type WVEntry struct {
	Classifier string
	Number     string
	Title      string
	Incipit    string
	ID         string // used for lilypond
	RowNumber  int    // used for lilypond
}

type Note struct {
	Message string
	Author  string
	DateSTR string
}

type CurrentSingle struct {
	Down   string
	Up     string
	Insert string
	Delete string
	Rows   int

	ID            string // used for edit page
	ComposerFirst string `json:"first"`
	ComposerLast  string `json:"last"`
	ComposerBirth int    `json:"birth"`
	ComposerDeath int    `json:"death"`
	Notes         []Note
	WVList        []WVEntry
	Lock          string `json:"lock"` // basic lock, prevents from further editing
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

	ComposerInfo.ID = id

	var CSVList [][]string
	CSVList, err = csv.NewReader(bytes.NewReader(fileList)).ReadAll()
	if err != nil {
		return nil, err
	}
	log.Println(CSVList)

	/*
		Parse CSV list into []WVEntry
	*/

	var AllWVEntries []WVEntry

	for rowNumber, row := range CSVList {
		if len(row) != WVEntryRowLength {
			return nil, errors.New("CSV: row with invalid length found")
		}

		log.Println(row)
		var newEntry WVEntry
		newEntry.Classifier = row[0]
		newEntry.Number = row[1]
		newEntry.Title = row[2]
		newEntry.Incipit = row[3]
		newEntry.RowNumber = rowNumber
		newEntry.ID = id

		if err != nil {
			return nil, errors.New("CSV: non-number found in column 1")
		}

		AllWVEntries = append(AllWVEntries, newEntry)
	}

	ComposerInfo.WVList = AllWVEntries

	/*
		Parse csv notes (written by the contributors)
	*/

	var CSVNotesList [][]string
	CSVNotesList, err = csv.NewReader(bytes.NewReader(fileNotes)).ReadAll()
	if err != nil {
		return nil, err
	}

	var allNotes []Note

	for _, row := range CSVNotesList {
		var n Note
		n.Author = row[0]
		n.DateSTR = row[1]
		n.Message = row[2]

		allNotes = append(allNotes, n)
	}

	ComposerInfo.Notes = allNotes

	log.Println(fileNotes, ComposerInfo)

	/*
		Add up down insert and delete chars for the edit page
	*/

	ComposerInfo.Up = UpChar
	ComposerInfo.Down = DownChar
	ComposerInfo.Insert = InsertChar
	ComposerInfo.Delete = DeleteChar
	ComposerInfo.Rows = HowManyRowsToAddAtATime

	return &ComposerInfo, nil
}
