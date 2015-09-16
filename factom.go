// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Glue code between BTCD code & Factom.

package btcd

import (
	"errors"
	"fmt"
	"os"

	//	"github.com/FactomProject/btcd/chaincfg"
	//	"github.com/FactomProject/btcutil"

	cp "github.com/FactomProject/FactomCode/controlpanel"
	"github.com/FactomProject/FactomCode/database"
	"github.com/FactomProject/btcd/wire"
)

var _ = fmt.Printf

var (
	local_Server *server
	db           database.Db              // database
	inMsgQueue   chan wire.FtmInternalMsg //incoming message queue for factom application messages
	outMsgQueue  chan wire.FtmInternalMsg //outgoing message queue for factom application messages

	inCtlMsgQueue  chan wire.FtmInternalMsg //incoming message queue for factom control messages
	outCtlMsgQueue chan wire.FtmInternalMsg //outgoing message queue for factom control messages
)

// start up Factom queue(s) managers/processors
// this is to be called within the btcd's main code
func factomForkInit(s *server) {
	// tweak some config options
	cfg.DisableCheckpoints = true

	local_Server = s // local copy of our server pointer

	// Write outgoing factom messages into P2P network
	go func() {
		for msg := range outMsgQueue {
			switch msg.(type) {
			case *wire.MsgInt_DirBlock:
				dirBlock, _ := msg.(*wire.MsgInt_DirBlock)
				iv := wire.NewInvVect(wire.InvTypeFactomDirBlock, dirBlock.ShaHash)
				s.RelayInventory(iv, nil)

			case wire.Message:
				wireMsg, _ := msg.(wire.Message)
				//s.BroadcastMessage(wireMsg)
				if ClientOnly {
					//fmt.Println("broadcasting from client.")
					s.BroadcastMessage(wireMsg)
				} else {
					if _, ok := msg.(*wire.MsgAcknowledgement); ok {
						//fmt.Println("broadcasting from server.")
						s.BroadcastMessage(wireMsg)
					}
				}

			default:
				panic(fmt.Sprintf("bad outMsgQueue message received: %v", msg))
			}
			/*      peerInfoResults := server.PeerInfo()
			        for peerInfo := range peerInfoResults{
			          fmt.Printf("PeerInfo:%+v", peerInfo)

			        }*/
		}
	}()

	go func() {
		for msg := range outCtlMsgQueue {

			fmt.Printf("in range outCtlMsgQueue, msg:%+v\n", msg)

			msgEom, _ := msg.(*wire.MsgInt_EOM)

			//			switch msgEom.Command() {
			switch msg.Command() {

			case wire.CmdInt_EOM:

				switch msgEom.EOM_Type {

				case wire.END_MINUTE_10:
					panic(errors.New("unhandled END_MINUTE_10"))

					/*
						// block building, return the hash of the new one via doneFB (via hook)
						generateFactoidBlock(msgEom.NextDBlockHeight)
						fmt.Println("***********************")
						fmt.Println("***********************")
					*/

				default:
					panic(errors.New("unhandled EOM type"))
				}

			default:
				panic(errors.New("unhandled CmdInt_EOM"))
			}

			/*
				switch msg.EOM_Type {

				case wire.END_MINUTE_10:
					//util.Trace("EOM10")
				default:
					//util.Trace("default")
				}
			*/

			/*
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
			*/
		}
	}()
}

func Start_btcd(
	ldb database.Db,
	inMsgQ chan wire.FtmInternalMsg,
	outMsgQ chan wire.FtmInternalMsg,
	inCtlMsgQ chan wire.FtmInternalMsg,
	outCtlMsgQ chan wire.FtmInternalMsg,
	user, pass string, clientMode bool) {

	factomdUser = user
	factomdPass = pass

	ClientOnly = clientMode

	if ClientOnly {
		cp.CP.AddUpdate(
			"FactomMode", // tag
			"system",     // Category
			"Factom Mode: Full Node (Client)", // Title
			"", // Message
			0)
		fmt.Println("\n\n>>>>>>>>>>>>>>>>>  CLIENT MODE <<<<<<<<<<<<<<<<<<<<<<<\n\n")
	} else {
		cp.CP.AddUpdate(
			"FactomMode",                    // tag
			"system",                        // Category
			"Factom Mode: Federated Server", // Title
			"", // Message
			0)
		fmt.Println("\n\n>>>>>>>>>>>>>>>>>  SERVER MODE <<<<<<<<<<<<<<<<<<<<<<<\n\n")
	}

	db = ldb

	inMsgQueue = inMsgQ
	outMsgQueue = outMsgQ

	inCtlMsgQueue = inCtlMsgQ
	outCtlMsgQueue = outCtlMsgQ

	// Use all processor cores.
	//runtime.GOMAXPROCS(runtime.NumCPU())

	// Up some limits.
	//if err := limits.SetLimits(); err != nil {
	//	os.Exit(1)
	//}

	// Call serviceMain on Windows to handle running as a service.  When
	// the return isService flag is true, exit now since we ran as a
	// service.  Otherwise, just fall through to normal operation.
	/*if runtime.GOOS == "windows" {
		isService, err := winServiceMain()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if isService {
			os.Exit(0)
		}
	}
	*/

	// Work around defer not working after os.Exit()
	if err := btcdMain(nil); err != nil {
		os.Exit(1)
	}
}

