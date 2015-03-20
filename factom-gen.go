// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// temp code to test out the block generating functions

package main

import (
	"fmt"
	//	"math/rand"
	"encoding/hex"
	"time"

	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcd/chaincfg"
	"github.com/FactomProject/btcutil"
)

// generateBlocks is a worker that is controlled by the miningWorkerController.
// It is self contained in that it creates block templates and attempts to solve
// them while detecting when it is performing stale work and reacting
// accordingly by generating a new block template.  When a block is solved, it
// is submitted.
//
// It must be run as a goroutine.
func test_generateBlocks() {
	minrLog.Tracef("Starting generate blocks worker")
	util.Trace()

	var cpum CPUMiner
	m := &cpum

	// Start a ticker which is used to signal checks for stale work and
	// updates to the speed monitor.
	ticker := time.NewTicker(time.Second * hashUpdateSecs)
	defer ticker.Stop()

	// No point in searching for a solution before the chain is
	// synced.  Also, grab the same lock as used for block
	// submission, since the current block will be changing and
	// this would otherwise end up building a new block template on
	// a block that is in the process of becoming stale.
	m.submitBlockLock.Lock()

	// Choose a payment address at random.
	//	rand.Seed(time.Now().UnixNano())
	//	payToAddr := cfg.miningAddrs[rand.Intn(len(cfg.miningAddrs))]
	payToAddr := newAddressPubKey(decodeHex("0461cbdcc5409fb4b" +
		"4d42b51d33381354d80e550078cb532a34bf" +
		"a2fcfdeb7d76519aecc62770f5b0e4ef8551" +
		"946d8a540911abe3e7854a26f39f58b25c15" +
		"342af"))

	// Create a new block template using the available transactions
	// in the memory pool as a source of transactions to potentially
	// include in the block.
	template, err := NewBlockTemplate(local_Server.txMemPool, payToAddr)
	m.submitBlockLock.Unlock()
	if err != nil {
		errStr := fmt.Sprintf("Failed to create new block "+
			"template: %v", err)
		minrLog.Errorf(errStr)
		util.Trace()
		return
	}

	curHeight := int64(0)

	// Attempt to solve the block.  The function will exit early
	// with false when conditions that trigger a stale block, so
	// a new block template can be generated.  When the return is
	// true a solution was found, so submit the solved block.
	//	if m.solveBlock(template.block, curHeight+1, ticker, quit) {
	if m.solveBlock(template.block, curHeight+1, ticker, make(chan struct{})) {
		block := btcutil.NewBlock(template.block)
		m.submitBlock(block)
	}

	m.workerWg.Done()
	minrLog.Tracef("Generate blocks worker done")
	util.Trace()
}

func test_timer() {
	count := int32(0)

	for {
		time.Sleep(time.Second * 5)
		count++

		fmt.Println("===================================")
		fmt.Println(count, count%2)
		if 0 == (count % 2) {
			fmt.Println("evensec")

			/*
				test_generateBlocks() // CheckConnectBlock not returning yet, called from NewBlockTemplate
			*/

		}
		fmt.Println("===================================")
	}
}

// newAddressPubKey returns a new btcutil.AddressPubKey from the provided
// serialized public key.  It panics if an error occurs.  This is only used in
// the tests as a helper since the only way it can fail is if there is an error
// in the test source code.
func newAddressPubKey(serializedPubKey []byte) btcutil.Address {
	addr, err := btcutil.NewAddressPubKey(serializedPubKey,
		&chaincfg.MainNetParams)
	if err != nil {
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
