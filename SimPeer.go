// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

var _ = fmt.Print
var _ = bytes.Compare

type SimPeer struct {
	// A connection to this node:
	FromName string
	ToName   string
	// Channels that define the connection:
	BroadcastOut chan []byte
	BroadcastIn  chan []byte
}

var _ interfaces.IPeer = (*SimPeer)(nil)

func (f *SimPeer) Len() int {
	return len(f.BroadcastIn)
}

func (f *SimPeer) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	f.BroadcastOut = make(chan []byte, 10000)
	return f
}

func (f *SimPeer) GetNameFrom() string {
	return f.FromName
}
func (f *SimPeer) GetNameTo() string {
	return f.ToName
}

func (f *SimPeer) Send(msg interfaces.IMsg) error {
	data, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	f.BroadcastOut <- data
	return nil
}

// Non-blocking return value from channel.
func (f *SimPeer) Recieve() (interfaces.IMsg, error) {
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

func AddSimPeer(fnodes []*FactomNode, i1 int, i2 int) {
	if i1 >= len(fnodes) || i2 >= len(fnodes) {
		return
	}

	fmt.Println("AddPeer(fnodes,", i1, i2, ")")

	f1 := fnodes[i1]
	f2 := fnodes[i2]

	peer12 := new(SimPeer).Init(f1.State.FactomNodeName, f2.State.FactomNodeName).(*SimPeer)
	peer21 := new(SimPeer).Init(f2.State.FactomNodeName, f1.State.FactomNodeName).(*SimPeer)
	peer12.BroadcastIn = peer21.BroadcastOut
	peer21.BroadcastIn = peer12.BroadcastOut

	f1.Peers = append(f1.Peers, peer12)
	f2.Peers = append(f2.Peers, peer21)

	for _, p := range f1.Peers {
		fmt.Printf("Peer f1 from %s to %s\n", p.GetNameTo(), p.GetNameFrom())
	}
	for _, p := range f2.Peers {
		fmt.Printf("Peer f2 from %s to %s\n", p.GetNameTo(), p.GetNameFrom())
	}

}
