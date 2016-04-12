// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pair"
	"github.com/go-mangos/mangos/transport/tcp"
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

var _ interfaces.IPeer = (*RemotePeer)(nil)

func (f *RemotePeer) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	return f
}

// I hope this isn't needed.
// func (f *NetPeer) AddExistingConnection(conn mangos.f.Socketet) {
// 	f.f.Socketet = conn
// }

// Serves on a default poort, incremented each time its called.
func RemoteServe(fnodes []*FactomNode) {
	// Increment the port so every connection is on a differnet port
	ServePort += 1
	fmt.Printf("%d -- $$$$$$$$$$$$$$$$$$$$$$$$$$ RemotePeer.RemoteServe port: %d \n", os.Getpid(), ServePort)
	RemoteServeOnPort(fnodes, ServePort)
}

// Serve:  Connects us to fnode 0 and starts listening on ServePort (which is incremented for each server started)
// Returns: Address on which we are serving.
func RemoteServeOnPort(fnodes []*FactomNode, port int) {
	f1 := fnodes[0]
	fmt.Printf("%d -- RemotePeer.RemoteServeOnPort CHECKPOINT ENTRY\n", os.Getpid())
	// Mangos implementation:
	address := fmt.Sprintf("%s:%d", "tcp://127.0.0.1", port)
	fmt.Printf("%d -- RemotePeer.RemoteServeOnPort listening on address: %s \n", os.Getpid(), address)

	peer := new(RemotePeer).Init(f1.State.FactomNodeName, address).(*RemotePeer)
	fmt.Printf("%d -- RemotePeer.RemoteServeOnPort CHECKPOINT ALPO\n", os.Getpid())
	if err := peer.Connect(server, address); nil == err {
		fmt.Printf("%d -- RemotePeer.RemoteServeOnPort CHECKPOINT CHEX\n", os.Getpid())
		f1.Peers = append(f1.Peers, peer)
		fmt.Printf("%d -- RemotePeer.RemoteServeOnPort CHECKPOINT LIFE\n", os.Getpid())
	}
	// fmt.Printf("%d -- peers: %+v\n", os.Getpid(), f1.Peers)
}

// Connects:  Connects us to fnode 0 and dials out to address, creating a TCP connection
func RemoteConnect(fnodes []*FactomNode, address string) error {
	f1 := fnodes[0]

	// Mangos implementation:
	fmt.Printf("%d -- RemotePeer.RemoteConnect connecting to address: %s\n(should be in form of tcp://127.0.0.1:1234)\n", os.Getpid(), address)

	peer := new(RemotePeer).Init(f1.State.FactomNodeName, address).(*RemotePeer)
	fmt.Printf("%d -- RemotePeer.RemoteConnect CHECKPOINT BETA\n", os.Getpid())

	if err := peer.Connect(client, address); nil == err {
		fmt.Printf("%d -- RemotePeer.RemoteConnect CHECKPOINT KAPPA\n", os.Getpid())

		f1.Peers = append(f1.Peers, peer)
		fmt.Printf("%d -- RemotePeer.RemoteConnect CHECKPOINT GAMMA\n", os.Getpid())

	} else {
		fmt.Printf("%d -- remotePeer.RemoteConnect: Failed to connect to remote peer at address: %s\n", os.Getpid(), address)
		return err
	}
	fmt.Printf("%d -- RemotePeer.RemoteConnect CHECKPOINT LEAVING\n", os.Getpid())

	// fmt.Printf("%d -- peers: %+v\n", os.Getpid(), f1.Peers)
	return nil
}

