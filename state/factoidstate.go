// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Defines the state for factoid.  By using the proper
// interfaces, the functionality of factoid can be imported
// into any framework.
package state

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/factoid/block"
	"github.com/FactomProject/factomd/common/factoid/block/coinbase"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var FACTOID_CHAINID_HASH = primitives.NewHash(constants.FACTOID_CHAINID)

type FactoidState struct {
	State           interfaces.IState
	FactoshisPerEC  uint64
	CurrentBlock    interfaces.IFBlock
	Wallet          interfaces.ISCWallet
	NumTransactions int
	FctBalances     map[[32]byte]int64
	ECBalances      map[[32]byte]int64
}

var _ interfaces.IFactoidState = (*FactoidState)(nil)

func (fs *FactoidState) EndOfPeriod(period int) {
	fs.GetCurrentBlock().EndOfPeriod(period)
}

func (fs *FactoidState) GetWallet() interfaces.ISCWallet {
	return fs.Wallet
}

func (fs *FactoidState) SetWallet(w interfaces.ISCWallet) {
	fs.Wallet = w
}

func (fs *FactoidState) GetCurrentBlock() interfaces.IFBlock {
	return fs.CurrentBlock
}

// When we are playing catchup, adding the transaction block is a pretty
// useful feature.
func (fs *FactoidState) AddTransactionBlock(blk interfaces.IFBlock) error {

	if err := blk.Validate(); err != nil {
		return err
	}

	transactions := blk.GetTransactions()
	for _, trans := range transactions {
		err := fs.UpdateTransaction(trans)
		if err != nil {
			return err
		}
	}
	fs.CurrentBlock = blk
	fs.SetFactoshisPerEC(blk.GetExchRate())

	return nil
}

// Checks the transaction timestamp for validity in being included in the current
// No node has any responsiblity to forward on transactions that do not fall within
// the timeframe around a block defined by TRANSACTION_PRIOR_LIMIT and TRANSACTION_POST_LIMIT
func (fs *FactoidState) ValidateTransactionAge(trans interfaces.ITransaction) error {
	tsblk := fs.GetCurrentBlock().GetCoinbaseTimestamp()
	if tsblk < 0 {
		return fmt.Errorf("Block has no coinbase transaction at this time")
	}

	tstrans := int64(trans.GetMilliTimestamp())

	if tsblk-tstrans > constants.TRANSACTION_PRIOR_LIMIT {
		return fmt.Errorf("Transaction is too old to be included in the current block")
	}

	if tstrans-tsblk > constants.TRANSACTION_POST_LIMIT {
		return fmt.Errorf("Transaction is dated too far in the future to be included in the current block")
	}
	return nil
}

// Only add valid transactions to the current
func (fs *FactoidState) AddTransaction(index int, trans interfaces.ITransaction) error {
	if err := fs.Validate(index, trans); err != nil {
		return err
	}
	if err := fs.ValidateTransactionAge(trans); err != nil {
		return err
	}
	if err := fs.UpdateTransaction(trans); err != nil {
		return err
	}
	if err := fs.CurrentBlock.AddTransaction(trans); err != nil {
		return err
	}

	return nil
}

func (fs *FactoidState) GetFctBalance(address [32]byte) int64 {
	v := fs.FctBalances[address]
	return v
}

func (fs *FactoidState) GetECBalance(address [32]byte) int64 {
	v := fs.ECBalances[address]
	return v
}

// Assumes validation has already been done.
func (fs *FactoidState) UpdateTransaction(trans interfaces.ITransaction) error {
	if fs.FctBalances == nil {
		fs.FctBalances = map[[32]byte]int64{}
		fs.ECBalances  = map[[32]byte]int64{}
	}
	for _, input := range trans.GetInputs() {
		fs.FctBalances[input.GetAddress().Fixed()] = fs.FctBalances[input.GetAddress().Fixed()] - int64(input.GetAmount())
	}
	for _, output := range trans.GetOutputs() {
		fs.FctBalances[output.GetAddress().Fixed()] = fs.FctBalances[output.GetAddress().Fixed()] + int64(output.GetAmount())
	}
	for _, ecoutput := range trans.GetECOutputs() {
		ecs := int64(ecoutput.GetAmount())/int64(fs.FactoshisPerEC)
		fs.ECBalances[ecoutput.GetAddress().Fixed()] = fs.ECBalances[ecoutput.GetAddress().Fixed()] + ecs
	}
	
	fs.NumTransactions++

	return nil
}

// End of Block means packing the current block away, and setting
// up the next
func (fs *FactoidState) ProcessEndOfBlock(state interfaces.IState) {
	var hash, hash2 interfaces.IHash

	if fs.GetCurrentBlock() == nil {
		panic("Invalid state on initialization")
	}

	hash = fs.CurrentBlock.GetHash()
	hash2 = fs.CurrentBlock.GetLedgerKeyMR()

	state.GetCurrentDirectoryBlock().GetDBEntries()[2].SetKeyMR(hash)

	if err := state.GetDB().SaveFactoidBlockHead(fs.CurrentBlock); err != nil {
		panic(err)
	}

	state.SetPrevFactoidKeyMR(hash)

	fs.CurrentBlock = block.NewFBlock(fs.GetFactoshisPerEC(), state.GetDBHeight()+1)

	t := coinbase.GetCoinbase(primitives.GetTimeMilli())
	err := fs.CurrentBlock.AddCoinbase(t)
	if err != nil {
		panic(err.Error())
	}
	fs.UpdateTransaction(t)

	if hash != nil {
		fs.CurrentBlock.SetPrevKeyMR(hash.Bytes())
		fs.CurrentBlock.SetPrevLedgerKeyMR(hash2.Bytes())
	}

}

// Returns an error message about what is wrong with the transaction if it is
// invalid, otherwise you are good to go.
func (fs *FactoidState) Validate(index int, trans interfaces.ITransaction) error {
	err := fs.CurrentBlock.ValidateTransaction(index, trans)
	if err != nil {
		return err
	}

	var sums = make(map[[32]byte]uint64, 10)  // Look at the sum of an address's inputs
	for _, input := range trans.GetInputs() { //    to a transaction.
		bal, err := factoid.ValidateAmounts(sums[input.GetAddress().Fixed()], input.GetAmount())
		if err != nil {
			return err
		}
		if int64(bal) > fs.FctBalances[input.GetAddress().Fixed()] {
			return fmt.Errorf("Not enough funds in input addresses for the transaction")
		}
		sums[input.GetAddress().Fixed()] = bal
	}
	return nil
}

func (fs *FactoidState) GetFactoshisPerEC() uint64 {
	return fs.FactoshisPerEC
}

func (fs *FactoidState) SetFactoshisPerEC(factoshisPerEC uint64) {
	fs.FactoshisPerEC = factoshisPerEC
}
