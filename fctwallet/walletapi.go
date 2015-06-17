// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/state/stateinit"
	"github.com/hoisie/web"
)

var _ = fct.Address{}

const (
    httpOK  = 200
    httpBad = 400
)

var (
	cfg = fct.ReadConfig().Wallet
    ipAddress        = cfg.Address
    portNumber       = cfg.Port
    applicationName  = "Factom/fctwallet"
    dataStorePath    = cfg.DataFile
    refreshInSeconds = cfg.RefreshInSeconds
    
    ipaddressFD      = "localhost:"
    portNumberFD     = "8088"
)

var factoidState = stateinit.NewFactoidState("/tmp/factoid_wallet_bolt.db")

var server = web.NewServer()

 
func Start() {
    // Balance
    // localhost:8089/v1/factoid-balance/<name or address>
    // Returns the balance of factoids at that address, or the address tied to
    // the given name.
    server.Get("/v1/factoid-balance/([^/]+)", handleFactoidBalance)
    
    // Generate Address
    // localhost:8089/v1/factoid-generate-address/<name>
    // Generate an address, and tie it to the given name within the wallet.  You can
    // use the name for the address in this API
    server.Get("/v1/factoid-generate-address/([^/]+)", handleFactoidGenerateAddress)
    
    // New Transaction
    // localhost:8089/v1/factoid-new-transaction/<key>
    // Use the key in subsequent calls to add inputs, outputs, ecoutputs, and to
    // sign and submit the transaction. Returns Success == true if all is well.
    // Multiple transactions can be in process.  Only one transaction per key.
    // Once the transaction has been submitted or deleted, the key can be reused.
    server.Get("/v1/factoid-new-transaction/([^/]+)", handleFactoidNewTransaction)
    
    // Add Input
    // localhost:8089/v1/factoid-add-input/?key=<key>&name=<name or address>&amount=<amount>
    // Add an input to a transaction in process.  Start with new-transaction.
    server.Get("/v1/factoid-add-input/(.*)", handleFactoidAddInput)
    
    // Add Output
    // localhost:8089/v1/factoid-add-output/?key=<key>&name=<name or address>&amount=<amount>
    // Add an output to a transaction in process.  Start with new-transaction.
    server.Get("/v1/factoid-add-output/(.*)", handleFactoidAddOutput)
    
    // Add Entry Credit Output
    // localhost:8089/v1/factoid-add-ecoutput/?key=<key>&name=<name or address>&amount=<amount>
    // Add an ecoutput to a transaction in process.  Start with new-transaction.
    server.Get("/v1/factoid-add-ecoutput/(.*)", handleFactoidAddECOutput)
    
    // Sign Transaction
    // localhost:8089/v1/factoid-sign-transaction/<key>
    // If the transaction validates structure wise and all signatures can be applied, 
    // then all inputs are signed, and returns success = true
    // Otherwise returns false.
    // Note that this doesn't check that the inputs can cover the transaction.  Use validate
    // to do that.
    server.Get("/v1/factoid-sign-transaction/(.*)", handleFactoidSignTransaction)
    
    // Validate
    // localhost:8089/v1/factoid-validate/<key>
    // Validates amounts and that all required signatures are applied, returns success = true
    // Otherwise returns false.
    server.Get("/v1/factoid-validate/(.*)", handleFactoidValidate)
    
    // Submit
    // localhost:8089/v1/factoid-submit/
    // Put the key for the transaction in {Transaction string}
    server.Post("/v1/factoid-submit/", handleFactoidSubmit)
    
    // Get Fee
    // localhost:8089/v1/factoid-get-fee/
    // Get the Transaction fee
    server.Get("/v1/factoid-get-fee/", handleGetFee)
    
    go server.Run(fmt.Sprintf("%s:%d", ipAddress, portNumber))
}   
 
func main() {
    Start()
    for { 
        time.Sleep(time.Second)
    }    
}