// Connect sets us up with a scoket connection, type indicates whether we dial in (as client) or listen (as server). address is the URL.
func (f *RemotePeer) Connect(connectionType int, address string) error {
	var err error
	err = nil

	if f.Socket, err = pair.NewSocket(); err != nil {
		fmt.Printf("%d -- RemotePeer.Connect error from pair.NewSocket() for %s :\n %+v\n\n", os.Getpid(), address, err)
	}
	// f.Socket.AddTransport(ipc.NewTransport()) // ipc works on a single machine we want to at least simulate a full network connection.
	f.Socket.AddTransport(tcp.NewTransport())

	switch connectionType {
	case server:
		if err = f.Socket.Listen(address); err != nil {
			fmt.Printf("%d -- RemotePeer.Connect error from pair.Listen() for %s :\n %+v\n\n", os.Getpid(), address, err)
		} else {
			fmt.Printf("%d -- RemotePeer.Connect LISTENING ON for %s :\n", os.Getpid(), address)
		}

	case client:
		if err = f.Socket.Dial(address); err != nil {
			fmt.Printf("%d -- RemotePeer.Connect error from pair.Dial() for %s :\n %+v\n\n", os.Getpid(), address, err)
		} else {
			fmt.Printf("%d -- RemotePeer.Connect DIALED IN for %s :\n", os.Getpid(), address)
			msg := "HEARTBEAT"
			if err = f.Socket.Send([]byte(msg)); err != nil {
				fmt.Printf("%d -- RemotePeer.Connect ###### error from f.Socket.Send(data) for:\n %+v\n\n", os.Getpid(), msg)
			}
		}
	}
	fmt.Printf("%d -- RemotePeer.Connect CHECKPOINT ZETA\n", os.Getpid())
	return err
}

func (f *RemotePeer) GetNameFrom() string {
	return f.FromName
}
func (f *RemotePeer) GetNameTo() string {
	return f.ToName
}

func (f *RemotePeer) Send(msg interfaces.IMsg) error {
	if 1 == rand.Intn(send_freq) {
		fmt.Printf("%d -- RemotePeer.SEND %s -> %s\t %s \n", os.Getpid(), f.FromName, f.ToName, msg)
	}
	// fmt.Printf("RemotePeer.Send for:\n %+v\n\n", msg)

	data, err := msg.MarshalBinary()
	if err != nil {
		fmt.Printf("%d -- RemotePeer.Send !!!!!!!!!!!! FAILED TO MARSHALL BINARY for:\n %+v\n\n", os.Getpid(), msg)

		return err
	}

	if err = f.Socket.Send(data); err != nil {
		fmt.Printf("%d -- RemotePeer.Send error from f.Socket.Send(data) for:\n %+v\n\n", os.Getpid(), msg)
	}
	return err
}

// Non-blocking return value from channel.
func (f *RemotePeer) Recieve() (interfaces.IMsg, error) {
	// if 1 == rand.Intn(recieve_freq) {
	// 	fmt.Printf("%d -- RemotePeer.RECIEVE %s -> %s\n", os.Getpid(), f.FromName, f.ToName)
	// }
	// 100ms Timeout
	f.Socket.SetOption(mangos.OptionRecvDeadline, 100*time.Millisecond)
	// Minimal blocking
	// f.Socket.SetOption(mangos.OptionRecvDeadline, 1*time.Millisecond)
	var data []byte
	var err error
	if data, err = f.Socket.Recv(); err == nil {
		// if len(data) > 0 {
		msg, err := messages.UnmarshalMessage(data)
		if nil == err {
			fmt.Printf("%d -- RemotePeer.Recieve $$$$$$$$$$$$ GOT VALID MESSAGE:\t %+v\n", os.Getpid(), msg)
		} else {
			fmt.Printf("%d -- RemotePeer.Recieve Got invalid MESSAGE:\t %+v\n", os.Getpid(), string(data))
			// }
			return msg, err
		}
	} else {
		fmt.Printf("%d -- RemotePeer.Recieve error:\t %+v\n", os.Getpid(), err)

	}
	beat := "HEARTBEAT"
	if err = f.Socket.Send([]byte(beat)); err != nil {
		fmt.Printf("%d -- RemotePeer.Connect ###### error from f.Socket.Send(data) for:\n %+v\n\n", os.Getpid(), beat)
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
