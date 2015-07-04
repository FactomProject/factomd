// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import(
    "bytes"
    "encoding/json"
    "net/http"    
    "encoding/hex"
    "fmt"
    "strconv"
    "strings"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/wallet"
)  

/************************************************************
 * NewTransaction
 ************************************************************/

type NewTransaction struct {
    ICommand
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
func (NewTransaction) Execute (state State, args []string) error {
    
    if len(args)!= 2 {
        return fmt.Errorf("Invalid Parameters")
    }
    key := args[1]
    
    // Make sure we don't already have a transaction in process with this key
    t := state.fs.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
    if t != nil {
        return fmt.Errorf("Duplicate key: '%s'", key)
    }
    // Create a transaction
    t = state.fs.GetWallet().CreateTransaction(state.fs.GetTimeNano())
    // Save it with the key
    state.fs.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), t)

    fmt.Println("Beginning Transaction ",key)
    return nil
}
    
func (NewTransaction) Name() string {
    return "NewTransaction"
}

func (NewTransaction) ShortHelp() string {
    return "NewTransaction <key> -- Begins the construction of a transaction.\n"+
           "                        Subsequent modifications must reference the key."
}

func (NewTransaction) LongHelp() string {
    return `
NewTransaction <key>                Begins the construction of a transaction.
                                    The <key> is any token without whitespace up to
                                    32 characters in length that can be used in 
                                    AddInput, AddOutput, AddECOutput, Sign, and
                                    Submit commands to construct and submit 
                                    transactions.
`
}


/************************************************************
 * AddInput
 ************************************************************/

type AddInput struct {
    ICommand
}
// AddInput <key> <name|address> amount
//
//

