// Copyright (c) 2014 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// temp code to test out the block generating functions

package btcd

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcd/blockchain"
	"github.com/FactomProject/btcd/chaincfg"
	"github.com/FactomProject/btcd/wire"
	"github.com/FactomProject/btcutil"
)

type CPUMINER struct {
	sync.Mutex
	server          *server
	submitBlockLock sync.Mutex
}

const (
	MaxPayoutAmount = blockchain.BaseSubsidy // in Satoshis
)

// generateBlocks is a worker that is controlled by the miningWorkerController.
// It is self contained in that it creates block templates and attempts to solve
// them while detecting when it is performing stale work and reacting
// accordingly by generating a new block template.  When a block is solved, it
// is submitted.
//
// It must be run as a goroutine.
func generateFactoidBlock(h uint32) {
	minrLog.Infof("Starting generate blocks worker")
	util.Trace()

	var cpum CPUMINER
	m := &cpum
	m.server = local_Server

	/*
		// Start a ticker which is used to signal checks for stale work and
		// updates to the speed monitor.
		ticker := time.NewTicker(time.Second * hashUpdateSecs)
		defer ticker.Stop()
	*/

	// No point in searching for a solution before the chain is
	// synced.  Also, grab the same lock as used for block
	// submission, since the current block will be changing and
	// this would otherwise end up building a new block template on
	// a block that is in the process of becoming stale.
	m.submitBlockLock.Lock()

	// Create a new block template using the available transactions
	// in the memory pool as a source of transactions to potentially
	// include in the block.

	//	payto := wire.RCDHash{} // for Phase 1 pay to a non-spendable address, FIXME later

	// comment out later, testing payout to a random address
	var payto wire.RCDHash
	randRCD := make([]byte, wire.RCDHashSize)
	rand.Read(randRCD)
	fmt.Println("randRCD len=", len(randRCD))
	copy(payto[:], randRCD)

	// 03264ef3197ff0b843e9f83362341e8cc135913bcbbdbabd62d049acaeb2e55944
	//	sha, err := wire.NewShaHashFromStr("264ef3197ff0b843e9f83362341e8cc135913bcbbdbabd62d049acaeb2e55944")
	sha, err := wire.NewShaHashFromStr("4459e5b2aeac49d062bdbabdcb3b9135c18c1e346233f8e943b8f07f19f34e26")

	if nil == err {
		util.Trace("mywallet sha=" + spew.Sdump(sha))
		copy(payto[:], sha[:])
		util.Trace("mywallet payto=" + spew.Sdump(payto))
	}

	template, err := factom_NewBlockTemplate(local_Server.txMemPool, payto, h)
	m.submitBlockLock.Unlock()

	util.Trace()

	if err != nil {
		errStr := fmt.Sprintf("Failed to create new block "+
			"template: %v", err)
		minrLog.Errorf(errStr)
		util.Trace()
		return
	}

	util.Trace()
	curHeight := int64(0)

	block := btcutil.NewBlock(template.block)
	fmt.Println("NewBlock= ", spew.Sdump(block))

	successful := m.submitBlock(block)

	//	m.workerWg.Done()
	minrLog.Infof("Generate blocks worker done; height= %d", curHeight)

	blockSha, _ := block.Sha()

	if successful {
		factomIngressBlock_hook(blockSha)
	}

	util.Trace()
}

// newAddressPubKey returns a new btcutil.AddressPubKey from the provided
// serialized public key.  It panics if an error occurs.  This is only used in
// the tests as a helper since the only way it can fail is if there is an error
// in the test source code.
func newAddressPubKey(serializedPubKey []byte) btcutil.Address {
	addr, err := btcutil.NewAddressPubKey(serializedPubKey,
		&chaincfg.MainNetParams)
	if err != nil {
		fmt.Println(err)
		panic("invalid public key in test source")
	}

	return addr
}

func decodeHex(hexStr string) []byte {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		panic("invalid hex string in test source: err " + err.Error() +
			", hex: " + hexStr)
	}

	return b
}

