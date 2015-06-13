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
)


func GetBalance(adr fct.IAddress) int64 {
    
    type Balance struct {
        Balance int64
    }
    
    key := string(hex.EncodeToString(adr.Bytes()))
    
    str := fmt.Sprintf("http://%s/v1/factoid-balance/%s", ipaddressFD+portNumberFD, key)
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
        
    b := new(Balance)
    if err := json.Unmarshal(body, b); err != nil {
        return -3
    }
        
    return b.Balance
    
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
    v := GetBalance(fct.NewAddress(adr))
    if v < 0 {
        fmt.Println("Unknown or bad address: ",v)
        v = 0
    }
    
    b.Balance = uint64(v)
    
    if p, err := json.Marshal(b); err != nil {
        ctx.WriteHeader(httpBad)
        return
    } else {
        ctx.Write(p)
    }
    
}