func (AddInput) Execute (state State, args []string) error {
    
    if len(args)!= 4 {
        return fmt.Errorf("Invalid Parameters")
    }
    key := args[1]
    adr := args[2]
    amt := args[3]
    
    ib := state.fs.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key))
    trans, ok := ib.(fct.ITransaction)
    if ib == nil || !ok {
        return fmt.Errorf("Unknown Transaction")
    }
    
    var addr fct.IAddress 
    if !fct.ValidateFUserStr(adr) {
        if len(adr) != 64 {
            if len(adr)>32 {
                return fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
            
            we := state.fs.GetDB().GetRaw([]byte(fct.W_NAME),[]byte(adr))
            
            if (we != nil){
                we2 := we.(wallet.IWalletEntry)
                addr,_ = we2.GetAddress()
                adr = hex.EncodeToString(addr.Bytes())
            }else{
                return fmt.Errorf("Name is undefined.")
            }
        }else {
            if badHexChar.FindStringIndex(adr)!=nil {
                return fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
        }
    } else {
        addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
    }
    amount,_ := fct.ConvertFixedPoint(amt)
    bamount,_ := strconv.ParseInt(amount,10,64)
    err := state.fs.GetWallet().AddInput(trans, addr, uint64(bamount))
    
    if err != nil {return err}
    
    fmt.Println("Added Input of ",amt," to be paid from ",args[2], 
        fct.ConvertFctAddressToUserStr(addr))
    return nil
}

func (AddInput) Name() string {
    return "AddInput"
}

func (AddInput) ShortHelp() string {
    return "AddInput <key> <name/address> <amount> -- Adds an input to a transaction.\n"+
           "                              the key should be created by NewTransaction, and\n"+
           "                              and the address and amount should come from your\n"+
           "                              wallet."
}

func (AddInput) LongHelp() string {
    return `
AddInput <key> <name|addr> <amt>    <key>       created by a previous NewTransaction call
                                    <name|addr> A Valid Name in your Factoid Address 
                                                book, or a valid Factoid Address
                                    <amt>       to be sent from the specified address to the
                                                outputs of this transaction.
`
}

/************************************************************
 * AddOutput
 ************************************************************/
type AddOutput struct {
    ICommand
}

// AddOutput <key> <name|address> amount
//
//

func (AddOutput) Execute (state State, args []string) error {
    
    if len(args)!= 4 {
        return fmt.Errorf("Invalid Parameters")
    }
    key := args[1]
    adr := args[2]
    amt := args[3]
    
    ib := state.fs.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key))
    trans, ok := ib.(fct.ITransaction)
    if ib == nil || !ok {
        return fmt.Errorf("Unknown Transaction")
    }
    
    var addr fct.IAddress 
    if !fct.ValidateFUserStr(adr) {
        if len(adr) != 64 {
            if len(adr)>32 {
                return fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
            
            we := state.fs.GetDB().GetRaw([]byte(fct.W_NAME),[]byte(adr))
            
            if (we != nil){
                we2 := we.(wallet.IWalletEntry)
                addr,_ = we2.GetAddress()
                adr = hex.EncodeToString(addr.Bytes())
            }else{
                return fmt.Errorf("Name is undefined.")
            }
        }else {
            if badHexChar.FindStringIndex(adr)!=nil {
                return fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
        }
    } else {
        addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
    }
    amount,_ := fct.ConvertFixedPoint(amt)
    bamount,_ := strconv.ParseInt(amount,10,64)
    err := state.fs.GetWallet().AddOutput(trans, addr, uint64(bamount))
    if err != nil {return err}
    
    fmt.Println("Added Output of ",amt," to be paid to ",args[2], 
                fct.ConvertFctAddressToUserStr(addr))
    
    return nil
}

func (AddOutput) Name() string {
    return "AddOutput"
}

func (AddOutput) ShortHelp() string {
    return "AddOutput <k> <n> <amount> -- Adds an output to a transaction.\n"+
           "                              the key <k> should be created by NewTransaction.\n"+
           "                              The address or name <n> can come from your address\n"+
           "                              book."
}

func (AddOutput) LongHelp() string {
    return `
AddOutput <key> <n|a> <amt>         <key>  created by a previous NewTransaction call
                                    <n|a>  A Valid Name in your Factoid Address 
                                           book, or a valid Factoid Address 
                                    <amt>  to be used to purchase Entry Credits at the
                                           current exchange rate.
`
}

/************************************************************
 * AddECOutput
 ************************************************************/
type AddECOutput struct {
    ICommand
}

// AddECOutput <key> <name|address> amount
//
// Buy Entry Credits

func (AddECOutput) Execute (state State, args []string) error {
    
    if len(args)!= 4 {
        return fmt.Errorf("Invalid Parameters")
    }
    key := args[1]
    adr := args[2]
    amt := args[3]
    
    ib := state.fs.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key))
    trans, ok := ib.(fct.ITransaction)
    if ib == nil || !ok {
        return fmt.Errorf("Unknown Transaction")
    }
    
    var addr fct.IAddress 
    if !fct.ValidateECUserStr(adr) {
        if len(adr) != 64 {
            if len(adr)>32 {
                return fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
            
            we := state.fs.GetDB().GetRaw([]byte(fct.W_NAME),[]byte(adr))
            
            if (we != nil){
                we2 := we.(wallet.IWalletEntry)
                addr,_ = we2.GetAddress()
                adr = hex.EncodeToString(addr.Bytes())
            }else{
                return fmt.Errorf("Name is undefined.")
            }
        }else {
            if badHexChar.FindStringIndex(adr)!=nil {
                return fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
        }
    } else {
        addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
    }
    amount,_ := fct.ConvertFixedPoint(amt)
    bamount,_ := strconv.ParseInt(amount,10,64)
    err := state.fs.GetWallet().AddECOutput(trans, addr, uint64(bamount))
    if err != nil {return err}
    
    fmt.Println("Added Output of ",amt," to be paid to ",args[2], 
                fct.ConvertECAddressToUserStr(addr))
    
    return nil
}

func (AddECOutput) Name() string {
    return "AddECOutput"
}

func (AddECOutput) ShortHelp() string {
    return "AddECOutput <k> <n> <amount> -- Adds an Entry Credit output (ecoutput)to a \n"+
           "                              transaction <k>.  The Entry Credits are assigned to\n"+
           "                              the address <n>.  The output <amount> is specified in\n"+
           "                              factoids, and purchases Entry Credits according to\n"+
           "                              the current exchange rage."
}

func (AddECOutput) LongHelp() string {
    return `
AddECOutput <key> <n|a> <amt>       <key>  created by a previous NewTransaction call
                                    <n|a>  Name or Address to hold the Entry Credits
                                    <amt>  Amount of Factoids to be used in this purchase.  Note
                                           that the exchange rate between Factoids and Entry
                                           Credits varies.
`
}

/************************************************************
 * Sign
 ************************************************************/
type Sign struct {
    ICommand
}

// Sign <k>
//
// Sign the given transaction identified by the given key
func (Sign) Execute (state State, args []string) error {
    if len(args)!= 2 {
        return fmt.Errorf("Invalid Parameters")
    }
    key := args[1]
    // Get the transaction
    ib := state.fs.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key))
    trans, ok := ib.(fct.ITransaction)
    if !ok {
        return fmt.Errorf("Invalid Parameters")
    }
    
    valid, err := state.fs.GetWallet().Validate(trans)
    if err != nil {
        return err
    }
    if !valid {
        return fmt.Errorf("Invalid Transaction")
    }
    
    valid, err = state.fs.GetWallet().SignInputs(trans)
    if err != nil {
        return err
    }
    if !valid {
        return fmt.Errorf("Error signing the transaction")
    }
    
    // Update our map with our new transaction to the same key.  Otherwise, all
    // of our work will go away!
    state.fs.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)
    
    return nil
    
}

func (Sign) Name() string {
    return "Sign"
}

func (Sign) ShortHelp() string {
    return "Sign <k> -- Sign the transaction given by the key <k>"
}

