// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
    "time"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/fctwallet/handlers"   
)

var _ = fct.Address{}

var server = web.NewServer()

func Start() {
    // Balance
    // localhost:8089/v1/factoid-balance/<name or address>
    // Returns the balance of factoids at that address, or the address tied to
    // the given name.
    server.Get("/v1/factoid-balance/([^/]+)", handlers.HandleFactoidBalance)
    
    // Balance
    // localhost:8089/v1/factoid-balance/<name or address>
    // Returns the balance of entry credits at that address, or the address tied to
    // the given name.
    server.Get("/v1/entry-credit-balance/([^/]+)", handlers.HandleEntryCreditBalance)
    
    // Generate Address
    // localhost:8089/v1/factoid-generate-address/<name>
    // Generate an address, and tie it to the given name within the wallet. You
    // can use the name for the address in this API
    server.Get("/v1/factoid-generate-address/([^/]+)", handlers.HandleFactoidGenerateAddress)

    // Generate Entry Credit Address
    // localhost:8089/v1/factoid-generate-ec-address/<name>
    // Generate an address, and tie it to the given name within the wallet. You
    // can use the name for the address in this API
    server.Get("/v1/factoid-generate-ec-address/([^/]+)", handlers.HandleFactoidGenerateECAddress)

    // New Transaction
    // localhost:8089/v1/factoid-new-transaction/<key>
    // Use the key in subsequent calls to add inputs, outputs, ecoutputs, and to
    // sign and submit the transaction. Returns Success == true if all is well.
    // Multiple transactions can be in process.  Only one transaction per key.
    // Once the transaction has been submitted or deleted, the key can be
    // reused.
    server.Post("/v1/factoid-new-transaction/([^/]+)", handlers.HandleFactoidNewTransaction)

    // Delete Transaction
    // localhost:8089/v1/factoid-new-transaction/<key>
    // Remove the key of a transaction in flight.  If it doesn't exist, then 
    // nobody cares.
    server.Post("/v1/factoid-delete-transaction/([^/]+)", handlers.HandleFactoidDeleteTransaction)
    
    // Add Input
    // localhost:8089/v1/factoid-add-input/?key=<key>&name=<name or address>
    // Add the fee for this transaction to the input specified by the name or address.
    // If the name or address is not an input to this transaction, then an error 
    // is posted.
    server.Post("/v1/factoid-add-fee/(.*)", handlers.HandleFactoidAddFee)
    
    // Add Input
    // localhost:8089/v1/factoid-add-input/?key=<key>&name=<name or address>&amount=<amount>
    // Add an input to a transaction in process.  Start with new-transaction.
    server.Post("/v1/factoid-add-input/(.*)", handlers.HandleFactoidAddInput)
    
    // Add Output
    // localhost:8089/v1/factoid-add-output/?key=<key>&name=<name or address>&amount=<amount>
    // Add an output to a transaction in process.  Start with new-transaction.
    server.Post("/v1/factoid-add-output/(.*)", handlers.HandleFactoidAddOutput)
    
    // Add Entry Credit Output
    // localhost:8089/v1/factoid-add-ecoutput/?key=<key>&name=<name or address>&amount=<amount>
    // Add an ecoutput to a transaction in process.  Start with new-transaction.
    server.Post("/v1/factoid-add-ecoutput/(.*)", handlers.HandleFactoidAddECOutput)
    
    // Sign Transaction
    // localhost:8089/v1/factoid-sign-transaction/<key>
    // If the transaction validates structure wise and all signatures can be
    // applied, then all inputs are signed, and returns success = true
    // Otherwise returns false. Note that this doesn't check that the inputs
    // can cover the transaction.  Use validate to do that.
    server.Post("/v1/factoid-sign-transaction/(.*)", handlers.HandleFactoidSignTransaction)
    
    // Setup
    // localhost:8089/v1/factoid-setup/<key>
    // hashes the given data to create a new seed from which to generate addresses.
    // The point is to create unique and secure addresses for this user.
    server.Post("/v1/factoid-setup/(.*)", handlers.HandleFactoidSetup)
    
	// Commit Chain
	// localhost:8089/v1/commit-chain/
	// sign a binary Chain Commit with an entry credit key and submit it to the
	// factomd server
	server.Post("/v1/commit-chain/([^/]+)", handlers.HandleCommitChain)

	// Commit Entry
	// localhost:8089/v1/commit-entry/
	// sign a binary Entry Commit with an entry credit key and submit it to the
	// factomd server
	server.Post("/v1/commit-entry/([^/]+)", handlers.HandleCommitEntry)

    // Submit
    // localhost:8089/v1/factoid-submit/
    // Put the key for the transaction in {Transaction string}
    server.Post("/v1/factoid-submit/(.*)", handlers.HandleFactoidSubmit)
    
    // Validate
    // localhost:8089/v1/factoid-validate/<key>
    // Validates amounts and that all required signatures are applied, returns success = true
    // Otherwise returns false.
    server.Get("/v1/factoid-validate/(.*)", handlers.HandleFactoidValidate)
    
    // Get Fee
    // localhost:8089/v1/factoid-get-fee/
    // Get the Transaction fee
    server.Get("/v1/factoid-get-fee/", handlers.HandleGetFee)
    
    // Get Address List
    // localhost:8089/v1/factoid-get-addresses/
    server.Get("/v1/factoid-get-addresses/", handlers.HandleGetAddresses)
    
    // Get transactions
    // localhost:8089/v1/factoid-get-addresses/
    server.Get("/v1/factoid-get-transactions/", handlers.HandleGetTransactions)
    
    
    go server.Run(fmt.Sprintf("%s:%d", handlers.IpAddress, handlers.PortNumber))
}   


func main() {
    Start()
    for { 
        time.Sleep(time.Second)
    }    
}
