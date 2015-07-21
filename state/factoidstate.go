// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Defines the state for factoid.  By using the proper
// interfaces, the functionality of factoid can be imported
// into any framework.
package state

import (
    "fmt"
    "time"
    "bytes"
    fct "github.com/FactomProject/factoid"
    "github.com/FactomProject/factoid/block"
    "github.com/FactomProject/factoid/wallet"
    db "github.com/FactomProject/factoid/database"
)

var _ = time.Sleep

type IFactoidState interface {
    // Set the database for the Coin State.  This is where
    // we manage the balances for transactions.  We also look
    // for previous blocks here.
    SetDB(db.IFDatabase)  
    GetDB() db.IFDatabase
    
    // Load the address state of Factoids
    LoadState() error 
    
    // Get the wallet used to help manage the Factoid State in
    // some applications.
    GetWallet() wallet.ISCWallet
    SetWallet(wallet.ISCWallet) 
    
    // The Exchange Rate for Entry Credits in Factoshis per
    // Entry Credits
    GetFactoshisPerEC() uint64
    SetFactoshisPerEC(uint64)
    
    // Get the current transaction block
    GetCurrentBlock() block.IFBlock
    
    // Update balance updates the balance for a Factoid address in
    // the database.  Note that we take an int64 to allow debits
    // as well as credits
    UpdateBalance(address fct.IAddress, amount int64)  error
    
    // Update balance updates the balance for an Entry Credit address 
    // in the database.  Note that we take an int64 to allow debits
    // as well as credits
    UpdateECBalance(address fct.IAddress, amount int64)  error
    
    // Use Entry Credits, which lowers their balance
    UseECs(address fct.IAddress, amount uint64) error
    
    // Return the Factoid balance for an address
    GetBalance(address fct.IAddress) uint64
    
    // Return the Entry Credit balance for an address
    GetECBalance(address fct.IAddress) uint64
    
    // Add a transaction block.  Useful for catching up with the network.
    AddTransactionBlock(block.IFBlock) error
    
    // Return the Factoid block with this hash.  If unknown, returns
    // a null.
    GetTransactionBlock(fct.IHash) block.IFBlock
    // Put a Factoid block with this hash into the database.
    PutTransactionBlock(fct.IHash, block.IFBlock) 
    
    // Time is something that can vary across multiple systems, and
    // must be controlled in order to build reliable, repeatable
    // tests.  Therefore, no node should directly querry system
    // time.  
    GetTimeMilli() uint64    // Count of milliseconds from Jan 1,1970
    GetTime() uint64         // Count of seconds from Jan 1, 1970
    
    // Validate transaction
    // Return zero len string if the balance of an address covers each input
    Validate(fct.ITransaction) error
    
    // Check the transaction timestamp for to ensure it can be included 
    // in the current block.  Transactions that are too old, or dated to
    // far in the future cannot be included in the current block
    ValidateTransactionAge(trans fct.ITransaction) error
    
    // Update Transaction just updates the balance sheet with the
    // addition of a transaction.
    UpdateTransaction(fct.ITransaction) error
    
    // Add a Transaction to the current block.  The transaction is
    // validated against the address balances, which must cover The
    // inputs.  Returns true if the transaction is added.
    AddTransaction(fct.ITransaction) error
    
    // Process End of Minute.  
    ProcessEndOfMinute()
    
    // Process End of Block.
    ProcessEndOfBlock() // to be replaced by ProcessEndOfBlock2
    ProcessEndOfBlock2(uint32)    
    
    // Get the current Directory Block Height
    GetDBHeight() uint32
    
    // Set the End of Period.  Currently, each block in Factom is broken
    // into ten, one minute periods.
    EndOfPeriod(period int)
}

type FactoidState struct {
    IFactoidState
    database db.IFDatabase
    factoshisPerEC uint64
    currentBlock block.IFBlock
    dbheight uint32
    wallet wallet.ISCWallet
}

var _ IFactoidState = (*FactoidState)(nil)

