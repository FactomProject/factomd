package main

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/FactomProject/electiontesting/controller"
	"github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/imessage"
)

func main() {
	recurse(3, 3, 120)
}

// newElections will return an array of elections (1 per leader) and an array
// of volunteers messages to kick things off.
//		Params:
//			feds int   Number of Federated Nodes
//			auds int   Number of Volunteers
//			noDisplay  Passing a true here will reduce memory consumption, as it is a debugging tool
//
//		Returns:
//			controller *Controller  This can used for debugging (Printing votes)
//			elections []*election   Nodes you can execute on (returns msg, statchange)
//			volmsgs   []*VoluntMsg	Volunteer msgs you can start things with
func newElections(feds, auds int, noDisplay bool) (*controller.Controller, []*election.Election, []*mymsg) {
	con := controller.NewController(feds, auds)

	if noDisplay {
		for _, e := range con.Elections {
			e.Display = nil
		}
		con.GlobalDisplay = nil
	}
	var msgs []*mymsg
	fmt.Println("Starting")
	for _, v := range con.Volunteers {
		for i, _ := range con.Elections {
			my := new(mymsg)
			my.leaderIdx = i
			my.msg = v
			msgs = append(msgs, my)
			fmt.Println(my.msg.String(), my.leaderIdx)
		}
	}
	return con, con.Elections, msgs
}

type mymsg struct {
	leaderIdx int
	msg       imessage.IMessage
}

var solutions = 0
var breadth = 0

var cuts []int

func dive(msgs []*mymsg, leaders []*election.Election, depth int, limit int) {
	depth++
	if depth > limit {
		fmt.Print("Breath ", breadth)
		for _, v := range cuts {
			fmt.Print(v, " ")
		}
		fmt.Println()
		breadth++
		return
	}
	done := 0
	for _,ldr := range leaders {
		if ldr.Committed { done ++}
	}
	if done > len(leaders)/2 {
		cuts[depth]++
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>> Solution Found @ ",depth)
		breadth++
		solutions++
		return
	}

//	fmt.Println("Height: ",depth)
//	for _,v := range msgs {
//		fmt.Println(v.msg, " to ",v.leaderIdx)
//	}

	fmt.Println("===============", depth, "solutions so far ", solutions)
	for _,ldr := range leaders {
		fmt.Println(ldr.Display.String())
	}
	for d, v := range msgs {
		msgs2 := append(msgs[0:d], msgs[d+1:]...)
		ml2 := len(msgs2)
		cl := CloneElection(leaders[v.leaderIdx])
		msg, changed := leaders[v.leaderIdx].Execute(v.msg)

		if changed {
			if msg != nil {
				for i, _ := range leaders {
					if i != v.leaderIdx {
						my := new(mymsg)
						my.leaderIdx = i
						my.msg = msg
						msgs2 = append(msgs2, my)
					}
				}
			}
			dive(msgs2, leaders, depth, limit)
			msgs2 = msgs2[:ml2]
		} else {
			for len(cuts) <= depth {
				cuts = append(cuts, 0)
			}
			cuts[depth]++
		}
		leaders[v.leaderIdx] = cl
	}
}

func recurse(auds int, feds int, limit int) {

	_, leaders, msgs := newElections(feds, auds, false)

	dive(msgs, leaders, 0, limit)
}

// reuse encoder/decoder so we don't recompile the struct definition
var enc *gob.Encoder
var dec *gob.Decoder

func init() {
	buff := new(bytes.Buffer)
	enc = gob.NewEncoder(buff)
	dec = gob.NewDecoder(buff)
}

func CloneElection(src *election.Election) *election.Election {
	dst := new(election.Election)
	err := enc.Encode(src)
	if err != nil {
		panic(err)
	}
	err = dec.Decode(dst)
	if err != nil {
		panic(err)
	}
	return dst
}
