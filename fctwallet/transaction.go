// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
    "fmt"
    "bytes"
    "strconv"
    "io/ioutil"
    "net/http"
    "encoding/hex"
    "encoding/json"
    "github.com/hoisie/web"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/wallet"
)

/******************************************
 * Helper Functions
 ******************************************/

// True is sccuess! False is failure
func reportResults(ctx *web.Context, v bool) {
    b := struct{Success bool}{v}
    if p, err := json.Marshal(b); err != nil {
        ctx.WriteHeader(httpBad)
        return
    } else {
        ctx.Write(p)
    }
}

func getTransaction(ctx *web.Context, key string) (trans fct.ITransaction, err error) {
    // Now get the transaction.  If we don't have a transaction by the given keys
    // thre is nothing we can do.  Now we *could* create the transaaction and tie
    // it to the key.  Something to think about.
    ib := factoidState.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key))
    trans, ok := ib.(fct.ITransaction)
    if ib == nil || !ok {
        fmt.Println("Unknown Transaction: '",key,"'")
        err = fmt.Errorf("Unknown Transaction")
        reportResults(ctx,false)
        ctx.WriteHeader(httpBad)
        return
    }
    return
}

// &key=<key>&name=<name or address>&amount=<amount>
func getParams(ctx *web.Context, params string) (
    trans fct.ITransaction, 
    key string, 
    name string, 
    address fct.IAddress, 
    amount int64 , 
    err error) {
    
    key = ctx.Params["key"]
    name = ctx.Params["name"]
    StrAmount := ctx.Params["amount"]
    
    if len(key)==0 || len(name)==0 || len(StrAmount)==0 {
        fmt.Println("Missing Parameters: key='",key,"' name='",name,"' amount='",StrAmount,"'")
        reportResults(ctx,false)
        err = fmt.Errorf("Missing Parameters")
        ctx.WriteHeader(httpBad)
        return
    }
    
    amount, err = strconv.ParseInt(StrAmount,10,64)
    if err != nil {
        fmt.Println("Error parsing amount.\n",err)
        err = fmt.Errorf("Error parsing amount")
        reportResults(ctx,false)
        ctx.WriteHeader(httpBad)
        return
    }
    
    // Get the transaction
    trans, err = getTransaction(ctx,key)
    if err != nil { return }
    
    // Get the input/output/ec address.  Which could be a name.  First look and see if it is
    // a name.  If it isn't, then look and see if it is an address.  Someone could
    // do a weird Address as a name and fool the code, but that seems unlikely.
    // Could check for that some how, but there are many ways around such checks.
    
    we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME_HASH),[]byte(name))
    if we != nil {
        address,err = we.(wallet.IWalletEntry).GetAddress()
        if err != nil || address == nil {
            fmt.Println("Should not get an error geting a address from a Wallet Entry")
            err = fmt.Errorf("Wallet Entry failed to provide address")
            reportResults(ctx,false)
            ctx.WriteHeader(httpBad)
            return
        }
    }else{
        badr, er := hex.DecodeString(name)
        if er != nil || len(badr) != fct.ADDRESS_LENGTH {
            fmt.Println("Bad or unknown address: "+name)
            err = fmt.Errorf("Bad or unknown address")
            reportResults(ctx,false)
            ctx.WriteHeader(httpBad)
            return
        }
        address = fct.NewAddress(badr)
    }
    
    return
}
    

// New Transaction:  key --
// We create a new transaction, and track it with the user supplied key.  The user
// can then use this key to make subsequent calls to add inputs, outputs, and to sign.
// Then they can submit the transaction.
//
// When the transaction is submitted, we clear it from our working memory.  Multiple
// transactions can be under construction at one time, but they need their own keys.
// Once a transaction is either submitted or deleted, the key can be reused.
//
func  handleFactoidNewTransaction(ctx *web.Context, key string) {

    // Make sure we have a key
    if len(key) == 0 {
        fmt.Println("Missing transaction key")
        reportResults(ctx,false)
        ctx.WriteHeader(httpBad)
        return
    }
    // Make sure we don't already have a transaction in process with this key
    t := factoidState.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key))
    if t != nil {
        fmt.Println("Duplicate key: '",key,"'")
        reportResults(ctx,false)
        ctx.WriteHeader(httpBad)
        return
    }
    // Create a transaction
    t = factoidState.GetWallet().CreateTransaction()
    // Save it with the key
    factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key),t)
        
    reportResults(ctx,true)
    
}