func(fs *FactoidState) EndOfPeriod(period int) {
    fs.GetCurrentBlock().EndOfPeriod(period)
}


func(fs *FactoidState) GetWallet() wallet.ISCWallet {
    return fs.wallet
}

func(fs *FactoidState) SetWallet(w wallet.ISCWallet) {
    fs.wallet = w
}

func(fs *FactoidState) GetCurrentBlock() block.IFBlock {
    return fs.currentBlock
}

func(fs *FactoidState) GetDBHeight() uint32 {
    return fs.dbheight
}

// When we are playing catchup, adding the transaction block is a pretty
// useful feature.
func(fs *FactoidState) AddTransactionBlock(blk block.IFBlock) error  {
    transactions := blk.GetTransactions()
    for _,trans := range transactions {
        err := fs.UpdateTransaction(trans)
        if err != nil {
            return err
        }
    }
    fs.currentBlock=blk
    fs.SetFactoshisPerEC(blk.GetExchRate())
    return nil
}

// Checks the transaction timestamp for validity in being included in the current block.
// No node has any responsiblity to forward on transactions that do not fall within
// the timeframe around a block defined by TRANSACTION_PRIOR_LIMIT and TRANSACTION_POST_LIMIT
func(fs *FactoidState) ValidateTransactionAge(trans fct.ITransaction) error {
    tsblk   := fs.GetCurrentBlock().GetCoinbaseTimestamp()
    if tsblk < 0  { return fmt.Errorf("Block has no coinbase transaction at this time") }
    
    tstrans := int64(trans.GetMilliTimestamp())
    
    if tsblk - tstrans  > fct.TRANSACTION_PRIOR_LIMIT {
        return fmt.Errorf("Transaction is too old to be included in the current block")
    }
    
    if tstrans - tsblk  > fct.TRANSACTION_POST_LIMIT {
        return fmt.Errorf("Transaction is dated too far in the future to be included in the current block")
    }
    return nil
}

// Only add valid transactions to the current block.
func(fs *FactoidState) AddTransaction(trans fct.ITransaction) error {
    if err := fs.Validate(trans);                     err != nil { return err }
    if err := fs.UpdateTransaction(trans);            err != nil { return err }
    if err := fs.ValidateTransactionAge(trans);       err != nil { return err }   
    if err := fs.currentBlock.AddTransaction(trans);  err != nil { return err }
    return nil
}

// Assumes validation has already been done.
func(fs *FactoidState) UpdateTransaction(trans fct.ITransaction) error {
    for _,input := range trans.GetInputs() {
        fs.UpdateBalance(input.GetAddress(), - int64(input.GetAmount()))
    }
    for _,output := range trans.GetOutputs() {
        fs.UpdateBalance(output.GetAddress(), int64(output.GetAmount()))
    }
    for _,ecoutput := range trans.GetECOutputs() {
        fs.UpdateECBalance(ecoutput.GetAddress(), int64(ecoutput.GetAmount()))
    }
    return nil
}
 
func(fs *FactoidState) ProcessEndOfMinute() {
}

// End of Block means packing the current block away, and setting 
// up the next block.
func(fs *FactoidState) ProcessEndOfBlock(){
    var hash,hash2 fct.IHash

    if fs.currentBlock != nil {             // If no blocks, the current block is nil
        hash = fs.currentBlock.GetHash()
        hash2 = fs.currentBlock.GetFullHash()
        fs.PutTransactionBlock(hash,fs.currentBlock)
        fs.PutTransactionBlock(fct.FACTOID_CHAINID_HASH,fs.currentBlock)
    }
    
    fs.dbheight += 1
    fs.currentBlock = block.NewFBlock(fs.GetFactoshisPerEC(),fs.dbheight)
 
    t := block.GetCoinbase(fs.GetTimeMilli())
    err := fs.currentBlock.AddCoinbase(t)
    if err !=nil {
        panic(err.Error())
    }
    fs.UpdateTransaction(t)
    
    if hash != nil {
        fs.currentBlock.SetPrevKeyMR(hash.Bytes())
        fs.currentBlock.SetPrevFullHash(hash2.Bytes())
    }
    
}

