// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simulation

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/FactomProject/factomd/common/globals"

	"github.com/FactomProject/factomd/fnode"

	"math/rand"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
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

	bytesOut int // Bytes sent out
	bytesIn  int // Bytes received

	Last int64 // Last time reset (nano seconds)

	RateOut int // Rate of Bytes output per ms
	RateIn  int // Rate of Bytes input per ms
}

var _ interfaces.IPeer = (*SimPeer)(nil)

// Bytes sent out per second from this peer
func (f *SimPeer) BytesOut() int {
	return f.RateOut
}

// Bytes received per second from this peer
func (f *SimPeer) BytesIn() int {
	return f.RateIn
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

func (f *SimPeer) Initialize(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	f.BroadcastOut = make(chan *SimPacket, 10000)
	f.Last = time.Now().UnixNano()
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
	delta := (now - f.Last) / 1000000000 // Make delta seconds
	if delta < 5 {
		// Wait atleast 5 seconds.
		return
	}
	f.RateIn = int(int64(f.bytesIn) / delta)
	f.RateOut = int(int64(f.bytesOut) / delta)
	f.bytesIn = 0
	f.bytesOut = 0
	f.Last = now
}

func (f *SimPeer) Send(msg interfaces.IMsg) error {

	data, err := msg.MarshalBinary()
	f.bytesOut += len(data)
	f.computeBandwidth()
	if err != nil {
		return err
	}

	go func() {
		if f.Delay > 0 {
			// Sleep some random number of milliseconds, then send the packet
			time.Sleep(time.Duration(rand.Intn(int(f.Delay))) * time.Millisecond)
		}
		packet := SimPacket{data: data, sent: time.Now().UnixNano() / 1000000}
		f.BroadcastOut <- &packet
	}()

	return nil
}

// Non-blocking return value from channel.
func (f *SimPeer) Receive() (interfaces.IMsg, error) {

	// We want a packet from the network
	var packet *SimPacket

	// However, we do not want to wait if one isn't there.
	select {
	case packet = <-f.BroadcastIn:
	default:
		return nil, nil // Nothing to do
	}

	// Count the overhead of packets
	f.bytesIn += len(packet.data)
	f.computeBandwidth()

	// Unmarshal our message, and throw it a way if we have an error.
	msg, err := msgsupport.UnmarshalMessage(packet.data)
	if err != nil {
		return nil, err
	}

	// All is good.  Return our message.
	return msg, err

}

func AddSimPeer(fnodes []*fnode.FactomNode, i1 int, i2 int) {
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

	fmt.Println(i1, " -- ", i2)

	peer12 := new(SimPeer).Initialize(f1.State.FactomNodeName, f2.State.FactomNodeName).(*SimPeer)
	peer21 := new(SimPeer).Initialize(f2.State.FactomNodeName, f1.State.FactomNodeName).(*SimPeer)
	peer12.BroadcastIn = peer21.BroadcastOut
	peer21.BroadcastIn = peer12.BroadcastOut

	f1.Peers = append(f1.Peers, peer12)
	f2.Peers = append(f2.Peers, peer21)

	//for _, p := range f1.Peers {
	//	fmt.Printf("%s's peer: %s\n", p.GetNameFrom(), p.GetNameTo())
	//}

}

// construct a simulated network
func BuildNetTopology(p *globals.FactomParams) {
	nodes := fnode.GetFnodes()

	switch p.Net {
	case "file":
		file, err := os.Open(p.Fnet)
		if err != nil {
			panic(fmt.Sprintf("File network.txt failed to open: %s", err.Error()))
		} else if file == nil {
			panic(fmt.Sprint("File network.txt failed to open, and we got a file of <nil>"))
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var a, b int
			var s string
			_, _ = fmt.Sscanf(scanner.Text(), "%d %s %d", &a, &s, &b)
			if s == "--" {
				AddSimPeer(nodes, a, b)
			}
		}
	case "square":
		side := int(math.Sqrt(float64(p.Cnt)))

		for i := 0; i < side; i++ {
			AddSimPeer(nodes, i*side, (i+1)*side-1)
			AddSimPeer(nodes, i, side*(side-1)+i)
			for j := 0; j < side; j++ {
				if j < side-1 {
					AddSimPeer(nodes, i*side+j, i*side+j+1)
				}
				AddSimPeer(nodes, i*side+j, ((i+1)*side)+j)
			}
		}
	case "long":
		fmt.Println("Using long Network")
		for i := 1; i < p.Cnt; i++ {
			AddSimPeer(nodes, i-1, i)
		}
		// Make long into a circle
	case "loops":
		fmt.Println("Using loops Network")
		for i := 1; i < p.Cnt; i++ {
			AddSimPeer(nodes, i-1, i)
		}
		for i := 0; (i+17)*2 < p.Cnt; i += 17 {
			AddSimPeer(nodes, i%p.Cnt, (i+5)%p.Cnt)
		}
		for i := 0; (i+13)*2 < p.Cnt; i += 13 {
			AddSimPeer(nodes, i%p.Cnt, (i+7)%p.Cnt)
		}
	case "alot":
		n := len(nodes)
		for i := 0; i < n; i++ {
			AddSimPeer(nodes, i, (i+1)%n)
			AddSimPeer(nodes, i, (i+5)%n)
			AddSimPeer(nodes, i, (i+7)%n)
		}

	case "alot+":
		n := len(nodes)
		for i := 0; i < n; i++ {
			AddSimPeer(nodes, i, (i+1)%n)
			AddSimPeer(nodes, i, (i+5)%n)
			AddSimPeer(nodes, i, (i+7)%n)
			AddSimPeer(nodes, i, (i+13)%n)
		}

	case "tree":
		index := 0
		row := 1
	treeloop:
		for i := 0; true; i++ {
			for j := 0; j <= i; j++ {
				AddSimPeer(nodes, index, row)
				AddSimPeer(nodes, index, row+1)
				row++
				index++
				if index >= len(nodes) {
					break treeloop
				}
			}
			row += 1
		}
	case "circles":
		circleSize := 7
		index := 0
		for {
			AddSimPeer(nodes, index, index+circleSize-1)
			for i := index; i < index+circleSize-1; i++ {
				AddSimPeer(nodes, i, i+1)
			}
			index += circleSize

			AddSimPeer(nodes, index, index-circleSize/3)
			AddSimPeer(nodes, index+2, index-circleSize-circleSize*2/3-1)
			AddSimPeer(nodes, index+3, index-(2*circleSize)-circleSize*2/3)
			AddSimPeer(nodes, index+5, index-(3*circleSize)-circleSize*2/3+1)

			if index >= len(nodes) {
				break
			}
		}
	default:
		fmt.Println("Didn't understand network type. Known types: mesh, long, circles, tree, loops.  Using a Long Network")
		for i := 1; i < p.Cnt; i++ {
			AddSimPeer(nodes, i-1, i)
		}

	}

	var colors = []string{"95cde5", "b01700", "db8e3c", "ffe35f"}

	if len(nodes) > 2 {
		for i, s := range nodes {
			fmt.Printf("%d {color:#%v, shape:dot, label:%v}\n", i, colors[i%len(colors)], s.State.FactomNodeName)
		}
		fmt.Printf("Paste the network info above into http://arborjs.org/halfviz to visualize the network\n")
	}
}
