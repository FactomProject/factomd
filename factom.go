// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Glue code between BTCD code & Factom.

package btcd

import (
	"fmt"
	"github.com/FactomProject/FactomCode/factomd"
	"github.com/FactomProject/FactomCode/util"
	"github.com/FactomProject/FactomCode/wallet"
	"github.com/FactomProject/btcd/wire"
	//	"github.com/FactomProject/btcutil"
)

var (
	local_Server *server
)

// start up Factom queue(s) managers/processors
// this is to be called within the btcd's main code
func factomForkInit(s *server) {
	// tweak some config options
	cfg.DisableCheckpoints = true

	local_Server = s // local copy of our server pointer

	// Write outgoing factom messages into P2P network
	go func() {
		for msg := range factomd.OutMsgQueue {
			wireMsg, ok := msg.(wire.Message)
			if ok {
				s.BroadcastMessage(wireMsg)
			}
			/*      peerInfoResults := server.PeerInfo()
			        for peerInfo := range peerInfoResults{
			          fmt.Printf("PeerInfo:%+v", peerInfo)

			        }*/
		}
	}()

	/*
	   go func() {
	     for msg := range inRpcQueue {
	       fmt.Printf("in range inRpcQueue, msg:%+v\n", msg)
	       switch msg.Command() {
	       case factomwire.CmdTx:
	         InMsgQueue <- msg //    for testing
	         server.blockManager.QueueTx(msg.(*factomwire.MsgTx), nil)
	       case factomwire.CmdConfirmation:
	         server.blockManager.QueueConf(msg.(*factomwire.MsgConfirmation), nil)

	       default:
	         inMsgQueue <- msg
	         outMsgQueue <- msg
	       }
	     }
	   }()
	*/
}

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

/*
// func (pl *ProcessList) AddFtmTxToProcessList(msg wire.Message, msgHash *wire.ShaHash) error {
func fakehook1(msg wire.Message, msgHash *wire.ShaHash) error {
	return nil
}

func factom_PL_hook(tx *btcutil.Tx, label string) error {
	util.Trace("label= " + label)

	_ = fakehook1(tx.MsgTx(), tx.Sha())

	return nil
}
*/

// for Jack
func global_DeleteMemPoolEntry(hash *wire.ShaHash) {
	// TODO: ensure mutex-protection
}

func FactomSetupOverrides() {
	//	factomd.FactomOverride.TxIgnoreMissingParents = true
	factomd.FactomOverride.TxOrphansInsteadOfMempool = true
	factomd.FactomOverride.BlockDisableChecks = true
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

	if cfg.RegressionTest || cfg.SimNet || cfg.Generate {
		panic(100)
	}

	util.Trace()
}

// feed all incoming Txs to the inner Factom code (for Jack)
// TODO: do this after proper mempool/orphanpool/validity triangulation & checks
func factomIngressTx_hook(tx *wire.MsgTx) error {
	util.Trace()

	ecmap := make(map[wire.ShaHash]uint64)

	txid, _ := tx.TxSha()
	hash, _ := wire.NewShaHash(wire.Sha256(txid.Bytes()))

	ecmap[*hash] = 1

	// Use wallet's public key to add EC??
	sig := wallet.SignData(nil)
	hash2 := new(wire.ShaHash)
	hash2.SetBytes((*sig.Pub.Key)[:])

	ecmap[*hash2] = 1000000

	txHash, _ := tx.TxSha()
	fo := &wire.MsgInt_FactoidObj{tx, &txHash, ecmap}

	fmt.Println("ecmap len =", len(ecmap))

	factomd.InMsgQueue <- fo

	return nil
}
