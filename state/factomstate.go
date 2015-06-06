// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Defines the state for simplecoin.  By using the proper
// interfaces, the functionality of simplecoin can be imported
// into any framework.
package state

import (
    "fmt"
    "time"
    "bytes"
    sc "github.com/FactomProject/simplecoin"
    "github.com/FactomProject/simplecoin/block"
    db "github.com/FactomProject/simplecoin/database"
)

var _ = time.Sleep

type IFactomState interface {
    // Set the database for the Coin State.  This is where
    // we manage the balances for transactions.  We also look
    // for previous blocks here.
    SetDB(db.ISCDatabase)       
    // The Exchange Rate for Entry Credits in Factoshis per
    // Entry Credits
    GetFactoshisPerEC() uint64
    SetFactoshisPerEC(uint64)
    // Update balance updates the balance for a Factoid address in
    // the database.  Note that we take an int64 to allow debits
    // as well as credits
    UpdateBalance(address sc.IAddress, amount int64)  error
    // Update balance updates the balance for an Entry Credit address 
    // in the database.  Note that we take an int64 to allow debits
    // as well as credits
    UpdateECBalance(address sc.IAddress, amount uint64)  error
    // Use Entry Credits, which lowers their balance
    UseECs(address sc.IAddress, amount uint64) error
    // Return the Factoid balance for an address
    GetBalance(address sc.IAddress) uint64
    // Return the Entry Credit balance for an address
    GetECBalance(address sc.IAddress) uint64
    // Add a transaction block.  Useful for catching up with the network.
    AddTransactionBlock(block.ISCBlock) error
    // Return the Factoid block with this hash.  If unknown, returns
    // a null.
    GetTransactionBlock(sc.IHash) block.ISCBlock
    // Put a Factoid block with this hash into the database.
    PutTransactionBlock(sc.IHash, block.ISCBlock) 
    // Time is something that can vary across multiple systems, and
    // must be controlled in order to build reliable, repeatable
    // tests.  Therefore, no node should directly querry system
    // time.  
    GetTimeNano() int64    // Count of nanoseconds from Jan 1,1970
    GetTime() int64        // Count of seconds from Jan 1, 1970
    // Validate transaction
    // Return true if the balance of an address covers each input
    Validate(sc.ITransaction) bool
    // Add a Transaction to the current block.  The transaction is
    // validated against the address balances, which must cover The
    // inputs.  Returns true if the transaction is added.
    AddTransaction(sc.ITransaction) bool
    // Process End of Minute.  
    ProcessEndOfMinute()
    // Process End of Block.
    ProcessEndOfBlock()
    // Get the current Directory Block Height
}

type FactomState struct {
    IFactomState
    database db.ISCDatabase
    factoshisPerEC uint64
    currentBlock block.ISCBlock
    dbheight uint32
}

var _ IFactomState = (*FactomState)(nil)

func(fs *FactomState) GetDBHeight() uint32 {
    return fs.dbheight
}

func(fs *FactomState) AddTransaction(trans sc.ITransaction) bool {
    if !fs.Validate(trans) { return false }
    for _,input := range trans.GetInputs() {
        fs.UpdateBalance(input.GetAddress(), - int64(input.GetAmount()))
    }
    for _,output := range trans.GetOutputs() {
        fs.UpdateBalance(output.GetAddress(), int64(output.GetAmount()))
    }
    for _,ecoutput := range trans.GetOutECs() {
        fs.UpdateECBalance(ecoutput.GetAddress(), ecoutput.GetAmount())
    }
    return true
}
 
func(fs *FactomState) ProcessEndOfMinute() {
}

// End of Block means packing the current block away, and setting 
// up the next block.
func(fs *FactomState) ProcessEndOfBlock(){
    fs.PutTransactionBlock(fs.currentBlock.GetHash(),fs.currentBlock)
    fs.PutTransactionBlock(sc.FACTOID_CHAINID_HASH,fs.currentBlock)
    fs.dbheight += 1
    fs.currentBlock = block.NewSCBlock(fs.GetFactoshisPerEC(),fs.dbheight)
}

// When we are playing catchup, adding the transaction block is a pretty
// usful feature.
func(fs *FactomState) AddTransactionBlock(blk block.ISCBlock) error  {
    transactions := blk.GetTransactions()
    for _,trans := range transactions {
        inputs := trans.GetInputs()
        for _, input := range inputs {
            amount  := input.GetAmount()
            err := fs.UpdateBalance(input.GetAddress(), -int64(amount))
            if err != nil { return err }
        }
        outputs := trans.GetOutputs()
        for _, output := range outputs {
            amount  := output.GetAmount()
            err := fs.UpdateBalance(output.GetAddress(), int64(amount))
            if err != nil { return err }
        }
        ecoutputs := trans.GetOutECs()
        for _, ecoutput := range ecoutputs {
            amount  := ecoutput.GetAmount()
            err := fs.UpdateBalance(ecoutput.GetAddress(), int64(amount))
            if err != nil { return err }
        }
    }
    return nil
}

