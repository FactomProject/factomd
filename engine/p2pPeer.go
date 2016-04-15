// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/bus"
	"github.com/go-mangos/mangos/transport/tcp"
)

var _ = fmt.Print

var (
	p2pSocket mangos.Socket // BUGBUG JAYJAY TODO This is a global. This needs to migrate to p2p package.
	// Frequency of issuing debug print statements in netowkr code-- 2 = %100, 100 = %1 of the time.
	send_freq    = 2
	recieve_freq = 2
)

type P2PPeer struct {
	ToName   string
	FromName string
}

var _ interfaces.IPeer = (*P2PPeer)(nil)

func (f *P2PPeer) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	return f
}

// I hope this isn't needed.
// func (f *NetPeer) AddExistingConnection(conn mangos.f.Socketet) {
// 	f.f.Socketet = conn
// }

// // Serves on a default poort, incremented each time its called.
// func RemoteServe(fnodes []*FactomNode) {
// 	// Increment the port so every connection is on a differnet port
// 	ServePort += 1
// 	fmt.Printf("%d -- $$$$$$$$$$$$$$$$$$$$$$$$$$ P2PPeer.RemoteServe port: %d \n", os.Getpid(), ServePort)
// 	RemoteServeOnPort(fnodes, ServePort)
// }

// // Serve:  Connects us to fnode 0 and starts listening on ServePort (which is incremented for each server started)
// // Returns: Address on which we are serving.
// func RemoteServeOnPort(fnodes []*FactomNode, port int) {
// 	f1 := fnodes[0]
// 	fmt.Printf("%d -- P2PPeer.RemoteServeOnPort CHECKPOINT ENTRY\n", os.Getpid())
// 	// Mangos implementation:
// 	address := fmt.Sprintf("%s:%d", "tcp://127.0.0.1", port)
// 	fmt.Printf("%d -- P2PPeer.RemoteServeOnPort listening on address: %s \n", os.Getpid(), address)

// 	peer := new(P2PPeer).Init(f1.State.FactomNodeName, address).(*P2PPeer)
// 	fmt.Printf("%d -- P2PPeer.RemoteServeOnPort CHECKPOINT ALPO\n", os.Getpid())
// 	if err := peer.Connect(server, address); nil == err {
// 		fmt.Printf("%d -- P2PPeer.RemoteServeOnPort CHECKPOINT CHEX\n", os.Getpid())
// 		f1.Peers = append(f1.Peers, peer)
// 		fmt.Printf("%d -- P2PPeer.RemoteServeOnPort CHECKPOINT LIFE\n", os.Getpid())
// 	}
// 	// fmt.Printf("%d -- peers: %+v\n", os.Getpid(), f1.Peers)
// }

// // Connects:  Connects us to fnode 0 and dials out to address, creating a TCP connection
// func RemoteConnect(fnodes []*FactomNode, address string) error {
// 	f1 := fnodes[0]

// 	// Mangos implementation:
// 	fmt.Printf("%d -- P2PPeer.RemoteConnect connecting to address: %s\n(should be in form of tcp://127.0.0.1:1234)\n", os.Getpid(), address)

// 	peer := new(P2PPeer).Init(f1.State.FactomNodeName, address).(*P2PPeer)
// 	fmt.Printf("%d -- P2PPeer.RemoteConnect CHECKPOINT BETA\n", os.Getpid())

// 	if err := peer.Connect(client, address); nil == err {
// 		fmt.Printf("%d -- P2PPeer.RemoteConnect CHECKPOINT KAPPA\n", os.Getpid())

// 		f1.Peers = append(f1.Peers, peer)
// 		fmt.Printf("%d -- P2PPeer.RemoteConnect CHECKPOINT GAMMA\n", os.Getpid())

// 	} else {
// 		fmt.Printf("%d -- P2PPeer.RemoteConnect: Failed to connect to remote peer at address: %s\n", os.Getpid(), address)
// 		return err
// 	}
// 	fmt.Printf("%d -- P2PPeer.RemoteConnect CHECKPOINT LEAVING\n", os.Getpid())

