// Copyright (c) 2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package btcd

import (
	"container/heap"
	"container/list"
	//	"crypto/rand"
	//	"fmt"
	"time"

	"github.com/FactomProject/btcd/blockchain"
	"github.com/FactomProject/btcd/database"
	//	"github.com/FactomProject/btcd/txscript"
	"github.com/FactomProject/btcd/wire"
	"github.com/FactomProject/btcutil"

	"github.com/FactomProject/FactomCode/util"
)

const (
	// generatedBlockVersion is the version of the block being generated.
	// It is defined as a constant here rather than using the
	// wire.BlockVersion constant since a change in the block version
	// will require changes to the generated block.  Using the wire constant
	// for generated block version could allow creation of invalid blocks
	// for the updated version.
	generatedBlockVersion = 2

	// minHighPriority is the minimum priority value that allows a
	// transaction to be considered high priority.
	minHighPriority = btcutil.SatoshiPerBitcoin * 144.0 / 250

	// blockHeaderOverhead is the max number of bytes it takes to serialize
	// a block header and max possible transaction count.
	blockHeaderOverhead = wire.MaxBlockHeaderPayload + wire.MaxVarIntPayload

	// coinbaseFlags is added to the coinbase script of a generated block
	// and is used to monitor BIP16 support as well as blocks that are
	// generated via btcd.
	coinbaseFlags = "/P2SH/btcd/"

	// standardScriptVerifyFlags are the script flags which are used when
	// executing transaction scripts to enforce additional checks which
	// are required for the script to be considered standard.  These checks
	// help reduce issues related to transaction malleability as well as
	// allow pay-to-script hash transactions.  Note these flags are
	// different than what is required for the consensus rules in that they
	// are more strict.
	/*
		standardScriptVerifyFlags = txscript.ScriptBip16 |
			txscript.ScriptCanonicalSignatures |
			txscript.ScriptStrictMultiSig |
			txscript.ScriptDiscourageUpgradableNops
	*/
)

// txPrioItem houses a transaction along with extra information that allows the
// transaction to be prioritized and track dependencies on other transactions
// which have not been mined into a block yet.
type txPrioItem struct {
	tx       *btcutil.Tx
	fee      int64
	priority float64
	feePerKB float64

	// dependsOn holds a map of transaction hashes which this one depends
	// on.  It will only be set when the transaction references other
	// transactions in the memory pool and hence must come after them in
	// a block.
	dependsOn map[wire.ShaHash]struct{}
}

// txPriorityQueueLessFunc describes a function that can be used as a compare
// function for a transaction priority queue (txPriorityQueue).
type txPriorityQueueLessFunc func(*txPriorityQueue, int, int) bool

// txPriorityQueue implements a priority queue of txPrioItem elements that
// supports an arbitrary compare function as defined by txPriorityQueueLessFunc.
type txPriorityQueue struct {
	lessFunc txPriorityQueueLessFunc
	items    []*txPrioItem
}

// Len returns the number of items in the priority queue.  It is part of the
// heap.Interface implementation.
func (pq *txPriorityQueue) Len() int {
	return len(pq.items)
}

// Less returns whether the item in the priority queue with index i should sort
// before the item with index j by deferring to the assigned less function.  It
// is part of the heap.Interface implementation.
func (pq *txPriorityQueue) Less(i, j int) bool {
	return pq.lessFunc(pq, i, j)
}

