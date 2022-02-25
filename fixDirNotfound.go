/*
Some necessary directories which are not included in the
repository (because they begin as empty) will lead to
panics while running. Alleviate this by creating these
directories if they are not present.
*/
package main

import (
	"fmt"
	"io/fs"
	"os"
)

var NecessaryDirs []string = []string{
	"current",
	"lilypond",
	"submissions",
}

var FailDirs []string = []string{
	/*
		If these directories are not found,
		fail straight away (something went
		very wrong).
	*/
	"rootstatic",
	"template",
}

func CheckForNeededDirs() error {
	/*
		Will create necessary directories.
		Note that after this function is called
		the daemon will continue to run and not
		crash or return an error, UNLESS an
		odd error (i.e. not a "directory doesn't
		exist" error) occurs.

		(print a warning if a directory needs
		to be added)
	*/

	for _, d := range NecessaryDirs {
		/*
			Use OS package to try to open
			directory, will return error if
			doesn't exist.
		*/
		_, err := os.ReadDir(d)
		if err == nil {
			continue
		}

		if os.IsNotExist(err) {
			/*
				Create the directory.
			*/
			var filemode fs.FileMode = 0777
			if err2 := os.Mkdir(d, filemode); err2 != nil {
				return err2
			}

			fmt.Println("NOTE: Directory", d, "is required but not present. Creating directory now.")

		} else {
			/*
				Weird...we got another error, return that.
			*/
			return err
		}

	}

	for _, f := range FailDirs {
		_, err := os.ReadDir(f)
		if err != nil {
			fmt.Println("FATAL ERROR, critical directory", f, "not found!")
			os.Exit(2)
		}

	}
	return nil
}
