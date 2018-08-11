// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Defines the state for factoid.  By using the proper
// interfaces, the functionality of factoid can be imported
// into any framework.
package state

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"runtime/debug"
	"sort"
	"time"

	"github.com/FactomProject/factomd/common/adminBlock"
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

type elementSortable []*element

func (slice elementSortable) Len() int {
	return len(slice)
}

func (slice elementSortable) Less(i, j int) bool {
	return bytes.Compare(slice[i].adr[:], slice[j].adr[:]) < 0
}

func (slice elementSortable) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type stringSortable []string

func (v stringSortable) Len() int           { return len(v) }
func (v stringSortable) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v stringSortable) Less(i, j int) bool { return v[i] < v[j] }

type element struct {
	adr [32]byte
	v   int64
}

func (fs *FactoidState) GetMultipleECBalances(singleAdd [32]byte) (uint32, uint32, int64, int64, string) {
	currentHeight := fs.DBHeight
	heighestSavedHeight := fs.State.GetHighestSavedBlk()
	errNotAcc := ""

	PermBalance, pok := fs.State.ECBalancesP[singleAdd] // Gets the Balance of the EC address

	if fs.State.ECBalancesPapi != nil {
		if savedBal, ok := fs.State.ECBalancesPapi[singleAdd]; ok {
			PermBalance = savedBal
		}
	}

	pl := fs.State.ProcessLists.Get(currentHeight)
	pl.ECBalancesTMutex.Lock()
	TempBalance, tok := pl.ECBalancesT[singleAdd] // Gets the Temp Balance of the EC address
	pl.ECBalancesTMutex.Unlock()

	if tok != true && pok != true {
		TempBalance = 0
		PermBalance = 0
		errNotAcc = "Address has not had a transaction"
	} else if tok == true && pok == false {
		PermBalance = 0
		errNotAcc = ""
	} else if tok == false && pok == true {
		TempBalance = PermBalance
	}

	if fs.State.IgnoreDone != true || fs.State.DBFinished != true {
		return 0, 0, 0, 0, "Not fully booted"
	}

	return currentHeight, heighestSavedHeight, TempBalance, PermBalance, errNotAcc
}

func (fs *FactoidState) GetMultipleFactoidBalances(singleAdd [32]byte) (uint32, uint32, int64, int64, string) {
	currentHeight := fs.DBHeight
	heighestSavedHeight := fs.State.GetHighestSavedBlk()
	errNotAcc := ""

	PermBalance, pok := fs.State.FactoidBalancesP[singleAdd] // Gets the Balance of the Factoid address

	if fs.State.FactoidBalancesPapi != nil {
		if savedBal, ok := fs.State.FactoidBalancesPapi[singleAdd]; ok {
			PermBalance = savedBal
		}
	}

	pl := fs.State.ProcessLists.Get(currentHeight)
	pl.FactoidBalancesTMutex.Lock()
	TempBalance, tok := pl.FactoidBalancesT[singleAdd] // Gets the Temp Balance of the Factoid address
	pl.FactoidBalancesTMutex.Unlock()

	if tok != true && pok != true {
		TempBalance = 0
		PermBalance = 0
		errNotAcc = "Address has not had a transaction"
	} else if tok == true && pok == false {
		PermBalance = 0
		errNotAcc = ""
	} else if tok == false && pok == true {
		TempBalance = PermBalance
	}

	if fs.State.IgnoreDone != true || fs.State.DBFinished != true {
		return 0, 0, 0, 0, "Not fully booted"
	}

	return currentHeight, heighestSavedHeight, TempBalance, PermBalance, errNotAcc
}

func (fs *FactoidState) GetFactiodAccounts(params interface{}) (uint32, []string) {
	name := fs.State.FactoidBalancesP
	height := fs.DBHeight
	list := make([]string, 0, len(name))

	for k, _ := range name {
		y := primitives.Hash(k)
		z := interfaces.IAddress(interfaces.IHash(&y))
		e := primitives.ConvertFctAddressToUserStr(z)
		list = append(list, e)
	}

	sort.Sort(stringSortable(list))

	return height, list
}

