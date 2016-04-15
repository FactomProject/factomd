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
}

var _ interfaces.IPeer = (*P2PPeer)(nil)

func (f *P2PPeer) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	return f
}

func (f *P2PPeer) GetNameFrom() string {
	return f.FromName
}
func (f *P2PPeer) GetNameTo() string {
	return f.ToName
}
func (f *P2PPeer) Send(msg interfaces.IMsg) error {
	data, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	f.BroadcastOut <- data
	return nil
}

// Non-blocking return value from channel.
func (f *P2PPeer) Recieve() (interfaces.IMsg, error) {
	select {
	case data, ok := <-f.BroadcastIn:
		if ok {
			msg, err := messages.UnmarshalMessage(data)
			return msg, err
		}
	default:
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
func P2PNetworkStart(address string, peers string) {
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

	note("P2PNetworkStart- waiting for peers to connect")
	time.Sleep(time.Second)
	note("P2PNetworkStart- spawning heartbeat")
	go heartbeat(address)

}

func heartbeat(address string) {
	beat := ""
	for i := 0; i < 500; i++ {
		beat = fmt.Sprintf("Heartbeat FROM %s. Beat #%d", address, i)
		sendP2P([]byte(beat))
		recieveP2P()
		i += i
		time.Sleep(time.Second)
	}
}

func sendP2P(msg []byte) {
	if err := p2pSocket.Send(msg); err != nil {
		note("sendP2P.Send ERROR: %s", err.Error())
	}
}

func recieveP2P() []byte {
	data, err := p2pSocket.Recv()
	if err != nil {
		note("recieveP2P.Recv ERROR: %s", err.Error())
	} else {
		note("recieveP2P RECEIVED \"%s\"", string(data))
	}
	return data
}

// Thought process:
// X leader listens, follower connects.
// X Change message format to binary
// X Make this file P2PPeer  and make it work like iPeer
// X we listen always on the given port (And we dial out to the peers we know about) (this requires we be probably in VMs)
// X no leadership awareness in p2p
// X Go back to Pauls' Send/Recieve from SimPeer
// X Verify sample code heart beat
// X Split out the send and recieve functions from sample code (no channels)
// X Verify heartbeat still works
// -- Make the send and recieve functions from run as goroutines and work on channels (STILL WITH HeARTBEAT SAMPLE CODE)
// -- Switch the channels over to the ones that P2PPeer uses (copied from simpeers)

// -- Split out the P@PNetworkStart and Send/Recoeve into a P2PNetowrk File

// Add a config file in .factom (peers.json?) and read it for a list of peers to connect to.

// - Make this no longer an iPEer. Make proxy iPeer
// -  Setup Channels between the P2P network and the rest of the stuff.  Maybe an iPeer that talks over the
//      channel to the P2P network stuff, so that we have process isolation of some sort.
// -- Add simple discovery (maybe scan 192.168.1.1-192.168.1.256 for connections.