// 	// fmt.Printf("%d -- peers: %+v\n", os.Getpid(), f1.Peers)
// 	return nil
// }

// // Connect sets us up with a scoket connection, type indicates whether we dial in (as client) or listen (as server). address is the URL.
// func (f *P2PPeer) Connect(connectionType int, address string) error {
// 	var err error
// 	err = nil

// 	if f.Socket, err = pair.NewSocket(); err != nil {
// 		fmt.Printf("%d -- P2PPeer.Connect error from pair.NewSocket() for %s :\n %+v\n\n", os.Getpid(), address, err)
// 	}
// 	// f.Socket.AddTransport(ipc.NewTransport()) // ipc works on a single machine we want to at least simulate a full network connection.
// 	f.Socket.AddTransport(tcp.NewTransport())

// 	switch connectionType {
// 	case server:
// 		if err = f.Socket.Listen(address); err != nil {
// 			fmt.Printf("%d -- P2PPeer.Connect error from pair.Listen() for %s :\n %+v\n\n", os.Getpid(), address, err)
// 		} else {
// 			fmt.Printf("%d -- P2PPeer.Connect LISTENING ON for %s :\n", os.Getpid(), address)
// 		}

// 	case client:
// 		if err = f.Socket.Dial(address); err != nil {
// 			fmt.Printf("%d -- P2PPeer.Connect error from pair.Dial() for %s :\n %+v\n\n", os.Getpid(), address, err)
// 		} else {
// 			fmt.Printf("%d -- P2PPeer.Connect DIALED IN for %s :\n", os.Getpid(), address)
// 			msg := "HEARTBEAT"
// 			if err = f.Socket.Send([]byte(msg)); err != nil {
// 				fmt.Printf("%d -- P2PPeer.Connect ###### error from f.Socket.Send(data) for:\n %+v\n\n", os.Getpid(), msg)
// 			}
// 		}
// 	}
// 	fmt.Printf("%d -- P2PPeer.Connect CHECKPOINT ZETA\n", os.Getpid())
// 	return err
// }

func (f *P2PPeer) GetNameFrom() string {
	return f.FromName
}
func (f *P2PPeer) GetNameTo() string {
	return f.ToName
}

func (f *P2PPeer) Send(msg interfaces.IMsg) error {
	if 1 == rand.Intn(send_freq) {
		fmt.Printf("%d -- P2PPeer.SEND %s -> %s\t %s \n", os.Getpid(), f.FromName, f.ToName, msg)
	}
	// fmt.Printf("P2PPeer.Send for:\n %+v\n\n", msg)

	// data, err := msg.MarshalBinary()
	// if err != nil {
	// 	fmt.Printf("%d -- P2PPeer.Send !!!!!!!!!!!! FAILED TO MARSHALL BINARY for:\n %+v\n\n", os.Getpid(), msg)

	// 	return err
	// }

	// if err = f.Socket.Send(data); err != nil {
	// 	fmt.Printf("%d -- P2PPeer.Send error from f.Socket.Send(data) for:\n %+v\n\n", os.Getpid(), msg)
	// }
	// return err

	return nil
}

