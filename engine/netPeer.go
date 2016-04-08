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
	basePort = 9000
)

type NetPeer struct {
	Socket   mangos.Socket
	ToName   string
	FromName string
}

const ( // iota is reset to 0
	server = iota // c0 == 0
	client = iota // c1 == 1
)

var _ interfaces.IPeer = (*NetPeer)(nil)

// I hope this isn't needed.
// func (f *NetPeer) AddExistingConnection(conn mangos.f.Socketet) {
// 	f.f.Socketet = conn
// }

// Connect sets us up with a scoket connection, type indicates whether we dial in (as client) or listen (as server). address is the URL.
func (f *NetPeer) Connect(connectionType int, address string) error {
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

func (f *NetPeer) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	return f
}

// Setup f.Socketet acceptor on port
// func Listen(service, port int)  {
//     f.amServer = true
//     portStr := fmt.sprintf(":%d", port)
//     f.Listen, err := net.Listen(service, portStr)
//     if err != nil {
//         fmt.Fprintf("netPeer.Listen error setting up listener on %s service for port %s:\n %+v", service, portStr, err)
//     }
//     // This version assumes only one connection, so we take the first one we get.
//     conn, err := f.Listen.Accept()
//     if err != nil {
//         fmt.Fprintf("netPeer.Listen error in Listener.Accept() call on %s service for port %s:\n %+v", service, portStr, err)
//     }
//     f.Conn = conn
// }

func AddNetPeer(fnodes []*FactomNode, i1 int, i2 int) {
	// Ignore out of range, and connections to self.
	if i1 < 0 ||
		i2 < 0 ||
		i1 >= len(fnodes) ||
		i2 >= len(fnodes) ||
		i1 == i2 {
		return
	}

	// If the connection already exists, ignore
	for _, p1 := range fnodes[i1].Peers {
		for _, p2 := range fnodes[i2].Peers {
			if p1.Equals(p2) {
				return
			}
		}
	}

	if i1 >= len(fnodes) || i2 >= len(fnodes) {
		return
	}

	f1 := fnodes[i1]
	f2 := fnodes[i2]

	// Increment the port so every connection is on a differnet port
	basePort += 1

	fmt.Println("netPeer.AddNetPeer Connecting", f1.State.FactomNodeName, f2.State.FactomNodeName)

	peer12 := new(NetPeer).Init(f1.State.FactomNodeName, f2.State.FactomNodeName).(*NetPeer)
	peer21 := new(NetPeer).Init(f2.State.FactomNodeName, f1.State.FactomNodeName).(*NetPeer)

	// Mangos implementation:
	address := fmt.Sprintf("%s:%d", "tcp://127.0.0.1", basePort)
	fmt.Println("netPeer.AddNetPeer Connecting to address: ", address)

	peer12.Connect(server, address)
	peer21.Connect(client, address)

	// // 1->2
	// // peer 2 needs to set up listener:
	// go peer21.Listen("tcp", port2)
	// // peer 1 dials into peer 2:
	// address := fmt.Sprintf("%s:%s", host, port2)
	// err = peer12.ConnectTCP(address)
	// if err != nil {
	//     fmt.Fprintf("netPeer.AddNetPeer ################## Error connecting to: %+v", address)
	// }

	f1.Peers = append(f1.Peers, peer12)
	f2.Peers = append(f2.Peers, peer21)

	for _, p := range f1.Peers {
		fmt.Printf("%s's peer: %s\n", p.GetNameFrom(), p.GetNameTo())
	}

}

func (f *NetPeer) GetNameFrom() string {
	return f.FromName
}
func (f *NetPeer) GetNameTo() string {
	return f.ToName
}

func (f *NetPeer) Send(msg interfaces.IMsg) error {
	// fmt.Printf("netPeer.Send: %+v\n", msg)

	data, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	if err = f.Socket.Send(data); err != nil {
		fmt.Printf("netPeer.Send error from f.Socket.Send(data) for:\n %+v\n\n", msg)
	}
	return err
}

// Non-blocking return value from channel.
func (f *NetPeer) Recieve() (interfaces.IMsg, error) {
	var data []byte
	var err error
	if data, err = f.Socket.Recv(); err == nil {
		if len(data) > 0 {
			msg, err := messages.UnmarshalMessage(data)
			// fmt.Printf("netPeer.Recieve $$$$$$$$$$$$ GOT MESSAGE:\n %+v\n\n", msg)

			return msg, err
		}
	}

	return nil, nil
}

// Is this connection equal to parm connection
func (f *NetPeer) Equals(ff interfaces.IPeer) bool {
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

// Returns the number of messages waiting to be read
func (f *NetPeer) Len() int {
	//TODO IMPLEMENT JAYJAY
	fmt.Printf("NetPeer.Len Not implemented.")
	// Sim Peer:
	//	return len(f.BroadcastIn)
	// Broadcase in is the Sim Peer channel.  We have a way to see how many TCP MEssages?
	return 1
}
