/*

NOTE: HOW TO ADD ADMIN CONSOLE COMMANDS

1. Add a case to the "switch argv[0]" statement in ExecuteAdminCommand
2. (Recommended) call a new function which will handle this command, and pass argv ([]string) to it. This function should return a string.
3. From the switch case, return a string.


*/
package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	BCRYPTCOST         = 10
	SubmissionsDirPath = "./submissions/"

	ADMINHELPMESSAGE = `Available commands:
ls - list all verified submissions
vsub <id> - View a submission
vedit <id> - View an edit
asub <id> - Accept a submission`
)

type AdminConsoleOutput struct {
	Command string
	Output  string
}

func AdminListCommand(argv []string) string {

	/*
		allow use of -a flag to also list unverified entries
	*/
	showUnverified := false
	for _, arg := range argv {
		if arg == "-a" {
			showUnverified = true
		}
	}

	allFiles, err := os.ReadDir(SubmissionsDirPath)
	if err != nil {
		return err.Error()
	}

	var listOfSubmissions []os.DirEntry

	for _, file := range allFiles {
		/*
			Don't list all files, only list those
			which are of type *.verified (or if
			-a flag, *.verified or *.unverified
		*/
		name := file.Name()

		if strings.HasSuffix(name, ".verified") || (showUnverified && strings.HasSuffix(name, ".unverified")) {
			listOfSubmissions = append(listOfSubmissions, file)
		}
	}

	output := new(bytes.Buffer)

	for _, sub := range listOfSubmissions {
		/*
			Prints out information about each submission line by line to output (*output)
		*/
		fname := sub.Name()
		subpath := SubmissionsDirPath + fname
		subContents, err := os.ReadFile(subpath)
		if err != nil {
			fmt.Fprintf(output, "%s - error: %s\n", subpath, err.Error())
			continue
		}

		/*
			Unmarshal the contents of the file to show
			the submitters name and email in the listing
		*/
		cstruct := new(V1UploadUglySanitizedInput)

		err = json.Unmarshal(subContents, cstruct)
		if err != nil {
			fmt.Fprintf(output, "%s - error: %s\n", subpath, err.Error())
			continue
		}

		submitterEmail := cstruct.SubmitEmail
		submitterName := cstruct.SubmitName
		composerLast := cstruct.ComposerLast

		fmt.Fprintf(output, "%s - ", submitterEmail)
		fmt.Fprintf(output, "%s: ", submitterName)
		fmt.Fprintf(output, "%s ", fname)
		fmt.Fprintf(output, "(%s)", composerLast)
		fmt.Fprintf(output, "\n")

	}
	return output.String()
}

func AdminGetSubmissionFromSnippet(id string) (*os.DirEntry, string) {
	/*
		NOTE: user only has to enter a small
		segment of the id which will lead to a unique
		submission. If the snipped provided is part of
		two submissions IDs, then those should be
		printed, not the submisions themselves
	*/
	allSubmissionFiles, err := os.ReadDir(SubmissionsDirPath)
	if err != nil {
		return nil, "./submissions/ error " + err.Error()
	}

	/*
		validSubmissions will be all the files in
		./submissions/ which will pass the REGEX
		check for the submitted id snippet.

		Hopefully this will be length one, if
		it is length zero or > 1, then we have
		a problem.
	*/
	var validSubmissions []os.DirEntry

	for _, file := range allSubmissionFiles {
		name := file.Name()
		if strings.HasPrefix(name, "submission.") &&
			(strings.HasSuffix(name, ".verified") ||
				strings.HasSuffix(name, ".unverified")) &&
			strings.Contains(name, id) {
			// This is a verified or unverified submission
			// and the id snippet matches
			validSubmissions = append(validSubmissions, file)
		}
	}

	/* Check if len(validSubmissions) is not equal to 1 */
	if len(validSubmissions) == 0 {
		return nil, "ERROR: no submissions found.\n\nvsub <id>"
	} else if len(validSubmissions) > 1 {
		buf := new(bytes.Buffer)
		fmt.Fprintln(buf, "ERROR: the id snippet matches multiple submissions.")
		fmt.Fprintln(buf)
		for _, f := range validSubmissions {
			fmt.Fprintln(buf, f.Name())
		}
		return nil, buf.String()

	}

	/*
		The snippet correctly matches exactly
		one submission. Process this and show.
	*/
	return &validSubmissions[0], "" // type fs.DirEntry (os)
}

