package main

import (
	"github.com/FactomProject/factomd/p2p"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/common/interfaces"
	"flag"
	"github.com/FactomProject/factomd/common/primitives"
	"time"
	"fmt"
)

var p2pProxy *engine.P2PProxy

var old map[[32]byte]interfaces.IMsg

func InitNetwork() {


	networkPortOverridePtr := flag.String("networkPort", "8108", "Address for p2p network to listen on.")
	peersPtr := flag.String("peers", "", "Array of peer addresses. ")
	netdebugPtr := flag.Int("netdebug", 0, "0-5: 0 = quiet, >0 = increasing levels of logging")
	exclusivePtr := flag.Bool("exclusive", false, "If true, we only dial out to special/trusted peers.")

	flag.Parse()

	port := *networkPortOverridePtr
	peers := *peersPtr
	netdebug := *netdebugPtr
	exclusive := *exclusivePtr

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
		msg,err := p2pProxy.Recieve()
		if err == nil && msg != nil && old[msg.GetHash().Fixed()]== nil {
			old[msg.GetHash().Fixed()]= msg
			bounce,ok := msg.(*Bounce)
			if ok {
				bounce.stamps = append(bounce.stamps,primitives.NewTimestampNow())
				p2pProxy.Send(msg)
				fmt.Println(msg.String())
			}
		}
	}
}


func main() {
	InitNetwork()

	bounce := new(Bounce)
	bounce.Timestamp = primitives.NewTimestampNow()
	p2pProxy.Send(bounce)

	go listen()

	time.Sleep(100*time.Second)

}