// Non-blocking return value from channel.
func (f *P2PPeer) Recieve() (interfaces.IMsg, error) {
	if 1 == rand.Intn(recieve_freq) {
		fmt.Printf("%d -- P2PPeer.RECIEVE %s -> %s\n", os.Getpid(), f.FromName, f.ToName)
	}
	// 100ms Timeout
	// f.Socket.SetOption(mangos.OptionRecvDeadline, 100*time.Millisecond)
	// // Minimal blocking
	// // f.Socket.SetOption(mangos.OptionRecvDeadline, 1*time.Millisecond)
	// var data []byte
	// var err error
	// if data, err = f.Socket.Recv(); err == nil {
	// 	// if len(data) > 0 {
	// 	msg, err := messages.UnmarshalMessage(data)
	// 	if nil == err {
	// 		fmt.Printf("%d -- P2PPeer.Recieve $$$$$$$$$$$$ GOT VALID MESSAGE:\t %+v\n", os.Getpid(), msg)
	// 	} else {
	// 		fmt.Printf("%d -- P2PPeer.Recieve Got invalid MESSAGE:\t %+v\n", os.Getpid(), string(data))
	// 		// }
	// 		return msg, err
	// 	}
	// }
	// beat := "HEARTBEAT"
	// if err = f.Socket.Send([]byte(beat)); err != nil {
	// 	fmt.Printf("%d -- P2PPeer.Connect ###### error from f.Socket.Send(data) for:\n %+v\n\n", os.Getpid(), beat)
	// }

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

// Mangos example code below:

///// PAIR EXAMPLE:

// func sendMessage(sock mangos.Socket, payload []byte) {
// 	fmt.Printf("%d: SENDING \"%s\"\n", os.Getpid(), string(payload))
// 	if err := sock.Send(payload); err != nil {
// 		die("failed sending: %s", err)
// 	}
// }

// func recvMessage(sock mangos.Socket) {
// 	var msg []byte
// 	var err error
// 	if msg, err = sock.Recv(); err == nil {
// 		fmt.Printf("%d -- RECEIVED: \"%s\"\n", os.Getpid(), string(msg))
// 	}
// }

// func heartbeat(sock mangos.Socket, name string) {
// 	for {
// 		sock.SetOption(mangos.OptionRecvDeadline, 100*time.Millisecond)
// 		recvMessage(sock)
// 		time.Sleep(time.Second)
// 		sendMessage(sock, []byte(name))
// 	}
// }

// func listen(url string) {
// 	var err error
// 	if err = p2pSocket.Listen(url); err != nil {
// 		die("can't listen on pair socket: %s", err.Error())
// 	}
// 	go heartbeat(p2pSocket, "Listener Says Hello")
// }

// func dial(url string) {
// 	var err error

// 	// sock.AddTransport(ipc.NewTransport())
// 	p2pSocket.AddTransport(tcp.NewTransport())
// 	if err = p2pSocket.Dial(url); err != nil {
// 		die("can't dial on pair socket: %s", err.Error())
// 	}
// 	go heartbeat(p2pSocket, "Caller says hello")
// }
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
	time.Sleep(time.Second * 3)

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
	var msg []byte
	var err error
	for i := 0; i < 500; i++ {
		beat = fmt.Sprintf("Heartbeat FROM %s. Beat #%d", address, i)
		// note("%s: SENDING >>>>>>>>>> '%s' ONTO BUS\n", address, beat)
		if err = p2pSocket.Send([]byte(beat)); err != nil {
			note("heartbeat.Send ERROR: %s", err.Error())
		}
		if msg, err = p2pSocket.Recv(); err != nil {
			note("heartbeat.Recv ERROR: %s", err.Error())
		} else {
			note("RECEIVED \"%s\" FROM BUS", string(msg))
		}
		i += i
		time.Sleep(time.Second * 1)
	}
}

// Thought process:
// X leader listens, follower connects.
// X Change message format to binary
// X Make this file P2PPeer  and make it work like iPeer
// -- Split out the P@PNetworkStart and Send/Recoeve into a P2PNetowrk File
// XX we listen always on the given port (And we dial out to the peers we know about) (this requires we be probably in VMs)
// XX no leadership awareness in p2p

// Add a config file in .factom (peers.json?) and read it for a list of peers to connect to.

// - Make this no longer an iPEer. Make proxy iPeer
// -  Setup Channels between the P2P network and the rest of the stuff.  Maybe an iPeer that talks over the
//      channel to the P2P network stuff, so that we have process isolation of some sort.
// -- Add simple discovery (maybe scan 192.168.1.1-192.168.1.256 for connections.
