// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bufio"
	"fmt"
	"os"
	"time"

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
	BroadcastOut chan factomMessage // ToNetwork from factomd
	BroadcastIn  chan factomMessage // FromNetwork for Factomd

	ToNetwork   chan p2p.Parcel // From p2pProxy to the p2p Controller
	FromNetwork chan p2p.Parcel // Parcels from the network for the application

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

func (f *P2PProxy) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	f.BroadcastOut = make(chan factomMessage, p2p.StandardChannelSize)
	f.BroadcastIn = make(chan factomMessage, p2p.StandardChannelSize)
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
	}
	if len(f.BroadcastOut) < 10000 {
		f.BroadcastOut <- message
	}
	return nil
}

// Non-blocking return value from channel.
func (f *P2PProxy) Recieve() (interfaces.IMsg, error) {
	select {
	case data, ok := <-f.BroadcastIn:
		if ok {
			msg, err := messages.UnmarshalMessage(data.message)
			if nil == err {
				msg.SetNetworkOrigin(data.peerHash)
			}
			if 1 < f.debugMode {
				f.logMessage(msg, true) // NODE_TALK_FIX
				fmt.Printf(".")
			}
			return msg, err
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

func (p *P2PProxy) startProxy() {
	if 1 < p.debugMode {
		go p.ManageLogging()
	}
	go p.ManageOutChannel() // Bridges between network format Parcels and factomd messages (incl. addressing to peers)
	go p.ManageInChannel()
}

// NODE_TALK_FIX
func (p *P2PProxy) stopProxy() {
	if 0 < p.debugMode {
		p.logging <- "stop"
	}
}

type messageLog struct {
	hash     string // string(GetMsgHash().Bytes())
	received bool   // true if logging a recieved message, false if sending
	time     int64
	target   string // the id of the targetted node (value may only have local meaning)
}

func (p *P2PProxy) logMessage(msg interfaces.IMsg, received bool) {
	if 1 < p.debugMode {
		hash := fmt.Sprintf("%x", msg.GetMsgHash().Bytes())
		time := time.Now().Unix()
		ml := messageLog{hash: hash, received: received, time: time}
		p.logging <- ml
	}
}

func (p *P2PProxy) ManageLogging() {
	note("setting up message logging")
	file, err := os.OpenFile("message_log.csv", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	p.logFile = *file
	if nil != err {
		note("Unable to open logging file. %v", err)
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
			line := fmt.Sprintf("%s, %t, %d, %s, %d\n", message.hash, message.received, message.time, message.target, elapsedMinutes)
			_, err := p.logWriter.Write([]byte(line))
			if nil != err {
				note("Error writing to logging file. %v", err)
				panic("Error writing to logging file")
			}
		default:
			break
		}
	}
	p.logWriter.Flush()
	defer p.logFile.Close()
}

// manageOutChannel takes messages from the f.broadcastOut channel and sends them to the network.
func (f *P2PProxy) ManageOutChannel() {
	for data := range f.BroadcastOut {
		// Wrap it in a parcel and send it out channel ToNetwork.
		parcel := p2p.NewParcel(p2p.CurrentNetwork, data.message)
		parcel.Header.Type = p2p.TypeMessage
		parcel.Header.TargetPeer = data.peerHash
		f.ToNetwork <- *parcel
	}
}

// manageInChannel takes messages from the network and stuffs it in the f.BroadcastIn channel
func (f *P2PProxy) ManageInChannel() {
	for data := range f.FromNetwork {
		message := factomMessage{message: data.Payload, peerHash: data.Header.TargetPeer}
		f.BroadcastIn <- message
	}
}

// ProxyStatusReport: Report the status of the peer channels
func (f *P2PProxy) ProxyStatusReport(fnodes []*FactomNode) {
	time.Sleep(time.Second * 3) // wait for things to spin up
	for {
		time.Sleep(time.Second * 6)
		listenTo := 0
		if listenTo >= 0 && listenTo < len(fnodes) {
			fmt.Printf("   %s\n", fnodes[listenTo].State.GetFactomNodeName())
		}
		now := time.Now().Format("01/02/2006 15:04:05")

		note("%s     ToNetwork Queue:   %d", now, len(f.ToNetwork))
		note("%s   FromNetwork Queue:   %d", now, len(f.FromNetwork))
		note("%s  BroadcastOut Queue:   %d", now, len(f.BroadcastOut))
		note("%s   BroadcastIn Queue:   %d", now, len(f.BroadcastIn))
	}
}

func PeriodicStatusReport(fnodes []*FactomNode) {
	time.Sleep(time.Second * 2) // wait for things to spin up
	for {
		time.Sleep(time.Second * 5)
		fmt.Println("-------------------------------------------------------------------------------")
		fmt.Println("-------------------------------------------------------------------------------")
		for _, f := range fnodes {
			f.State.Status = true
		}
		time.Sleep(100 * time.Millisecond)
		for _, f := range fnodes {
			fmt.Printf("%8s %s \n", f.State.FactomNodeName, f.State.ShortString())
		}

		listenTo := 0
		if listenTo >= 0 && listenTo < len(fnodes) {
			fmt.Printf("   %s\n", fnodes[listenTo].State.GetFactomNodeName())
			fmt.Printf("      InMsgQueue             %d\n", len(fnodes[listenTo].State.InMsgQueue()))
			fmt.Printf("      AckQueue               %d\n", len(fnodes[listenTo].State.AckQueue()))
			fmt.Printf("      MsgQueue               %d\n", len(fnodes[listenTo].State.MsgQueue()))
			fmt.Printf("      TimerMsgQueue          %d\n", len(fnodes[listenTo].State.TimerMsgQueue()))
			fmt.Printf("      NetworkOutMsgQueue     %d\n", len(fnodes[listenTo].State.NetworkOutMsgQueue()))
			fmt.Printf("      NetworkInvalidMsgQueue %d\n", len(fnodes[listenTo].State.NetworkInvalidMsgQueue()))
			fmt.Printf("      HoldingQueue          %d\n", len(fnodes[listenTo].State.Holding))
		}
		fmt.Println("-------------------------------------------------------------------------------")
		fmt.Println("-------------------------------------------------------------------------------")
	}
}

func note(format string, v ...interface{}) {
	fmt.Fprintln(os.Stdout, fmt.Sprintf("%d:", os.Getpid()), fmt.Sprintf(format, v...))
}
