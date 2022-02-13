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
	"os/exec"
	"strings"
	"time"
)

const (
	BCRYPTCOST         = 10
	SubmissionsDirPath = "./submissions/"
	RemovedEmailStr    = ""

	ADMINHELPMESSAGE = `Available commands:
ls - list all verified submissions
vsub <id> - View a submission
vedit <id> - View an edit
asub <id> - Accept a submission
rsub <id> - Reject a submission
testemail <address> - Send an email to test SMTP settings
help - list all available commands`
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

	var allVerified uint = 0
	var allUnverified uint = 0

	var listOfSubmissions []os.DirEntry

	for _, file := range allFiles {
		/*
			Don't list all files, only list those
			which are of type *.verified (or if
			-a flag, *.verified or *.unverified
		*/
		name := file.Name()

		isVerified := strings.HasSuffix(name, ".verified")
		isUnverified := strings.HasSuffix(name, ".unverified")

		if isVerified || (showUnverified && isUnverified) {
			listOfSubmissions = append(listOfSubmissions, file)
		}
		if isVerified {
			allVerified++
		}
		if isUnverified {
			allUnverified++
		}
	}

	output := new(bytes.Buffer)

	fmt.Fprintln(output, allVerified+allUnverified, "total submissions")
	fmt.Fprintln(output, allVerified, "verified")
	fmt.Fprintln(output, allUnverified, "unverified")

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
	a, b := AdminGetAnyFileTypeFromSnippet(id, "submission")
	return a, b
}

func AdminGetEditFromSnippet(id string) (*os.DirEntry, string) {
	a, b := AdminGetAnyFileTypeFromSnippet(id, "edit")
	return a, b
}

func AdminGetAnyFileTypeFromSnippet(id string, filetype string) (*os.DirEntry, string) {
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
		if strings.HasPrefix(name, filetype+".") &&
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
		fmt.Fprintf(buf, "%s,\t", c.Number)
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
		WVListCSV[i] = make([]string, WVEntryRowLength)
		WVListCSV[i][0] = entry.Classifier
		WVListCSV[i][1] = entry.Number
		WVListCSV[i][2] = entry.Title
		WVListCSV[i][3] = entry.Incipit
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
		return AdminViewEdit(argv)
	case "ve":
		return AdminViewEdit(argv)
	case "asub":
		return AdminAcceptSubmission(argv)
	case "aedit":
		return AdminAcceptEdit(argv)
	case "rsub":
		return AdminRejectSubmission(argv)
	case "testemail":
		return AdminTestEmail(argv)
	case "help":
		fallthrough
	default:
		return ADMINHELPMESSAGE
	}
}

func GetAcceptEditPatchCommand(file, patch string) *exec.Cmd {
	/*
		file and patch are filenames
	*/
	return exec.Command("patch", "-V", "t", "-b", "-f", file, patch)
}

func AdminAcceptEdit(argv []string) string {
	if len(argv) < 2 {
		return "asub <id>"
	}

	id := argv[1]
	submissionp, errorMessage := AdminGetEditFromSnippet(id)
	if errorMessage != "" {
		return errorMessage
	}

	submission := *submissionp
	if len(argv) < 3 || argv[2] != "confirm" {
		return "About to accept " + submission.Name() + "\nAre you sure you want to do this? Type asub <id> confirm"
	}

	fullFileName := SubmissionsDirPath + submission.Name()

	contents, err := os.ReadFile(fullFileName)
	if err != nil {
		return "read file error: " + err.Error()
	}

	/*
		Unmarshal contents into the struct
	*/

	var sub V1UploadEditUglyBodyOutput

	err = json.Unmarshal(contents, &sub)
	if err != nil {
		return "submission json parsing error: " + err.Error()
	}

	/*
		Get the filename which is being edited.
	*/

	csvid := sub.ID
	csvFileName := "./current/" + csvid + ".csv"

	if _, err := os.ReadFile(csvFileName); err != nil {
		return "file to be edited read error: " + err.Error()
	}

	/*
		Write the patch to a temp file, to be passed to the
	*/

	var thePatch []byte = sub.Diff

	PatchTempFile, err := os.CreateTemp("", "*")
	if err != nil {
		return "patch temp file error: " + err.Error()
	}
	defer os.Remove(PatchTempFile.Name())

	if _, err = PatchTempFile.Write(thePatch); err != nil {
		return "write patch to temp file error: " + err.Error()
	}

	PatchTempFile.Close()

	cmd := GetAcceptEditPatchCommand(csvFileName, PatchTempFile.Name())

	output, err := cmd.CombinedOutput()

	if err != nil {
		return string(output) + "\n" + err.Error()
	}

	/*
		Change the filename of the edit
		submission to accepted.
	*/

	sub.SubmitEmail = "" // erase email

	var remarshal []byte
	remarshal, err = json.MarshalIndent(sub, "", "  ")
	if err != nil {
		return "remarshal error: " + err.Error()

	}

	editFilenameSplit := strings.Split(fullFileName, ".")
	newEditFileName := strings.Join(editFilenameSplit[0:len(editFilenameSplit)-1], ".") + ".accepted"

	newFile, err := os.Create(newEditFileName)
	if err != nil {
		return "new edit file create error: " + err.Error()
	}

	if _, err = newFile.Write(remarshal); err != nil {
		return "new edit file write error: " + err.Error()
	}

	newFile.Close()
	if err = os.Remove(fullFileName); err != nil {
		return "error removing previous edit submission file: " + err.Error()
	}

	return string(output)

}

