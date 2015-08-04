// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/state/stateinit"
   
)

const (
    httpOK  = 200
    httpBad = 400
)


var (
    cfg = fct.ReadConfig().Wallet
    IpAddress        = cfg.Address
    PortNumber       = cfg.Port
    applicationName  = "Factom/fctwallet"
    dataStorePath    = cfg.DataFile
    refreshInSeconds = cfg.RefreshInSeconds
    
    ipaddressFD      = "localhost:"
    portNumberFD     = "8088"
    
    databasefile     = "factoid_wallet_bolt.db"
)

var factoidState = stateinit.NewFactoidState(cfg.BoltDBPath + databasefile)