// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.



package handlers

import (
    
    "fmt"   
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
)

var _ = fct.Address{}

func  HandleFactoidGenerateAddress(ctx *web.Context, name string) {

    msg, ok := ValidateKey(name) 
    if !ok {
        reportResults(ctx, msg, false)
        return
    }
    
    adr, err := factoidState.GetWallet().GenerateFctAddress([]byte(name),1,1)
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return
    }
    
    adrstr := fct.ConvertFctAddressToUserStr(adr)
        
    reportResults(ctx,adrstr,true)
      
}

func  HandleFactoidGenerateECAddress(ctx *web.Context, name string) {
    
    msg, ok := ValidateKey(name) 
    if !ok {
        reportResults(ctx, msg, false)
        return
    }
    
    type faddress struct {
        Address string
    }
    
    adr, err := factoidState.GetWallet().GenerateECAddress([]byte(name))
    if err != nil {
        str := fmt.Sprintf("Error: %s",err)
        reportResults(ctx, str, false)
        return
    }
    
    adrstr := fct.ConvertECAddressToUserStr(adr)
    
    reportResults(ctx, adrstr, true)
    
}
