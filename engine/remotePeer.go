// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"time"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pair"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

var _ = fmt.Print

var (
	ServePort = 9000
)

type RemotePeer struct {
	Socket   mangos.Socket
	ToName   string
	FromName string
}

// These are defined in netPeer:
// const ( // iota is reset to 0
// 	server = iota // c0 == 0
// 	client = iota // c1 == 1
// )

var _ interfaces.IPeer = (*RemotePeer)(nil)

// I hope this isn't needed.
// func (f *NetPeer) AddExistingConnection(conn mangos.f.Socketet) {
// 	f.f.Socketet = conn
// }

// Connect sets us up with a scoket connection, type indicates whether we dial in (as client) or listen (as server). address is the URL.
func (f *RemotePeer) Connect(connectionType int, address string) error {
	var err error
	err = nil

	if f.Socket, err = pair.NewSocket(); err != nil {
		fmt.Printf("netPeer.Connect error from pair.NewSocket() for %s :\n %+v\n\n", address, err)
	}
	f.Socket.AddTransport(ipc.NewTransport()) // ipc works on a single machine we want to at least simulate a full network connection.
	f.Socket.AddTransport(tcp.NewTransport())

	switch connectionType {
	case server:
		if err = f.Socket.Listen(address); err != nil {
			fmt.Printf("netPeer.Connect error from pair.Listen() for %s :\n %+v\n\n", address, err)
		} else {
			fmt.Printf("netPeer.Connect LISTENING ON for %s :\n", address)
		}

	case client:
		if err = f.Socket.Dial(address); err != nil {
			fmt.Printf("netPeer.Connect error from pair.Dial() for %s :\n %+v\n\n", address, err)
		} else {
			fmt.Printf("netPeer.Connect DIALED IN for %s :\n", address)
		}
	}
	// 100ms Timeout
	// f.Socket.SetOption(mangos.OptionRecvDeadline, 100*time.Millisecond)
	// Minimal blocking
	f.Socket.SetOption(mangos.OptionRecvDeadline, 1*time.Millisecond)

	return err
}

// func (f *NetPeer) ConnectTCP(address string) error {
// 	return f.Connect("tcp", address)
// }

// func (f *NetPeer) ConnectUDP(address string) error {
// 	return f.Connect("udp", address)
// }

func (f *RemotePeer) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	return f
}

// Serves on a default poort, incremented each time its called.
func RemoteServe(fnodes []*FactomNode) {
	// Increment the port so every connection is on a differnet port
	ServePort += 1

	RemoteServeOnPort(fnodes, ServePort)
}

// Serve:  Connects us to fnode 0 and starts listening on ServePort (which is incremented for each server started)
// Returns: Address on which we are serving.
func RemoteServeOnPort(fnodes []*FactomNode, port int) {
	f1 := fnodes[0]

	// Mangos implementation:
	address := fmt.Sprintf("%s:%d", "tcp://127.0.0.1", port)
	fmt.Println("RemotePeer.RemoteServe listening on address: ", address)

	peer := new(RemotePeer).Init(f1.State.FactomNodeName, address).(*RemotePeer)
	if err := peer.Connect(server, address); nil == err {
		f1.Peers = append(f1.Peers, peer)
	}
	for _, p := range f1.Peers {
		fmt.Printf("%s's peer: %s\n", p.GetNameFrom(), p.GetNameTo())
	}
}

// Connects:  Connects us to fnode 0 and dials out to address, creating a TCP connection
func RemoteConnect(fnodes []*FactomNode, address string) error {
	f1 := fnodes[0]

	// Mangos implementation:
	fmt.Printf("RemotePeer.RemoteConnect connecting to address: %s\n(should be in form of tcp://127.0.0.1:1234)\n", address)

	peer := new(RemotePeer).Init(f1.State.FactomNodeName, address).(*RemotePeer)
	if err := peer.Connect(client, address); nil == err {
		f1.Peers = append(f1.Peers, peer)
	} else {
		fmt.Printf("remotePeer.RemoteConnect: Failed to connect to remote peer at address: %s", address)
		return err
	}
	for _, p := range f1.Peers {
		fmt.Printf("%s's peer: %s\n", p.GetNameFrom(), p.GetNameTo())
	}
    return nil
}

func (f *RemotePeer) GetNameFrom() string {
	return f.FromName
}
func (f *RemotePeer) GetNameTo() string {
	return f.ToName
}

func (f *RemotePeer) Send(msg interfaces.IMsg) error {
	// fmt.Printf("RemotePeer.Send for:\n %+v\n\n", msg)

	data, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	if err = f.Socket.Send(data); err != nil {
		fmt.Printf("RemotePeer.Send error from f.Socket.Send(data) for:\n %+v\n\n", msg)
	}
	return err
}

// Non-blocking return value from channel.
func (f *RemotePeer) Recieve() (interfaces.IMsg, error) {
	var data []byte
	var err error
	if data, err = f.Socket.Recv(); err == nil {
		if len(data) > 0 {
			msg, err := messages.UnmarshalMessage(data)
			// fmt.Printf("RemotePeer.Recieve $$$$$$$$$$$$ GOT MESSAGE:\n %+v\n\n", msg)

			return msg, err
		}
	}

	return nil, nil
}

// Is this connection equal to parm connection
func (f *RemotePeer) Equals(ff interfaces.IPeer) bool {
	f2, ok := ff.(*NetPeer)
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
// func (f *RemotePeer) Len() int {
// 	//TODO IMPLEMENT JAYJAY
// 	fmt.Printf("RemotePeer.Len Not implemented.")
// 	// Sim Peer:
// 	//	return len(f.BroadcastIn)
// 	// Broadcase in is the Sim Peer channel.  We have a way to see how many TCP MEssages?
// 	return 1
// }