func  handleFactoidAddInput(ctx *web.Context, parms string) {
    trans, key, _, address, amount, err := getParams(ctx, parms);
    if err != nil {
        return
    }
    
    // Add our new input
    err = factoidState.GetWallet().AddInput(trans,address,uint64(amount))
    if err != nil {
        fmt.Println("Failed to add input")
        reportResults(ctx,false)
        ctx.WriteHeader(httpBad)
        return
    }
   
    // Update our map with our new transaction to the same key.  Otherwise, all of
    // our work will go away!
    factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key),trans)
    
    reportResults(ctx,true)    
}

func  handleFactoidAddOutput(ctx *web.Context, parms string) {
    trans, key, _, address, amount, err := getParams(ctx, parms);
    if err != nil {
        return
    }
    
    // Add our new Output
    err = factoidState.GetWallet().AddOutput(trans,address,uint64(amount))
    if err != nil {
        fmt.Println("Failed to add input")
        reportResults(ctx,false)
        ctx.WriteHeader(httpBad)
        return
    }
    
    // Update our map with our new transaction to the same key.  Otherwise, all of
    // our work will go away!
    factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key),trans)
    
    reportResults(ctx,true)    
}

func  handleFactoidAddECOutput(ctx *web.Context, parms string) {
    trans, key, _, address, amount, err := getParams(ctx, parms);
    if err != nil {
        return
    }
    
    // Add our new Entry Credit Output
    err = factoidState.GetWallet().AddECOutput(trans,address,uint64(amount))
    if err != nil {
        fmt.Println("Failed to add input")
        reportResults(ctx,false)
        ctx.WriteHeader(httpBad)
        return
    }
    
    // Update our map with our new transaction to the same key.  Otherwise, all of
    // our work will go away!
    factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key),trans)
    
    reportResults(ctx,true)    
}

func  handleFactoidSignTransaction(ctx *web.Context, key string) {
    
    // Get the transaction
    trans, err := getTransaction(ctx,key)
    if err != nil { return }
    
    valid, err := factoidState.GetWallet().Validate(trans)
    if !valid || err != nil {
        reportResults(ctx,false)
        return
    }
    
    valid, err = factoidState.GetWallet().SignInputs(trans)
    if err != nil || !valid {
        reportResults(ctx,false)
        return
    }

    // Update our map with our new transaction to the same key.  Otherwise, all of
    // our work will go away!
    factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key),trans)
    
    reportResults(ctx,true)    
}

func handleFactoidSubmit(ctx *web.Context) {
    
    type transaction struct {
        Transaction string
    }
    t := new(transaction)
    if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
        ctx.WriteHeader(httpBad)
        return
    } else {
        if err := json.Unmarshal(p, t); err != nil {
            ctx.WriteHeader(httpBad)
            return
        }
    }
    
    // Get the transaction
    trans, err := getTransaction(ctx,t.Transaction)
    
    if err != nil { return }
    
    valid := factoidState.GetWallet().ValidateSignatures(trans)
    if !valid {
        reportResults(ctx,false)
        return
    }
    
    // Okay, transaction is good, so marshal and send to factomd!
    data, err := trans.MarshalBinary()
    if err != nil {
        reportResults(ctx,false)
        return
    }
    
    transdata := string(hex.EncodeToString(data))
        
    s := struct{Transaction string}{transdata}
    
    j, err := json.Marshal(s)
    if err != nil {
        reportResults(ctx,false)
        return
    }
        
    resp, err := http.Post(
        fmt.Sprintf("http://%s/v1/factoid-submit/", ipaddressFD+portNumberFD),
                           "application/json",
                           bytes.NewBuffer(j))
    
    fmt.Println(bytes.NewBuffer(j)) /////////////////////////////////////////////
    
    if err != nil {
        fmt.Println("Error coming back from server ")
        return 
    }
    resp.Body.Close()
}
   
func handleGetFee(ctx *web.Context) {
    
    str := fmt.Sprintf("http://%s/v1/factoid-get-fee/", ipaddressFD+portNumberFD)
    resp, err := http.Get(str)
    if err != nil {
        fmt.Println("\n",str)
        reportResults(ctx,false)
        return 
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        reportResults(ctx,false)
        return 
    }
    resp.Body.Close()
    
    type x struct { Fee int64 }
    b := new(x)
    if err := json.Unmarshal(body, b); err != nil {
        fmt.Println(err)
        reportResults(ctx,false)
        return 
    }
    
    ctx.Write(body)
    
}   
    
func handleFactoidValidate(ctx *web.Context) {
}