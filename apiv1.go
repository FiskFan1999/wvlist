package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thanhpk/randstr"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
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

type V1UploadEditUglyBodyInput struct {
	ID              string
	Notes           string
	SubmitName      string
	SubmitEmail     string
	CompositionList [][]string
}

type V1UploadEditUglyBodyOutput struct {
	ID          string
	Notes       string
	SubmitName  string
	SubmitEmail string
	Diff        []byte
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

	case "verifyedit":
		if len(argv) != 2 {
			http.Error(w, "Bad request: usage /api/v1/verify/<id>/<password>", 400)
			return
		}
		id := argv[0]
		password := argv[1]

		BadRequestMessage := "Bad Request: This submission already verified, submission file not found, or password incorrect"
		fmt.Println(BadRequestMessage)
		mainFileName := "./submissions/edit." + id + ".unverified"
		mainFileNameIfAccepted := "./submissions/edit." + id + ".verified"
		PasswordFileName := "./submissions/edit." + id + ".password"

		_, err := os.ReadFile(mainFileName)
		if err != nil && os.IsNotExist(err) {
			//http.Error(w, BadRequestMessage, 400)
			http.Error(w, "MainFileReadError", 400)
			return
		}
		//If that test succeeds, check the password

		passwordFileText, err := os.ReadFile(PasswordFileName)
		if err != nil && os.IsNotExist(err) {
			//http.Error(w, BadRequestMessage, 400)
			http.Error(w, "passwordfile read error", 400)
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
			//http.Error(w, BadRequestMessage, 400)
			http.Error(w, "incorrect password", 400)
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

	case "uploadeditugly":
		if r.Method != "POST" {
			SendJSONSuccessOrErrorMessage(w, false, "405 Only POST method is allowed", 405)
			return
		}
		if len(argv) != 1 {
			SendJSONErrorMessage(w, "400 Bad Request (required syntax /api/v1/uploadeditugly/ID)")
			return
		}

		id := argv[0]
		fmt.Println("id", id)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			SendJSONErrorMessage(w, "Error while reading POST body")
		}
		fmt.Println(string(body))

		bodyUnm := new([]map[string]string)
		err = json.Unmarshal(body, bodyUnm)
		if err != nil {
			SendJSONErrorMessage(w, "Error while unmarshaling POST body")
			return
		}
		var eug V1UploadEditUglyBodyInput
		eug.ID = id

		/*
			Iterate through the slice of [string]string
			maps and fill in the information appropriately
		*/

		ccclea := func() bool {
			/*
				Check for composition list empty array

				return TRUE if the length is EQUAL TO 0 (would panic)
				return FALSE if the length is NOT equal to 0 (fine)

			*/
			return len(eug.CompositionList) == 0
		}

		var clCurrentRow int = -1
		/*
			clCurrentRow:
			which index of eug.CompositionList to add elements
			add one each time a slice is appended to [][]string.
			should be equal to len(eug.CompositionList) - 1.
		*/

		for _, keyvalue := range *bodyUnm {
			key, kok := keyvalue["name"]
			value, vok := keyvalue["value"]
			if !kok || !vok {
				SendJSONErrorMessage(w, "Error: one item in serialized array does not contain \"name\" and \"value\" keys.")
				return
			}

			switch key {
			case "notes":
				if eug.Notes != "" {
					SendJSONErrorMessage(w, "Error: Notes key specified multiple times.")
					return
				}
				eug.Notes = value

			case "submitname":
				if eug.SubmitName != "" {
					SendJSONErrorMessage(w, "Error: submitname key specified multiple times.")
					return
				}
				eug.SubmitName = value

			case "email":
				if eug.SubmitEmail != "" {
					SendJSONErrorMessage(w, "Error: email key specified multiple times.")
					return
				}
				eug.SubmitEmail = value

			case "classification":
				// first in each line, append a new row.

				eug.CompositionList = append(eug.CompositionList, make([]string, WVEntryRowLength))
				clCurrentRow++

				eug.CompositionList[clCurrentRow][0] = value

			case "number":
				if ccclea() {
					SendJSONErrorMessage(w, "Error: illegal POST syntax: number declared but no classification")
					return
				}

				if eug.CompositionList[clCurrentRow][1] != "" {
					SendJSONErrorMessage(w, "Error: number key specified multiple times for same row.")
					return
				}

				eug.CompositionList[clCurrentRow][1] = value

			case "title":
				if ccclea() {
					SendJSONErrorMessage(w, "Error: illegal POST syntax: title declared but no classification")
					return
				}

				if eug.CompositionList[clCurrentRow][2] != "" {
					SendJSONErrorMessage(w, "Error: title key specified multiple times for same row.")
					return
				}

				eug.CompositionList[clCurrentRow][2] = value

			case "incipit":
				if ccclea() {
					SendJSONErrorMessage(w, "Error: illegal POST syntax: incipit declared but no classification")
					return
				}

				if eug.CompositionList[clCurrentRow][3] != "" {
					SendJSONErrorMessage(w, "Error: title key specified multiple times for same row.")
					return
				}

				eug.CompositionList[clCurrentRow][3] = value

			default:
				/*
					Unknown key passed, fail.
				*/
				SendJSONErrorMessage(w, "Unknown key in serialized array passed.")
				return

			} // end of switch state
		}

		fmt.Printf("%+v\n", eug)

		/*
			Marshal the [][]string eug.CompositionList into
			a bytes.Buffer of csv, which will be compared
			against the already existing csv to make a diff

		*/

		buf := new(bytes.Buffer)

		writer := csv.NewWriter(buf)
		writer.WriteAll(eug.CompositionList)
		if writer.Error() != nil {
			SendJSONErrorMessage(w, "csv write error: "+writer.Error().Error())
			return
		}

		// write this to a temp file

		tmpFileNewSub, err := os.CreateTemp("", "*.csv")
		defer os.Remove(tmpFileNewSub.Name())
		if err != nil {
			SendJSONErrorMessage(w, "csv temp file error: "+err.Error())
			return
		}

		_, err = tmpFileNewSub.Write(buf.Bytes())
		tmpFileNewSub.Close()
		if err != nil {
			SendJSONErrorMessage(w, "csv temp file error: "+err.Error())
			return
		}

		/*
			Read from the original CSV file to make the diff
		*/

		originalSubmissionCSVFilename := "./current/" + eug.ID + ".csv"
		newSubmissionCSVFilename := tmpFileNewSub.Name()
		fmt.Println(originalSubmissionCSVFilename, newSubmissionCSVFilename)

		// compute the diff between these two files

		cmd := APIV1EditUglyGetDiff(originalSubmissionCSVFilename, newSubmissionCSVFilename)
		output, _ := cmd.CombinedOutput()
		fmt.Println(string(output))

		/*
			// note, diff seems to exit code 1 when showing the diff
			if err != nil {
				SendJSONErrorMessage(w, "Error while calculating diff: "+err.Error())
				return
			}
		*/

		/*
			Write all this to a new output struct and marshall that to a file to be read
			later.
		*/

		var out V1UploadEditUglyBodyOutput

		out.SubmitName = eug.SubmitName
		out.SubmitEmail = eug.SubmitEmail
		out.ID = eug.ID
		out.Notes = eug.Notes
		out.Diff = output

		file, err := os.CreateTemp("./submissions", "edit.*.unverified")
		if err != nil {
			fmt.Println(err.Error())
			SendJSONInternalErrorMessage(w, "500 Internal server error: "+err.Error())
			return
		}
		defer file.Close()

		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			SendJSONInternalErrorMessage(w, "500 Internal server error: "+err.Error())
			return
		}
		fmt.Println(string(b))

		_, err = file.Write(b)
		if err != nil {
			SendJSONInternalErrorMessage(w, "500 Internal server error: "+err.Error())
			return
		}

		// write password
		fileNoEnding := strings.TrimRight(file.Name(), ".unverified")
		fileIndex := strings.TrimLeft(fileNoEnding, "./submissions/edit.")
		verifyPassword := randstr.Hex(16)
		fmt.Println("password", (verifyPassword))
		passwordFilename := fileNoEnding + ".password"
		err = os.WriteFile(passwordFilename, []byte(verifyPassword), 0666)
		if err != nil {
			fmt.Println("password file error", err.Error())
			SendJSONInternalErrorMessage(w, "500 Internal server error (password file): "+err.Error())
			return
		}

		/*
			Send SMTP email containing password
		*/

		if err = Apiv1SentSmtpEmailForEditUgly(out.SubmitName, out.SubmitEmail, fileIndex, verifyPassword); err != nil {
			SendJSONInternalErrorMessage(w, "SMTP Mail error: "+err.Error())
			return
		}

		SendJSONSuccessMessage(w, "The edit submission has been processed. Thank you!")
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
	//w.WriteHeader(status)

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

func APIV1EditUglyGetDiff(original, edited string) *exec.Cmd {
	return exec.Command("diff", original, edited)
}
