/*

This is called by the homepage function
and is used by the home page template

*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gosimple/unidecode"
	"os"
	"sort"
	"strings"
)

type FullListIndex struct {
	Name string
	Path string
}

func GetAllLists() (FullIndexList []FullListIndex) {

	dir, err := os.ReadDir("current")
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range dir {
		fname := file.Name()
		if !strings.HasSuffix(fname, ".json") {
			continue
		}
		// get the full name from opening the file
		file, _ := os.ReadFile("./current/" + fname)
		var CS *CurrentSingle // from getCurrentList.go
		CS = new(CurrentSingle)
		json.Unmarshal(file, CS)

		var i FullListIndex
		i.Name = CS.ComposerLast + ", " + CS.ComposerFirst
		i.Path = strings.TrimRight(fname, ".json")

		FullIndexList = append(FullIndexList, i)
	}

	/*
		TODO: sort these alphabetically by .Name (will not be
		alphabetically sorted)
	*/

	// Sort by alphabetical name
	sort.SliceStable(FullIndexList, func(i, j int) bool {
		return (strings.ToLower(unidecode.Unidecode(FullIndexList[i].Name)) <
			strings.ToLower(unidecode.Unidecode(FullIndexList[j].Name)))
	})

	return
}
