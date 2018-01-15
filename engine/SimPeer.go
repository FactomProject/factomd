// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"math/rand"
	"time"
	"github.com/FactomProject/factomd/util/atomic"
)

var _ = fmt.Print
var _ = bytes.Compare

type SimPacket struct {
	data []byte
	sent int64 // Time in milliseconds
}

type SimPeer struct {
	// A connection to this node:
	FromName string
	ToName   string
	// Channels that define the connection:
	BroadcastOut chan *SimPacket
	BroadcastIn  chan *SimPacket

	// Delay in Milliseconds
	Delay    int64 // The maximum delay
	DelayUse int64 // We actually select a random delay for each data element.
	// Were we hold delayed packets
	Delayed *SimPacket

	bytesOut atomic.AtomicInt // Bytes sent out
	bytesIn  atomic.AtomicInt // Bytes Received

	Last atomic.AtomicInt64 // Last time reset (nano seconds)

	RateOut atomic.AtomicInt // Rate of Bytes output per ms
	RateIn  atomic.AtomicInt // Rate of Bytes input per ms
}

var _ interfaces.IPeer = (*SimPeer)(nil)

// Bytes sent out per second from this peer
func (f *SimPeer) BytesOut() int {
	return f.RateOut.Load()
}

// Bytes Received per second from this peer
func (f *SimPeer) BytesIn() int {
	return f.RateIn.Load()
}

func (*SimPeer) Weight() int {
	// A SimPeer only represents itself
	return 1
}

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
	f.BroadcastOut = make(chan *SimPacket, 10000)
	f.Last.Store(time.Now().UnixNano())
	return f
}

func (f *SimPeer) GetNameFrom() string {
	return f.FromName
}
func (f *SimPeer) GetNameTo() string {
	return f.ToName
}

func (f *SimPeer) computeBandwidth() {
	now := time.Now().UnixNano()
	delta := (now - f.Last.Load()) / 1000000000 // Make delta seconds
	if delta < 5 {
		// Wait at least 5 seconds.
		return
	}
	f.RateIn.Store(int(int64(f.bytesIn) / delta))
	f.RateOut.Store(int(int64(f.bytesOut) / delta))
	f.bytesIn.Store(0)
	f.bytesOut.Store(0)
	f.Last.Store(now)
}

func (f *SimPeer) Send(msg interfaces.IMsg) error {
	data, err := msg.MarshalBinary()
	f.bytesOut.Store(f.bytesOut.Load() + len(data))
	f.computeBandwidth()
	if err != nil {
		fmt.Println("ERROR on Send: ", err)
		return err
	}
	if len(f.BroadcastOut) < 9000 {
		packet := SimPacket{data: data, sent: time.Now().UnixNano() / 1000000}
		f.BroadcastOut <- &packet
	}
	return nil
}

// Non-blocking return value from channel.
func (f *SimPeer) Receive() (interfaces.IMsg, error) {
	if f.Delayed == nil {
		select {
		case packet, ok := <-f.BroadcastIn:
			if ok {
				f.Delayed = packet
			}
		default:
			return nil, nil // Nothing to do
		}
		if f.Delay > 0 {
			f.DelayUse = rand.Int63n(f.Delay)
		} else {
			f.DelayUse = 0
		}

	}

	now := time.Now().UnixNano() / 1000000

	if f.Delayed != nil && now-f.Delayed.sent > f.DelayUse {
		data := f.Delayed.data
		f.Delayed = nil
		msg, err := messages.UnmarshalMessage(data)
		if err != nil {
			fmt.Printf("SimPeer ERROR: %s %x %s\n", err.Error(), data[:8], messages.MessageName(data[0]))
		}

		f.bytesIn.Store(f.bytesIn.Load() + len(data))
		f.computeBandwidth()
		return msg, err
	} else {
		// fmt.Println("dddd Delay: ", now-f.Delayed.sent)
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
		fmt.Printf("AddSimPeer(%p,%v,%v) error: Unresonable value\n", fnodes, i1, i2)
		return
	}

	// If the connection already exists, ignore
	for _, p1 := range fnodes[i1].Peers {
		for _, p2 := range fnodes[i2].Peers {
			if p1.Equals(p2) {
				// this happens normally in some configs because the links are picked by formula
				// and as the node count changes the %Node Count can wrap the links back around
				// to nodes that already have links
			    //	fmt.Printf("AddSimPeer(%p,%v,%v) error: Duplicate\n", fnodes, i1, i2)
				return
			}
		}
	}

	f1 := fnodes[i1]
	f2 := fnodes[i2]

	fmt.Println(i1, " -- ", i2)

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