// Swap swaps the items at the passed indices in the priority queue.  It is
// part of the heap.Interface implementation.
func (pq *txPriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

// Push pushes the passed item onto the priority queue.  It is part of the
// heap.Interface implementation.
func (pq *txPriorityQueue) Push(x interface{}) {
	pq.items = append(pq.items, x.(*txPrioItem))
}

// Pop removes the highest priority item (according to Less) from the priority
// queue and returns it.  It is part of the heap.Interface implementation.
func (pq *txPriorityQueue) Pop() interface{} {
	n := len(pq.items)
	item := pq.items[n-1]
	pq.items[n-1] = nil
	pq.items = pq.items[0 : n-1]
	return item
}

// SetLessFunc sets the compare function for the priority queue to the provided
// function.  It also invokes heap.Init on the priority queue using the new
// function so it can immediately be used with heap.Push/Pop.
func (pq *txPriorityQueue) SetLessFunc(lessFunc txPriorityQueueLessFunc) {
	pq.lessFunc = lessFunc
	heap.Init(pq)
}

// txPQByPriority sorts a txPriorityQueue by transaction priority and then fees
// per kilobyte.
func txPQByPriority(pq *txPriorityQueue, i, j int) bool {
	// Using > here so that pop gives the highest priority item as opposed
	// to the lowest.  Sort by priority first, then fee.
	if pq.items[i].priority == pq.items[j].priority {
		return pq.items[i].feePerKB > pq.items[j].feePerKB
	}
	return pq.items[i].priority > pq.items[j].priority

}

// txPQByFee sorts a txPriorityQueue by fees per kilobyte and then transaction
// priority.
func txPQByFee(pq *txPriorityQueue, i, j int) bool {
	// Using > here so that pop gives the highest fee item as opposed
	// to the lowest.  Sort by fee first, then priority.
	if pq.items[i].feePerKB == pq.items[j].feePerKB {
		return pq.items[i].priority > pq.items[j].priority
	}
	return pq.items[i].feePerKB > pq.items[j].feePerKB
}

// newTxPriorityQueue returns a new transaction priority queue that reserves the
// passed amount of space for the elements.  The new priority queue uses either
// the txPQByPriority or the txPQByFee compare function depending on the
// sortByFee parameter and is already initialized for use with heap.Push/Pop.
// The priority queue can grow larger than the reserved space, but extra copies
// of the underlying array can be avoided by reserving a sane value.
func newTxPriorityQueue(reserve int, sortByFee bool) *txPriorityQueue {
	pq := &txPriorityQueue{
		items: make([]*txPrioItem, 0, reserve),
	}
	if sortByFee {
		pq.SetLessFunc(txPQByFee)
	} else {
		pq.SetLessFunc(txPQByPriority)
	}
	return pq
}

// BlockTemplate houses a block that has yet to be solved along with additional
// details about the fees and the number of signature operations for each
// transaction in the block.
type BlockTemplate struct {
	block           *wire.MsgBlock
	fees            []int64
	sigOpCounts     []int64
	height          int64
	validPayAddress bool
}

// minInt is a helper function to return the minimum of two ints.  This avoids
// a math import and the need to cast to floats.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// mergeTxStore adds all of the transactions in txStoreB to txStoreA.  The
// result is that txStoreA will contain all of its original transactions plus
// all of the transactions in txStoreB.
func mergeTxStore(txStoreA blockchain.TxStore, txStoreB blockchain.TxStore) {
	for hash, txDataB := range txStoreB {
		if txDataA, exists := txStoreA[hash]; !exists ||
			(txDataA.Err == database.ErrTxShaMissing &&
				txDataB.Err != database.ErrTxShaMissing) {

			txStoreA[hash] = txDataB
		}
	}
}

/*
// standardCoinbaseScript returns a standard script suitable for use as the
// signature script of the coinbase transaction of a new block.  In particular,
// it starts with the block height that is required by version 2 blocks and adds
// the extra nonce as well as additional coinbase flags.
func standardCoinbaseScript(nextBlockHeight int64, extraNonce uint64) ([]byte, error) {
	return txscript.NewScriptBuilder().AddInt64(nextBlockHeight).
		AddUint64(extraNonce).AddData([]byte(coinbaseFlags)).Script()
}
*/

// createCoinbaseTx returns a coinbase transaction paying an appropriate subsidy
// based on the passed block height to the provided address.  When the address
// is nil, the coinbase transaction will instead be redeemable by anyone.
//
// See the comment for NewBlockTemplate for more information about why the nil
// address handling is useful.
func createCoinbaseTx(nonce uint32, addr wire.RCDHash) (*btcutil.Tx, error) {

	util.Trace()

	tx := wire.NewMsgTx()

	tx.AddTxOut(&wire.TxOut{
		Value:   MaxPayoutAmount,
		RCDHash: addr,
	})

	/* TEST ONLY: add an Entry Credit payout
	tx.AddECOut(&wire.TxEntryCreditOut{
		Value:    12345,
		ECpubkey: wire.ECPubKey{},
	})
	*/

	/*
		tx.AddTxIn(&wire.TxIn{
			// Coinbase transactions have no inputs, so previous outpoint is
			// zero hash and max index.
			PreviousOutPoint: *wire.NewOutPoint(&wire.ShaHash{},
				uint32(nextBlockHeight)),
		})

		randHashBytes := make([]byte, wire.HashSize)
		nn, err0 := rand.Read(randHashBytes)
		fmt.Println(nn, err0, randHashBytes)

		newsha, _ := wire.NewShaHash(randHashBytes)
	*/

	tx.AddTxIn(&wire.TxIn{
		// Coinbase transactions have no inputs, so previous outpoint is
		// zero hash and max index.
		//		PreviousOutPoint: *wire.NewOutPoint(&wire.ShaHash{},
		//		PreviousOutPoint: *wire.NewOutPoint(newsha,
		//			wire.MaxPrevOutIndex),
		PreviousOutPoint: *wire.NewOutPoint(&wire.ShaHash{},
			nonce),
	})

	/*
		randBytes := make([]byte, wire.PubKeySize)
		n, err := rand.Read(randBytes)
		fmt.Println("randBytes: ", n, err, randBytes)

		//	var pubkeys []wire.PubKey
		pubkeys := make([]wire.PubKey, 1)

		copy(pubkeys[0][:], randBytes)

		tx.AddRCD(&wire.RCDreveal{PubKey: pubkeys})
	*/

	return btcutil.NewTx(tx), nil
}

/*
// calcPriority returns a transaction priority given a transaction and the sum
// of each of its input values multiplied by their age (# of confirmations).
// Thus, the final formula for the priority is:
// sum(inputValue * inputAge) / adjustedTxSize
func calcPriority(tx *btcutil.Tx, serializedTxSize int, inputValueAge float64) float64 {
	// In order to encourage spending multiple old unspent transaction
	// outputs thereby reducing the total set, don't count the constant
	// overhead for each input as well as enough bytes of the signature
	// script to cover a pay-to-script-hash redemption with a compressed
	// pubkey.  This makes additional inputs free by boosting the priority
	// of the transaction accordingly.  No more incentive is given to avoid
	// encouraging gaming future transactions through the use of junk
	// outputs.  This is the same logic used in the reference
	// implementation.
	//
	// The constant overhead for a txin is 41 bytes since the previous
	// outpoint is 36 bytes + 4 bytes for the sequence + 1 byte the
	// signature script length.
	//
	// A compressed pubkey pay-to-script-hash redemption with a maximum len
	// signature is of the form:
	// [OP_DATA_73 <73-byte sig> + OP_DATA_35 + {OP_DATA_33
	// <33 byte compresed pubkey> + OP_CHECKSIG}]
	//
	// Thus 1 + 73 + 1 + 1 + 33 + 1 = 110
	overhead := 0
	for _, txIn := range tx.MsgTx().TxIn {
		// Max inputs + size can't possibly overflow here.
		overhead += 41 + minInt(110, len(txIn.SignatureScript))
	}

	if overhead >= serializedTxSize {
		return 0.0
	}

	return inputValueAge / float64(serializedTxSize-overhead)
}
*/

// spendTransaction updates the passed transaction store by marking the inputs
// to the passed transaction as spent.  It also adds the passed transaction to
// the store at the provided height.
func spendTransaction(txStore blockchain.TxStore, tx *btcutil.Tx, height int64) error {
	for _, txIn := range tx.MsgTx().TxIn {
		originHash := &txIn.PreviousOutPoint.Hash
		originIndex := txIn.PreviousOutPoint.Index
		if originTx, exists := txStore[*originHash]; exists {
			originTx.Spent[originIndex] = true
		}
	}

	txStore[*tx.Sha()] = &blockchain.TxData{
		Tx:          tx,
		Hash:        tx.Sha(),
		BlockHeight: height,
		Spent:       make([]bool, len(tx.MsgTx().TxOut)),
		Err:         nil,
	}

	return nil
}

// logSkippedDeps logs any dependencies which are also skipped as a result of
// skipping a transaction while generating a block template at the trace level.
func logSkippedDeps(tx *btcutil.Tx, deps *list.List) {
	if deps == nil {
		return
	}

	for e := deps.Front(); e != nil; e = e.Next() {
		item := e.Value.(*txPrioItem)
		minrLog.Tracef("Skipping tx %s since it depends on %s\n",
			item.tx.Sha(), tx.Sha())
	}
}

// minimumMedianTime returns the minimum allowed timestamp for a block building
// on the end of the current best chain.  In particular, it is one second after
// the median timestamp of the last several blocks per the chain consensus
// rules.
func minimumMedianTime(chainState *chainState) (time.Time, error) {
	chainState.Lock()
	defer chainState.Unlock()
	if chainState.pastMedianTimeErr != nil {
		return time.Time{}, chainState.pastMedianTimeErr
	}

	return chainState.pastMedianTime.Add(time.Second), nil
}

// medianAdjustedTime returns the current time adjusted to ensure it is at least
// one second after the median timestamp of the last several blocks per the
// chain consensus rules.
func medianAdjustedTime(chainState *chainState) (time.Time, error) {
	chainState.Lock()
	defer chainState.Unlock()
	if chainState.pastMedianTimeErr != nil {
		return time.Time{}, chainState.pastMedianTimeErr
	}

	// The timestamp for the block must not be before the median timestamp
	// of the last several blocks.  Thus, choose the maximum between the
	// current time and one second after the past median time.  The current
	// timestamp is truncated to a second boundary before comparison since a
	// block timestamp does not supported a precision greater than one
	// second.
	newTimestamp := time.Unix(time.Now().Unix(), 0)
	minTimestamp := chainState.pastMedianTime.Add(time.Second)
	if newTimestamp.Before(minTimestamp) {
		newTimestamp = minTimestamp
	}

	return newTimestamp, nil
}

// UpdateBlockTime updates the timestamp in the header of the passed block to
// the current time while taking into account the median time of the last
// several blocks to ensure the new time is after that time per the chain
// consensus rules.  Finally, it will update the target difficulty if needed
// based on the new time for the test networks since their target difficulty can
// change based upon time.
func UpdateBlockTime(msgBlock *wire.MsgBlock, bManager *blockManager) error {
	// The new timestamp is potentially adjusted to ensure it comes after
	// the median time of the last several blocks per the chain consensus
	// rules.
	newTimestamp, err := medianAdjustedTime(&bManager.chainState)
	if err != nil {
		return err
	}
	msgBlock.Header.Timestamp = newTimestamp

	/*
		// If running on a network that requires recalculating the difficulty,
		// do so now.
		if activeNetParams.ResetMinDifficulty {
			difficulty, err := bManager.CalcNextRequiredDifficulty(newTimestamp)
			if err != nil {
				return err
			}
			msgBlock.Header.Bits = difficulty
		}
	*/

	return nil
}

// UpdateExtraNonce updates the extra nonce in the coinbase script of the passed
// block by regenerating the coinbase script with the passed value and block
// height.  It also recalculates and updates the new merkle root that results
// from changing the coinbase script.
func UpdateExtraNonce(msgBlock *wire.MsgBlock, blockHeight int64, extraNonce uint64) error {
	util.Trace()
	/*
		coinbaseScript, err := standardCoinbaseScript(blockHeight, extraNonce)
		if err != nil {
			return err
		}
		if len(coinbaseScript) > blockchain.MaxCoinbaseScriptLen {
			return fmt.Errorf("coinbase transaction script length "+
				"of %d is out of range (min: %d, max: %d)",
				len(coinbaseScript), blockchain.MinCoinbaseScriptLen,
				blockchain.MaxCoinbaseScriptLen)
		}
		msgBlock.Transactions[0].TxIn[0].SignatureScript = coinbaseScript
	*/

	// TODO(davec): A btcutil.Block should use saved in the state to avoid
	// recalculating all of the other transaction hashes.
	// block.Transactions[0].InvalidateCache()

	// Recalculate the merkle root with the updated extra nonce.
	block := btcutil.NewBlock(msgBlock)
	merkles := blockchain.BuildMerkleTreeStore(block.Transactions())
	msgBlock.Header.MerkleRoot = *merkles[len(merkles)-1]
	return nil
}