func GetMapHash(dbheight uint32, bmap map[[32]byte]int64) interfaces.IHash {
	list := make([]*element, 0, len(bmap))

	for k, v := range bmap {
		e := new(element)
		copy(e.adr[:], k[:])
		e.v = v
		list = append(list, e)
	}
	// GoLang > 1.8
	//sort.Slice(list, func(i, j int) bool { return bytes.Compare(list[i].adr[:], list[j].adr[:]) < 0 })
	// GoLang < 1.8
	sort.Sort(elementSortable(list))

	var buff primitives.Buffer
	if err := binary.Write(&buff, binary.BigEndian, &dbheight); err != nil {
		return nil
	}

	for _, e := range list {
		_, err := buff.Write(e.adr[:])
		if err != nil {
			return nil
		}
		if err := binary.Write(&buff, binary.BigEndian, &e.v); err != nil {
			return nil
		}
	}

	h := primitives.Sha(buff.Bytes())

	return h
}

func (fs *FactoidState) GetBalanceHash(includeTemp bool) interfaces.IHash {
	h1 := GetMapHash(fs.DBHeight, fs.State.FactoidBalancesP)
	h2 := GetMapHash(fs.DBHeight, fs.State.ECBalancesP)
	h3 := h1
	h4 := h2
	if includeTemp {
		pl := fs.State.ProcessLists.Get(fs.DBHeight)
		if pl == nil {
			return primitives.NewZeroHash()
		}
		pl.ECBalancesTMutex.Lock()
		pl.FactoidBalancesTMutex.Lock()
		h3 = GetMapHash(fs.DBHeight, pl.FactoidBalancesT)
		h4 = GetMapHash(fs.DBHeight, pl.ECBalancesT)
		pl.ECBalancesTMutex.Unlock()
		pl.FactoidBalancesTMutex.Unlock()
	}
	var b []byte
	b = append(b, h1.Bytes()...)
	b = append(b, h2.Bytes()...)
	if includeTemp {
		b = append(b, h3.Bytes()...)
		b = append(b, h4.Bytes()...)
	}
	r := primitives.Sha(b)
	// Debug aid for Balance Hashes
	// fmt.Printf("%8d %x\n", fs.DBHeight, r.Bytes()[:16])
	return r
}

// Reset this Factoid state to an empty state at a dbheight following the
// given dbstate.
func (fs *FactoidState) Reset(dbstate *DBState) {
	ht := dbstate.DirectoryBlock.GetHeader().GetDBHeight()
	if fs.DBHeight > ht+1 {
		fs.DBHeight = ht

		dbstate := fs.State.DBStates.Get(int(fs.DBHeight))

		fBlock := factoid.NewFBlock(dbstate.FactoidBlock)
		fBlock.SetExchRate(dbstate.FinalExchangeRate)

		fs.CurrentBlock = fBlock

		t := fs.GetCoinbaseTransaction(fs.CurrentBlock.GetDatabaseHeight(), dbstate.NextTimestamp)

		fs.State.FactoshisPerEC = dbstate.FinalExchangeRate
		fs.State.LeaderTimestamp = dbstate.NextTimestamp

		err := fs.CurrentBlock.AddCoinbase(t)
		if err != nil {
			panic(err.Error())
		}
		fs.UpdateTransaction(true, t)

		fs.DBHeight++
	}
}

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
		t := fs.GetCoinbaseTransaction(fs.CurrentBlock.GetDatabaseHeight(), fs.State.GetLeaderTimestamp())
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
// No node has any responsibility to forward on transactions that do not fall within
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
	if err := fs.CurrentBlock.AddTransaction(trans); err != nil {
		return err
	}
	if err := fs.UpdateTransaction(true, trans); err != nil {
		return err
	}
	return nil
}

