// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.



package main

import (
    
    "fmt"   
    "encoding/hex"
    "encoding/json"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
)

var _ = fct.Address{}

func  handleFactoidGenerateAddress(ctx *web.Context, keymr string) {

    type faddress struct {
        Address string
    }
    
    adr, err := factoidState.GetWallet().GenerateAddress([]byte(keymr),1,1)
    if err != nil {
        fmt.Println("Error: ",err)
        ctx.WriteHeader(httpBad)
        return
    }
    
    a := new(faddress)
    
    adrstr := hex.EncodeToString(adr.Bytes())
    
    a.Address = adrstr
    
    if p, err := json.Marshal(a); err != nil {
        ctx.WriteHeader(httpBad)
        return
    } else {
        ctx.Write(p)
    }  
    
}
