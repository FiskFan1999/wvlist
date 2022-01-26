package main

type ParamsStr struct {
	PlaintextPort uint
	TLSPort       uint
	DebugModeTLS  bool
	FullCert      string
	PrivCert      string
	ConfigPath    string
}

var Params *ParamsStr