func (fs *FactoidState) GetFactoidBalance(address [32]byte) int64 {
	return fs.State.GetF(true, address)
}

func (fs *FactoidState) GetECBalance(address [32]byte) int64 {
	return fs.State.GetE(true, address)
}

func (fs *FactoidState) UpdateECTransaction(rt bool, trans interfaces.IECBlockEntry) error {
	switch trans.ECID() {
	case constants.ECIDServerIndexNumber:
		return nil
	case constants.ECIDMinuteNumber:
		return nil
	case constants.ECIDBalanceIncrease:
		return nil
	case constants.ECIDChainCommit:
		t := trans.(*entryCreditBlock.CommitChain)
		v := fs.State.GetE(rt, t.ECPubKey.Fixed()) - int64(t.Credits)
		if (fs.DBHeight > 97886 || fs.State.GetNetworkID() != constants.MAIN_NETWORK_ID) && v < 0 {
			return fmt.Errorf("%29s dbht %d: Not enough ECs (%d) to cover a chain commit (%d)",
				fs.State.GetFactomNodeName(),
				fs.DBHeight,
				fs.State.GetE(rt, t.ECPubKey.Fixed()),
				t.Credits)
		}
		fs.State.PutE(rt, t.ECPubKey.Fixed(), v)
		fs.State.NumTransactions++
		fs.State.Replay.IsTSValid(constants.INTERNAL_REPLAY, t.GetSigHash(), t.GetTimestamp())
		fs.State.Replay.IsTSValid(constants.NETWORK_REPLAY, t.GetSigHash(), t.GetTimestamp())
	case constants.ECIDEntryCommit:
		t := trans.(*entryCreditBlock.CommitEntry)
		v := fs.State.GetE(rt, t.ECPubKey.Fixed()) - int64(t.Credits)
		if (fs.DBHeight > 97886 || fs.State.GetNetworkID() != constants.MAIN_NETWORK_ID) && v < 0 {
			return fmt.Errorf("%29s dbht %d: Not enough ECs (%d) to cover a entry commit (%d)",
				fs.State.GetFactomNodeName(),
				fs.DBHeight,
				fs.State.GetE(rt, t.ECPubKey.Fixed()),
				t.Credits)
		}
		fs.State.PutE(rt, t.ECPubKey.Fixed(), v)
		fs.State.NumTransactions++
		fs.State.Replay.IsTSValid(constants.INTERNAL_REPLAY, t.GetSigHash(), t.GetTimestamp())
		fs.State.Replay.IsTSValid(constants.NETWORK_REPLAY, t.GetSigHash(), t.GetTimestamp())
	default:
		return fmt.Errorf("Unknown EC Transaction")
	}

	return nil
}

