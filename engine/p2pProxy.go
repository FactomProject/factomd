// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bufio"
	"fmt"
	"os"
	"time"

	// "github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/p2p"
)

var _ = fmt.Print

var ()

type P2PProxy struct {
	// A connection to this node:
	ToName   string
	FromName string
	// Channels that define the connection:
	BroadcastOut chan interface{} // factomMessage ToNetwork from factomd
	BroadcastIn  chan interface{} // factomMessage FromNetwork for Factomd

	ToNetwork   chan interface{} // p2p.Parcel From p2pProxy to the p2p Controller
	FromNetwork chan interface{} // p2p.Parcel Parcels from the network for the application

	logFile   os.File
	logWriter bufio.Writer
	debugMode int
	logging   chan interface{} // NODE_TALK_FIX
}

type factomMessage struct {
	message  []byte
	peerHash string
}

var _ interfaces.IPeer = (*P2PProxy)(nil)

func (f *P2PProxy) Weight() int {
	// should return the number of connections this peer represents.  For now, just say a lot
	return 10
}

func (f *P2PProxy) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	f.BroadcastOut = make(chan interface{}, p2p.StandardChannelSize)
	f.BroadcastIn = make(chan interface{}, p2p.StandardChannelSize)
	f.logging = make(chan interface{}, p2p.StandardChannelSize)
	return f
}
func (f *P2PProxy) SetDebugMode(netdebug int) {
	f.debugMode = netdebug
}

func (f *P2PProxy) GetNameFrom() string {
	return f.FromName
}

func (f *P2PProxy) GetNameTo() string {
	return f.ToName
}

func (f *P2PProxy) Send(msg interfaces.IMsg) error {
	f.logMessage(msg, false) // NODE_TALK_FIX
	data, err := msg.MarshalBinary()
	if err != nil {
		fmt.Println("ERROR on Send: ", err)
		return err
	}
	message := factomMessage{message: data, peerHash: msg.GetNetworkOrigin()}
	if !msg.IsPeer2Peer() {
		message.peerHash = ""
	} else {
		fmt.Printf("%s Sending directed to: %s message: %+v\n", time.Now().String(), msg.GetNetworkOrigin(), msg.String())
	}
	p2p.BlockFreeChannelSend(f.BroadcastOut, message)
	return nil
}

// Non-blocking return value from channel.
func (f *P2PProxy) Recieve() (interfaces.IMsg, error) {
	select {
	case data, ok := <-f.BroadcastIn:
		if ok {
			switch data.(type) {
			case factomMessage:
				fmessage := data.(factomMessage)
				msg, err := messages.UnmarshalMessage(fmessage.message)
				if nil == err {
					msg.SetNetworkOrigin(fmessage.peerHash)
				}
				if  1 < f.debugMode {
					f.logMessage(msg, true) // NODE_TALK_FIX
				}
				return msg, err
			default:
				fmt.Printf("Garbage on f.BroadcastIn. %+v", data)
			}
		}
	default:
	}
	return nil, nil
}

// Is this connection equal to parm connection
func (f *P2PProxy) Equals(ff interfaces.IPeer) bool {
	f2, ok := ff.(*P2PProxy)
	if !ok {
		return false
	} // Different peer type can't be equal
	// Check If this is another connection from here to there
	if f.FromName == f2.FromName && f.ToName == f2.FromName {
		return true
	}
	// Check if this is a connection from there to here
	if f.FromName == f2.ToName && f.ToName == f2.FromName {
		return true
	}
	return false
}

// Returns the number of messages waiting to be read
func (f *P2PProxy) Len() int {
	return len(f.BroadcastIn)
}

//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////

func (p *P2PProxy) StartProxy() {
	if 1 < p.debugMode {
		go p.ManageLogging()
	}
	go p.ManageOutChannel() // Bridges between network format Parcels and factomd messages (incl. addressing to peers)
	go p.ManageInChannel()
}

