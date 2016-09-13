// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/p2p"
	"math/rand"
	"time"
)

var p2pProxy *engine.P2PProxy

var old map[[32]byte]interfaces.IMsg
var oldcnt int

var broadcastSent int
var broadcastReceived int
var p2pSent int
var p2pReceived int
var p2pRequestSent int
var p2pRequestReceived int

var name string
var isp2p bool

func InitNetwork() {

	go engine.StartProfiler()

	namePtr := flag.String("name", fmt.Sprintf("%d", rand.Int()), "Name for this node")
	networkPortOverridePtr := flag.String("networkPort", "8108", "Address for p2p network to listen on.")
	peersPtr := flag.String("peers", "", "Array of peer addresses. ")
	netdebugPtr := flag.Int("netdebug", 0, "0-5: 0 = quiet, >0 = increasing levels of logging")
	exclusivePtr := flag.Bool("exclusive", false, "If true, we only dial out to special/trusted peers.")
	deadlinePtr := flag.Int64("deadline", 1, "Deadline for Reads and Writes to conn.")
	p2pPtr := flag.Bool("p2p", false, "Test p2p messages (default to false)")
	flag.Parse()

	name = *namePtr
	port := *networkPortOverridePtr
	peers := *peersPtr
	netdebug := *netdebugPtr
	exclusive := *exclusivePtr
	p2p.Deadline = time.Duration(*deadlinePtr) * time.Millisecond
	isp2p = *p2pPtr

	old = make(map[[32]byte]interfaces.IMsg, 0)
	connectionMetricsChannel := make(chan interface{}, p2p.StandardChannelSize)
	ci := p2p.ControllerInit{
		Port:                     port,
		PeersFile:                "peers.json",
		Network:                  1,
		Exclusive:                exclusive,
		SeedURL:                  "",
		SpecialPeers:             peers,
		ConnectionMetricsChannel: connectionMetricsChannel,
	}
	p2pNetwork := new(p2p.Controller).Init(ci)
	p2pNetwork.StartNetwork()
	// Setup the proxy (Which translates from network parcels to factom messages, handling addressing for directed messages)
	p2pProxy = new(engine.P2PProxy).Init("testnode", "P2P Network").(*engine.P2PProxy)
	p2pProxy.FromNetwork = p2pNetwork.FromNetwork
	p2pProxy.ToNetwork = p2pNetwork.ToNetwork
	p2pProxy.SetDebugMode(netdebug)

	if netdebug > 0 {
		p2pNetwork.StartLogging(uint8(netdebug))
	} else {
		p2pNetwork.StartLogging(uint8(0))
	}
	p2pProxy.StartProxy()
	// Command line peers lets us manually set special peers
	p2pNetwork.DialSpecialPeersString("")
}

func listen() {

	for {
		msg, err := p2pProxy.Recieve()
		if err != nil || msg == nil {
			time.Sleep(1 * time.Millisecond)
			continue
		}
		time.Sleep(1 * time.Millisecond)

		bounce, ok1 := msg.(*messages.Bounce)
		bounceReply, ok2 := msg.(*messages.BounceReply)

		if old[msg.GetHash().Fixed()] == nil {
			old[msg.GetHash().Fixed()] = msg
			if ok1 && len(bounce.Stamps) < 5{
				if isp2p {
					for i:=0; i<200; i++ {
						bounceReply = new(messages.BounceReply)
						bounceReply.Number = int32(i)
						bounceReply.Name = name
						bounceReply.Timestamp = bounce.Timestamp
						bounceReply.Stamps = append(bounce.Stamps, primitives.NewTimestampNow())

						bounceReply.SetOrigin(bounce.GetOrigin())
						bounceReply.SetNetworkOrigin(bounce.GetNetworkOrigin())

						p2pProxy.Send(bounceReply)
						old[msg.GetHash().Fixed()] = msg

						p2pSent++
					}
					p2pRequestReceived++
				} else {
					bounce.Stamps = append(bounce.Stamps, primitives.NewTimestampNow())
					p2pProxy.Send(msg)
					old[msg.GetHash().Fixed()] = msg
					broadcastReceived++
					broadcastSent++
				}
			}
			if false && ok2 && len(bounceReply.Stamps) < 5 {
				bounceReply.Stamps = append(bounceReply.Stamps, primitives.NewTimestampNow())
				p2pProxy.Send(msg)
				old[msg.GetHash().Fixed()] = msg
				p2pReceived++
				p2pSent++
			}
			fmt.Println("    ", msg.String())

		} else {
			oldcnt++
			fmt.Println("OLD:", msg.String())
		}

	}
}

func main() {
	InitNetwork()

	time.Sleep(10 * time.Second)
	fmt.Println ("Starting...")
	time.Sleep(3 * time.Second)
	go listen()

	for {
		bounce := new(messages.Bounce)
		bounce.Number = int32(p2pRequestSent + 1)
		bounce.Name = name
		bounce.Timestamp = primitives.NewTimestampNow()
		bounce.Stamps = append(bounce.Stamps, primitives.NewTimestampNow())
		if isp2p {
			bounce.SetPeer2Peer(true)
			p2pRequestSent++
		} else {
			broadcastSent++
		}
		p2pProxy.Send(bounce)
		old[bounce.GetHash().Fixed()] = bounce

		if isp2p {
			fmt.Printf("netTest(%s): Reads: %d errs %d Writes %d errs %d  ::p2p:: request sent: %d request recieved %d sent: %d received: %d\n",
				name,
				p2p.Reads, p2p.ReadsErr,
				p2p.Writes, p2p.WritesErr,
				p2pRequestSent, p2pRequestReceived,
				p2pSent, p2pReceived)

		} else {
			fmt.Printf("netTest(%s): Reads: %d errs %d Writes %d errs %d  ::: broadcast sent: %d broadcast received: %d\n",
				name,
				p2p.Reads, p2p.ReadsErr,
				p2p.Writes, p2p.WritesErr,
				broadcastSent, broadcastReceived)
		}
		time.Sleep(20 * time.Second)
	}
}
