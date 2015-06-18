// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.



package main

import (
    
    "fmt"   
    "encoding/json"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
)

var _ = fct.Address{}

func  handleFactoidGenerateAddress(ctx *web.Context, name string) {
    
    type faddress struct {
        Address string
    }
    
    adr, err := factoidState.GetWallet().GenerateAddress([]byte(name),1,1)
    if err != nil {
        fmt.Println("Error: ",err)
        reportResults(ctx,false)
        return
    }
    
    a := new(faddress)
    
    adrstr := ConvertFAddressToUserStr(adr)
    a.Address = adrstr
    if p, err := json.Marshal(a); err != nil {
        reportResults(ctx,false)
        return
    } else {
        ctx.Write(p)
    }  
    
}