func AdminViewSubmission(argv []string) string {
	if len(argv) < 2 {
		return "vsub <id>"
	}

	id := argv[1]

	buf := new(bytes.Buffer)

	var substr V1UploadUglySanitizedInput

	submissionp, errorMessage := AdminGetSubmissionFromSnippet(id)
	if errorMessage != "" {
		return errorMessage
	}
	submission := *submissionp

	subcontents, err := os.ReadFile(SubmissionsDirPath + submission.Name())
	if err != nil {
		fmt.Fprintln(buf, "read file error:", err.Error())
		return buf.String()
	}

	err = json.Unmarshal(subcontents, &substr)
	if err != nil {
		fmt.Fprintln(buf, "read file error:", err.Error())
		return buf.String()
	}

	fmt.Fprintln(buf, "    Submission:", submission.Name())
	fmt.Fprintln(buf, "  Submitted by:", substr.SubmitName)
	fmt.Fprintln(buf, " Email address:", substr.SubmitEmail)
	fmt.Fprintln(buf, " Composer Last:", substr.ComposerLast)
	fmt.Fprintln(buf, "Composer First:", substr.ComposerFirst)
	fmt.Fprintln(buf, "Composer Birth:", substr.ComposerBirth)
	fmt.Fprintln(buf, "Composer Death:", substr.ComposerDeath)
	fmt.Fprintln(buf, "Notes:")
	fmt.Fprintln(buf, substr.Notes)
	fmt.Fprintln(buf, "--------------------------------")

	for _, c := range substr.CompositionList {
		//c is type WVEntry
		fmt.Fprintf(buf, "%s,\t", c.Classifier)
		fmt.Fprintf(buf, "%d,\t", c.Number)
		fmt.Fprintf(buf, "%s,\t", c.Extra)
		fmt.Fprintf(buf, "<t>%s</t>,\t", c.Title)
		fmt.Fprintf(buf, "<i>%s</i>,\t", c.Incipit)

		fmt.Fprintln(buf)

	}

	return buf.String()
}

func AdminAcceptSubmission(argv []string) string {
	if len(argv) < 2 {
		return "asub <id>"
	}

	id := argv[1]
	submissionp, errorMessage := AdminGetSubmissionFromSnippet(id)
	if errorMessage != "" {
		return errorMessage
	}

	submission := *submissionp
	if len(argv) < 3 || argv[2] != "confirm" {
		return "About to accept " + submission.Name() + "\nAre you sure you want to do this? Type asub <id> confirm"
	}

	/*
		Continue, and add the submission.
	*/

	subcontents, err := os.ReadFile(SubmissionsDirPath + submission.Name())
	if err != nil {
		return "read file error " + err.Error()
	}

	var substr V1UploadUglySanitizedInput
	err = json.Unmarshal(subcontents, &substr)
	if err != nil {
		return "json unmarshal error " + err.Error()
	}

	var entry CurrentSingle

	/*
		Convert the items in substr to entry
		Remember to not store the email at all
		in the final form
	*/

	entry.ComposerFirst = substr.ComposerFirst
	entry.ComposerLast = substr.ComposerLast
	entry.ComposerBirth = substr.ComposerBirth
	entry.ComposerDeath = substr.ComposerDeath
	entry.Lock = ""

	var note Note
	note.Message = substr.Notes
	note.Author = substr.SubmitName
	note.DateSTR = GetCurrentDateStr()

	entry.Notes = []Note{note}

	/*
		Get file to upload to
	*/

	newJsonFile, err := os.CreateTemp("./current/", "*.json")
	if err != nil {
		return ".json file create temp error: " + err.Error()
	}
	defer newJsonFile.Close()

	entryJson, err := json.MarshalIndent(entry, "", " ")
	if err != nil {
		return ".json marshall: " + err.Error()
	}

	_, err = newJsonFile.Write(entryJson)
	if err != nil {
		return "json write error: " + err.Error()
	}

	// for now, write notes to .notes csv file also.

	notesCsvFileName := strings.TrimRight(newJsonFile.Name(), ".json") + ".notes"
	// note that this includes the directory, is a full relative path

	notesCsvFile, err := os.Create(notesCsvFileName)
	if err != nil {
		return "notes csv file error: " + err.Error()
	}
	defer notesCsvFile.Close()

	notesCsvListFull := make([][]string, 1)
	notesCsvListFull[0] = make([]string, 3)
	notesCsvListFull[0][0] = note.Author
	notesCsvListFull[0][1] = note.DateSTR
	notesCsvListFull[0][2] = note.Message

	buf := new(bytes.Buffer)

	w := csv.NewWriter(buf)
	w.WriteAll(notesCsvListFull)
	if w.Error() != nil {
		return "csv notes write error: " + w.Error().Error()
	}

	_, err = notesCsvFile.Write(buf.Bytes())
	if err != nil {
		return "notes csv file write error: " + err.Error()
	}

	/*
		Write wv entries to .csv file
	*/

	WVList := substr.CompositionList

	WVListCSV := make([][]string, len(WVList))

	for i, entry := range WVList {
		WVListCSV[i] = make([]string, 5)
		WVListCSV[i][0] = entry.Classifier
		WVListCSV[i][1] = strconv.Itoa(entry.Number)
		WVListCSV[i][2] = entry.Extra
		WVListCSV[i][3] = entry.Title
		WVListCSV[i][4] = entry.Incipit
	}

	buf2 := new(bytes.Buffer)

	w2 := csv.NewWriter(buf2)
	w2.WriteAll(WVListCSV)
	if w2.Error() != nil {
		return "csv notes write error: " + w2.Error().Error()
	}

	WVCsvFileName := strings.TrimRight(newJsonFile.Name(), ".json") + ".csv"
	// note that this includes the directory, is a full relative path

	WVCsvFile, err := os.Create(WVCsvFileName)
	if err != nil {
		return "wv entry csv file error: " + err.Error()
	}
	defer WVCsvFile.Close()

	_, err = WVCsvFile.Write(buf2.Bytes())
	if err != nil {
		return "wv entry csv file write error: " + err.Error()
	}

	/*
		Everything else having gone smoothly, change from .verified to .accepted
		(do this via remarshalling)
	*/
	substr.SubmitEmail = ""

	oldsubmissionfile := SubmissionsDirPath + submission.Name()
	newsubmissionfile := oldsubmissionfile + ".accepted"
	newSubmissionContents, err := json.Marshal(substr)
	if err != nil {
		return "submission file rename json error: " + err.Error() + "\nBut the submission acception was successful"
	}
	newsubmissionfilefile, err := os.Create(newsubmissionfile)
	if err != nil {
		return "submission file create error: " + err.Error() + "\nBut the submission acception was successful"
	}
	defer newsubmissionfilefile.Close()

	_, err = newsubmissionfilefile.Write(newSubmissionContents)
	if err != nil {
		return "submission file move write to file error: " + err.Error() + "\nBut the submission acception was successful"
	}

	os.Remove(oldsubmissionfile)
	stringssplit := strings.Split(oldsubmissionfile, ".")
	os.Remove(strings.Join(stringssplit[:len(stringssplit)-1], ".") + ".password")

	return "Acceptfully accepted."
}

