// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
    "regexp"
	"bytes"
    "strings"
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

var badChar,_ = regexp.Compile("[^A-Za-z0-9_-]")
var badHexChar,_ = regexp.Compile("[^A-Fa-f0-9]")

type Response struct {
    Response string
    Success bool
}

func ValidateKey(key string) (msg string, valid bool) {
    if len(key) > fct.ADDRESS_LENGTH     { 
        return "Key is too long.  Keys must be less than 32 characters", false     
    }
    if badChar.FindStringIndex(key)!=nil { 
        str := fmt.Sprintf("The key or name '%s' contains invalid characters.\n"+
          "Keys and names are restricted to alphanumeric characters,\n"+
          "minuses (dashes), and underscores", key)
        return str, false
    }
    return "", true
}

// True is sccuess! False is failure.  The Response is what the CLI
// should report.
func reportResults(ctx *web.Context, response string , success bool) {
    b := Response {
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
    
    msg, valid := ValidateKey(key)
    if !valid {
        return nil, fmt.Errorf(msg)
    }
    
    // Now get the transaction.  If we don't have a transaction by the given
    // keys there is nothing we can do.  Now we *could* create the transaaction
    // and tie it to the key.  Something to think about.
    ib := factoidState.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key))
    
    trans, ok := ib.(fct.ITransaction)
    if ib == nil || !ok {
        return nil, fmt.Errorf("Unknown Transaction: %s",key)
    }
    return 
}

// &key=<key>&name=<name or address>&amount=<amount>
// If no amount is specified, a zero is returned.
func getParams_(ctx *web.Context, params string, ec bool) (
    trans fct.ITransaction, 
    key string, 
    name string, 
    address fct.IAddress, 
    amount int64 , 
    ok bool) {
    
    key = ctx.Params["key"]
    name = ctx.Params["name"]
    StrAmount := ctx.Params["amount"]
    
    if len(StrAmount)== 0 {
        StrAmount = "0"
    }
    
    if len(key)==0 || len(name)==0 {
        str := fmt.Sprintln("Missing Parameters: key='",key,"' name='",name,"' amount='",StrAmount,"'")
        reportResults(ctx,str,false)
        ok = false
        return 
    }
    
    msg, valid := ValidateKey(key)
    if !valid {
        reportResults(ctx,msg,false)
        ok = false
        return 
    }
     
    amount, err := strconv.ParseInt(StrAmount,10,64)
    if err != nil {
        str := fmt.Sprintln("Error parsing amount.\n",err)
        reportResults(ctx,str,false)
        ok = false
        return 
    }
    
    // Get the transaction
    trans, err = getTransaction(ctx,key)
    if err != nil { 
        reportResults(ctx,"Failure to locate the transaction",false)
        ok = false
        return 
    }
    
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
                ok = false
                return 
            }
            ok = true
            return 
        }
    }
    if (!ec && !fct.ValidateFUserStr(name)) || (ec && !fct.ValidateECUserStr(name)) {
        reportResults(ctx,fmt.Sprintf("The address specified isn't defined or is invalid: %s",name),false)
        ctx.WriteHeader(httpBad)
        ok = false
        return 
    }
    baddr := fct.ConvertUserStrToAddress(name)
    
    address = fct.NewAddress(baddr)
    
    ok = true
    return 
}

/*************************************************************************
 * Handler Functions
 *************************************************************************/

// Setup:  seed --
// Setup creates the 10 fountain Factoid Addresses, then sets address
// generation to be unique for this wallet.  You CAN call setup multiple
// times, but once the Fountain addresses are created, Setup only changes
// the seed.
//
// Setup must be called once before you do anything else with the wallet.
//