func(fs *FactomState) LoadState() error  {
    var hashes []sc.IHash
    hashes = append(hashes, sc.NewHash(sc.FACTOID_CHAINID))
    iblk := fs.GetTransactionBlock(sc.NewHash(sc.FACTOID_CHAINID))
    var blk block.ISCBlock
    for {
        if iblk == nil {
            sc.Prtln("No Genesis Block detected.  Adding Genesis Block")
            gb := block.GetGenesisBlock(1000000,10,200000000000)
            gb.SetDBHeight(fs.dbheight)
            fs.PutTransactionBlock(gb.GetHash(),gb)
            fs.PutTransactionBlock(sc.FACTOID_CHAINID_HASH,gb)
            iblk = gb
        }
        blk, ok := iblk.(block.ISCBlock)
        if !ok {return fmt.Errorf("Block not found or not formated properly") }
        if bytes.Compare(blk.GetPrevBlock().Bytes(),sc.ZERO_HASH) == 0 { break }
        hashes = append(hashes, blk.GetHash())
        iblk = fs.GetTransactionBlock(blk.GetPrevBlock())
    }
    for i := len(hashes)-1; i>=0; i-- {
        iblk = fs.GetTransactionBlock(hashes[i])
        blk = iblk.(block.ISCBlock)
        err := fs.AddTransactionBlock(blk)
        if err != nil { return err }
    }
    return nil
}
        

func(fs *FactomState) Validate(trans sc.ITransaction) bool  {
    for _, input := range trans.GetInputs() {
        bal := fs.GetBalance(input.GetAddress())
        if input.GetAmount()>bal { return false }
    }
    return true;
}


func(fs *FactomState) GetFactoshisPerEC() uint64 {
    return fs.factoshisPerEC
}

func(fs *FactomState) SetFactoshisPerEC(factoshisPerEC uint64){
    fs.factoshisPerEC = factoshisPerEC
}

func(fs *FactomState) PutTransactionBlock(hash sc.IHash, trans block.ISCBlock) {
    fs.database.Put(sc.DB_FACTOID_BLOCKS, hash, trans)
}

func(fs *FactomState) GetTransactionBlock(hash sc.IHash) block.ISCBlock {
    trans := fs.database.Get(sc.DB_FACTOID_BLOCKS, hash)
    return trans.(block.ISCBlock)
}

func(fs *FactomState) GetTime64() int64 {
    return time.Now().UnixNano()
}

func(fs *FactomState) GetTime32() int64 {
    return time.Now().Unix()
}

func(fs *FactomState) SetDB(database db.ISCDatabase){
    fs.database = database
}

func(fs *FactomState) GetDB() db.ISCDatabase {
    return fs.database 
}

// Any address that is not defined has a zero balance.
func(fs *FactomState) GetBalance(address sc.IAddress) uint64 {
    balance := uint64(0)
    b  := fs.database.GetRaw([]byte(sc.DB_F_BALANCES),address.Bytes())
    if b != nil  {
        balance = b.(*FSbalance).number
    }
    return balance
}

// Update balance throws an error if your update will drive the balance negative.
func(fs *FactomState) UpdateBalance(address sc.IAddress, amount int64) error {
    nbalance := int64(fs.GetBalance(address))+amount
    if nbalance < 0 {return fmt.Errorf("New balance cannot be negative")}
    balance := uint64(nbalance)
    fs.database.PutRaw([]byte(sc.DB_F_BALANCES),address.Bytes(),&FSbalance{number: balance})
    return nil
} 


// Add to Entry Credit Balance.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits. Using lowers the Entry Credit Balance.
func(fs *FactomState) AddToECBalance(address sc.IAddress, amount uint64) error {
    ecs := amount/fs.GetFactoshisPerEC()
    balance := fs.GetBalance(address)+ecs
    fs.database.PutRaw([]byte(sc.DB_EC_BALANCES),address.Bytes(),&FSbalance{number: balance})
    return nil
}    
// Use Entry Credits.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits.  Using lowers the Entry Credit Balance.
func(fs *FactomState) UseECs(address sc.IAddress, amount uint64) error {
    balance := fs.GetBalance(address)-amount
    if balance < 0 { return fmt.Errorf("Overdraft of Entry Credits attempted.") }
    fs.database.PutRaw([]byte(sc.DB_EC_BALANCES),address.Bytes(),&FSbalance{number: balance})
    return nil
}      
    
// Any address that is not defined has a zero balance.
func(fs *FactomState) GetECBalance(address sc.IAddress) uint64 {
    balance := uint64(0)
    b  := fs.database.GetRaw([]byte(sc.DB_EC_BALANCES),address.Bytes())
    if b != nil  {
        balance = b.(*FSbalance).number
    }
    return balance
}
    
        
