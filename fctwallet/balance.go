// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

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


func GetBalance(adr string) int64 {
    
    if !ValidateFUserStr(adr) {
        we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME_HASH),[]byte(adr))
    
        if (we != nil){
            we2 := we.(wallet.IWalletEntry)
            addr,_ := we2.GetAddress()
            adr = hex.EncodeToString(addr.Bytes())
        }
    }else{
        baddr := ConvertUserStrToAddress(adr)
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

func  handleFactoidBalance(ctx *web.Context, adr string) {
    
    v := GetBalance(adr)
    if v < 0 {
        fmt.Println("Unknown or bad address: ",v)
        reportResults(ctx,false)
        return
        v = 0
    }
    
    type factoidbal struct {
        Balance uint64
    }
    
    b := new(factoidbal)
    b.Balance = uint64(v)
    
    if p, err := json.Marshal(b); err != nil {
        reportResults(ctx,false)
        return
    } else {
        ctx.Write(p)
    }
    
}

