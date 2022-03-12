package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var Commit string
var Version string

func main() {

	LilypondFilesToMake = make(chan LilypondFileToMakeStr, 256)
	go LilypondWriteIncipitsFromChannel()

	EmailCoolDown = make(map[string]time.Time)

	if Version == "" {
		Version = "Unreleased"
	}
	//Full Params with falue from flags

	Params = new(ParamsStr)

	flag.Uint64Var(&Params.PlaintextPort, "p", 6060, "Plaintext port, set to 0 to disable.")
	flag.Uint64Var(&Params.TLSPort, "t", 0, "TLS port, set to 0 to disable.")
	flag.BoolVar(&Params.DebugModeTLS, "d", false, "Debug mode: listens to TLS port over plaintext. Should not use in prod.")
	flag.StringVar(&Params.FullCert, "k", "", "Key, path to fullchain.pem")
	flag.StringVar(&Params.PrivCert, "x", "", "Secret, path to privkey.pem")
	flag.StringVar(&Params.ConfigPath, "c", "./config.json", "Config, path to config.json")

	flag.Parse()

	// Load config
	FullConfig = new(ConfigStr)
	if err := RehashConfig(); err != nil {
		panic(err)
	}
	FullConfig.Commit = Commit
	FullConfig.Version = Version

	/*
		Check for required directories
		This function also handles the
		creation of these directories.
	*/
	if err := CheckForNeededDirs(); err != nil {
		fmt.Println("ERROR while checking for required directories:", err.Error())
		os.Exit(1)
	}

	/*
		Check for lilypond
	*/
	LilypondVer, err := CheckForLilypondAtStart()
	if err != nil {
		fmt.Print("---------------\nERROR while initializing: lilypond error:\n")
		fmt.Println(err.Error())
		fmt.Print("The incipits and lilypond sandbox will NOT work properly.\n---------------\n")
	} else {
		fmt.Printf("%s", LilypondVer)
	}

	argv := flag.Args()

	if len(argv) == 0 {
		argv = []string{""}
	}

	switch argv[0] {
	case "sendemail":
		if len(argv) != 2 {
			fmt.Println("./wvlist sendemail <to>")
			return
		}
		to := argv[1]
		fmt.Println("Sending email to " + to)
		buf := SendTestSMTPEmail(to)
		fmt.Println(buf.String())

	case "run":
		/*
			Operate the server and listen as normal
		*/
		//Load config
		/*
			Create seperate plaintext and TLS muxers
			(plaintext mux will disable operator
			console)
		*/
		pmux := http.NewServeMux()
		tmux := http.NewServeMux()

		pmux.HandleFunc("/", HomePage(false))
		tmux.HandleFunc("/", HomePage(true))

		pmux.HandleFunc("/view/", ViewPage)
		tmux.HandleFunc("/view/", ViewPage)

		pmux.HandleFunc("/submit/", SubmitPage)
		tmux.HandleFunc("/submit/", SubmitPage)

		pmux.HandleFunc("/edit/", GetEditPage)
		tmux.HandleFunc("/edit/", GetEditPage)

		pmux.HandleFunc("/lilysand/", LilypondSandbox)
		tmux.HandleFunc("/lilysand/", LilypondSandbox)

		pmux.HandleFunc("/incipit/", GetLilypond)
		tmux.HandleFunc("/incipit/", GetLilypond)

		pmux.HandleFunc("/api/v1/", APIv1Handler)
		tmux.HandleFunc("/api/v1/", APIv1Handler)

		pmux.HandleFunc("/admin/", AdminConsolePlaintextHandler)
		tmux.HandleFunc("/admin/", AdminConsole)

		// run plain and tls listeners concurrently
		wg := new(sync.WaitGroup)
		wg.Add(1)
		if Params.PlaintextPort != 0 {
			wg.Add(1)
			go func() {
				http.ListenAndServe(":"+strconv.FormatUint(Params.PlaintextPort, 10), pmux)
				wg.Done()
			}()
		}
		if Params.TLSPort != 0 {
			wg.Add(1)
			go func() {
				if Params.DebugModeTLS {
					http.ListenAndServe(":"+strconv.FormatUint(Params.TLSPort, 10), tmux)
				} else {
					http.ListenAndServeTLS(":"+strconv.FormatUint(Params.TLSPort, 10), Params.FullCert, Params.PrivCert, tmux)
				}
				wg.Done()
			}()
		}
		wg.Done()
		wg.Wait()
	case "password":
		if len(argv) == 2 {
			MakePasswordHashCommand(argv[1])
		} else {
			MakePasswordHashCommand("")
		}
	default:
		fmt.Println(`./wvlist run
./wvlist sendemail <to>
./wvlist password [password]`)
	}
}