func (Sign) LongHelp() string {
    return `
Sign <key>                          Signs the transaction specified by the given key.
                                    Each input is found within the wallet, and if 
                                    we have the private key for that input, we 
                                    sign for that input.  
                                    
                                    Transctions can have inputs from multiple parties.
                                    In this case, the inputs can be signed by each
                                    party by first creating all the inputs and 
                                    outputs for a transaction.  Then signing your
                                    inputs.  Exporting the transaction.  Then
                                    sending it to the other party or parties for
                                    their signatures.
`
}

/************************************************************
 * Submit
 ************************************************************/
type Submit struct {
    ICommand
}

// Submit <k>
//
// Submit the given transaction identified by the given key
func (Submit) Execute (state State, args []string) error {

    if len(args)!= 2 {
        return fmt.Errorf("Invalid Parameters")
    }
    key := args[1]
    // Get the transaction
    ib := state.fs.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(key))
    trans, ok := ib.(fct.ITransaction)
    if !ok {
        return fmt.Errorf("Invalid Parameters")
    }
    
    valid, err := state.fs.GetWallet().Validate(trans)
    if err != nil {
        return err
    }
    if !valid {
        return fmt.Errorf("Invalid transaction")
    }
    
    ok = state.fs.GetWallet().ValidateSignatures(trans)
    if !ok {
        return fmt.Errorf("Not all signatures have been validated")
    }
    
    // Okay, transaction is good, so marshal and send to factomd!
    data, err := trans.MarshalBinary()
    if err != nil {
        return err
    }
    
    transdata := string(hex.EncodeToString(data))
    
    s := struct{ Transaction string }{transdata}
    
    j, err := json.Marshal(s)
    if err != nil {
        return err
    }
    
    resp, err := http.Post(
        fmt.Sprintf("http://%s/v1/factoid-submit/", state.GetServer()),
                           "application/json",
                           bytes.NewBuffer(j))
    
    if err != nil {
        return fmt.Errorf("Error coming back from server ")
    }
    resp.Body.Close()
    
    // Clear out the transaction
    state.fs.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), nil)
    
    fmt.Println("Transaction",key,"Submitted")
    
    return nil
}

func (Submit) Name() string {
    return "Submit"
}

func (Submit) ShortHelp() string {
    return "Submit <k> -- Submit the transaction given by the key <k>"
}

func (Submit) LongHelp() string {
    return `
Submit <key>                          Submits the transaction specified by the given key.
                                      Each input in the transaction must have  a valid
                                      signature, or Submit will reject the transaction.
`
}

/************************************************************
 * Print <v>
 ************************************************************/
type Print struct {
    ICommand
}

// Print <v1> <v2> ...
//
// Print Stuff.  We will add to this over time.  Right now, if <v> = a transaction
// key, it prints that transaction.

func (Print) Execute (state State, args []string) error {
    fmt.Println()
    for i, v := range args {
        if i == 0 {continue}
        
        ib := state.fs.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS),[]byte(v))
        trans, ok := ib.(fct.ITransaction)
        if ib != nil && ok {
            fmt.Println(trans)
            v, err := GetRate(state)
            if err != nil {fmt.Println(err); continue }
            fee, err := trans.CalculateFee(uint64(v))
            if err != nil {fmt.Println(err); continue }
            fmt.Println("Required Fee:       ", strings.TrimSpace(fct.ConvertDecimal(fee)))
            tin,  ok1 := trans.TotalInputs()
            tout, ok2 := trans.TotalOutputs()
            if ok1 && ok2 {
                cfee := int64(tin)- int64(tout)
                sign := ""
                if cfee < 0 { sign = "-"; cfee = -cfee } 
                fmt.Print("Fee You are paying: ", 
                        sign, strings.TrimSpace(fct.ConvertDecimal(uint64(cfee))),"\n")
            }else{
                if !ok1 {
                    fmt.Println("One or more of your inputs are invalid")
                }
                if !ok2 {
                    fmt.Println("One or more of your outputs are invalid")
                }
            }
            binary, err := trans.MarshalBinary()
            if err != nil {fmt.Println(err); continue }
            fmt.Println("Transaction Size:   ", len(binary))
            continue
        }
        
        switch strings.ToLower(v) {
            case "currentblock":
                fmt.Println(state.fs.GetCurrentBlock())
            case "rate":
                v, err := GetRate(state)
                if err != nil {fmt.Println(err); continue }
                fmt.Println("Factoids to buy one Entry Credit: ",
                            fct.ConvertDecimal(uint64(v)))
            case "height":  
                fmt.Println("Directory block height is: ",state.fs.GetDBHeight())
            default :
                fmt.Println("Unknown: ", v)
        }
    }
    
    return nil
}

func (Print) Name() string {
    return "Print"
}

func (Print) ShortHelp() string {
    return "Print <v1> <v2> ...  Prints the specified transaction(s) or the exchange rate."
}

func (Print) LongHelp() string {
    return `
Print <v1> <v2> ...                  Prints the specified values.  If <v> is a key for 
                                     a transaction, it will print said transaction.
      v = rate                       Print the number of factoids required to buy one
                                     one entry credit
`
}
