package main

import (
	"os"
)

func CreateTemp(dir, pattern string) (file *os.File, err error) {
	/*
		os.CreateTemp creates files with bad perms.
		This function wraps around os.CreateTemp and sets perms
		to 0666 (will get masked by whatever perms ./wvlist
		is running with). This way, all files will be editably by
		the admin without sudo.
	*/
	file, err = os.CreateTemp(dir, pattern)
	if err != nil {
		return
	}

	err = file.Chmod(0666)

	return
}
