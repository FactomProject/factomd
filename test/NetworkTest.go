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
var msgcnt int
var bounces int
var name string

func InitNetwork() {

	go engine.StartProfiler()

	namePtr := flag.String("name", fmt.Sprintf("%d", rand.Int()), "Name for this node")
	networkPortOverridePtr := flag.String("networkPort", "8108", "Address for p2p network to listen on.")
	peersPtr := flag.String("peers", "", "Array of peer addresses. ")
	netdebugPtr := flag.Int("netdebug", 0, "0-5: 0 = quiet, >0 = increasing levels of logging")
	exclusivePtr := flag.Bool("exclusive", false, "If true, we only dial out to special/trusted peers.")

	flag.Parse()

	name = *namePtr
	port := *networkPortOverridePtr
	peers := *peersPtr
	netdebug := *netdebugPtr
	exclusive := *exclusivePtr

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
	prtone := false
	for {
		msg, err := p2pProxy.Recieve()
		if err != nil || msg == nil {
			if !prtone {
				if err != nil {
					fmt.Println(err.Error())
				}
			}
			prtone = true
			time.Sleep(1 * time.Millisecond)
			continue
		}

		if old[msg.GetHash().Fixed()] == nil {
			prtone = false
			old[msg.GetHash().Fixed()] = msg
			bounce, ok := msg.(*messages.Bounce)
			if ok && len(bounce.Stamps)<100 {
				bounce.Stamps = append(bounce.Stamps, primitives.NewTimestampNow())
				p2pProxy.Send(msg)
				fmt.Println(msg.String())
			}
			bounces++
		} else {
			oldcnt++
		}
	}
}

func main() {
	InitNetwork()

	go listen()

	for {
		if msgcnt < 10 {
			bounce := new(messages.Bounce)
			bounce.Number = int32(msgcnt)
			bounce.Name = name
			bounce.Timestamp = primitives.NewTimestampNow()
			p2pProxy.Send(bounce)
			msgcnt++
		}
		fmt.Printf("bbbb Summary: Reads: %d errs %d Writes %d errs %d Msg Sent %d Msg Received %d\n",
			p2p.Reads, p2p.ReadsErr,
			p2p.Writes, p2p.WritesErr,
			msgcnt, bounces)
		time.Sleep(20 * time.Second)
	}

}
