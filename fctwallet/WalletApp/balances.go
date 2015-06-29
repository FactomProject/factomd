// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import(
    "fmt" 
    "regexp"
    "strconv"
    "bytes"
    "strings"
    "net/http"
    "io/ioutil"    
    "encoding/hex"
    "encoding/json"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/wallet"
)

/***********************************************
 * General Support Functions
 ***********************************************/


var badChar,_ = regexp.Compile("[^A-Za-z0-9_]")
var badHexChar,_ = regexp.Compile("[^A-Fa-f0-9]")

// Get the Factoshis per Entry Credit Rate
func GetRate(state IState) (int64, error) {
    str := fmt.Sprintf("http://%s/v1/factoid-get-fee/", state.GetServer())
    resp, err := http.Get(str)
    if err != nil {
        return 0, err
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return 0, err
    }
    resp.Body.Close()
    
    type x struct { Fee int64 }
    b := new(x)
    if err = json.Unmarshal(body, b); err != nil {
        return 0, err
    }
        
    return b.Fee, nil
    
}   


func FctBalance(state State, adr string) (int64, error) {
    
    if !fct.ValidateFUserStr(adr) {
        if len(adr) != 64 {
            if len(adr)>32 {
                return 0, fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
            
            we := state.fs.GetDB().GetRaw([]byte(fct.W_NAME),[]byte(adr))
            
            if (we != nil){
                we2 := we.(wallet.IWalletEntry)
                addr,_ := we2.GetAddress()
                adr = hex.EncodeToString(addr.Bytes())
            }else{
                return 0, fmt.Errorf("Name is undefined.")
            }
        }else {
            if badHexChar.FindStringIndex(adr)!=nil {
                return 0, fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
        }
    } else {
        baddr := fct.ConvertUserStrToAddress(adr)
        adr = hex.EncodeToString(baddr)
    }
    
   
    
    str := fmt.Sprintf("http://%s/v1/factoid-balance/%s", state.GetServer(), adr)
    resp, err := http.Get(str)
    if err != nil {
        fmt.Println("\n",str)
        return 0, fmt.Errorf("Communication Error with Factom Client")
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return 0, fmt.Errorf("Read Error with Factom Client")
    }
    resp.Body.Close()
    
    type Balance struct { Balance int64 }
    b := new(Balance)
    if err := json.Unmarshal(body, b); err != nil {
        return 0, fmt.Errorf("Parsing Error on data returned by Factom Client")
    }
    
    return b.Balance, nil
    
}

func ECBalance(state State, adr string) (int64, error) {
    
    if !fct.ValidateECUserStr(adr) {
        if len(adr) != 64 {
            if len(adr)>32 {
                return 0, fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
            
            we := state.fs.GetDB().GetRaw([]byte(fct.W_NAME),[]byte(adr))
            
            if (we != nil){
                we2 := we.(wallet.IWalletEntry)
                addr,_ := we2.GetAddress()
                adr = hex.EncodeToString(addr.Bytes())
            }else{
                return 0, fmt.Errorf("Name is undefined.")
            }
        }else {
            if badHexChar.FindStringIndex(adr)!=nil {
                return 0, fmt.Errorf("Invalid Name.  Name is too long: %v characters",len(adr))
            }
        }
    } else {
        baddr := fct.ConvertUserStrToAddress(adr)
        adr = hex.EncodeToString(baddr)
    }
    
    
    str := fmt.Sprintf("http://%s/v1/entry-credit-balance/%s", state.GetServer(), adr)
    resp, err := http.Get(str)
    if err != nil {
        return 0, fmt.Errorf("Communication Error with Factom Client")
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return 0, fmt.Errorf("Read Error with Factom Client")
    }
    resp.Body.Close()
    
    type Balance struct { Balance int64 }
    b := new(Balance)
    if err := json.Unmarshal(body, b); err != nil {
        fmt.Println("-4::",err)
        return 0, fmt.Errorf("Parsing Error on data returned by Factom Client")
    }
    
    return b.Balance, nil
}

func GetBalances(state State) [] byte{
    keys, values := state.fs.GetDB().GetKeysValues([]byte(fct.W_NAME))
    
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
            bal,_ := ECBalance(state, adr)
            ecBalances = append(ecBalances,strconv.FormatInt(bal,10))
        }else{
            address, err := we.GetAddress()
            if err != nil { continue }
            adr = fct.ConvertFctAddressToUserStr(address)
            fctAddresses = append(fctAddresses,adr)
            fctKeys = append(fctKeys, string(k))
            bal,_ := FctBalance(state, adr)
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

/*************************************************************
 * Balance
 *************************************************************/

type Balance struct {
    ICommand
}

func (Balance) Execute (state State, args []string) (err error) {
    if len(args) != 3 {
        return fmt.Errorf("Wrong number of parameters")
    }
    
    var bal int64
    switch strings.ToLower(args[1]) {
        case "ec" :
            bal, err = ECBalance(state, strings.ToLower(args[2]))
        case "fct" :
            bal, err = FctBalance(state, strings.ToLower(args[2]))
        default :
            return fmt.Errorf("Invalid parameters")
    }
    if err != nil {
        return err
    }
    fmt.Println(args[2],"=",fct.ConvertDecimal(uint64(bal)))
    return nil
}
    
func (Balance) Name() string {
    return "Balance"
}

func (Balance) ShortHelp() string {
    return "Balance <ec|fct> <name|address> -- Returns the Entry Credits or Factoids at the\n"+
           "                                   specified name or address"
}

func (Balance) LongHelp() string {
    return `
Balance <ec|fct> <name|address>     ec      -- an Entry Credit address balance
                                    fct     -- a Factoid address balance
                                    name    -- Look up address by its name
                                    address -- specify the address directly
`
}

/*************************************************************
 * New Address
 *************************************************************/

type NewAddress struct {
    ICommand
}

func (NewAddress) Execute (state State, args []string) (err error) {
    
   
    
    if len(args) != 3 { 
        return fmt.Errorf("Incorrect Number of Arguments")
    }
    if len(args[2])>32 {
        return fmt.Errorf("Name of address is too long")
    }
    if badChar.FindStringIndex(args[2])!=nil {
        return fmt.Errorf("Invalid name. Names must be alphanumeric or underscores")
    }
    
    var adr fct.IAddress
    switch strings.ToLower(args[1]) {
        case "ec":
            adr, err = state.fs.GetWallet().GenerateECAddress([]byte(args[2]))
            if err != nil { return err}
            fmt.Println(args[2],"=",fct.ConvertECAddressToUserStr(adr))
        case "fct":
            adr, err = state.fs.GetWallet().GenerateFctAddress([]byte(args[2]),1,1)
            if err != nil { return err}
            fmt.Println(args[2],"=",fct.ConvertFctAddressToUserStr(adr))
        default:
            return fmt.Errorf("Invalid Parameters")
    }

    return nil
}

func (NewAddress) Name() string {
    return "NewAddress"
}

func (NewAddress) ShortHelp() string {
    return "NewAddress <ec|fct> <name> -- Returns a new Entry Credit or Factoid address"
}

func (NewAddress) LongHelp() string {
    return `
NewAddress <ec|fct> <name>          <ec>   Generates a new Entry Credit Address and
                                           saves it with the given <name>
                                    <fct>  Generates a new Factom Address and saves it
                                           with the given <name>
                                    <name> Names must be made up of alphanumeric 
                                           characters or underscores.  They cannot be
                                           more than 32 characters in length.  
`
}

/*************************************************************
 * Balances
 *************************************************************/

type Balances struct {
    ICommand
}

func (Balances) Execute (state State, args []string) (err error) {
    
    if len(args) != 1 { 
        return fmt.Errorf("Balances takes no arguments")
    }
    
    fmt.Println(string(GetBalances(state)))  
    return nil
}

func (Balances) Name() string {
    return "Balances"
}

func (Balances) ShortHelp() string {
    return "Balances -- Returns the balances of the Factoids and Entry Credits"+
           "            in this wallet, or tracked by this wallet."
}

func (Balances) LongHelp() string {
    return `
Balances                            Returns the Factoid and Entry Credit names
                                    and balances for the address in or tracked by
                                    this wallet.
`
}
