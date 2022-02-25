/*
Check if the lilypond library is
linked correctly (binary is
listed correctly in config.json)
and give a warning if it is not
(but do not stop the program
immediately)
*/
package main

import (
	"os/exec"
)

func CheckForLilypondAtStart() (output []byte, err error) {
	command := exec.Command(FullConfig.LilypondPath, "--version")
	output, err = command.CombinedOutput()
	return
}
