/*

This function will be called when the function
GetLilypond 404's. Run the lilypond executable
to create the file.

*/
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"text/template"
	"time"
)

const (
	LILYPOND_TEMPLATE_FILE = "lilypond_template"
)

type LilypondTemplateInput struct {
	LilypondVersion string
	Score           string
}

func GetLilypondExec(in, out, dir string) (*exec.Cmd, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()
	cmd := exec.CommandContext(ctx, "lilypond", "-dbackend=eps", "-dsafe", "--png", "-o", out, in)
	cmd.Dir = dir
	return cmd, cancel
}

func CreateLilypondIncipit(originalScore, filename string) {
	fmt.Println("Now making incipit with score " + originalScore)

	// Load the lilypond template
	tmp, err := template.ParseFiles(LILYPOND_TEMPLATE_FILE)
	if err != nil {
		fmt.Println("CreateLilypondIncipit error: ", err)
		return
	}

	// Fill out a LilypondTemplateInput struct
	// and fill with appropriate values

	var t *LilypondTemplateInput = new(LilypondTemplateInput)
	t.LilypondVersion = FullConfig.LilypondVersion
	t.Score = originalScore

	var buf bytes.Buffer

	tmp.Execute(&buf, t)

	fmt.Println(buf.String())

	// save this to a temporary file

	tmpFile, err := os.CreateTemp("", "lilyfile")
	if err != nil {
		tmpFile.Close()
		fmt.Println("os.CreateTemp error", err)
		return
	}
	defer tmpFile.Close()
	//defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.WriteString(buf.String()); err != nil {
		fmt.Println("tmpFile.WriteString error", err)
		return
	}
	fmt.Println("written to file", tmpFile.Name())

	command, cancel := GetLilypondExec(tmpFile.Name(), filename, "./lilypond")
	defer cancel()
	combinedOutput, err := command.CombinedOutput()
	fmt.Println(string(combinedOutput))
	if err != nil {
		fmt.Println("combined output error", err)
		return
	}

}
