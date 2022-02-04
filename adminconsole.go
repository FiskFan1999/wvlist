package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
	"html/template"
	"net/http"
	"os"
	"strings"
)

const (
	BCRYPTCOST = 10

	ADMINHELPMESSAGE = `Available commands:
ls - list all verified submissions
view <id> - View a submission
accept <id> - Accept a submission`
)

type AdminConsoleOutput struct {
	Command string
	Output  string
}

func ExecuteAdminCommand(command string) string {
	argv := strings.Split(command, " ")

	switch argv[0] {
	case "ls":
		return "list"
	case "view":
		return "view"
	case "accept":
		return "accept"
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

func MakePasswordHashCommand() {
	/*
		Function is run by calling ./wvlist password
	*/
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
		MakePasswordHashCommand() // re run
	}

	if string(passwd1) != string(passwd2) {
		fmt.Println("ERROR: you did not enter the same password.")
		MakePasswordHashCommand() // re run
	}

	hash, err := bcrypt.GenerateFromPassword(passwd1, BCRYPTCOST)
	if err != nil {
		fmt.Println("BCRYPT ERROR:", err.Error())
		return
	}
	fmt.Println("\n\nhash: (enter the following has in the \"password\" field of the admin table in config.json)")
	fmt.Println()
	fmt.Println(string(hash))
	fmt.Println()
}
