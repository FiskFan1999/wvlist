package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thanhpk/randstr"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type V1UploadUglySanitizedInput struct {
	ComposerFirst   string
	ComposerLast    string
	ComposerBirth   int
	ComposerDeath   int
	Notes           string
	SubmitName      string
	SubmitEmail     string
	CompositionList []WVEntry
}

func APIv1Handler(w http.ResponseWriter, r *http.Request) {
	var argvraw []string
	for _, arg := range strings.Split(r.URL.Path, "/")[3:] {
		if len(strings.TrimSpace(arg)) != 0 {
			argvraw = append(argvraw, strings.ToLower(arg))
		}
	}

	command := argvraw[0]
	argv := argvraw[1:]

	switch command {
	case "verify":
		if len(argv) != 2 {
			http.Error(w, "Bad request: usage /api/v1/verify/<id>/<password>", 400)
			return
		}
		id := argv[0]
		password := argv[1]

		BadRequestMessage := "Bad Request: This submission already verified, submission file not found, or password incorrect"
		mainFileName := "./submissions/submission." + id + ".unverified"
		mainFileNameIfAccepted := "./submissions/submission." + id + ".verified"
		PasswordFileName := "./submissions/submission." + id + ".password"

		_, err := os.ReadFile(mainFileName)
		if err != nil && os.IsNotExist(err) {
			http.Error(w, BadRequestMessage, 400)
			return
		}
		//If that test succeeds, check the password

		passwordFileText, err := os.ReadFile(PasswordFileName)
		if err != nil && os.IsNotExist(err) {
			http.Error(w, BadRequestMessage, 400)
			return
		}

		if string(passwordFileText) == password {
			// Password accepted.
			// change the filename to accepted submission
			if linkerr := os.Rename(mainFileName, mainFileNameIfAccepted); linkerr != nil {
				http.Error(w, linkerr.Error(), 500)
				return
			}

		} else {
			http.Error(w, BadRequestMessage, 400)
			return
		}

		http.Redirect(w, r, "/", 200)

	case "uploadugly":
		if r.Method != "POST" {
			SendJSONSuccessOrErrorMessage(w, false, "405 Only POST method is allowed", 405)
			return
		}
		if len(argv) != 0 {
			SendJSONErrorMessage(w, "400 Bad Request (no path after /api/v1/uploadugly)")
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			SendJSONInternalErrorMessage(w, "500 Internal server error: "+err.Error())
			return
		}

		var bodySliceMaps []map[string]string

		err = json.Unmarshal(body, &bodySliceMaps)
		if err != nil {
			SendJSONInternalErrorMessage(w, "500 Internal server error: "+err.Error())
			return
		}

		if err = UploadUglyCheckForCorrectPostBody(bodySliceMaps); err != nil {
			SendJSONErrorMessage(w, "400 bad request: "+err.Error())
			return
		}

		var sanitizedInputStr V1UploadUglySanitizedInput

		for _, row := range bodySliceMaps {
			key, ok1 := row["name"]
			value, ok2 := row["value"]
			if !ok1 || !ok2 {
				SendJSONErrorMessage(w, "400 Unknown key: "+row["name"])
				return
			}

			switch key {
			case "composerFirst":
				sanitizedInputStr.ComposerFirst = value
			case "composerLast":
				sanitizedInputStr.ComposerLast = value
			case "birth":
				sanitizedInputStr.ComposerBirth, err = strconv.Atoi(value)
				if err != nil {
					sanitizedInputStr.ComposerBirth = 0
				}
			case "death":
				sanitizedInputStr.ComposerDeath, err = strconv.Atoi(value)
				if err != nil {
					sanitizedInputStr.ComposerDeath = 0
				}
			case "notes":
				sanitizedInputStr.Notes = value
			case "submitname":
				sanitizedInputStr.SubmitName = value
			case "email":
				sanitizedInputStr.SubmitEmail = value
			case "classification":
				var newWVEntry WVEntry
				sanitizedInputStr.CompositionList = append(sanitizedInputStr.CompositionList, newWVEntry)
				index := len(sanitizedInputStr.CompositionList) - 1
				sanitizedInputStr.CompositionList[index].Classifier = value
			case "number":
				index := len(sanitizedInputStr.CompositionList) - 1
				sanitizedInputStr.CompositionList[index].Number = value
			case "title":
				index := len(sanitizedInputStr.CompositionList) - 1
				sanitizedInputStr.CompositionList[index].Title = value
			case "incipit":
				index := len(sanitizedInputStr.CompositionList) - 1
				sanitizedInputStr.CompositionList[index].Incipit = value
			default:
				fmt.Println("unknown key", row["name"])
				if false {
					SendJSONErrorMessage(w, "400 Unknown key: "+row["name"])
					return
				}
			}
		}
		fmt.Printf("%+v\n", sanitizedInputStr)

		/*
			Marshall sanitizedInputStr to a new tempfile in /submissions/
		*/

		file, err := os.CreateTemp("./submissions", "submission.*.unverified")
		if err != nil {
			fmt.Println(err.Error())
			SendJSONInternalErrorMessage(w, "500 Internal server error: "+err.Error())
			return
		}
		defer file.Close()

		fileNoEnding := strings.TrimRight(file.Name(), ".unverified")
		fileIndex := strings.TrimLeft(fileNoEnding, "./submissions/submission.")

		marshaled, err := json.Marshal(sanitizedInputStr)
		if err != nil {
			fmt.Println(err.Error())
			SendJSONInternalErrorMessage(w, "500 Internal server error: "+err.Error())
			return
		}

		if _, err = file.Write(marshaled); err != nil {
			fmt.Println(err.Error())
			SendJSONInternalErrorMessage(w, "500 Internal server error: "+err.Error())
			return
		}

		// Generate a random sequence of bytes to be send for the email verification

		verifyPassword := randstr.Hex(16)
		fmt.Println("password", (verifyPassword))
		passwordFilename := fileNoEnding + ".password"
		err = os.WriteFile(passwordFilename, []byte(verifyPassword), 0666)
		if err != nil {
			fmt.Println("password file error", err.Error())
			SendJSONInternalErrorMessage(w, "500 Internal server error (password file): "+err.Error())
			return
		}

		if strings.TrimSpace(sanitizedInputStr.SubmitEmail) != "" {
			if err = Apiv1SendSmtpEmailForSubmitUgly(sanitizedInputStr, fileIndex, verifyPassword); err != nil {
				SendJSONErrorMessage(w, err.Error())
				return
			}
		}
		SendJSONSuccessMessage(w, "Thank you for your submission.")

	default:
		http.Error(w, "404 Not Found", 404)
		return
	}
}

