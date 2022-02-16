package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"
)

type ConfigAdminInfo struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	HideEmail bool   `json:"hideemail"`
	Username  string `json:"username"`
	Hash      string `json:"hash"`
}

type ConfigStr struct {
	Name            string `json:"name"`
	Hostname        string `json:"hostname"`
	TorAddress      string `json:"tor_address"`
	LilypondVersion string `json:"lilypond_version"`
	LilypondTimeout string `json:"lilypond_timeout"`
	LilyTimeStr     time.Duration
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

	if err = json.Unmarshal(contents, FullConfig); err != nil {
		return err
	}

	FullConfig.LilyTimeStr, err = time.ParseDuration(FullConfig.LilypondTimeout)
	if err != nil {
		return err
	}
	return nil

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