func HandleFactoidSetup(ctx *web.Context, seed string) {
    // Make sure we have a seed.
    if len(seed) == 0 {
        msg := "You must supply some random seed. For example (don't use this!)\n"+
        "factom-cli setup 'woe!#in31!%234ng)%^&$%oeg%^&*^jp45694a;gmr@#t4 q34y'\n"+
        "would make a nice seed.  The more random the better.\n\n"+
        "Note that if you create an address before you call Setup, you must\n"+
        "use those address(s) as you access the fountians."
        
        reportResults(ctx,msg,false)
    }
    setFountian := false
    keys,_ := factoidState.GetDB().GetKeysValues([]byte(fct.W_NAME)) 
    if len(keys) == 0 {
        setFountian = true
        for i:=1; i<=10; i++ {
            name := fmt.Sprintf("%02d-Fountain",i)
            _, err := factoidState.GetWallet().GenerateFctAddress([]byte(name),1,1)
            if err != nil {
                reportResults(ctx,err.Error(),false)
                return
            }
        }
    }

    seedprime := fct.Sha([]byte(fmt.Sprintf("%s%v",seed,time.Now().UnixNano()))).Bytes()
    factoidState.GetWallet().NewSeed(seedprime)
        
    if setFountian {
        reportResults(ctx,"New seed set, fountain addresses defined", true)
    }else{
        reportResults(ctx,"New seed set, no fountain addresses defined", true)
    }
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
		return
	}
	
	msg, valid := ValidateKey(key)
    if !valid {
        reportResults(ctx, msg, false)
        return
    }
    
	// Make sure we don't already have a transaction in process with this key
	t := factoidState.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	if t != nil {
		str := fmt.Sprintln("Duplicate key: '", key, "'")
        reportResults(ctx, str, false)
		return
	}
	// Create a transaction
	t = factoidState.GetWallet().CreateTransaction(factoidState.GetTimeMilli())
	// Save it with the key
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), t)

	reportResults(ctx,"Success building a transaction", true)
}

// Delete Transaction:  key --
// Remove a transaction rather than sign and submit the transaction.  Sometimes
// you just need to throw one a way, and rebuild it.
//
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

func HandleFactoidAddFee(ctx *web.Context, parms string) {
    trans, key, _, address, _, ok := getParams_(ctx, parms, false)
    if !ok {
        return
    }
    
    name := ctx.Params["name"]   // This is the name the user used.
    
    {
        ins,err  := trans.TotalInputs();  if err!=nil { reportResults(ctx,err.Error(), false) } 
        outs,err := trans.TotalOutputs(); if err!=nil { reportResults(ctx,err.Error(), false) } 
        ecs,err  := trans.TotalECs();     if err!=nil { reportResults(ctx,err.Error(), false) } 
    
        if ins != outs+ecs {
            msg := fmt.Sprintf(
            "Addfee requires that all the inputs balance the outputs.\n"+
            "The total inputs of your transaction are              %s\n"+
            "The total outputs + ecoutputs of your transaction are %s",
            fct.ConvertDecimal(ins), fct.ConvertDecimal(outs+ecs))
            
            reportResults(ctx,msg,false)
            return
        }
    }
        
    msg, ok := ValidateKey(key) 
    if !ok {
        reportResults(ctx, msg, false)
        return
    }
    
    fee,err := GetFee(ctx)
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return
    }
    
    transfee,err := trans.CalculateFee(uint64(fee))
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return
    }
    
    adr, err := factoidState.GetWallet().GetAddressHash(address)
    if err != nil {
        reportResults(ctx, err.Error(),false)
        return
    }
    
    for _,input := range trans.GetInputs() {
        
        if input.GetAddress().IsSameAs(adr) {
            amt,err := fct.ValidateAmounts(input.GetAmount(), transfee)  
            if err != nil {
                reportResults(ctx,err.Error(),false)
                return
            }
            input.SetAmount(amt)
            reportResults(ctx, fmt.Sprintf("Added %s to %s",fct.ConvertDecimal(uint64(transfee)),name), true)
            return
        }
    }
    reportResults(ctx, fmt.Sprintf("%s is not an input to the transaction.",key), false)
    return
}
 