func ExecuteAdminCommand(command string) string {
	argvRaw := strings.Split(command, " ")

	/*
		Sanitize argv
	*/
	var argv []string

	for _, arg := range argvRaw {
		if arg != "" {
			argv = append(argv, arg)
		}
	}

	/*
		Protect against an empty command
		which panics
	*/
	if len(argv) == 0 {
		argv = []string{"help"}
	}

	switch argv[0] {
	case "ls":
		return AdminListCommand(argv)
	case "vsub":
		return AdminViewSubmission(argv)
	case "vs":
		return AdminViewSubmission(argv) // alias to same command
	case "vedit":
		return "vedit"
	case "asub":
		return AdminAcceptSubmission(argv)
	case "help":
		fallthrough
	default:
		return ADMINHELPMESSAGE
	}
}

func FailAdminAuth(w http.ResponseWriter) {
	/*
		Authentication error. Reply with WWW-Authenticate
		header and 401 error.
	*/
	w.Header().Set(`WWW-Authenticate`, `Basic realm="admin"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func AdminConsole(w http.ResponseWriter, r *http.Request) {
	if !AdminConsoleCheckAuth(w, r) {
		return
	}

	/*
		Handle a successfull authentication here
	*/
	htmlTemplate, err := template.ParseFiles("template/admin.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var out AdminConsoleOutput
	commandr := r.URL.Query()["command"]
	if len(commandr) == 1 {

		out.Command = commandr[0]
		out.Output = ExecuteAdminCommand(out.Command)
		fmt.Println(out.Command)
	}

	htmlTemplate.Execute(w, out)
}

func AdminConsoleCheckAuth(w http.ResponseWriter, r *http.Request) bool {
	u, p, ok := r.BasicAuth()

	if !ok {
		FailAdminAuth(w)
		return false
	}

	// cycle through all admins and passwords and see if it matches.

	passed := false
	allAdmins := FullConfig.Admins
	for _, adm := range allAdmins {
		username := adm.Username
		hash := adm.Hash

		if username != u {
			continue
		}

		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) == nil {
			passed = true
			break
		} else {
			continue
		}
	}
	if !passed {
		FailAdminAuth(w)
		return false
	}

	return true
}

func AdminConsolePlaintextHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Admin console is forbidden over plaintext.", 403)
}

func MakePasswordHashCommand(password string) {
	/*
		Function is run by calling ./wvlist password
	*/
	var passwd1 []byte
	var passwd2 []byte

	if len(password) != 0 {
		passwd1 = []byte(password)
		passwd2 = []byte(password)
	} else {
		fmt.Printf("Password: ")
		passwd1, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("\nRe-enter: ")
		passwd2, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Println()
		if len(passwd1) == 0 || len(passwd2) == 0 {
			fmt.Println("ERROR: password may not be empty.")
			MakePasswordHashCommand("") // re run
		}
	}

	if string(passwd1) != string(passwd2) {
		fmt.Println("ERROR: you did not enter the same password.")
		MakePasswordHashCommand("") // re run
	}

	hash, err := bcrypt.GenerateFromPassword(passwd1, BCRYPTCOST)
	if err != nil {
		fmt.Println("BCRYPT ERROR:", err.Error())
		return
	}
	if len(password) == 0 {
		fmt.Println("\n\nhash: (enter the following has in the \"password\" field of the admin table in config.json)")
		fmt.Println()
	}
	fmt.Println(string(hash))
	if len(password) == 0 {
		fmt.Println()
	}
}

func GetCurrentDateStr() string {
	now := time.Now()
	return now.Format("Jan 2, 2006")
}
