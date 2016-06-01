// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
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
	BroadcastOut chan []byte // ToNetwork from factomd
	BroadcastIn  chan []byte // FromNetwork for Factomd

	ToNetwork   chan p2p.Parcel // Parcels from the application for us to route
	FromNetwork chan p2p.Parcel // Parcels from the network for the application

	debugMode int
}

var _ interfaces.IPeer = (*P2PProxy)(nil)

func (f *P2PProxy) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	f.BroadcastOut = make(chan []byte, 10000)
	f.BroadcastIn = make(chan []byte, 10000)
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
	data, err := msg.MarshalBinary()
	if err != nil {
		fmt.Println("ERROR on Send: ", err)
		return err
	}
	if len(f.BroadcastOut) < 10000 {
		f.BroadcastOut <- data
	}
	return nil
}

// Non-blocking return value from channel.
func (f *P2PProxy) Recieve() (interfaces.IMsg, error) {
	select {
	case data, ok := <-f.BroadcastIn:
		if ok {
			msg, err := messages.UnmarshalMessage(data)
			if 0 < f.debugMode {
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
	go p.ManageOutChannel() // Bridges between network format Parcels and factomd messages (incl. addressing to peers)
	go p.ManageInChannel()
}

// Note: BUGBUG - the NetworkProcessorNet / state has a channel of bad messages "Bad message queue?"
// We need to rpocess these messages and get / give demerits.
// Paul says its ok to punt on this for the alpha
//

// manageOutChannel takes messages from the f.broadcastOut channel and sends them to the network.
func (f *P2PProxy) ManageOutChannel() {
	for data := range f.BroadcastOut {
		// Wrap it in a parcel and send it out channel ToNetwork.
		parcel := p2p.NewParcel(p2p.CurrentNetwork, data)
		parcel.Header.Type = p2p.TypeMessage
		// BUGBUG JAYJAY TODO -- Load the target peer from the message, if there is one, int the parcel
		// so it can be sent as a directed message
		f.ToNetwork <- *parcel
	}
}

// manageInChannel takes messages from the network and stuffs it in the f.BroadcastIn channel
func (f *P2PProxy) ManageInChannel() {
	for data := range f.FromNetwork {
		// BUGBUG JAYJAY TODO Here is where you copy the connecton ID into the Factom message?
		message := data.Payload
		f.BroadcastIn <- message
	}
}

// ProxyStatusReport: Report the status of the peer channels
func (f *P2PProxy) ProxyStatusReport(fnodes []*FactomNode) {
	time.Sleep(time.Second * 3) // wait for things to spin up
	for {
		time.Sleep(time.Second * 20)
		listenTo := 0
		if listenTo >= 0 && listenTo < len(fnodes) {
			fmt.Printf("   %s\n", fnodes[listenTo].State.GetFactomNodeName())
		}
		note("     ToNetwork Queue:   %d", len(f.ToNetwork))
		note("   FromNetwork Queue:   %d", len(f.FromNetwork))
		note("  BroadcastOut Queue:   %d", len(f.BroadcastOut))
		note("   BroadcastIn Queue:   %d", len(f.BroadcastIn))
	}
}

func PeriodicStatusReport(fnodes []*FactomNode) {
	time.Sleep(time.Second * 2) // wait for things to spin up
	for {
		time.Sleep(time.Second * 15)
		fmt.Println("-------------------------------------------------------------------------------")
		fmt.Println("-------------------------------------------------------------------------------")
		for _, f := range fnodes {
			f.State.SetString()
			fmt.Printf("%8s %s\n", f.State.FactomNodeName, f.State.ShortString())
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
		}
		fmt.Println("-------------------------------------------------------------------------------")
		fmt.Println("-------------------------------------------------------------------------------")
	}
}

func note(format string, v ...interface{}) {
	fmt.Fprintln(os.Stdout, fmt.Sprintf("%d:", os.Getpid()), fmt.Sprintf(format, v...))
}