func AdminViewEdit(argv []string) string {

	if len(argv) != 2 {
		return "vedit <id>"
	}

	id := argv[1]

	submissionp, errorMessage := AdminGetEditFromSnippet(id)
	if errorMessage != "" {
		return errorMessage
	}

	submission := *submissionp

	fullFileName := SubmissionsDirPath + submission.Name()

	contents, err := os.ReadFile(fullFileName)
	if err != nil {
		return "os.ReadFile error " + err.Error()
	}

	/*
		Unmarshal contents into the struct
	*/

	var sub V1UploadEditUglyBodyOutput

	err = json.Unmarshal(contents, &sub)

	/*
		Get some information about what the submission
		is editing.
	*/

	composer, err := ParseCurrentSingle(sub.ID)
	if err != nil {
		return "error: this submission has an ID that does not link back to any existing submission."
	}

	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "Submission author: %s\n", sub.SubmitName)
	fmt.Fprintf(buf, "Submission email: %s\n", sub.SubmitEmail)

	fmt.Fprintf(buf, "\nComposer: %s %s\n", composer.ComposerFirst, composer.ComposerLast)

	fmt.Fprintf(buf, "\n%s\n", sub.Diff)

	return buf.String()
}

func AdminRejectSubmission(argv []string) string {
	id := argv[1]
	submissionp, errorMessage := AdminGetSubmissionFromSnippet(id)
	if errorMessage != "" {
		return errorMessage
	}

	submission := *submissionp

	/*
		Rewrite .verified or .unverified file to .rejected (remove email)
	*/

	filename := SubmissionsDirPath + submission.Name()
	origContents, err := os.ReadFile(filename)
	if err != nil {
		return err.Error()
	}

	input := new(V1UploadUglySanitizedInput)
	if err = json.Unmarshal(origContents, input); err != nil {
		return err.Error()
	}

	// remove email
	input.SubmitEmail = RemovedEmailStr

	output, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return err.Error()
	}

	OFNS := strings.Split(filename, ".")
	outputFileName := strings.Join(OFNS[0:len(OFNS)-1], ".") + ".rejected"

	outFileStr, err := os.Create(outputFileName)
	if err != nil {
		return err.Error()
	}

	if _, err = outFileStr.Write(output); err != nil {
		return err.Error()
	}

	/*
		If new file is successfully writte, remove original file
	*/

	if err = os.Remove(filename); err != nil {
		return err.Error()
	}
	// also remove password file

	passwordFileName := strings.Join(OFNS[0:len(OFNS)-1], ".") + ".password"
	if err = os.Remove(passwordFileName); err != nil {
		return err.Error()
	}

	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "Submission %s successfully rejected.", filename)
	return buf.String()

}

func AdminTestEmail(argv []string) string {
	if len(argv) != 2 {
		return "testemail <address> - send an email to test SMTP settings"
	}

	to := argv[1]
	return SendTestSMTPEmail(to).String()
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
