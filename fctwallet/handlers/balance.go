// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
    "fmt"
    "strconv"
    "net/http"
    "io/ioutil"    
    "encoding/hex"
    "encoding/json"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/wallet"
)


func FctBalance(adr string) (int64, error) {
    
    if !fct.ValidateFUserStr(adr) {
        msg, ok := ValidateKey(adr)
        if !ok {
            return 0, fmt.Errorf("%s",msg)
        }
        
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
        return 0, err
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return 0, err
    }
    resp.Body.Close()
    
    type x struct { 
        Response string
        Success  bool
    }
    b := new(x)
    if err := json.Unmarshal(body, b); err != nil {
        return 0, err
    }
    
    if !b.Success {
        return 0, fmt.Errorf("%s",b.Response)
    }
    
    v, err := strconv.ParseInt(b.Response,10,64)
    if err != nil {
        return 0,err
    }
    
    return v, nil
    
}

func ECBalance(adr string) (int64, error) {
    
    if !fct.ValidateECUserStr(adr) {
        msg, ok := ValidateKey(adr)
        if !ok {
            return 0, fmt.Errorf("%s",msg)
        }
        
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
        return 0, err
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return 0, err
    }
    resp.Body.Close()
    
    type x struct { 
        Response string
        Success  bool
    }
    b := new(x)
    if err := json.Unmarshal(body, b); err != nil {
        return 0, err
    }
    
    if !b.Success {
        return 0, fmt.Errorf("%s",b.Response)
    }
    
    v, err := strconv.ParseInt(b.Response,10,64)
    if err != nil {
        return 0,err
    }
    
    return v, nil
}

func  HandleEntryCreditBalance(ctx *web.Context, adr string) {    

    v,err := ECBalance( adr)
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return
    }
    str := fmt.Sprintf("%d",v)
    reportResults(ctx,str,true)
    
}


func  HandleFactoidBalance(ctx *web.Context, adr string) {
    
    v,err := FctBalance(adr)
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return
    }
    
    str := fmt.Sprintf("%d",v)
    reportResults(ctx,str,true)
    
}
