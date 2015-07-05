// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
    "time"

	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"github.com/hoisie/web"
)

/******************************************
 * Helper Functions
 ******************************************/

func GetTimeNano() uint64 {
    return uint64(time.Now().UnixNano())
}


// True is sccuess! False is failure.  The Response is what the CLI
// should report.
func reportResults(ctx *web.Context, response string , success bool) {
    b := struct{
            Response string; 
            Success bool
         } {
            Response: response, 
            Success:  success,
         }
    if p, err := json.Marshal(b); err != nil {
        ctx.WriteHeader(httpBad)
        return
    } else {
        ctx.Write(p)
    }
}

func getTransaction(ctx *web.Context, key string) (trans fct.ITransaction, err error) {
    // Now get the transaction.  If we don't have a transaction by the given
    // keys there is nothing we can do.  Now we *could* create the transaaction
    // and tie it to the key.  Something to think about.
    ib := factoidState.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key))
    trans, ok := ib.(fct.ITransaction)
    if ib == nil || !ok {
        str := fmt.Sprintf("Unknown Transaction: %s",key)
        reportResults(ctx,str,false)
        ctx.WriteHeader(httpBad)
        return
    }
    return
}

// &key=<key>&name=<name or address>&amount=<amount>
func getParams(ctx *web.Context, params string, ec bool) (
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
        str := fmt.Sprintln("Missing Parameters: key='",key,"' name='",name,"' amount='",StrAmount,"'")
        reportResults(ctx,str,false)
        return
    }
        
    amount, err = strconv.ParseInt(StrAmount,10,64)
    if err != nil {
        str := fmt.Sprintln("Error parsing amount.\n",err)
        reportResults(ctx,str,false)
        return
    }
    
    // Get the transaction
    trans, err = getTransaction(ctx,key)
    if err != nil { return }
    
    // Get the input/output/ec address.  Which could be a name.  First look and see if it is
    // a name.  If it isn't, then look and see if it is an address.  Someone could
    // do a weird Address as a name and fool the code, but that seems unlikely.
    // Could check for that some how, but there are many ways around such checks.
    
    if len(name)<= fct.ADDRESS_LENGTH {
        we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
        if we != nil {
            address,err = we.(wallet.IWalletEntry).GetAddress()
            if err != nil || address == nil {
                reportResults(ctx,"Should not get an error geting a address from a Wallet Entry",false)
                return
            }
            return
        }
    }
    if (!ec && !fct.ValidateFUserStr(name)) || (ec && !fct.ValidateECUserStr(name)) {
        reportResults(ctx,"Badly formed address",false)
        ctx.WriteHeader(httpBad)
        return
    }
    baddr := fct.ConvertUserStrToAddress(name)
    
    address = fct.NewAddress(baddr)
    
    return
}

// New Transaction:  key --
// We create a new transaction, and track it with the user supplied key.  The
// user can then use this key to make subsequent calls to add inputs, outputs,
// and to sign. Then they can submit the transaction.
//
// When the transaction is submitted, we clear it from our working memory.
// Multiple transactions can be under construction at one time, but they need
// their own keys. Once a transaction is either submitted or deleted, the key
// can be reused.
func HandleFactoidNewTransaction(ctx *web.Context, key string) {
	// Make sure we have a key
	if len(key) == 0 {
        reportResults(ctx, "Missing transaction key", false)
		ctx.WriteHeader(httpBad)
		return
	}
	// Make sure we don't already have a transaction in process with this key
	t := factoidState.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	if t != nil {
		str := fmt.Sprintln("Duplicate key: '", key, "'")
		reportResults(ctx,str, false)
		return
	}
	// Create a transaction
	t = factoidState.GetWallet().CreateTransaction(GetTimeNano())
	// Save it with the key
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), t)

	reportResults(ctx,"Success building a transaction", true)
}