func HandleFactoidAddInput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, ok := getParams_(ctx, parms, false)
    
    if !ok {
		return
	}
    msg, ok := ValidateKey(key) 
    if !ok {
        reportResults(ctx, msg, false)
        return
    }
    
    // First look if this is really an update
    for _,input := range trans.GetInputs() {
        if input.GetAddress().IsSameAs(address) {
            oldamt := input.GetAmount()
            input.SetAmount(uint64(amount))
            reportResults(ctx, fmt.Sprintf("Input was %s\n"+
                                           "Now is    %s",
                fct.ConvertDecimal(oldamt),
                fct.ConvertDecimal(uint64(amount))), true)
            return
        }
    }
    
	// Add our new input
	err := factoidState.GetWallet().AddInput(trans, address, uint64(amount))
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
	trans, key, _, address, amount, ok := getParams_(ctx, parms, false)
	if !ok {
		return
	}
	
	msg, ok := ValidateKey(key) 
    if !ok {
        reportResults(ctx, msg, false)
        return
    }
    
    // First look if this is really an update
    for _,output := range trans.GetOutputs() {
        if output.GetAddress().IsSameAs(address) {
            oldamt := output.GetAmount()
            output.SetAmount(uint64(amount))
            reportResults(ctx, fmt.Sprintf("Input was %s\n"+
                                           "Now is    %s",
                fct.ConvertDecimal(oldamt),
                fct.ConvertDecimal(uint64(amount))), true)
            return
        }
    }    
	// Add our new Output
	err := factoidState.GetWallet().AddOutput(trans, address, uint64(amount))
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
	trans, key, _, address, amount, ok := getParams_(ctx, parms, true)
	if !ok {
		return
	}

	msg, ok := ValidateKey(key) 
    if !ok {
        reportResults(ctx, msg, false)
        return
    }
    // First look if this is really an update
    for _,ecoutput := range trans.GetECOutputs() {
        if ecoutput.GetAddress().IsSameAs(address) {
            oldamt := ecoutput.GetAmount()
            ecoutput.SetAmount(uint64(amount))
            reportResults(ctx, fmt.Sprintf("Input was %s\n"+
            "Now is    %s",
            fct.ConvertDecimal(oldamt),
                                           fct.ConvertDecimal(uint64(amount))), true)
            return
        }
    }     
	// Add our new Entry Credit Output
	err := factoidState.GetWallet().AddECOutput(trans, address, uint64(amount))
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
    
    msg, ok := ValidateKey(key) 
    if !ok {
        reportResults(ctx, msg, false)
        return
    }
    
    // Get the transaction
    trans, err := getTransaction(ctx, key)
    if err != nil {
        reportResults(ctx, "Failed to get the transaction", false)
        return
    }

    err = factoidState.GetWallet().Validate(1,trans)
    if err != nil {
    	reportResults(ctx, err.Error(), false)
    	return
    }

    valid, err := factoidState.GetWallet().SignInputs(trans)
    if err != nil {
        reportResults(ctx,err.Error(), false)
    }
    if !valid {
        reportResults(ctx,"Not all inputs are signed", false)
    }
    // Update our map with our new transaction to the same key.  Otherwise, all
    // of our work will go away!
    factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

    reportResults(ctx, "Success signing transaction", true)    
}

func HandleFactoidSubmit(ctx *web.Context, jsonkey string) {
    type submitReq struct {
        Transaction string
    }
    
    in := new(submitReq)
    json.Unmarshal([]byte(jsonkey), in)
    
    key := in.Transaction
    // Get the transaction
    trans, err := getTransaction(ctx, key)
    if err != nil {
        reportResults(ctx, err.Error(), false)
        return
    }

    err = factoidState.GetWallet().ValidateSignatures(trans)
    if err != nil {
    	reportResults(ctx, err.Error() ,false)
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

    if err != nil {
        reportResults(ctx,err.Error(),false)
    	return
    }
    
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        reportResults(ctx,err.Error(),false)
    }
    
    resp.Body.Close()

    r := new(Response)
    if err := json.Unmarshal(body, r); err != nil {
        reportResults(ctx,err.Error(),false)
    }

    if r.Success {
        factoidState.GetDB().DeleteKey([]byte(fct.DB_BUILD_TRANS), []byte(key))
        reportResults(ctx,r.Response,true)
    }else{
        reportResults(ctx,r.Response,false)
    }
    
}
   
func GetFee(ctx *web.Context) (int64, error) {
    str := fmt.Sprintf("http://%s/v1/factoid-get-fee/", ipaddressFD+portNumberFD)
    resp, err := http.Get(str)
    if err != nil {
        return 0, err
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        resp.Body.Close()
        return 0, err
    }
    resp.Body.Close()

    type x struct { Fee int64 }
    b := new(x)
    if err := json.Unmarshal(body, b); err != nil {
        return 0, err
    }
    
    return b.Fee, nil
}   