// Assumes validation has already been done.
func (fs *FactoidState) UpdateTransaction(rt bool, trans interfaces.ITransaction) error {

	// First check all inputs are good.
	for _, input := range trans.GetInputs() {
		adr := input.GetAddress().Fixed()
		oldv := fs.State.GetF(rt, adr)
		v := oldv - int64(input.GetAmount())
		if v < 0 {
			return fmt.Errorf("%29s dbht %d: Not enough factoids (%d) to cover a transaction (%d)",
				fs.State.GetFactomNodeName(),
				fs.DBHeight,
				oldv,
				input.GetAmount())
		}
	}
	// Then update the state for all inputs.
	for _, input := range trans.GetInputs() {
		adr := input.GetAddress().Fixed()
		oldv := fs.State.GetF(rt, adr)
		v := oldv - int64(input.GetAmount())
		fs.State.PutF(rt, adr, v)
	}
	// Then log that the transaction has been seen and processed.
	fs.State.Replay.IsTSValid(constants.INTERNAL_REPLAY, trans.GetSigHash(), trans.GetTimestamp())
	fs.State.Replay.IsTSValid(constants.NETWORK_REPLAY, trans.GetSigHash(), trans.GetTimestamp())

	for _, output := range trans.GetOutputs() {
		adr := output.GetAddress().Fixed()
		oldv := fs.State.GetF(rt, adr)
		fs.State.PutF(rt, adr, oldv+int64(output.GetAmount()))
	}
	for _, ecOut := range trans.GetECOutputs() {
		ecbal := int64(ecOut.GetAmount()) / int64(fs.State.FactoshisPerEC)
		fs.State.PutE(rt, ecOut.GetAddress().Fixed(), fs.State.GetE(rt, ecOut.GetAddress().Fixed())+ecbal)
	}
	fs.State.NumTransactions++
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

	leaderTS := fs.State.GetLeaderTimestamp()

	t := fs.GetCoinbaseTransaction(fs.CurrentBlock.GetDatabaseHeight(), leaderTS)

	dbstate := fs.State.DBStates.Get(int(fs.DBHeight))
	if dbstate != nil {
		dbstate.FinalExchangeRate = fs.State.GetFactoshisPerEC()
		dbstate.NextTimestamp = leaderTS
	}

	err := fs.CurrentBlock.AddCoinbase(t)
	if err != nil {
		panic(err.Error())
	}
	fs.UpdateTransaction(true, t)

	fs.DBHeight++
	fs.State.CurrentBlockStartTime = time.Now().UnixNano()
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
		curbal := fs.State.GetF(true, input.GetAddress().Fixed())
		if int64(bal) > curbal {
			return fmt.Errorf("%20s DBHT %d %s %d %s %d %s",
				fs.State.GetFactomNodeName(),
				fs.DBHeight, "Not enough funds in input addresses (", bal,
				") to cover the transaction (", curbal, ")")
		}
		sums[input.GetAddress().Fixed()] = bal
	}

	return nil
}

func (fs *FactoidState) GetCoinbaseTransaction(dbheight uint32, ftime interfaces.Timestamp) interfaces.ITransaction {
	coinbase := new(factoid.Transaction)
	coinbase.SetTimestamp(ftime)

	// Coinbases only have outputs on payout blocks.
	//	Payout blocks are every n blocks, where n is the coinbase frequency
	if dbheight > constants.COINBASE_ACTIVATION && // Coinbase code must be above activation
		dbheight != 0 && // Does not affect gensis
		(dbheight%constants.COINBASE_PAYOUT_FREQUENCY == 0 || dbheight%constants.COINBASE_PAYOUT_FREQUENCY == 1) && // Frequency of payouts
		// Cannot payout before a declaration (cannot grab below height 0)
		dbheight > constants.COINBASE_DECLARATION+constants.COINBASE_PAYOUT_FREQUENCY {
		// Grab the admin block 1000 blocks earlier
		descriptorHeight := dbheight - constants.COINBASE_DECLARATION
		ablock, err := fs.State.DB.FetchABlockByHeight(descriptorHeight)
		if err != nil {
			panic(fmt.Sprintf("When creating coinbase, admin block at height %d could not be retrieved", descriptorHeight))
		}

		abe := ablock.FetchCoinbaseDescriptor()
		if abe != nil {
			desc := abe.(*adminBlock.CoinbaseDescriptor)
			// Before we go through the outputs, we need to check if we have any
			// cancellations pending.
			m := make(map[uint32]struct{}, 0)
			list, ok := fs.State.IdentityControl.CanceledCoinbaseOutputs[descriptorHeight]
			if ok {
				// No longer need this
				delete(fs.State.IdentityControl.CanceledCoinbaseOutputs, descriptorHeight)
			}

			// Map contains all cancelled indices
			for _, v := range list {
				m[v] = struct{}{}
			}

			for i, o := range desc.Outputs {
				// Only elements not in map are ok
				if _, ok := m[uint32(i)]; !ok {
					coinbase.AddOutput(o.GetAddress(), o.GetAmount())
				}
			}
		}
	}

	return coinbase
}