// End of Block means packing the current block away, and setting 
// up the next block.
// this function is to replace the existing function: ProcessEndOfBlock
func(fs *FactoidState) ProcessEndOfBlock2(nextBlkHeight uint32) {
    var hash,hash2 fct.IHash
    
    if fs.currentBlock != nil {             // If no blocks, the current block is nil
        hash  = fs.currentBlock.GetHash()
        hash2 = fs.currentBlock.GetFullHash()
    }
    
    fs.currentBlock = block.NewFBlock(fs.GetFactoshisPerEC(), nextBlkHeight)

    t := block.GetCoinbase(fs.GetTimeMilli())
    err := fs.currentBlock.AddCoinbase(t)    
    if err !=nil {
        panic(err.Error())
    }
    fs.UpdateTransaction(t)
    
    if hash != nil {
        fs.currentBlock.SetPrevKeyMR(hash.Bytes())
        fs.currentBlock.SetPrevFullHash(hash2.Bytes())
    }
    
}

func(fs *FactoidState) LoadState() error  {
    var hashes []fct.IHash
    blk := fs.GetTransactionBlock(fct.FACTOID_CHAINID_HASH)
    // If there is no head for the Factoids in the database, we have an
    // uninitialized database.  We need to add the Genesis Block. TODO
    if blk == nil {
        fct.Prtln("No Genesis Block for Factoids detected.  Adding Genesis Block")
        gb := block.GetGenesisBlock(fs.GetTimeMilli(), 1000000,10,200000000000)
        fs.PutTransactionBlock(gb.GetHash(),gb)
        fs.PutTransactionBlock(fct.FACTOID_CHAINID_HASH,gb)
        err := fs.AddTransactionBlock(gb)
        if err != nil { 
            fct.Prtln("Failed to build initial state.\n",err); 
            return err 
        }
        fs.dbheight = 0
        fs.ProcessEndOfBlock()
        return nil
    }
    // First run back from the head back to the genesis block, collecting hashes.
    for {
        if blk == nil {return fmt.Errorf("Block not found or not formated properly") }
        hashes = append(hashes, blk.GetHash())
        if bytes.Compare(blk.GetPrevKeyMR().Bytes(),fct.ZERO_HASH) == 0 { 
            break 
        }
        tblk := fs.GetTransactionBlock(blk.GetPrevKeyMR())
        if tblk.GetHash().IsEqual(blk.GetPrevKeyMR()) != nil {
            return fmt.Errorf("Hash Failure!  Database must be rebuilt")
        }
        blk = tblk
    }

    // Now run forward, and build our accounting
    for i := len(hashes)-1; i>=0; i-- {
        blk = fs.GetTransactionBlock(hashes[i])
        if blk == nil { 
            return fmt.Errorf("Should never happen.  Block not found in LoadState") 
        }
        fct.Prt("Loading from disk block: ", blk.GetDBHeight(),"\r")
        err := fs.AddTransactionBlock(blk)  // updates accounting for this block
        if err != nil { 
            fct.Prtln("Failed to rebuild state.\n",err); 
            return err 
        }
        if i%50 == 0 {
            time.Sleep(10000)
        }
    }
    fs.dbheight = blk.GetDBHeight()+1
    fs.currentBlock = block.NewFBlock(fs.GetFactoshisPerEC(),fs.dbheight)
    fs.currentBlock.SetPrevKeyMR(blk.GetHash().Bytes())
    return nil
}
    
// Returns an error message about what is wrong with the transaction if it is
// invalid, otherwise you are good to go.
func(fs *FactoidState) Validate(trans fct.ITransaction) error  {
    err := fs.currentBlock.ValidateTransaction(trans) 
    if err != nil { return err }

    var sums = make(map[fct.IAddress]uint64,10)
    for _, input := range trans.GetInputs() {
        bal,err := fct.ValidateAmounts(
            sums[input.GetAddress()],           // Will be zero the first time around 
            input.GetAmount())                  // Get this amount, check against bounds
        if err != nil { 
            return err
        }
        if bal > fs.GetBalance(input.GetAddress()) {
            return fmt.Errorf("Not enough funds in input addresses for the transaction")
        }
        sums[input.GetAddress()] = bal
    }
    return nil
}


