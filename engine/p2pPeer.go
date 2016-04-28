// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/bus"
	"github.com/go-mangos/mangos/transport/tcp"
)

var _ = fmt.Print

var (
	p2pSocket mangos.Socket // BUGBUG JAYJAY TODO This is a global. This needs to migrate to p2p package.
)

type P2PPeer struct {
	// A connection to this node:
	ToName   string
	FromName string
	// Channels that define the connection:
	BroadcastOut chan []byte
	BroadcastIn  chan []byte
	testMode     bool
	debugMode    bool
}

var _ interfaces.IPeer = (*P2PPeer)(nil)

func (f *P2PPeer) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	f.BroadcastOut = make(chan []byte, 10000)
	f.BroadcastIn = make(chan []byte, 10000)
	f.testMode = false // When this is false, factomd is connected to the network.  When true, network is isolated, and a heartbeat test message sent over the network.
	return f
}
func (f *P2PPeer) SetDebugMode(netdebug bool) {

	f.debugMode = netdebug
}

func (f *P2PPeer) SetTestMode(test bool) {

	f.testMode = test
}
func (f *P2PPeer) GetNameFrom() string {
	return f.FromName
}
func (f *P2PPeer) GetNameTo() string {
	return f.ToName
}
func (f *P2PPeer) Send(msg interfaces.IMsg) error {
	if !f.testMode {
		// fmt.Printf("S")
		data, err := msg.MarshalBinary()
		if err != nil {
			die("Send error! %+v", err)
			return err
		}
		if len(f.BroadcastOut) < 1000 {
			f.BroadcastOut <- data
		}
	}
	return nil
}

// Non-blocking return value from channel.
func (f *P2PPeer) Recieve() (interfaces.IMsg, error) {
	if !f.testMode {
		select {
		case data, ok := <-f.BroadcastIn:
			if ok {
				msg, err := messages.UnmarshalMessage(data)
				if f.debugMode {
					fmt.Printf(".")
					// m, _ := msg.JSONString()
					// note("Recieve Successfully got a message %s", m)
				}
				return msg, err
			}
		default:
		}
	}
	return nil, nil
}

// Is this connection equal to parm connection
func (f *P2PPeer) Equals(ff interfaces.IPeer) bool {
	f2, ok := ff.(*P2PPeer)
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
func (f *P2PPeer) Len() int {
	//TODO IMPLEMENT JAYJAY
	fmt.Printf("P2PPeer.Len Not implemented.")
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
func P2PNetworkStart(address string, peers string, p2pProxy *P2PPeer) {
	var err error
	if p2pSocket, err = bus.NewSocket(); err != nil {
		die("P2PNetworkStart.NewSocket: %s", err)
	}
	p2pSocket.AddTransport(tcp.NewTransport())

	note("P2PNetworkStart- Start Listening on address: %s", address)

	// address := "tcp://127.0.0.1:40891"
	if err = p2pSocket.Listen(address); err != nil {
		die("P2PNetworkStart.Listen: %s", err.Error())
	}

	note("P2PNetworkStart- Sleep for a few seconds to let peers wake up.")
	// wait for everyone to start listening
	time.Sleep(time.Second)

	// Parse the peers into an array.
	parseFunc := func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c) && !unicode.IsPunct(c)
	}

	peerAddresses := strings.FieldsFunc(peers, parseFunc)
	note("P2PNetworkStart- our peers: %+v", peerAddresses)

	var x int
	for x = 0; x < len(peerAddresses); x++ {
		if err = p2pSocket.Dial(peerAddresses[x]); err != nil {
			note("P2PNetworkStart.Dial: %s", err.Error())
		}
	}

	// note("P2PNetworkStart- waiting for peers to connect")
	// time.Sleep(time.Second)

	go p2pProxy.ManageOutChannel()
	go p2pProxy.ManageInChannel()

	// // BIG SWITCH between test code and factomd.  We switch which gets hooked up to channels
	if p2pProxy.testMode {
		go heartbeat(p2pProxy)
	}

}

func heartbeat(p2pProxy *P2PPeer) {
	beat := ""
	i := 0
	for {
		// for i := 0; i < 500; i++ {
		beat = fmt.Sprintf("Heartbeat FROM %d. Beat #%d", os.Getpid(), i)
		p2pProxy.BroadcastOut <- []byte(beat)
		select {
		case data, ok := <-p2pProxy.BroadcastIn:
			if ok {
				note("Recieved message: %s", string(data))
			}
		default:
		}
		time.Sleep(time.Millisecond * 200)
		i++
	}
}

// this is a goroutine infinite loop
// manageOutChannel takes messages from the f.broadcastOut channel and sends them to the network.
func (f *P2PPeer) ManageOutChannel() {
	for {
		select {
		case data, ok := <-f.BroadcastOut:
			if ok {
				f.sendP2P(data)
			}
		default:
		}
		// time.Sleep(time.Millisecond * 100)
	}
}

// this is a goroutine infinite loop
// manageInChannel takes messages from the network and stuffs it in the f.BroadcastIn channel
func (f *P2PPeer) ManageInChannel() {
	for {
		data := f.recieveP2P()
		f.BroadcastIn <- data
		// time.Sleep(time.Millisecond * 100)
	}
}

func (f *P2PPeer) sendP2P(msg []byte) {
	if err := p2pSocket.Send(msg); err != nil {
		note("sendP2P.Send ERROR: %s", err.Error())
	}
}

func (f *P2PPeer) recieveP2P() []byte {
	data, err := p2pSocket.Recv()
	if err != nil {
		note("recieveP2P.Recv ERROR: %s", err.Error())
	}
	// if f.debugMode {
	// 	note("recieveP2P.Recv Successfully got a message")
	// }
	return data
}

func PeriodicStatusReport(fnodes []*FactomNode) {
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
