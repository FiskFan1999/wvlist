package main

type ParamsStr struct {
	PlaintextPort uint64
	TLSPort       uint64
	DebugModeTLS  bool
	FullCert      string
	PrivCert      string
	ConfigPath    string
}

var Params *ParamsStr
