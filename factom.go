// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Glue code between BTCD code & Factom.

package main

import (
	"fmt"
	"github.com/FactomProject/FactomCode/factomd"
	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/btcd/wire"
	"github.com/FactomProject/btcutil"
)

var (
	local_Server *server
)

/*
// Handle factom app imcoming msg
func (p *peer) handleBuyCreditMsg(msg *wire.MsgGetCredit) {
	util.Trace()

	// Add the msg to inbound msg queue
	inMsgQueue <- msg
}
*/

// Handle factom app imcoming msg
func (p *peer) handleCommitChainMsg(msg *wire.MsgCommitChain) {
	util.Trace()

	// Add the msg to inbound msg queue
	factomd.InMsgQueue <- msg
}

// Handle factom app imcoming msg
func (p *peer) handleRevealChainMsg(msg *wire.MsgRevealChain) {
	util.Trace()

	// Add the msg to inbound msg queue
	factomd.InMsgQueue <- msg
}

// Handle factom app imcoming msg
func (p *peer) handleCommitEntryMsg(msg *wire.MsgCommitEntry) {
	util.Trace()

	// Add the msg to inbound msg queue
	factomd.InMsgQueue <- msg
}

// Handle factom app imcoming msg
func (p *peer) handleRevealEntryMsg(msg *wire.MsgRevealEntry) {
	util.Trace()

	// Add the msg to inbound msg queue
	factomd.InMsgQueue <- msg
}

// returns true if the message should be relayed, false otherwise
func (p *peer) shallRelay(msg interface{}) bool {
	util.Trace()

	fmt.Println("shallRelay msg= ", msg)

	hash, _ := wire.NewShaHashFromStruct(msg)
	fmt.Println("shallRelay hash= ", hash)

	iv := wire.NewInvVect(wire.InvTypeFactomRaw, hash)

	fmt.Println("shallRelay iv= ", iv)

	if !p.isKnownInventory(iv) {
		p.AddKnownInventory(iv)

		return true
	}

	fmt.Println("******************* SHALL NOT RELAY !!!!!!!!!!! ******************")

	return false
}

// Call FactomRelay to relay/broadcast a Factom message (to your peers).
// The intent is to call this function after certain 'processor' checks been done.
func (p *peer) FactomRelay(msg wire.Message) {
	util.Trace()

	fmt.Println("FactomRelay msg= ", msg)

	// broadcast/relay only if hadn't been done for this peer
	if p.shallRelay(msg) {
		//		p.server.BroadcastMessage(msg, p)
		local_Server.BroadcastMessage(msg)
	}
}

// func (pl *ProcessList) AddFtmTxToProcessList(msg wire.Message, msgHash *wire.ShaHash) error {
func fakehook1(msg wire.Message, msgHash *wire.ShaHash) error {
	return nil
}

func factom_PL_hook(tx *btcutil.Tx, label string) error {
	util.Trace("label= " + label)

	_ = fakehook1(tx.MsgTx(), tx.Sha())

	return nil
}

// for Jack
func global_DeleteMemPoolEntry(hash *wire.ShaHash) {
	// TODO: ensure mutex-protection
}

func factomInitFork() {
	cfg.DisableCheckpoints = true
}

// check a few btcd-related flags for sanity in our fork
func (b *blockManager) factomChecks() {
	util.Trace()

	if b.headersFirstMode {
		panic(1)
	}

	if cfg.AddrIndex {
		panic(2)
	}

	if !cfg.DisableCheckpoints {
		panic(3)
	}

	if cfg.RegressionTest || cfg.TestNet3 || cfg.SimNet || cfg.Generate {
		panic(100)
	}
}
