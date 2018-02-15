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
	recurse(3, 3, 2)
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
	for _, v := range con.Volunteers {
		my := new(mymsg)
		for i, _ := range con.Elections {
			my.leaderIdx = i
			my.msg = v
		}
		msgs = append(msgs, my)
	}
	return con, con.Elections, msgs
}

type mymsg struct {
	leaderIdx int
	msg       imessage.IMessage
}

var breath = 0

var cuts []int

func dive(msgs []*mymsg, leaders []*election.Election, depth int, limit int) {
	depth++
	if depth > limit {
		fmt.Print("Breath ", breath)
		for _, v := range cuts {
			fmt.Print(v, " ")
		}
		fmt.Println()
		breath++
		return
	}

	fmt.Println(leaders[0].Display.Global.String())

	for d, v := range msgs {
		msgs2 := append(msgs[0:d], msgs[d+1:]...)
		ml2 := len(msgs2)
		cl := clone(leaders[v.leaderIdx])
		fmt.Println(v.msg.String())
		msg, changed := leaders[v.leaderIdx].Execute(v.msg)
		if changed {
			if msg != nil {
				for i, _ := range leaders {
					if i != v.leaderIdx {
						my := new(mymsg)
						my.leaderIdx = i
						my.msg = msg
						msgs2 = append(msgs2, v)
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

// need better reflect based deep copy
func clone(src *election.Election) *election.Election {
	dst := new(election.Election)
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	enc.Encode(src)
	dec.Decode(dst)
	return dst
}