// New Transaction:  key --
// We create a new transaction, and track it with the user supplied key.  The
// user can then use this key to make subsequent calls to add inputs, outputs,
// and to sign. Then they can submit the transaction.
//
// When the transaction is submitted, we clear it from our working memory.
// Multiple transactions can be under construction at one time, but they need
// their own keys. Once a transaction is either submitted or deleted, the key
// can be reused.
func HandleFactoidDeleteTransaction(ctx *web.Context, key string) {
    // Make sure we have a key
    if len(key) == 0 {
        reportResults(ctx,"Missing transaction key",false)  
        return
    }
    // Wipe out the key
    factoidState.GetDB().DeleteKey([]byte(fct.DB_BUILD_TRANS), []byte(key))
    reportResults(ctx, "Success deleting transaction",true)
}



func HandleFactoidAddInput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, err := getParams(ctx, parms, false)
	if err != nil {
		return
	}

	// Add our new input
	err = factoidState.GetWallet().AddInput(trans, address, uint64(amount))
	if err != nil {
        reportResults(ctx, "Failed to add input", false)
		return
	}

	// Update our map with our new transaction to the same key. Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	reportResults(ctx, "Success adding Input", true)
}

func HandleFactoidAddOutput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, err := getParams(ctx, parms, false)
	if err != nil {
		return
	}

	// Add our new Output
	err = factoidState.GetWallet().AddOutput(trans, address, uint64(amount))
	if err != nil {
        reportResults(ctx, "Failed to add output", false)
		return
	}

	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	reportResults(ctx, "Success adding output", true)
}

func HandleFactoidAddECOutput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, err := getParams(ctx, parms, true)
	if err != nil {
		return
	}

	// Add our new Entry Credit Output
	err = factoidState.GetWallet().AddECOutput(trans, address, uint64(amount))
	if err != nil {
        reportResults(ctx, "Failed to add input", false)
		return
	}

	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	reportResults(ctx,"Success adding Entry Credit Output", true)
}

func  HandleFactoidSignTransaction(ctx *web.Context, key string) {
    // Get the transaction
    trans, err := getTransaction(ctx, key)
    if err != nil {
    	return
    }

    valid, err := factoidState.GetWallet().Validate(trans)
    if !valid || err != nil {
    	reportResults(ctx, "Transaction is Invalid", false)
    	return
    }

    valid, err = factoidState.GetWallet().SignInputs(trans)
    if err != nil {
        str:= fmt.Sprintf("%s",err)
        reportResults(ctx,str, false)
    }
    if !valid {
    	reportResults(ctx,"Could not sign all inputs", false)
    	return
    }

    // Update our map with our new transaction to the same key.  Otherwise, all
    // of our work will go away!
    factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

    reportResults(ctx, "Success signing transaction", true)    
}

func HandleFactoidSubmit(ctx *web.Context) {
    type transaction struct {
    	Transaction string
    }
    t := new(transaction)
    if p, err := ioutil.ReadAll(ctx.Request.Body); err != nil {
    	ctx.WriteHeader(httpBad)
    	return
    } else {
    	if err := json.Unmarshal(p, t); err != nil {
            reportResults(ctx,"Failed to unmarshal the response from factomd",false)
    		return
    	}
    }

    // Get the transaction
    trans, err := getTransaction(ctx, t.Transaction)

    if err != nil {
    	return
    }

    valid := factoidState.GetWallet().ValidateSignatures(trans)
    if !valid {
    	reportResults(ctx, "Could not validate all the signatures of the transaction",false)
    	return
    }

    // Okay, transaction is good, so marshal and send to factomd!
    data, err := trans.MarshalBinary()
    if err != nil {
        reportResults(ctx,"Failed to marshal the transaction for factomd",false)
    	return
    }

    transdata := string(hex.EncodeToString(data))

    s := struct{ Transaction string }{transdata}

    j, err := json.Marshal(s)
    if err != nil {
        reportResults(ctx,"Failed to marshal the transaction for factomd",false)
        
    	return
    }

    resp, err := http.Post(
    	fmt.Sprintf("http://%s/v1/factoid-submit/", ipaddressFD+portNumberFD),
    	"application/json",
    	bytes.NewBuffer(j))

    fmt.Println(bytes.NewBuffer(j))

    if err != nil {
        str := fmt.Sprintf("%s",err)
        reportResults(ctx,str,false)
    	return
    }
    resp.Body.Close()

    factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(t.Transaction), nil)
    reportResults(ctx,"Transaction submitted",true)
    
}
   