/*
   mp.pool[*tx.Sha()] = &TxDesc{
     Tx:     tx,
     Added:  time.Now(),
     Height: height,
     Fee:    fee,
*/

// TxDescs returns a slice of descriptors for all the transactions in the pool.
// The descriptors are to be treated as read only.
//
// This function is safe for concurrent access.
func (mp *txMemPool) myDescs() []*TxDesc {
	mp.RLock()
	defer mp.RUnlock()

	util.Trace()

	descs := make([]*TxDesc, len(mp.orphans))
	i := 0
	for _, tx := range mp.orphans {
		descs[i] = &TxDesc{tx, time.Now(), 0, 0, 0}
		//		fmt.Println("i= ", spew.Sdump(tx))
		i++
	}

	util.Trace("collected " + fmt.Sprintf("%d", i) + " orphans as FAKE TXs for the block")

	return descs
}

// submitBlock submits the passed block to network after ensuring it passes all
// of the consensus validation rules.
func (m *CPUMINER) submitBlock(block *btcutil.Block) bool {
	m.submitBlockLock.Lock()
	defer m.submitBlockLock.Unlock()

	util.Trace()

	// Ensure the block is not stale since a new block could have shown up
	// while the solution was being found.  Typically that condition is
	// detected and all work on the stale block is halted to start work on
	// a new block, but the check only happens periodically, so it is
	// possible a block was found and submitted in between.
	latestHash, _ := m.server.blockManager.chainState.Best()
	msgBlock := block.MsgBlock()
	if !msgBlock.Header.PrevBlock.IsEqual(latestHash) {
		minrLog.Debugf("Block submitted via CPU miner with previous "+
			"block %s is stale", msgBlock.Header.PrevBlock)
		return false
	}

	// Process this block using the same rules as blocks coming from other
	// nodes.  This will in turn relay it to the network like normal.
	isOrphan, err := m.server.blockManager.bm_ProcessBlock(block, blockchain.BFNone)
	if err != nil {
		// Anything other than a rule violation is an unexpected error,
		// so log that error as an internal error.
		if _, ok := err.(blockchain.RuleError); !ok {
			minrLog.Errorf("Unexpected error while processing "+
				"block submitted via CPU miner: %v", err)
			return false
		}

		minrLog.Debugf("Block submitted via CPU miner rejected: %v", err)
		return false
	}
	if isOrphan {
		minrLog.Debugf("Block submitted via CPU miner is an orphan")
		return false
	}

	// The block was accepted.
	blockSha, _ := block.Sha()
	coinbaseTx := block.MsgBlock().Transactions[0].TxOut[0]
	txsha, _ := block.MsgBlock().Transactions[0].TxSha()
	minrLog.Infof("Block submitted via CPU miner accepted (hash %s, "+
		"amount %v; tx hash: %s)", blockSha, btcutil.Amount(coinbaseTx.Value), txsha)
	return true
}