func SendJSONSuccessOrErrorMessage(w http.ResponseWriter, success bool, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{}
	if success {
		resp["status"] = "success"
	} else {
		resp["status"] = "error"
	}
	resp["message"] = message
	respBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		panic("json marshal error" + err.Error())
	}
	w.Write(respBytes)

	w.WriteHeader(status)
}

func SendJSONInternalErrorMessage(w http.ResponseWriter, message string) {
	SendJSONSuccessOrErrorMessage(w, false, message, 500)
}

func SendJSONErrorMessage(w http.ResponseWriter, message string) {
	SendJSONSuccessOrErrorMessage(w, false, message, 400)
}

func SendJSONSuccessMessage(w http.ResponseWriter, message string) {
	SendJSONSuccessOrErrorMessage(w, true, message, 201)
}

func UploadUglyCheckForCorrectPostBody(body []map[string]string) error {
	/*
		TODO: check for correctly formatted input.
		Check that all fields have been sent even
		if some of them are empty, and check for
		correct order of
	*/
	requiredFields := [...]string{
		"composerLast",
	}

	checkForInts := [...]string{
		"birth", "death",
	}

	//fieldWatch := make(map[string]bool)

	for _, keyvalue := range body {
		/*
			keyvalue = {
				"name": "key",
				"value": "value"
			}
		*/
		key, ok := keyvalue["name"]

		if !ok {
			// the json map handled does not have a "name" key, which violates
			// what will be sent to us by AJAX serialized json
			return errors.New("Illegal json POSTed, does not follow AJAX serialized array syntax")
		}

		value, ok := keyvalue["value"]
		if !ok {
			// the json map handled does not have a "name" key, which violates
			// what will be sent to us by AJAX serialized json
			return errors.New("Illegal json POSTed, does not follow AJAX serialized array syntax")
		}
		for _, currentRequiredField := range requiredFields {
			if currentRequiredField == key && len(strings.TrimSpace(value)) == 0 {
				return errors.New("Field " + key + " is required.")
			}
		}

		for _, currentRequiredInt := range checkForInts {
			if currentRequiredInt == key {
				if _, err := strconv.Atoi(value); err != nil {
					// fails
					return errors.New("Field " + key + " must be an integer.")
				}
			}
		}

	}

	return nil
}
