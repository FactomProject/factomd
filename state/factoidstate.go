// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Defines the state for factoid.  By using the proper
// interfaces, the functionality of factoid can be imported
// into any framework.
package state

import (
	"fmt"
	"runtime/debug"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = debug.PrintStack

var FACTOID_CHAINID_HASH = primitives.NewHash(constants.FACTOID_CHAINID)

type FactoidState struct {
	DBHeight     uint32
	State        *State
	CurrentBlock interfaces.IFBlock
	Wallet       interfaces.ISCWallet
}

var _ interfaces.IFactoidState = (*FactoidState)(nil)

func (fs *FactoidState) EndOfPeriod(period int) {
	if period > 9 || period < 0 {
		panic(fmt.Sprintf("Minute is out of range: %d", period))
	}
	fs.GetCurrentBlock().EndOfPeriod(period)
}

func (fs *FactoidState) GetWallet() interfaces.ISCWallet {
	return fs.Wallet
}

func (fs *FactoidState) SetWallet(w interfaces.ISCWallet) {
	fs.Wallet = w
}

func (fs *FactoidState) GetCurrentBlock() interfaces.IFBlock {
	if fs.CurrentBlock == nil {
		fs.CurrentBlock = factoid.NewFBlock(nil)
		fs.CurrentBlock.SetExchRate(fs.State.GetFactoshisPerEC())
		fs.CurrentBlock.SetDBHeight(fs.DBHeight)
		t := factoid.GetCoinbase(fs.State.GetLeaderTimestamp())
		err := fs.CurrentBlock.AddCoinbase(t)
		if err != nil {
			panic(err.Error())
		}
		fs.UpdateTransaction(true, t)
	}
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
		err := fs.UpdateTransaction(false, trans)
		if err != nil {
			return err
		}
	}
	fs.CurrentBlock = blk
	//fs.State.SetFactoshisPerEC(blk.GetExchRate())

	return nil
}

func (fs *FactoidState) AddECBlock(blk interfaces.IEntryCreditBlock) error {
	transactions := blk.GetBody().GetEntries()

	for _, trans := range transactions {
		err := fs.UpdateECTransaction(false, trans)
		if err != nil {
			return err
		}
	}

	return nil
}