// factom_NewBlockTemplate returns a new block template that is ready to be solved
// using the transactions from the passed transaction memory pool and a coinbase
// that either pays to the passed address if it is not nil, or a coinbase that
// is redeemable by anyone if the passed address is nil.  The nil address
// functionality is useful since there are cases such as the getblocktemplate
// RPC where external mining software is responsible for creating their own
// coinbase which will replace the one generated for the block template.  Thus
// the need to have configured address can be avoided.
//
// The transactions selected and included are prioritized according to several
// factors.  First, each transaction has a priority calculated based on its
// value, age of inputs, and size.  Transactions which consist of larger
// amounts, older inputs, and small sizes have the highest priority.  Second, a
// fee per kilobyte is calculated for each transaction.  Transactions with a
// higher fee per kilobyte are preferred.  Finally, the block generation related
// configuration options are all taken into account.
//
// Transactions which only spend outputs from other transactions already in the
// block chain are immediately added to a priority queue which either
// prioritizes based on the priority (then fee per kilobyte) or the fee per
// kilobyte (then priority) depending on whether or not the BlockPrioritySize
// configuration option allots space for high-priority transactions.
// Transactions which spend outputs from other transactions in the memory pool
// are added to a dependency map so they can be added to the priority queue once
// the transactions they depend on have been included.
//
// Once the high-priority area (if configured) has been filled with transactions,
// or the priority falls below what is considered high-priority, the priority
// queue is updated to prioritize by fees per kilobyte (then priority).
//
// When the fees per kilobyte drop below the TxMinFreeFee configuration option,
// the transaction will be skipped unless there is a BlockMinSize set, in which
// case the block will be filled with the low-fee/free transactions until the
// block size reaches that minimum size.
//
// Any transactions which would cause the block to exceed the BlockMaxSize
// configuration option, exceed the maximum allowed signature operations per
// block, or otherwise cause the block to be invalid are skipped.
//
// Given the above, a block generated by this function is of the following form:
//
//   -----------------------------------  --  --
//  |      Coinbase Transaction         |   |   |
//  |-----------------------------------|   |   |
//  |                                   |   |   | ----- cfg.BlockPrioritySize
//  |   High-priority Transactions      |   |   |
//  |                                   |   |   |
//  |-----------------------------------|   | --
//  |                                   |   |
//  |                                   |   |
//  |                                   |   |--- cfg.BlockMaxSize
//  |  Transactions prioritized by fee  |   |
//  |  until <= cfg.TxMinFreeFee        |   |
//  |                                   |   |
//  |                                   |   |
//  |                                   |   |
//  |-----------------------------------|   |
//  |  Low-fee/Non high-priority (free) |   |
//  |  transactions (while block size   |   |
//  |  <= cfg.BlockMinSize)             |   |
//   -----------------------------------  --
func factom_NewBlockTemplate(mempool *txMemPool, payToAddress wire.RCDHash, globalHeight uint32) (*BlockTemplate, error) {
	util.Trace()

	blockManager := mempool.server.blockManager
	chainState := &blockManager.chainState
	//	chain := blockManager.blockChain

	// Extend the most recently known best block.
	chainState.Lock()
	prevHash := chainState.newestHash
	nextBlockHeight := chainState.newestHeight + 1
	chainState.Unlock()

	if uint32(nextBlockHeight) != globalHeight {
		panic(errors.New(fmt.Sprintf("ERROR: DirBlockHeight (globalHeight) is %d and FactoidChainHeight (nextBlockHeight) is %d\n", globalHeight, nextBlockHeight)))
	}

	fmt.Printf("nextBlockHeight= %d\n", nextBlockHeight)
	fmt.Println("prevHash= ", prevHash)

	coinbaseTx, err := createCoinbaseTx(uint32(nextBlockHeight), payToAddress)
	if err != nil {
		return nil, err
	}
	util.Trace(fmt.Sprintf("coinbase tx=%s", spew.Sdump(coinbaseTx)))

	numCoinbaseSigOps := int64(blockchain.CountSigOps(coinbaseTx))

	// Get the current memory pool transactions and create a priority queue
	// to hold the transactions which are ready for inclusion into a block
	// along with some priority related and fee metadata.  Reserve the same
	// number of items that are in the memory pool for the priority queue.
	// Also, choose the initial sort order for the priority queue based on
	// whether or not there is an area allocated for high-priority
	// transactions.

	var mempoolTxns []*TxDesc

	if !FactomOverride.TxOrphansInsteadOfMempool {
		util.Trace()
		mempoolTxns = mempool.TxDescs()
	} else {
		util.Trace()
		mempoolTxns = mempool.myDescs()
	}
	util.Trace()

	// Create a slice to hold the transactions to be included in the
	// generated block with reserved space.  Also create a transaction
	// store to house all of the input transactions so multiple lookups
	// can be avoided.
	blockTxns := make([]*btcutil.Tx, 0, len(mempoolTxns))
	blockTxns = append(blockTxns, coinbaseTx)

	blockTxStore := make(blockchain.TxStore)

	// dependers is used to track transactions which depend on another
	// transaction in the memory pool.  This, in conjunction with the
	// dependsOn map kept with each dependent transaction helps quickly
	// determine which dependent transactions are now eligible for inclusion
	// in the block once each transaction has been included.
	//	dependers := make(map[wire.ShaHash]*list.List)

	// Create slices to hold the fees and number of signature operations
	// for each of the selected transactions and add an entry for the
	// coinbase.  This allows the code below to simply append details about
	// a transaction as it is selected for inclusion in the final block.
	// However, since the total fees aren't known yet, use a dummy value for
	// the coinbase fee which will be updated later.
	txFees := make([]int64, 0, len(mempoolTxns))
	txSigOpCounts := make([]int64, 0, len(mempoolTxns))
	txFees = append(txFees, -1) // Updated once known
	txSigOpCounts = append(txSigOpCounts, numCoinbaseSigOps)

	util.Trace("Considering " + fmt.Sprintf("%d", len(mempoolTxns)) + " transactions for a new block !")
	minrLog.Infof("Considering %d mempool transactions for inclusion to "+
		"new block", len(mempoolTxns))

	// mempoolLoop:
	for _, txDesc := range mempoolTxns {
		// A block can't have more than one coinbase or contain
		// non-finalized transactions.
		tx := txDesc.Tx

		// Spend the transaction inputs in the block transaction store
		// and add an entry for it to ensure any transactions which
		// reference this one have it available as an input and can
		// ensure they aren't double spending.
		spendTransaction(blockTxStore, tx, nextBlockHeight)

		// Add the transaction to the block, increment counters, and
		// save the fees and signature operation counts to the block
		// template.
		blockTxns = append(blockTxns, tx)
	}

	blockSize := 0
	blockSigOps := 0
	totalFees := int64(0)

	util.Trace()

	// block size for the real transaction count and coinbase value with
	// the total fees accordingly.
	coinbaseTx.MsgTx().TxOut[0].Value += totalFees
	txFees[0] = -totalFees

	util.Trace()

	// Create a new block ready to be solved.

	// TODO: need the PrevHash3 (SHA3) calculation here............
	util.Trace("Need SHA3 calculations here.......................")

	merkles := blockchain.BuildMerkleTreeStore(blockTxns)
	var msgBlock wire.MsgBlock
	msgBlock.Header = wire.FBlockHeader{
		PrevBlock:  *prevHash,
		MerkleRoot: *merkles[len(merkles)-1],
		// PrevHash3:  <- Put the SHA3 here!
		//		Bits:       0,
	}

	for _, tx := range blockTxns {
		if err := msgBlock.AddTransaction(tx.MsgTx()); err != nil {
			util.Trace("ERROR AddTransaction")
			return nil, err
		}
	}
	util.Trace()

	// Finally, perform a full check on the created block against the chain
	// consensus rules to ensure it properly connects to the current best
	// chain with no issues.
	block := btcutil.NewBlock(&msgBlock)
	block.SetHeight(nextBlockHeight)

	if !FactomOverride.BlockDisableChecks {
		if err := blockManager.CheckConnectBlock(block); err != nil {
			return nil, err
		}
	}
	util.Trace()

	minrLog.Infof("Created new block template (%d transactions, %d in fees, %d signature operations, %d bytes)",
		len(msgBlock.Transactions), totalFees, blockSigOps, blockSize)

	util.Trace()
	return &BlockTemplate{
		block:       &msgBlock,
		fees:        txFees,
		sigOpCounts: txSigOpCounts,
		height:      nextBlockHeight,
		// TODO: good check to have; compile issue TBD; FIXME
		// cannot convert nil to type wire.RCDHash
		// use of untyped nil
		//		validPayAddress: payToAddress != nil,
		validPayAddress: true,
	}, nil
}