// Handle factom app imcoming msg
func (p *peer) handleCommitChainMsg(msg *wire.MsgCommitChain) {
	// Add the msg to inbound msg queue
	if !ClientOnly {
		inMsgQueue <- msg
	}
}

// Handle factom app imcoming msg
func (p *peer) handleRevealChainMsg(msg *wire.MsgRevealChain) {
	// Add the msg to inbound msg queue
	//inMsgQueue <- msg
	if !ClientOnly {
		inMsgQueue <- msg
	}
}

// Handle factom app imcoming msg
func (p *peer) handleCommitEntryMsg(msg *wire.MsgCommitEntry) {
	// Add the msg to inbound msg queue
	if !ClientOnly {
		inMsgQueue <- msg
	}
}

// Handle factom app imcoming msg
func (p *peer) handleRevealEntryMsg(msg *wire.MsgRevealEntry) {
	// Add the msg to inbound msg queue
	if !ClientOnly {
		inMsgQueue <- msg
	}
}

// Handle factom app imcoming msg
func (p *peer) handleAcknoledgementMsg(msg *wire.MsgAcknowledgement) {
	// Add the msg to inbound msg queue
	if !ClientOnly {
		inMsgQueue <- msg
	}
}

// returns true if the message should be relayed, false otherwise
func (p *peer) shallRelay(msg interface{}) bool {
	hash, _ := wire.NewShaHashFromStruct(msg)
	iv := wire.NewInvVect(wire.InvTypeFactomRaw, hash)

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
	//util.Trace("label= " + label)

	_ = fakehook1(tx.MsgTx(), tx.Sha())

	return nil
}

// for Jack
func global_DeleteMemPoolEntry(hash *wire.ShaHash) {
	// TODO: ensure mutex-protection
}

// check a few btcd-related flags for sanity in our fork
func (b *blockManager) factomChecks() {
	//util.Trace()

	if cfg.AddrIndex {
		panic(errors.New("AddrIndex must be disabled and it is NOT !!!"))
	}

	// DisableCheckpoints should always be set
	if !cfg.DisableCheckpoints {
		panic(errors.New("checkpoints must be disabled and they are NOT !!!"))
	}

	if cfg.RegressionTest || cfg.SimNet || cfg.Generate {
		panic(100)
	}

	if cfg.TestNet3 {
		panic(errors.New("TestNet mode is NOT SUPPORTED (remove the option from the command line or from the .conf file)!"))
	}

	//util.Trace()
}

// feed all incoming Txs to the inner Factom code (for Jack)
// TODO: do this after proper mempool/orphanpool/validity triangulation & checks
func factomIngressTx_hook(tx *wire.MsgTx) error {
	//util.Trace()

	ecmap := make(map[wire.ShaHash]uint64)

	txid, _ := tx.TxSha()
	hash, _ := wire.NewShaHash(wire.Sha256(txid.Bytes()))

	ecmap[*hash] = 1

	// Use wallet's public key to add EC??
	sig := wallet.SignData(nil)
	hash2 := new(wire.ShaHash)
	hash2.SetBytes((*sig.Pub.Key)[:])

	ecmap[*hash2] = 100

	txHash, _ := tx.TxSha()
	fo := &wire.MsgInt_FactoidObj{tx, &txHash, ecmap}

	fmt.Println("ecmap len =", len(ecmap))

	inMsgQueue <- fo

	return nil
}

func factomIngressBlock_hook(hash *wire.ShaHash) error {
	//util.Trace(fmt.Sprintf("hash: %s", hash))

	fbo := &wire.MsgInt_FactoidBlock{
		ShaHash: *hash}

	doneFBlockQueue <- fbo

	return nil
}
*/

/*
func ExtractPkScriptAddrs(pkScript []byte, chainParams *chaincfg.Params) ([]btcutil.Address, int, error) {
	oldWay := false

	//util.Trace("bytes= " + spew.Sdump(pkScript))

	var addrs []btcutil.Address
	var requiredSigs int

	if oldWay {
		// A pay-to-pubkey script is of the form:
		//  <pubkey> OP_CHECKSIG
		// Therefore the pubkey is the first item on the stack.
		// Skip the pubkey if it's invalid for some reason.
		requiredSigs = 1
		addr, err := btcutil.NewAddressPubKey(pkScript, chainParams)
		if err == nil {
			addrs = append(addrs, addr)
		}

	} else {

		// A pay-to-pubkey-hash script is of the form:
		//  OP_DUP OP_HASH160 <hash> OP_EQUALVERIFY OP_CHECKSIG
		// Therefore the pubkey hash is the 3rd item on the stack.
		// Skip the pubkey hash if it's invalid for some reason.
		requiredSigs = 1
		//	addr, err := btcutil.NewAddressPubKeyHash(pops[2].data,
		addr, err := btcutil.NewAddressPubKeyHash(pkScript, chainParams)
		if err == nil {
			addrs = append(addrs, addr)
		}
	}

	//util.Trace("addrs= " + spew.Sdump(addrs))

	return addrs, requiredSigs, nil
}

// PayToAddrScript creates a new script to pay a transaction output to a the
// specified address.
func PayToAddrScript(addr btcutil.Address) ([]byte, error) {
	scrAddr := addr.ScriptAddress()

	//util.Trace("scrAddr= " + spew.Sdump(scrAddr))

	return scrAddr, nil

	//	panic(errors.New("PayToAddrScript -- NOT IMPLEMENTED !!!"))

	//	return payToPubKeyHashScript(addr.ScriptAddress())
}
*/