// Checks the transaction timestamp for validity in being included in the current
// No node has any responsiblity to forward on transactions that do not fall within
// the timeframe around a block defined by TRANSACTION_PRIOR_LIMIT and TRANSACTION_POST_LIMIT
func (fs *FactoidState) ValidateTransactionAge(trans interfaces.ITransaction) error {
	tsblk := fs.GetCurrentBlock().GetCoinbaseTimestamp().GetTimeMilli()
	if tsblk < 0 {
		return fmt.Errorf("Block has no coinbase transaction at this time")
	}

	tstrans := trans.GetTimestamp().GetTimeMilli()

	if tsblk-tstrans > constants.TRANSACTION_PRIOR_LIMIT {
		return fmt.Errorf("Transaction is too old to be included in the current block")
	}

	if tstrans-tsblk > constants.TRANSACTION_POST_LIMIT {
		//	return fmt.Errorf("Transaction is dated too far in the future to be included in the current block")
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
	if err := fs.UpdateTransaction(true, trans); err != nil {
		return err
	}
	if err := fs.CurrentBlock.AddTransaction(trans); err != nil {
		if err == nil {
			// We assume validity has been done elsewhere.  We are maintaining the "seen" state of
			// all transactions here.
			fs.State.Replay.IsTSValid(constants.INTERNAL_REPLAY|constants.NETWORK_REPLAY, trans.GetHash(), trans.GetTimestamp())
			fs.State.Replay.IsTSValid(constants.NETWORK_REPLAY|constants.NETWORK_REPLAY, trans.GetHash(), trans.GetTimestamp())
		}
		return err
	}

	return nil
}

func (fs *FactoidState) GetFactoidBalance(address [32]byte) int64 {
	return fs.State.GetF(address)
}

func (fs *FactoidState) GetECBalance(address [32]byte) int64 {
	return fs.State.GetE(address)
}

func (fs *FactoidState) ResetBalances() {
	fs.State.FactoidBalancesP = map[[32]byte]int64{}
	fs.State.ECBalancesP = map[[32]byte]int64{}
	fs.State.FactoidBalancesT = map[[32]byte]int64{}
	fs.State.ECBalancesT = map[[32]byte]int64{}
	fs.State.NumTransactions = 0
}

func (fs *FactoidState) UpdateECTransaction(rt bool, trans interfaces.IECBlockEntry) error {

	switch trans.ECID() {
	case entryCreditBlock.ECIDServerIndexNumber:
		return nil

	case entryCreditBlock.ECIDMinuteNumber:
		return nil

	case entryCreditBlock.ECIDChainCommit:
		t := trans.(*entryCreditBlock.CommitChain)
		fs.State.PutE(rt, t.ECPubKey.Fixed(), fs.State.GetE(t.ECPubKey.Fixed())-int64(t.Credits))
		fs.State.NumTransactions++

	case entryCreditBlock.ECIDEntryCommit:
		t := trans.(*entryCreditBlock.CommitEntry)
		fs.State.PutE(rt, t.ECPubKey.Fixed(), fs.State.GetE(t.ECPubKey.Fixed())-int64(t.Credits))
		fs.State.NumTransactions++

	case entryCreditBlock.ECIDBalanceIncrease:
		t := trans.(*entryCreditBlock.IncreaseBalance)
		fs.State.PutE(rt, t.ECPubKey.Fixed(), fs.State.GetE(t.ECPubKey.Fixed())+int64(t.NumEC))
		fs.State.NumTransactions++

	default:
		return fmt.Errorf("Unknown EC Transaction")
	}

	return nil
}

// Assumes validation has already been done.
func (fs *FactoidState) UpdateTransaction(rt bool, trans interfaces.ITransaction) error {
	for _, input := range trans.GetInputs() {
		adr := input.GetAddress().Fixed()
		oldv := fs.State.GetF(adr)
		fs.State.PutF(rt, adr, oldv-int64(input.GetAmount()))
	}
	for _, output := range trans.GetOutputs() {
		adr := output.GetAddress().Fixed()
		oldv := fs.State.GetF(adr)
		fs.State.PutF(rt, adr, oldv+int64(output.GetAmount()))
	}
	for _, ecOut := range trans.GetECOutputs() {
		ecbal := int64(ecOut.GetAmount()) / int64(fs.State.FactoshisPerEC)
		fs.State.PutE(rt, ecOut.GetAddress().Fixed(), fs.State.GetE(ecOut.GetAddress().Fixed())+ecbal)
	}
	fs.State.NumTransactions++
	return nil
}

// Assumes validation has already been done.
func (fs *FactoidState) ClearRealTime() error {
	fs.State.FactoidBalancesT = map[[32]byte]int64{}
	fs.State.ECBalancesT = map[[32]byte]int64{}
	return nil
}

// End of Block means packing the current block away, and setting
// up the next
func (fs *FactoidState) ProcessEndOfBlock(state interfaces.IState) {
	if fs.GetCurrentBlock() == nil {
		panic("Invalid state on initialization")
	}

	// 	outstr := fs.CurrentBlock.String()
	// 	if len(outstr) < 10000 {
	//		if state.GetOut() {
	// 			fs.State.Println("888888888888888888  ",fs.State.GetFactomNodeName()," 8888888888888888888")
	// 			fs.State.Println(outstr)
	//		}
	// 	}

	fBlock := factoid.NewFBlock(fs.CurrentBlock)
	fBlock.SetExchRate(fs.State.GetFactoshisPerEC())

	fs.CurrentBlock = fBlock

	t := factoid.GetCoinbase(fs.State.GetLeaderTimestamp())
	err := fs.CurrentBlock.AddCoinbase(t)
	if err != nil {
		panic(err.Error())
	}
	fs.UpdateTransaction(true, t)

	// Monitor for changes in Identity
	dblk, _ := fs.State.DB.FetchDirectoryBlockHead()
	if dblk != nil {
		for _, dEntry := range dblk.GetDBEntries() {
			if isIdentityChain(dEntry.GetChainID(), fs.State.Identities) != -1 {
				eblk, err := fs.State.DB.FetchEBlock(dEntry.GetKeyMR())
				if err != nil {
					continue
				}
				LoadIdentityByEntryBlock(eblk, fs.State, true)
			}
		}
	}
	fs.DBHeight++
}

// Returns an error message about what is wrong with the transaction if it is
// invalid, otherwise you are good to go.
func (fs *FactoidState) Validate(index int, trans interfaces.ITransaction) error {

	var sums = make(map[[32]byte]uint64, 10)  // Look at the sum of an address's inputs
	for _, input := range trans.GetInputs() { //    to a transaction.
		bal, err := factoid.ValidateAmounts(sums[input.GetAddress().Fixed()], input.GetAmount())
		if err != nil {
			return err
		}
		if int64(bal) > fs.State.GetF(input.GetAddress().Fixed()) {
			return fmt.Errorf("Not enough funds in input addresses for the transaction")
		}
		sums[input.GetAddress().Fixed()] = bal
	}

	return nil
}
