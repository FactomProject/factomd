// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

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

func (f *SimPeer) Equals(ff interfaces.IPeer) bool {
	f2, ok := ff.(*SimPeer)
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
		fmt.Println("ERROR on Send: ", err)
		return err
	}
	if len(f.BroadcastOut) < 9000 {
		f.BroadcastOut <- data
	}
	return nil
}

// Non-blocking return value from channel.
func (f *SimPeer) Recieve() (interfaces.IMsg, error) {
	select {
	case data, ok := <-f.BroadcastIn:
		if ok {
			msg, err := messages.UnmarshalMessage(data)
			if err != nil {
				fmt.Printf("SimPeer ERROR: %s %x %s\n", err.Error(), data[:8], messages.MessageName(data[0]))
			}

			return msg, err
		}
	default:
	}
	return nil, nil
}

func AddSimPeer(fnodes []*FactomNode, i1 int, i2 int) {
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

	fmt.Println(f1.State.FactomNodeName, " -> ", f2.State.FactomNodeName)

	peer12 := new(SimPeer).Init(f1.State.FactomNodeName, f2.State.FactomNodeName).(*SimPeer)
	peer21 := new(SimPeer).Init(f2.State.FactomNodeName, f1.State.FactomNodeName).(*SimPeer)
	peer12.BroadcastIn = peer21.BroadcastOut
	peer21.BroadcastIn = peer12.BroadcastOut

	f1.Peers = append(f1.Peers, peer12)
	f2.Peers = append(f2.Peers, peer21)

	// 	for _, p := range f1.Peers {
	// 		fmt.Printf("%s's peer: %s\n", p.GetNameFrom(), p.GetNameTo())
	// 	}

}
