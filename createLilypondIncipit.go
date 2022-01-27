/*

This function will be called when the function
GetLilypond 404's. Run the lilypond executable
to create the file.

*/
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

type LilypondTemplateInput struct {
	LilypondVersion string
	Score           string
}

func GetLilypondExec(in, out string) *exec.Cmd {
	cmd := exec.Command("lilypond", "-dbackend=eps", "-dsafe", "--png", "-o", out, in)
	cmd.Dir = "./lilypond"
	return cmd
}

func CreateLilypondIncipit(originalScore, filename string) {
	fmt.Println("Now making incipit with score " + originalScore)

	// Load the lilypond template
	tmp, err := template.ParseFiles("lilypond_template")
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

	command := GetLilypondExec(tmpFile.Name(), filename)
	combinedOutput, err := command.CombinedOutput()
	fmt.Println(string(combinedOutput))
	if err != nil {
		fmt.Println("combined output error", err)
		return
	}

}
