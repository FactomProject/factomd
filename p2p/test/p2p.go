// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	. "github.com/factomproject/factomd/p2p"
)

// Note - we only need to expose the Listen port, so this can be tested
// on Caladan by simply connecting to the Caladan Cluster at a URL and Port.
// It will be effectively a random connection (we might connect to ourselves)
// So will need to check and release if we end up talking to ourself.

const (
	fullPeerManagement = iota
	simpleMinimalTest
)

func main() {
	mode := fullPeerManagement
	switch mode {
	case simpleMinimalTest:
		simpleTest()
	case fullPeerManagement:
		controllerManagedTest()
	}
}

func controllerManagedTest() {
	controller := new(Controller).Init("8088")
	controller.StartLogging(Debugging)
	// fmt.Printf("p2p.controllerManagedTest() %+v\n", " Checkpoint")
	controller.StartNetwork()
	// fmt.Printf("p2p.controllerManagedTest() %+v\n", " Network start called")
	// Send some custom messages.
	controller.DialPeer("127.0.0.1:8088") // start a connection to ourselves
	// fmt.Printf("p2p.controllerManagedTest() %+v\n", "DialPeer called")
	var message string
	count := 0

	for {
		message = fmt.Sprintf("Heartbeat #%d", count)
		ApplicationSend(controller, message)
		// fmt.Printf("p2p.controllerManagedTest() %+v\n", "ApplicationSend called")
		time.Sleep(time.Second * 2)
		fmt.Printf("      ToNetwork             %d\n", len(controller.ToNetwork))
		fmt.Printf("      FromNetwork           %d\n", len(controller.FromNetwork))
		responses := ApplicationRecieve(controller)
		// fmt.Printf("p2p.controllerManagedTest() %+v\n", "ApplicationRecieve called")
		for _, parcel := range responses {
			parcel.Print()
			fmt.Printf("Recieved: %s\n", parcel.Payload.(string))
		}
		count++
	}
}

func ApplicationSend(c *Controller, payload interface{}) {
	// fmt.Printf("p2p.ApplicationSend() %+v\n", payload)
	parcel := NewParcel(TestNet)
	parcel.Payload = payload
	c.ToNetwork <- *parcel
}

// ApplicationRecieve prints details of the recieved messages.
func ApplicationRecieve(c *Controller) []Parcel {
	// fmt.Printf("p2p.ApplicationRecieve() %+v\n", " ")
	var payloads []Parcel
	for 0 < len(c.FromNetwork) {
		parcel := <-c.FromNetwork
		// fmt.Printf("p2p.ApplicationRecieve() %+v\n", parcel)
		// parcel.Print()
		payloads = append(payloads, parcel)
	}
	return payloads
}

func simpleTest() {
	nodeA := new(Connection).Init()
	nodeB := new(Connection).Init()
	count := 0
	var message string

	for {
		message = fmt.Sprintf("Heartbeat #%d", count)
		SimpleSend(nodeA, message)
		SimpleSend(nodeB, message)
		for 0 < len(nodeA.ReceiveChannel) {
			SimpleReceive(nodeA)
		}
		for 0 < len(nodeB.ReceiveChannel) {
			SimpleReceive(nodeB)
		}

		time.Sleep(time.Second * 5)
		fmt.Printf("      NodeA.SendChannel             %d\n", len(nodeA.SendChannel))
		fmt.Printf("      NodeA.ReceiveChannel          %d\n", len(nodeA.ReceiveChannel))
		fmt.Printf("      NodeB.SendChannel             %d\n", len(nodeB.SendChannel))
		fmt.Printf("      NodeB.ReceiveChannel          %d\n", len(nodeB.ReceiveChannel))

	}
}

// SimpleSend sends a simple message, wrapping it up in a parce.
func SimpleSend(c *Connection, payload interface{}) {
	header := new(ParcelHeader).Init(TestNet) //.(*ParcelHeader)
	parcel := new(Parcel).Init(*header)       //.(*Parcel)
	parcel.Payload = payload
	c.SendChannel <- *parcel
}

// SimpleReceive prints details of the recieved messages.
func SimpleReceive(c *Connection) {
	parcel := <-c.ReceiveChannel
	parcel.Print()
}