func(fs *FactoidState) GetFactoshisPerEC() uint64 {
    return fs.factoshisPerEC
}

func(fs *FactoidState) SetFactoshisPerEC(factoshisPerEC uint64){
    fs.factoshisPerEC = factoshisPerEC
}

func(fs *FactoidState) PutTransactionBlock(hash fct.IHash, trans block.IFBlock) {
    fs.database.Put(fct.DB_FACTOID_BLOCKS, hash, trans)
}

func(fs *FactoidState) GetTransactionBlock(hash fct.IHash) block.IFBlock {
    transblk := fs.database.Get(fct.DB_FACTOID_BLOCKS, hash)
    if transblk == nil { return nil }
    return transblk.(block.IFBlock)
}

func(fs *FactoidState) GetTimeMilli() uint64 {
    return uint64(time.Now().UnixNano())/1000000  // 10^-9 >> 10^-3
}

func(fs *FactoidState) GetTime() uint64 {
    return uint64(time.Now().Unix())
}

func(fs *FactoidState) SetDB(database db.IFDatabase){
    fs.database = database
}

func(fs *FactoidState) GetDB() db.IFDatabase {
    return fs.database 
}

// Any address that is not defined has a zero balance.
func(fs *FactoidState) GetBalance(address fct.IAddress) uint64 {
    balance := uint64(0)
    b  := fs.database.GetRaw([]byte(fct.DB_F_BALANCES),address.Bytes())
    if b != nil  {
        balance = b.(*FSbalance).number
    }
    return balance
}

// Update balance throws an error if your update will drive the balance negative.
func(fs *FactoidState) UpdateBalance(address fct.IAddress, amount int64) error {
    nbalance := int64(fs.GetBalance(address))+amount
    if nbalance < 0 {return fmt.Errorf("New balance cannot be negative")}
    balance := uint64(nbalance)
    fs.database.PutRaw([]byte(fct.DB_F_BALANCES),address.Bytes(),&FSbalance{number: balance})
    return nil
} 

// Update ec balance throws an error if your update will drive the balance negative.
func(fs *FactoidState) UpdateECBalance(address fct.IAddress, amount int64) error {
    nbalance := int64(fs.GetBalance(address))+amount
    if nbalance < 0 {return fmt.Errorf("New balance cannot be negative")}
    balance := uint64(nbalance)
    fs.database.PutRaw([]byte(fct.DB_EC_BALANCES),address.Bytes(),&FSbalance{number: balance})
    return nil
} 

// Add to Entry Credit Balance.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits. Using lowers the Entry Credit Balance.
func(fs *FactoidState) AddToECBalance(address fct.IAddress, amount uint64) error {
    ecs := amount/fs.GetFactoshisPerEC()
    balance := fs.GetBalance(address)+ecs
    fs.database.PutRaw([]byte(fct.DB_EC_BALANCES),address.Bytes(),&FSbalance{number: balance})
    return nil
}    
// Use Entry Credits.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits.  Using lowers the Entry Credit Balance.
func(fs *FactoidState) UseECs(address fct.IAddress, amount uint64) error {
    balance := fs.GetBalance(address)-amount
    if balance < 0 { return fmt.Errorf("Overdraft of Entry Credits attempted.") }
    fs.database.PutRaw([]byte(fct.DB_EC_BALANCES),address.Bytes(),&FSbalance{number: balance})
    return nil
}      
    
// Any address that is not defined has a zero balance.
func(fs *FactoidState) GetECBalance(address fct.IAddress) uint64 {
    balance := uint64(0)
    b  := fs.database.GetRaw([]byte(fct.DB_EC_BALANCES),address.Bytes())
    if b != nil  {
        balance = b.(*FSbalance).number
    }
    return balance
}
    
        
