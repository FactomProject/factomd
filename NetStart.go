// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	"os"
	"time"
	"strconv"
)

var _ = fmt.Print

type FactomNode struct {
	State 	*state.State
	Peers 	[]interfaces.IPeer
	MLog	*MsgLog
}

func NetStart(s *state.State) {

	var fnodes []*FactomNode

	s.SetOut(false)

	fmt.Println(">>>>>>>>>>>>>>>>")
	fmt.Println(">>>>>>>>>>>>>>>> Net Sim Start!!!!!")
	fmt.Println(">>>>>>>>>>>>>>>>")

	pcfg, _, err := btcd.LoadConfig()
	if err != nil {
		log.Println(err.Error())
	}
	FactomConfigFilename := pcfg.FactomConfigFile

	if len(FactomConfigFilename) == 0 {
		FactomConfigFilename = util.GetConfigFilename("m2")
	}
	fmt.Println(fmt.Sprintf("factom config: %s", FactomConfigFilename))

	// Figure out how many nodes I am going to generate.  Default 10
	cnt := 10
	if len(os.Args) > 2 {
		cnt1, err := strconv.Atoi(os.Args[1])
		if err == nil && cnt1 != 0 {
			cnt = cnt1
		}
	}
	
	mLog := new(MsgLog)
	mLog.init(cnt)
	
	makeServer := func() *FactomNode {
		// All other states are clones of the first state.  Which this routine
		// gets passed to it.
		newState := s

		if len(fnodes) > 0 {
			number := fmt.Sprintf("%d", len(fnodes))
			newState = s.Clone(number).(*state.State)
			newState.Init()
		}

		fnode := new(FactomNode)
		fnode.State = newState
		fnodes = append(fnodes, fnode)
		fnode.MLog = mLog
		
		return fnode
	}

	startServers := func() {
		for i, fnode := range fnodes {
			if i > 0 { fnode.State.Init() }
			go NetworkProcessorNet(fnode)
			go loadDatabase(fnode.State)
			go Timer(fnode.State)
			go Validator(fnode.State)
		}
	}

	//************************************************
	// Actually setup the Network
	//************************************************

	s.LoadConfig(FactomConfigFilename)
	s.Init()
	
	
	// Make cnt Factom nodes
	for i := 0; i < cnt; i++ { 
		makeServer()
	}

	primes := []int { 7, 13, 17, 23,  37,  43, 47, 53, 67, 73, 83, 97, 103 }	
	
	var p1 []int
	// Pick 3 primes
	for _,p := range primes {
		a := cnt
		b := p
		if a < b { 
			a = p
			b = cnt
		}
		if a%b != 0 {
			p1 = append(p1,p)
			if len(p1) > 3 {
				break
			}
		}
	}
	
	fmt.Println("factomd <node count> <network config: mesh/long/loops>")
	
	if len (os.Args) > 2 {
		switch os.Args[2] {
			case "mesh" :
				
				h:=0
				for index, p := range p1 {
					fmt.Println()
					for i:= 0; i < cnt/(index+1); i++ {
						h2 := (h+p)%cnt
						AddSimPeer(fnodes, h,h2)
						h = h2
					}
				}
			case "long" :
				for i := 1; i < cnt; i++ {
					AddSimPeer(fnodes,i-1,i)
				}
			case "loops" :
				for i := 1; i < cnt; i++ {
					AddSimPeer(fnodes,i-1,i)
				}
				for i := 0; i+5 < cnt; i+=6 {
					AddSimPeer(fnodes,i,i+5)
				}
				for i := 0; i+7 < cnt; i+=3 {
					AddSimPeer(fnodes,i,i+7)
				}
		}
	}
	
	startServers()

	go wsapi.Start(fnodes[0].State)

	AddInterruptHandler(func() {
		fmt.Print("<Break>\n")
		fmt.Print("Gracefully shutting down the server...\n")
		for _, fnode := range fnodes {
			fmt.Print("Shutting Down: ", fnode.State.FactomNodeName, "\r\n")
			fnode.State.ShutdownChan <- 0
		}
		fmt.Print("Waiting...\r\n")
		time.Sleep(3 * time.Second)
		os.Exit(0)
	})

	// Web API runs independent of Factom Servers


	var _ = time.Sleep
	p := 0
	var _ = p
	for {
		fmt.Sprintf(">>>>>>>>>>>>>>")
		
		b := make([]byte, 10)
		var err error
		if _, err = os.Stdin.Read(b); err != nil {
			log.Fatal(err.Error())
		}
		for i, c := range b {
			if c <= 32 {
				b = b[:i]
				break
			}
		}
		v, err := strconv.Atoi(string(b))
		if err == nil && v >= 0 && v < len(fnodes) {
			fnodes[p].State.SetOut(false)
			p = v
			fnodes[p].State.SetOut(true)
			fmt.Print("\r\nSwitching to Node ", p,"\r\n")
			wsapi.SetState(fnodes[p].State)
		}else{
			if len(b) == 0 { b = append(b,'a') }
			switch b[0] {
				case 'a', 'A':
					fnodes[p].State.SetOut(false)
					for _,f := range fnodes {
						fmt.Printf("%8s %s\n",f.State.FactomNodeName, f.State.String())
					}
				case 27:
					fmt.Print("Gracefully shutting down the servers...\r\n")
					for _, fnode := range fnodes {
						fmt.Print("Shutting Down: ", fnode.State.FactomNodeName, "\r\n")
						fnode.State.ShutdownChan <- 0
					}
					fmt.Print("Waiting...\r\n")
					time.Sleep(time.Duration(len(fnodes)/8+1) * time.Second)
					fmt.Println()
					os.Exit(0)
				case 32:
					fnodes[p].State.SetOut(false)
					p++
					if p >= len(fnodes) {
						p = 0
					}
					fnodes[p].State.SetOut(true)
					fmt.Print("\r\nSwitching to Node ", p,"\r\n")
					wsapi.SetState(fnodes[p].State)
				case 's','S' :
					msg := messages.NewAddServerMsg(fnodes[p].State)
					fnodes[p].State.NetworkInMsgQueue() <- msg
					fnodes[p].State.SetOut(true)
					fmt.Println("Attempting to make",fnodes[p].State.GetFactomNodeName(),"a Leader")
				default:
			}
		}
	}

}
