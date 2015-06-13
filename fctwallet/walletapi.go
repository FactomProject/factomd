// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.



package main

import (
    "os"
    "strconv"    
    "encoding/hex"
    "encoding/json"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
)

var _ = fct.Address{}

const (
    httpOK  = 200
    httpBad = 400
)

var (
    portNumber       = 8089
    applicationName  = "Factom/fctwallet"
    dataStorePath    = "/tmp/fctwallet.dat"
    refreshInSeconds = 60
)


var server = web.NewServer()

 
func Start() {
    
    server.Get("/v1/factoid-balance/([^/]+)", handleFactoidBalance)
    server.Post("/v1/stop/?", handleStop)
    
    go server.Run("localhost:" + strconv.Itoa(portNumber))
}

func Stop() {
    server.Close()
}

func handleStop(ctx *web.Context, keymr string) {
    Stop()
    os.Exit(0)
}

func  handleFactoidBalance(ctx *web.Context, keymr string) {
    type factoidbal struct {
        Balance uint64
    }
    
    b := new(factoidbal)
    
    adr, err := hex.DecodeString(keymr)
    if err != nil { 
        fct.Prtln("Error: ",err)
    }
    if len(adr) != fct.ADDRESS_LENGTH {
        fct.Prtln("Error: Bad Address: ", keymr)
    }
    b.Balance = fs.GetBalance(fct.NewAddress(adr))
    
    
    if p, err := json.Marshal(b); err != nil {
        ctx.WriteHeader(httpBad)
        return
    } else {
        ctx.Write(p)
    }
    
    ctx.WriteHeader(httpOK)
}

func main() {
    Start()
    web.Run("0.0.0.0:9999")
}
