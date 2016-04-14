// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package p2p

import (
	"fmt"
	"os"
	"time"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pair"
	"github.com/go-mangos/mangos/transport/tcp"
)

var _ = fmt.Print

var (
	ServePort = 40891
)

const ( // iota is reset to 0
	server = iota // c0 == 0
	client = iota // c1 == 1
)

type P2PPeer struct {
	Socket   mangos.Socket
	ToName   string
	FromName string
}

// var _ interfaces.IPeer = (*P2PPeer)(nil)

// func (f *P2PPeer) Init(fromName, toName string) interfaces.IPeer {
// 	f.ToName = toName
// 	f.FromName = fromName
// 	return f
// }

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

// func (f *P2PPeer) GetNameFrom() string {
// 	return f.FromName
// }
// func (f *P2PPeer) GetNameTo() string {
// 	return f.ToName
// }

// func (f *P2PPeer) Send(msg interfaces.IMsg) error {
// 	if 1 == rand.Intn(send_freq) {
// 		fmt.Printf("%d -- P2PPeer.SEND %s -> %s\t %s \n", os.Getpid(), f.FromName, f.ToName, msg)
// 	}
// 	// fmt.Printf("P2PPeer.Send for:\n %+v\n\n", msg)

// 	data, err := msg.MarshalBinary()
// 	if err != nil {
// 		fmt.Printf("%d -- P2PPeer.Send !!!!!!!!!!!! FAILED TO MARSHALL BINARY for:\n %+v\n\n", os.Getpid(), msg)

// 		return err
// 	}

// 	if err = f.Socket.Send(data); err != nil {
// 		fmt.Printf("%d -- P2PPeer.Send error from f.Socket.Send(data) for:\n %+v\n\n", os.Getpid(), msg)
// 	}
// 	return err
// }

// // Non-blocking return value from channel.
// func (f *P2PPeer) Recieve() (interfaces.IMsg, error) {
// 	// if 1 == rand.Intn(recieve_freq) {
// 	// 	fmt.Printf("%d -- P2PPeer.RECIEVE %s -> %s\n", os.Getpid(), f.FromName, f.ToName)
// 	// }
// 	// 100ms Timeout
// 	f.Socket.SetOption(mangos.OptionRecvDeadline, 100*time.Millisecond)
// 	// Minimal blocking
// 	// f.Socket.SetOption(mangos.OptionRecvDeadline, 1*time.Millisecond)
// 	var data []byte
// 	var err error
// 	if data, err = f.Socket.Recv(); err == nil {
// 		// if len(data) > 0 {
// 		msg, err := messages.UnmarshalMessage(data)
// 		if nil == err {
// 			fmt.Printf("%d -- P2PPeer.Recieve $$$$$$$$$$$$ GOT VALID MESSAGE:\t %+v\n", os.Getpid(), msg)
// 		} else {
// 			fmt.Printf("%d -- P2PPeer.Recieve Got invalid MESSAGE:\t %+v\n", os.Getpid(), string(data))
// 			// }
// 			return msg, err
// 		}
// 	}
// 	beat := "HEARTBEAT"
// 	if err = f.Socket.Send([]byte(beat)); err != nil {
// 		fmt.Printf("%d -- P2PPeer.Connect ###### error from f.Socket.Send(data) for:\n %+v\n\n", os.Getpid(), beat)
// 	}

// 	return nil, nil
// }

// // Is this connection equal to parm connection
// func (f *P2PPeer) Equals(ff interfaces.IPeer) bool {
// 	f2, ok := ff.(*NetPeer)
// 	if !ok {
// 		return false
// 	} // Different peer type can't be equal

// 	// Check If this is another connection from here to there
// 	if f.FromName == f2.FromName && f.ToName == f2.FromName {
// 		return true
// 	}

// 	// Check if this is a connection from there to here
// 	if f.FromName == f2.ToName && f.ToName == f2.FromName {
// 		return true
// 	}
// 	return false
// }

// // Unused!
// // // Returns the number of messages waiting to be read
// func (f *P2PPeer) Len() int {
// 	//TODO IMPLEMENT JAYJAY
// 	fmt.Printf("P2PPeer.Len Not implemented.")
// 	// Sim Peer:
// 	//	return len(f.BroadcastIn)
// 	// Broadcase in is the Sim Peer channel.  We have a way to see how many TCP MEssages?
// 	return 1
// }

//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////

// Mangos example code below:
func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("%d -- ", os.Getpid()), fmt.Sprintf(format, v...))
	os.Exit(1)
}

func sendMessage(sock mangos.Socket, payload []byte) {
	fmt.Printf("%d: SENDING \"%s\"\n", os.Getpid(), string(payload))
	if err := sock.Send(payload); err != nil {
		die("failed sending: %s", err)
	}
}

func recvMessage(sock mangos.Socket) {
	var msg []byte
	var err error
	if msg, err = sock.Recv(); err == nil {
		fmt.Printf("%d -- RECEIVED: \"%s\"\n", os.Getpid(), string(msg))
	}
}

func sendRecv(sock mangos.Socket, name string) {
	for {
		sock.SetOption(mangos.OptionRecvDeadline, 100*time.Millisecond)
		recvMessage(sock)
		time.Sleep(time.Second)
		sendMessage(sock, []byte(name))
	}
}

func listen(url string) {
	var sock mangos.Socket
	var err error
	if sock, err = pair.NewSocket(); err != nil {
		die("can't get new pair socket: %s", err)
	}

	sock.AddTransport(tcp.NewTransport())
	if err = sock.Listen(url); err != nil {
		die("can't listen on pair socket: %s", err.Error())
	}
	sendRecv(sock, "Listener Says Hello")
}

func dial(url string) {
	var sock mangos.Socket
	var err error

	if sock, err = pair.NewSocket(); err != nil {
		die("can't get new pair socket: %s", err.Error())
	}
	// sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())
	if err = sock.Dial(url); err != nil {
		die("can't dial on pair socket: %s", err.Error())
	}
	sendRecv(sock, "Caller says hello")
}

// BUGBUG TODO JAYJAY - get rid of passing in leader, shouldn't matter.
func P2PNetworkStart(leader bool, address string) {
	// address := "tcp://127.0.0.1:40891"
	if leader {
		listen(address)
	} else {
		dial(address)
	}

	fmt.Fprintf(os.Stderr, "%d -- Usage: pair node0|node1 <URL>\n", os.Getpid())
	os.Exit(1)
}

// Thought process:
// X leader listens, follower connects.
// X Change message format to binary
// - Think about how to refactor this to be really peer connections.
// - Next step
//      -- we listen always on the given port (And we dial out to the peers we know about) (this requires we be probably in VMs)
//      -- no leadership awareness in p2p
// - Make this no longer an iPEer. Make proxy iPeer
// -  Setup Channels between the P2P network and the rest of the stuff.  Maybe an iPeer that talks over the
//      channel to the P2P network stuff, so that we have process isolation of some sort.
// -- Add simple discovery (maybe scan 192.168.1.1-192.168.1.256 for connections.