func HandleGetFee(ctx *web.Context) {
    str := fmt.Sprintf("http://%s/v1/factoid-get-fee/", ipaddressFD+portNumberFD)
    resp, err := http.Get(str)
    if err != nil {
        str := fmt.Sprintf("%s",err.Error())
        reportResults(ctx,str,false)
        return 
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        str := fmt.Sprintf("%s",err.Error())
        reportResults(ctx,str,false)
        return 
    }
    resp.Body.Close()
    
    type x struct { Fee int64 }
    b := new(x)
    if err := json.Unmarshal(body, b); err != nil {
        str := fmt.Sprintf("%s",err.Error())
        reportResults(ctx,str,false)
        return 
    }
    
    ctx.Write(body)
}   
   
func GetAddresses() [] byte{
    keys, values := factoidState.GetDB().GetKeysValues([]byte(fct.W_NAME))
    
    ecKeys := make([]string,0,len(keys))
    fctKeys := make([]string,0,len(keys))
    ecBalances := make([]string,0,len(keys))
    fctBalances := make([]string,0,len(keys))
    fctAddresses := make([]string,0,len(keys))
    ecAddresses := make([]string,0,len(keys))
    
    var maxlen int
    for i,k := range keys {
        if len(k) > maxlen {maxlen = len(k)}
        we,ok := values[i].(wallet.IWalletEntry)
        if !ok { 
            panic("Get Addresses finds the database corrupt.  Shouldn't happen")
        }
        var adr string
        if we.GetType() == "ec" {
            address, err := we.GetAddress()
            if err != nil { continue }
            adr = fct.ConvertECAddressToUserStr(address)
            ecAddresses = append(ecAddresses,adr)
            ecKeys = append(ecKeys, string(k))
            bal := ECBalance(adr)
            ecBalances = append(ecBalances,strconv.FormatInt(bal,10))
        }else{
            address, err := we.GetAddress()
            if err != nil { continue }
            adr = fct.ConvertFctAddressToUserStr(address)
            fctAddresses = append(fctAddresses,adr)
            fctKeys = append(fctKeys, string(k))
            bal := FctBalance(adr)
            sbal := fct.ConvertDecimal(uint64(bal))
            fctBalances = append(fctBalances,sbal)
        }
    }
    var out bytes.Buffer
    if len(fctKeys)>0 { out.WriteString("\n  Factoid Addresses\n\n") }
    fstr := fmt.Sprintf("%s%vs    %s38s %s14s\n","%",maxlen+4,"%","%")
    for i,key := range fctKeys {
        str := fmt.Sprintf(fstr,  key,  fctAddresses[i],   fctBalances[i]) 
        out.WriteString(str)
    }
    if len(ecKeys)>0 { out.WriteString("\n  Entry Credit Addresses\n\n") }
    for i,key := range ecKeys {
        str := fmt.Sprintf(fstr,  key,  ecAddresses[i],    ecBalances[i]) 
        out.WriteString(str)
    }
    
    return out.Bytes()
}
   
   
func   HandleGetAddresses  (ctx *web.Context) {
    
    type x struct {
        Body string
        Success bool
    }
    b := new(x)
    b.Body = string(GetAddresses())
    b.Success = true
    j, err := json.Marshal(b)
    if err != nil {
        str := fmt.Sprintf("%s",err.Error())
        reportResults(ctx,str,false)
        return
    }
    ctx.Write(j)
}    
   
   
func HandleFactoidValidate(ctx *web.Context) {
}
