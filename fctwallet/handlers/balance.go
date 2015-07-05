// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
    "fmt"
    "net/http"
    "io/ioutil"    
    "encoding/hex"
    "encoding/json"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/wallet"
)


func FctBalance(adr string) int64 {
    
    if !fct.ValidateFUserStr(adr) {
        we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME),[]byte(adr))
    
        if (we != nil){
            we2 := we.(wallet.IWalletEntry)
            addr,_ := we2.GetAddress()
            adr = hex.EncodeToString(addr.Bytes())
        }
    }else{
        baddr := fct.ConvertUserStrToAddress(adr)
        adr = hex.EncodeToString(baddr)
    }
    
    str := fmt.Sprintf("http://%s/v1/factoid-balance/%s", ipaddressFD+portNumberFD, adr)
    resp, err := http.Get(str)
    if err != nil {
        fmt.Println("\n",str)
        return -1
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return -2
    }
    resp.Body.Close()
        
    type Balance struct { Balance int64 }
    b := new(Balance)
    if err := json.Unmarshal(body, b); err != nil {
        return -3
    }
        
    return b.Balance
    
}

func ECBalance(adr string) int64 {
    
    if !fct.ValidateECUserStr(adr) {
        we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME),[]byte(adr))
        
        if (we != nil){
            we2 := we.(wallet.IWalletEntry)
            addr,_ := we2.GetAddress()
            adr = hex.EncodeToString(addr.Bytes())
        }
    }else{
        baddr := fct.ConvertUserStrToAddress(adr)
        adr = hex.EncodeToString(baddr)
    }
    
    str := fmt.Sprintf("http://%s/v1/entry-credit-balance/%s", ipaddressFD+portNumberFD, adr)
    resp, err := http.Get(str)
    if err != nil {
        fmt.Println("-2::",err)
        return -2
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("-3::",err)
        return -3
    }
    resp.Body.Close()
    
    type Balance struct { Balance int64 }
    b := new(Balance)
    if err := json.Unmarshal(body, b); err != nil {
        fmt.Println("-4::",err)
        return -4
    }
    
    return b.Balance
}

func  HandleEntryCreditBalance(ctx *web.Context, adr string) {    

    v := ECBalance(adr)
    if v < 0 {
        reportResults(ctx,"Unknown or bad address",false)
        return
    }
    
    type ecbal struct {
        Balance uint64
    }
    
    b := new(ecbal)
    b.Balance = uint64(v)
    
    if p, err := json.Marshal(b); err != nil {
        reportResults(ctx,"Failed to unmarshal the response from factomd", false)
        return
    } else {
        ctx.Write(p)
    }
    
}


func  HandleFactoidBalance(ctx *web.Context, adr string) {
    
    v := FctBalance(adr)
    if v < 0 {
        reportResults(ctx,"Unknown or bad address: ",false)
        return
    }
    
    type factoidbal struct {
        Balance uint64
    }
    
    b := new(factoidbal)
    b.Balance = uint64(v)
    
    if p, err := json.Marshal(b); err != nil {
        reportResults(ctx,"Failed to unmarshal the response from factomd", false)
        return
    } else {
        ctx.Write(p)
    }
    
}
