package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type ConfigAdminInfo struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	HideEmail bool   `json:"hideemail"`
	Username  string `json:"username"`
	Hash      string `json:"hash"`
}

type ConfigStr struct {
	Name            string            `json:"name"`
	LilypondVersion string            `json:"lilypond_version"`
	SmtpDestination string            `json:"smtp_destination"`
	SmtpPort        int               `json:"smtp_port"`
	SmtpUsername    string            `json:"smtp_username"`
	SmtpPassword    string            `json:"smtp_password"`
	SmtpAdminBCC    []string          `json:"smtp_adminbcc"`
	Admins          []ConfigAdminInfo `json:"admins"`
	Commit          string            // set by linker
	Version         string            // set by linker
}

var FullConfig *ConfigStr

func RehashConfig() error {
	/*
		Rehash Config file, read from
	*/
	var contents []byte
	var err error

	contents, err = ioutil.ReadFile(Params.ConfigPath)
	if err != nil {
		return err
	}

	var tmpConf ConfigStr

	if !verifyConfig(contents) {
		return errors.New("Invalid JSON")
	}

	err = json.Unmarshal(contents, &tmpConf)
	if err != nil {
		return err
	}

	return json.Unmarshal(contents, FullConfig)

}

func verifyConfig(contents []byte) bool {
	/*
		Will check that the json config
		file is a valid JSON file, and
		check that any necessary values are
		present
	*/
	if !json.Valid(contents) {
		return false
	}

	/*
		To do: verify each key in the json
		maps to the correct type (int, string, etc)
	*/

	return true
}
