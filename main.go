package main

import (
	"flag"
	"fmt"
)

func main() {
	//Full Params with falue from flags

	Params = new(ParamsStr)

	flag.UintVar(&Params.PlaintextPort, "p", 6060, "Plaintext port, set to 0 to disable.")
	flag.UintVar(&Params.TLSPort, "t", 0, "TLS port, set to 0 to disable.")
	flag.BoolVar(&Params.DebugModeTLS, "d", false, "Debug mode: listens to TLS port over plaintext. Should not use in prod.")
	flag.StringVar(&Params.FullCert, "k", "", "Key, path to fullchain.pem")
	flag.StringVar(&Params.PrivCert, "x", "", "Secret, path to privkey.pem")
	flag.StringVar(&Params.ConfigPath, "c", "./config.json", "Config, path to config.json")

	flag.Parse()

	//Load config
	FullConfig = new(ConfigStr)
	if err := RehashConfig(); err != nil {
		panic(err)
	}
	fmt.Println(FullConfig)
}