func HandleGetFee(ctx *web.Context) {
    str := fmt.Sprintf("http://%s/v1/factoid-get-fee/", ipaddressFD+portNumberFD)
    resp, err := http.Get(str)
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return 
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return 
    }
    resp.Body.Close()
    
    type x struct { Fee int64 }
    b := new(x)
    if err := json.Unmarshal(body, b); err != nil {
        reportResults(ctx,err.Error(),false)
        return 
    }
    
    ctx.Write(body)
}   


func GetAddresses() ([]byte) {
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
            bal,_ := ECBalance(adr)
            ecBalances = append(ecBalances,strconv.FormatInt(bal,10))
        }else{
            address, err := we.GetAddress()
            if err != nil { continue }
            adr = fct.ConvertFctAddressToUserStr(address)
            fctAddresses = append(fctAddresses,adr)
            fctKeys = append(fctKeys, string(k))
            bal,_ := FctBalance(adr)
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

func GetTransactions(ctx *web.Context) ([]byte, error) {
    exch,err := GetFee(ctx)
    if err != nil {
        return nil, err
    }

    // Get the transactions in flight.
    keys, values := factoidState.GetDB().GetKeysValues([]byte(fct.DB_BUILD_TRANS))
    
    for i:=0; i<len(keys)-1; i++ {
        for j:=0; j<len(keys)-i-1;j++ {
            if bytes.Compare(keys[j],keys[j+1])>0 {
                t := keys[j]
                keys[j]=keys[j+1]
                keys[j+1]=t
                t2 := values[j]
                values[j]=values[j+1]
                values[j+1]=t2
            }
            
        }
    }
    var out bytes.Buffer
    for i,key := range keys {
        if values[i]== nil {continue}
        trans := values[i].(fct.ITransaction)
        
        fee, _ := trans.CalculateFee(uint64(exch))  
        cprt := ""
        cin, err := trans.TotalInputs() 
        if err != nil {
            cprt = cprt + err.Error()
        }
        cout, err := trans.TotalOutputs() 
        if err != nil {
            cprt = cprt + err.Error()
        }
        cecout, err := trans.TotalECs() 
        if err != nil {
            cprt = cprt + err.Error()
        }
        
        if len(cprt)==0 {
            v := int64(cin) - int64(cout) - int64(cecout)
            sign := ""
            if v < 0 {
                sign = "-"
                v = -v
            }
            cprt = fmt.Sprintf(" Currently will pay: %s%s",
                        sign,
                        strings.TrimSpace(fct.ConvertDecimal(uint64(v))))
            if sign == "-" || fee > uint64(v) {
                cprt = cprt + "\n\nWARNING: Currently your transaction fee may be too low"
            }
        }
        
        out.WriteString(fmt.Sprintf("\n%25s:  Fee Due: %s  %s\n\n%s\n",
                                    key,
                                    strings.TrimSpace(fct.ConvertDecimal(fee)),
                                    cprt,
                                    values[i].String()))     
    }
    
    output := out.Bytes()
    // now look for the addresses, and replace them with our names. (the transactions
    // in flight also have a Factom address... We leave those alone.
    
    names, vs    := factoidState.GetDB().GetKeysValues([]byte(fct.W_NAME))
    
    for i,name := range names {
        we,ok := vs[i].(wallet.IWalletEntry)
        if !ok { return nil,fmt.Errorf("Database is corrupt") }

        address, err := we.GetAddress()
        if err != nil { continue }      // We shouldn't get any of these, but ignore them if we do.
        adrstr := []byte(hex.EncodeToString(address.Bytes()))
        
        output = bytes.Replace(output,adrstr,name,-1)
    }
    
    return output, nil
}


   
func   HandleGetAddresses  (ctx *web.Context) {
    b := new(Response)
    b.Response = string(GetAddresses())
    b.Success = true
    j, err := json.Marshal(b)
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return
    }
    ctx.Write(j)
}    
  
func   HandleGetTransactions  (ctx *web.Context) {
    b := new(Response)
    txt,err := GetTransactions(ctx)
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return
    }
    b.Response = string(txt)
    b.Success = true
    j, err := json.Marshal(b)
    if err != nil {
        reportResults(ctx,err.Error(),false)
        return
    }
    ctx.Write(j)
}      
   
func HandleFactoidValidate(ctx *web.Context) {
}


func HandleFactoidNewSeed(ctx *web.Context) {
}