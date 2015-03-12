// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/FactomCode/factomd"
	"github.com/FactomProject/FactomCode/util"
)

func realMain() {
	factomd.Factomd_init()
	factomd.Factomd_main()

	util.Trace()
	btcd_main()
}

func main() {
	fmt.Println("//////////////////////// Copyright 2015 Factom Foundation")
	fmt.Println("//////////////////////// Use of this source code is governed by the MIT")
	fmt.Println("//////////////////////// license that can be found in the LICENSE file.")

	realMain()
}

// start up Factom queue(s) managers/processors
func factomQueues(s *server) {
	// Write outgoing factom messages into P2P network
	go func() {
		for msg := range factomd.OutMsgQueue {
			s.BroadcastMessage(msg)
			/*      peerInfoResults := server.PeerInfo()
			        for peerInfo := range peerInfoResults{
			          fmt.Printf("PeerInfo:%+v", peerInfo)

			        }*/
		}
	}()

	/*
	   go func() {
	     for msg := range inRpcQueue {
	       fmt.Printf("in range inRpcQueue, msg:%+v\n", msg)
	       switch msg.Command() {
	       case factomwire.CmdTx:
	         InMsgQueue <- msg //    for testing
	         server.blockManager.QueueTx(msg.(*factomwire.MsgTx), nil)
	       case factomwire.CmdConfirmation:
	         server.blockManager.QueueConf(msg.(*factomwire.MsgConfirmation), nil)

	       default:
	         inMsgQueue <- msg
	         outMsgQueue <- msg
	       }
	     }
	   }()
	*/
}