// NODE_TALK_FIX
func (p *P2PProxy) stopProxy() {
	if 0 < p.debugMode {
		p2p.BlockFreeChannelSend(p.logging, "stop")
	}
}

type messageLog struct {
	hash     string // string(GetMsgHash().Bytes())
	received bool   // true if logging a recieved message, false if sending
	time     int64
	target   string // the id of the targetted node (value may only have local meaning)
	mtype    byte   /// message type (types defined in constants.go)
}

func (p *P2PProxy) logMessage(msg interfaces.IMsg, received bool) {
	if 1 < p.debugMode {
		// if constants.DBSTATE_MSG == msg.Type() {
		// fmt.Printf("AppMsgLogging: \n Type: %s \n Network Origin: %s \n Message: %s", msg.Type(), msg.GetNetworkOrigin(), msg.String())
		// }
		hash := fmt.Sprintf("%x", msg.GetMsgHash().Bytes())
		time := time.Now().Unix()
		ml := messageLog{hash: hash, received: received, time: time, mtype: msg.Type(), target: msg.GetNetworkOrigin()}
		p2p.BlockFreeChannelSend(p.logging, ml)
	}
}

func (p *P2PProxy) ManageLogging() {
	fmt.Printf("setting up message logging")
	file, err := os.OpenFile("message_log.csv", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	p.logFile = *file
	if nil != err {
		fmt.Printf("Unable to open logging file. %v", err)
		panic("unable to open logging file")
	}
	writer := bufio.NewWriter(&p.logFile)
	p.logWriter = *writer
	start := time.Now()
	for {
		item := <-p.logging
		switch item.(type) {
		case messageLog:
			message := item.(messageLog)
			elapsedMinutes := int(time.Since(start).Minutes())
			line := fmt.Sprintf("%d, %s, %t, %d, %s, %d\n", message.mtype, message.hash, message.received, message.time, message.target, elapsedMinutes)
			_, err := p.logWriter.Write([]byte(line))
			if nil != err {
				fmt.Printf("Error writing to logging file. %v", err)
				panic("Error writing to logging file")
			}
		default:
			fmt.Printf("Garbage on p.logging. %+v", item)
			break
		}
	}
	p.logWriter.Flush()
	defer p.logFile.Close()
}

// manageOutChannel takes messages from the f.broadcastOut channel and sends them to the network.
func (f *P2PProxy) ManageOutChannel() {
	for data := range f.BroadcastOut {
		switch data.(type) {
		case factomMessage:
			fmessage := data.(factomMessage)
			// Wrap it in a parcel and send it out channel ToNetwork.
			parcel := p2p.NewParcel(p2p.CurrentNetwork, fmessage.message)
			parcel.Header.Type = p2p.TypeMessage
			parcel.Header.TargetPeer = fmessage.peerHash
			p2p.BlockFreeChannelSend(f.ToNetwork, *parcel)
		default:
			fmt.Printf("Garbage on f.BrodcastOut. %+v", data)
		}
	}
}

// manageInChannel takes messages from the network and stuffs it in the f.BroadcastIn channel
func (f *P2PProxy) ManageInChannel() {
	for data := range f.FromNetwork {
		switch data.(type) {
		case p2p.Parcel:
			parcel := data.(p2p.Parcel)
			message := factomMessage{message: parcel.Payload, peerHash: parcel.Header.TargetPeer}
			p2p.BlockFreeChannelSend(f.BroadcastIn, message)
		default:
			fmt.Printf("Garbage on f.FromNetwork. %+v", data)
		}
	}
}

func (f *P2PProxy) PeriodicStatusReport(fnodes []*FactomNode) {
	time.Sleep(p2p.NetworkStatusInterval) // wait for things to spin up
	for {
		time.Sleep(p2p.NetworkStatusInterval)
		fmt.Println("\n\n\n")
		fmt.Println("-------------------------------------------------------------------------------")
		fmt.Println("-------------------------------------------------------------------------------")
		for _, f := range fnodes {
			f.State.Status = 1
		}
		time.Sleep(1 * time.Second)
		for _, f := range fnodes {
			fmt.Printf("%s \n\n", f.State.ShortString())
		}
		now := time.Now().Format("01/02/2006 15:04:05")
		listenTo := 0
		if listenTo >= 0 && listenTo < len(fnodes) {
			fmt.Printf("%s:\n", now)
			fmt.Printf("      InMsgQueue             %d\n", len(fnodes[listenTo].State.InMsgQueue()))
			fmt.Printf("      AckQueue               %d\n", len(fnodes[listenTo].State.AckQueue()))
			fmt.Printf("      MsgQueue               %d\n", len(fnodes[listenTo].State.MsgQueue()))
			fmt.Printf("      TimerMsgQueue          %d\n", len(fnodes[listenTo].State.TimerMsgQueue()))
			fmt.Printf("      NetworkOutMsgQueue     %d\n", len(fnodes[listenTo].State.NetworkOutMsgQueue()))
			fmt.Printf("      NetworkInvalidMsgQueue %d\n", len(fnodes[listenTo].State.NetworkInvalidMsgQueue()))
			fmt.Printf("      HoldingQueue           %d\n", len(fnodes[listenTo].State.Holding))
		}
		fmt.Printf("      ToNetwork Queue:       %d\n", len(f.ToNetwork))
		fmt.Printf("      FromNetwork Queue:     %d\n", len(f.FromNetwork))
		fmt.Printf("      BroadcastOut Queue:    %d\n", len(f.BroadcastOut))
		fmt.Printf("      BroadcastIn Queue:     %d\n", len(f.BroadcastIn))
		fmt.Println("-------------------------------------------------------------------------------")
		fmt.Println("-------------------------------------------------------------------------------")
	}
}

// const (
// 	EOM_MSG                       byte = iota // 0
// 	ACK_MSG                                   // 1
// 	FED_SERVER_FAULT_MSG                      // 2
// 	AUDIT_SERVER_FAULT_MSG                    // 3
// 	FULL_SERVER_FAULT_MSG                     // 4
// 	COMMIT_CHAIN_MSG                          // 5
// 	COMMIT_ENTRY_MSG                          // 6
// 	DIRECTORY_BLOCK_SIGNATURE_MSG             // 7
// 	EOM_TIMEOUT_MSG                           // 8
// 	FACTOID_TRANSACTION_MSG                   // 9
// 	HEARTBEAT_MSG                             // 10
// 	INVALID_ACK_MSG                           // 11
// 	INVALID_DIRECTORY_BLOCK_MSG               // 12

// 	REVEAL_ENTRY_MSG      // 13
// 	REQUEST_BLOCK_MSG     // 14
// 	SIGNATURE_TIMEOUT_MSG // 15
// 	MISSING_MSG           // 16
// 	MISSING_DATA          // 17
// 	DATA_RESPONSE         // 18
// 	MISSING_MSG_RESPONSE  //19

// 	DBSTATE_MSG          // 20
// 	DBSTATE_MISSING_MSG  // 21
// 	ADDSERVER_MSG        // 22
// 	CHANGESERVER_KEY_MSG // 23
// 	REMOVESERVER_MSG     // 24
// 	NEGOTIATION_MSG      // 25
// )

// var LoggingLevels = map[uint8]string{
// 	Silence:     "Silence",     // Say nothing. A log output with level "Silence" is ALWAYS printed.
// 	Significant: "Significant", // Significant things that should be printed, but aren't necessary errors.
// 	Fatal:       "Fatal",       // Log only fatal errors (fatal errors are always logged even on "Silence")
// 	Errors:      "Errors",      // Log all errors (many errors may be expected)
// 	Notes:       "Notes",       // Log notifications, usually significant events
// 	Debugging:   "Debugging",   // Log diagnostic info, pretty low level
// 	Verbose:     "Verbose",     // Log everything
// }
