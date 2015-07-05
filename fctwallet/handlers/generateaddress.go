// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.



package handlers

import (
    
    "fmt"   
    "encoding/json"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
)

var _ = fct.Address{}

func  HandleFactoidGenerateAddress(ctx *web.Context, name string) {

    type faddress struct {
        Address string
    }
    
    adr, err := factoidState.GetWallet().GenerateFctAddress([]byte(name),1,1)
    if err != nil {
        str := fmt.Sprintln("Error: %s",err)
        reportResults(ctx,str,false)
        return
    }
    
    a := new(faddress)
    
    adrstr := fct.ConvertFctAddressToUserStr(adr)
    a.Address = adrstr
    if p, err := json.Marshal(a); err != nil {
        reportResults(ctx,"Failed to unmarshal the response from factomd",false)
        return
    } else {
        fmt.Println("\n",p,"\n")
        ctx.Write(p)
    }  
    
}

func  HandleFactoidGenerateECAddress(ctx *web.Context, name string) {
    
    type faddress struct {
        Address string
    }
    
    adr, err := factoidState.GetWallet().GenerateECAddress([]byte(name))
    if err != nil {
        str := fmt.Sprintf("Error: %s",err)
        reportResults(ctx, str, false)
        return
    }
    
    a := new(faddress)
    
    adrstr := fct.ConvertECAddressToUserStr(adr)
    a.Address = adrstr
    if p, err := json.Marshal(a); err != nil {
        reportResults(ctx,"Failed to unmarshal the response from factomd",false)
        return
    } else {
        ctx.Write(p)
    }  
    
}
