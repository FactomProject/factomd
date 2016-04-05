// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"net"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

var _ = fmt.Print

var (
    basePort = 9000
)

// In the current pardigm Peer connections are one way, so while TCP sockets 
// are bidirectonal, each IPeer is either listening or recieving.

type NetPeer struct {
    bool    amServer 
	Conn    net.Conn
    Listen  net.Listener
	ToName   string
	FromName string
}

var _ interfaces.IPeer = (*NetPeer)(nil)

func (f *NetPeer) AddExistingConnection(conn net.Conn) {
	f.Conn = conn
}

func (f *NetPeer) Connect(service, address string) error {
	c, err := net.Dial(service, address)
	if err != nil {
		return err
	}
	f.Conn = c
	return nil
}

func (f *NetPeer) ConnectTCP(address string) error {
	return f.Connect("tcp", address)
}

func (f *NetPeer) ConnectUDP(address string) error {
	return f.Connect("udp", address)
}

func (f *NetPeer) Init(fromName, toName string) *NetPeer { // interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
    f.amServer = false
	return f
}

// Setup socket acceptor on port
func Listen(service, port int)  {
    f.amServer = true
    portStr := fmt.sprintf(":%d", port)
    f.Listen, err := net.Listen(service, portStr)
    if err != nil {
        fmt.Fprintf("netPeer.Listen error setting up listener on %s service for port %s:\n %+v", service, portStr, err)
    }
    // This version assumes only one connection, so we take the first one we get.
    conn, err := f.Listen.Accept()
    if err != nil {
        fmt.Fprintf("netPeer.Listen error in Listener.Accept() call on %s service for port %s:\n %+v", service, portStr, err)
    }
    f.Conn = conn
}
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

    port1 := basePort + i1
    port2 := basePort + i2
    
	fmt.Println("netPeer.AddNetPeer Connecting", f1.State.FactomNodeName, f2.State.FactomNodeName)

	peer12 := new(NetPeer).Init(f1.State.FactomNodeName, f2.State.FactomNodeName).(*NetPeer)
	peer21 := new(NetPeer).Init(f2.State.FactomNodeName, f1.State.FactomNodeName).(*NetPeer)
    
    // 1->2
    // peer 2 needs to set up listener:
    go peer21.Listen("tcp", port2)
    // peer 1 dials into peer 2:
    address := fmt.Sprintf("%s:%s", host, port2)
    err = peer12.ConnectTCP(address)
	if err != nil {
        fmt.Fprintf("netPeer.AddNetPeer ################## Error connecting to: %+v", address)
	}

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
	data, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = f.Conn.Write(data)
	return err
}

// Non-blocking return value from channel.
func (f *NetPeer) Recieve() (interfaces.IMsg, error) {
	data := make([]byte, 5000)

	n, err := f.Conn.Read(data)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		msg, err := messages.UnmarshalMessage(data)
		return msg, err
	}
	return nil, nil
}

// Is this connection equal to parm connection
func (f *NetPeer) Equals(IPeer peer) bool {
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
}

// type IPeer interface {
// 	Init(nameTo, nameFrom string) IPeer // Name of peer
// 	GetNameTo() string                  // Return the name of the peer
// 	GetNameFrom() string                // Return the name of the peer
// 	Send(IMsg) error                    // Send a message to this peer
// 	Recieve() (IMsg, error)             // Recieve a message from this peer; nil if no message is ready.
// 	Len() int                           // Returns the number of messages waiting to be read
// 	Equals(IPeer) bool                  // Is this connection equal to parm connection
// }
