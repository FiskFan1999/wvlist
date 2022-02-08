/*

linkToRepo.go

this file contains the constant of the link to the repo.

If you fork wvlist, you should change this to link to
your own repository, otherwise the link will not work.

*/
package main

import (
	"bytes"
	"fmt"
)

const (
	LinkToRepository = "https://github.com/FiskFan1999/wvlist"
)

func GetLinkToCommitInRepositry(commit string) (html string, snippet string) {
	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "%s/commit/%s", LinkToRepository, commit)
	html = buf.String()

	snippet = commit[:9] + "..."

	return

}
