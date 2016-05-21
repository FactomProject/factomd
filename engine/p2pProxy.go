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

	testMode  bool
	debugMode int
}

var _ interfaces.IPeer = (*P2PProxy)(nil)

func (f *P2PProxy) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	f.BroadcastOut = make(chan []byte, 10000)
	f.BroadcastIn = make(chan []byte, 10000)
	f.testMode = false // When this is false, factomd is connected to the network.  When true, network is isolated, and a heartbeat test message sent over the network.
	return f
}
func (f *P2PProxy) SetDebugMode(netdebug int) {

	f.debugMode = netdebug
}

func (f *P2PProxy) SetTestMode(test bool) {

	f.testMode = test
}
func (f *P2PProxy) GetNameFrom() string {
	return f.FromName
}
func (f *P2PProxy) GetNameTo() string {
	return f.ToName
}
func (f *P2PProxy) Send(msg interfaces.IMsg) error {
	if !f.testMode {
		// fmt.Printf("S")
		data, err := msg.MarshalBinary()
		if err != nil {
			die("Send error! %+v", err)
			return err
		}
		if len(f.BroadcastOut) < 10000 {
			f.BroadcastOut <- data
		}
	}
	return nil
}

// Non-blocking return value from channel.
func (f *P2PProxy) Recieve() (interfaces.IMsg, error) {
	if !f.testMode {
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

// Unused!
// // Returns the number of messages waiting to be read
func (f *P2PProxy) Len() int {
	//TODO IMPLEMENT JAYJAY
	fmt.Printf("P2PProxy.Len Not implemented.")
	// Sim Peer:
	//	return len(f.BroadcastIn)
	// Broadcase in is the Sim Peer channel.  We have a way to see how many TCP MEssages?
	return 1
}

//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%d:", os.Getpid()), fmt.Sprintf(format, v...))
	os.Exit(1)
}
func note(format string, v ...interface{}) {
	fmt.Fprintln(os.Stdout, fmt.Sprintf("%d:", os.Getpid()), fmt.Sprintf(format, v...))
}

///// BUS EXAMPLE

// BUGBUG TODO JAYJAY - switch to standard port, and read peers from peers.json.
// func P2PNetworkStart(address string, peers string, p2pProxy *P2PProxy) {
// 	var err error
// 	if p2pSocket, err = bus.NewSocket(); err != nil {
// 		die("P2PNetworkStart.NewSocket: %s", err)
// 	}
// 	p2pSocket.AddTransport(tcp.NewTransport())

// 	note("P2PNetworkStart- Start Listening on address: %s", address)

// 	// address := "tcp://127.0.0.1:40891"
// 	if err = p2pSocket.Listen(address); err != nil {
// 		die("P2PNetworkStart.Listen: %s", err.Error())
// 	}

// 	note("P2PNetworkStart- Sleep for a few seconds to let peers wake up.")
// 	// wait for everyone to start listening
// 	time.Sleep(time.Second)

// 	// Parse the peers into an array.
// 	parseFunc := func(c rune) bool {
// 		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
// 	}

// 	peerAddresses := strings.FieldsFunc(peers, parseFunc)
// 	note("P2PNetworkStart- our peers: %+v", peerAddresses)

// 	var x int
// 	for x = 0; x < len(peerAddresses); x++ {
// 		if err = p2pSocket.Dial(peerAddresses[x]); err != nil {
// 			note("P2PNetworkStart.Dial: %s", err.Error())
// 		}
// 	}

// 	// note("P2PNetworkStart- waiting for peers to connect")
// 	// time.Sleep(time.Second)

// 	// // BIG SWITCH between test code and factomd.  We switch which gets hooked up to channels
// 	if p2pProxy.testMode {
// 		go heartbeat(p2pProxy)
// 	}

// }

// func heartbeat(p2pProxy *P2PProxy) {
// 	beat := ""
// 	i := 0
// 	for {
// 		// for i := 0; i < 500; i++ {
// 		beat = fmt.Sprintf("Heartbeat FROM %d. Beat #%d", os.Getpid(), i)
// 		p2pProxy.BroadcastOut <- []byte(beat)
// 		select {
// 		case data, ok := <-p2pProxy.BroadcastIn:
// 			if ok {
// 				note("Recieved message: %s", string(data))
// 			}
// 		default:
// 		}
// 		time.Sleep(time.Millisecond * 200)
// 		i++
// 	}
// }

func (p *P2PProxy) startProxy() {
	go p.ManageOutChannel() // Bridges between network format Parcels and factomd messages (incl. addressing to peers)
	go p.ManageInChannel()
}

// Note: BUGBUG - the NetworkProcessorNet / state has a channel of bad messages "Bad message queue?"
// We need to rpocess these messages and get / give demerits.
// Paul says its ok to punt on this for the alpha
//

// this is a goroutine infinite loop
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

// this is a goroutine infinite loop
// manageInChannel takes messages from the network and stuffs it in the f.BroadcastIn channel
func (f *P2PProxy) ManageInChannel() {
	for data := range f.FromNetwork {
		// BUGBUG JAYJAY TODO Here is where you copy the connecton ID into the Factom message?
		message := data.Payload
		f.BroadcastIn <- message
	}
}

// // this is a goroutine infinite loop
// // manageOutChannel takes messages from the f.broadcastOut channel and sends them to the network.
// func (f *P2PProxy) ManageOutChannel() {
// 	for {
// 		select {
// 		case data, ok := <-f.BroadcastOut:
// 			if ok {
// 				f.sendP2P(data)
// 			}
// 		default:
// 		}
// 		// time.Sleep(time.Millisecond * 100)
// 	}
// }

// // this is a goroutine infinite loop
// // manageInChannel takes messages from the network and stuffs it in the f.BroadcastIn channel
// func (f *P2PProxy) ManageInChannel() {
// 	for {
// 		data := f.recieveP2P()
// 		f.BroadcastIn <- data
// 		// time.Sleep(time.Millisecond * 100)
// 	}
// }

// func (f *P2PProxy) sendP2P(msg []byte) {
// 	if err := p2pSocket.Send(msg); err != nil {
// 		note("sendP2P.Send ERROR: %s", err.Error())
// 	}
// }

// func (f *P2PProxy) recieveP2P() []byte {
// 	data, err := p2pSocket.Recv()
// 	if err != nil {
// 		note("recieveP2P.Recv ERROR: %s", err.Error())
// 	}
// 	// if f.debugMode {
// 	// 	note("recieveP2P.Recv Successfully got a message")
// 	// }
// 	return data
// }

// ProxyStatusReport: Report the status of the peer channels
func (f *P2PProxy) ProxyStatusReport() {
	time.Sleep(time.Second * 3) // wait for things to spin up
	for {
		time.Sleep(time.Second * 3)
		note("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&")
		note("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&")
		note("     ToNetwork Queue:   %d", len(f.ToNetwork))
		note("   FromNetwork Queue:   %d", len(f.FromNetwork))
		note("  BroadcastOut Queue:   %d", len(f.BroadcastOut))
		note("   BroadcastIn Queue:   %d", len(f.BroadcastIn))
		note("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&")
		note("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&")
	}
}

func PeriodicStatusReport(fnodes []*FactomNode) {
	time.Sleep(time.Second * 5) // wait for things to spin up
	for {
		time.Sleep(time.Second * 5)
		fmt.Println("-------------------------------------------------------------------------------")
		fmt.Println("-------------------------------------------------------------------------------")
		for _, f := range fnodes {
			fmt.Printf("%8s %s\n", f.State.FactomNodeName, f.State.ShortString())
		}
		listenTo := 0
		if listenTo >= 0 && listenTo < len(fnodes) {
			fmt.Printf("   %s\n", fnodes[listenTo].State.GetFactomNodeName())
			fmt.Printf("      InMsgQueue             %d\n", len(fnodes[listenTo].State.InMsgQueue()))
			fmt.Printf("      LeaderMsgQueue         %d\n", len(fnodes[listenTo].State.LeaderMsgQueue()))
			fmt.Printf("      TimerMsgQueue          %d\n", len(fnodes[listenTo].State.TimerMsgQueue()))
			fmt.Printf("      NetworkOutMsgQueue     %d\n", len(fnodes[listenTo].State.NetworkOutMsgQueue()))
			fmt.Printf("      NetworkInvalidMsgQueue %d\n", len(fnodes[listenTo].State.NetworkInvalidMsgQueue()))
		}
		fmt.Println("-------------------------------------------------------------------------------")
		fmt.Println("-------------------------------------------------------------------------------")
	}
}
