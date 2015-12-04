// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

// Glue code between BTCD code & Factom.

package btcd

import (
	//"errors"
	"fmt"
	//"os"

	"github.com/FactomProject/factomd/btcd/wire"
	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"
	//cp "github.com/FactomProject/factomd/controlpanel"
)

var _ = fmt.Printf

var (
	local_Server *server
	db           interfaces.DBOverlay     // database
	inMsgQueue   chan wire.FtmInternalMsg //incoming message queue for factom application messages
	outMsgQueue  chan wire.FtmInternalMsg //outgoing message queue for factom application messages

	inCtlMsgQueue  chan wire.FtmInternalMsg //incoming message queue for factom control messages
	outCtlMsgQueue chan wire.FtmInternalMsg //outgoing message queue for factom control messages
)

/*
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
				iv := wire.NewInvVect(wire.InvTypeFactomDirBlock, dirBlock.interfaces.IHash)
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

				default:
					panic(errors.New("unhandled EOM type"))
				}

			default:
				panic(errors.New("unhandled CmdInt_EOM"))
			}
		}
	}()
}

func Start_btcd(
	ldb interfaces.DBOverlay,
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

	// Work around defer not working after os.Exit()
	if err := btcdMain(nil); err != nil {
		os.Exit(1)
	}
}
*/

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
	hash, _ := NewShaHashFromStruct(msg)
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
