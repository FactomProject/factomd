// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
    "time"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/state/stateinit"
)

var _ = fct.Address{}

const (
    httpOK  = 200
    httpBad = 400
)

var (
    ipaddress        = "localhost:"
    portNumber       = "8089"
    applicationName  = "Factom/fctwallet"
    dataStorePath    = "/tmp/fctwallet.dat"
    refreshInSeconds = 60
    
    ipaddressFD      = "localhost:"
    portNumberFD     = "8088"
)

var factoidState = stateinit.NewFactoidState("/tmp/factoid_wallet_bolt.db")

var server = web.NewServer()

 
func Start() {
    
    server.Get("/v1/factoid-balance/([^/]+)", handleFactoidBalance)
    server.Get("/v1/factoid-generate-address/([^/]+)", handleFactoidGenerateAddress)
    
    go server.Run(ipaddress +portNumber)
}   
 
func main() {
    Start()
    for { 
        time.Sleep(time.Second)
    }    
}
