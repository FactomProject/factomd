package main

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/FactomProject/electiontesting/controller"
	"github.com/FactomProject/electiontesting/election"
	"github.com/FactomProject/electiontesting/imessage"

	"crypto/sha256"
)

var mirrorMap map[[32]byte][]byte

var solutions = 0
var breadth = 0
var loops = 0
var mirrors = 0
var depths []int
var solutionsAt []int
var mirrorsAt []int
var deadMessagesAt []int
var hitlimit int
var maxdepth int
var failure int

var globalRunNumber = 0

var extraPrints = true
var insanePrints = false

//================ main =================
func main() {
	recurse(2, 5, 100)
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

	global := con.Elections[0].Display.Global
	for i, ldr := range con.Elections {
		con.Elections[i] = CloneElection(ldr)
		con.Elections[i].Display.Global = global
	}

	return con, con.Elections, msgs
}

type mymsg struct {
	leaderIdx int
	msg       imessage.IMessage
}

// dive
// Pass a list of messages to process, to a set of leaders, at a current depth, with a particular limit.
// Provided a msgPath, and updated for recording purposes.
// Return success if we actually settled on a solution
// Have seen a success is useful for declaring a loop has been detected.
//
// Note that we actually dive 100 levels beyond our limit, and declare seeSuccess past our limit as proof we are
// in a loop.
// Hitting the limit and seeSuccess is proof of a loop that none the less can resolve.
func dive(msgs []*mymsg, leaders []*election.Election, depth int, limit int, msgPath []*mymsg) (limitHit bool, seeSuccess bool) {
	depths = incCounter(depths, depth)
	depth++
	if depth > limit {
		breadth++
		hitlimit++
		return true, false
	}

	if depth > maxdepth {
		maxdepth = depth
	}

	fmt.Println("=============== ",
		" Depth=", depth, "/", maxdepth,
		" Failures=", failure,
		" MsgQ=", len(msgs),
		", Mirrors=", mirrors, len(mirrorMap),
		", Hit the Limits=", hitlimit,
		" Breadth=", breadth,
		", solutions so far =", solutions,
		", global count= ", globalRunNumber,
		", loops detected=", loops)
	fmt.Printf("\n%20s", "Dead Messages")
	for i, v := range deadMessagesAt {
		if i%16 == 0 {
			fmt.Println()
		}
		fmt.Printf("%3d/%12d ", i, v)
	}
	fmt.Printf("\n%20s ", "Mirrors")
	fmt.Print("Mirrors       ", depth)
	for i, v := range mirrorsAt {
		fmt.Printf("%d/%d ", i, v)
	}
	fmt.Printf("\n%20s ", "Solutions")
	fmt.Print("Solutions     ", depth)
	for i, v := range solutionsAt {
		fmt.Printf("%d/%d ", i, v)
	}
	fmt.Printf("\n%20s ", "Depths")
	fmt.Print("Depths        ", depth)
	for i, v := range depths {
		fmt.Printf("%d/%d ", i, v)
	}
	fmt.Println()

	if extraPrints {
		// Lots of printing... Not necessary....
		fmt.Println(leaders[0].Display.Global.String())

		for _, ldr := range leaders {
			fmt.Println(ldr.Display.String())
		}

		if insanePrints {
			// Example of a run that has a werid msg state
			if globalRunNumber > -1 {
				fmt.Println("Leader 0")
				fmt.Println(leaders[0].PrintMessages())
				fmt.Println("Leader 1")
				fmt.Println(leaders[1].PrintMessages())
				fmt.Println("Leader 2")
				fmt.Println(leaders[2].PrintMessages())
			}
		}
	}

	done := 0
	for _, ldr := range leaders {
		if ldr.Committed {
			done++
		}
	}
	if done == len(leaders)/2+1 {
		solutionsAt = incCounter(solutionsAt, depth)
		if extraPrints {
			fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>> Solution Found @ ", depth)
		}
		breadth++
		solutions++
		return false, true

	}

	// Look for mirrorMap, but only after we have been going a bit.
	if depth > 70 {
		var hashes [][32]byte
		var strings []string
		for _, ldr := range leaders {
			bits := ldr.NormalizedString()
			if bits != nil {
				strings = append(strings, string(bits))
				h := Sha(bits)
				hashes = append(hashes, h)
			} else {
				panic("shouldn't happen")
			}
		}
		for i := 0; i < len(hashes)-1; i++ {
			for j := 0; j < len(hashes)-1-i; j++ {
				if bytes.Compare(hashes[j][:], hashes[j+1][:]) > 0 {
					hashes[j], hashes[j+1] = hashes[j+1], hashes[j+1]
					strings[j], strings[j+1] = strings[j+1], strings[j]
				}
			}
		}
		var all []byte
		var alls string
		for i, h := range hashes {
			all = append(all, h[:]...)
			alls += strings[i]
		}
		mh := Sha(all)
		if mirrorMap[mh] != nil {
			mirrors++
			breadth++
			mirrorsAt = incCounter(mirrorsAt, depth)
			return false, false
		}
		fmt.Println("All State: ", alls)
		mirrorMap[mh] = mh[:]
	}

	for d, v := range msgs {
		var msgs2 []*mymsg
		msgs2 = append(msgs2, msgs[0:d]...)
		msgs2 = append(msgs2, msgs[d+1:]...)
		ml2 := len(msgs2)

		globalRunNumber++

		cl := CloneElection(leaders[v.leaderIdx])

		//if !spewSame(cl, leaders[v.leaderIdx]) {
		//	fmt.Println("Clone Failed")
		//	debugClone(cl, leaders[v.leaderIdx])
		//	os.Exit(0)
		//}

		msg, changed := leaders[v.leaderIdx].Execute(v.msg, depth)

		msgPath2 := append(msgPath, v)

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
			gl := leaders[v.leaderIdx].Display.Global
			for _, ldr := range leaders {
				ldr.Display.Global = gl
			}
			// Recursive Dive
			lim, ss := dive(msgs2, leaders, depth, limit, msgPath2)
			_ = lim || ss
			seeSuccess = seeSuccess || ss
			limitHit = limitHit || lim
			for _, ldr := range leaders {
				ldr.Display.Global = cl.Display.Global
			}
			msgs2 = msgs2[:ml2]
		} else {
			deadMessagesAt = incCounter(deadMessagesAt, depth)
		}
		leaders[v.leaderIdx] = cl
	}

	if depth == limit-50 && limitHit {
		if seeSuccess {
			loops++
		} else {
			failure++
		}
		limitHit = false
	}

	return limitHit, seeSuccess
}

func incCounter(counter []int, depth int) []int {
	for len(counter) <= depth {
		counter = append(counter, 0)
	}
	counter[depth]++
	return counter
}

func recurse(auds int, feds int, limit int) {

	_, leaders, msgs := newElections(feds, auds, false)
	var msgpath []*mymsg
	dive(msgs, leaders, 0, limit, msgpath)
}

// reuse encoder/decoder so we don't recompile the struct definition
var enc *gob.Encoder
var dec *gob.Decoder

// LoopingDetected will the number of looping leaders
func LoopingDetected(global *election.Display) int {
	return global.DetectLoops()
}

func init() {
	buff := new(bytes.Buffer)
	enc = gob.NewEncoder(buff)
	dec = gob.NewDecoder(buff)
	mirrorMap = make(map[[32]byte][]byte, 10000)
}

func CloneElection(src *election.Election) *election.Election {
	return src.Copy()
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

// Create a Sha256 Hash from a byte array
func Sha(p []byte) [32]byte {
	b := sha256.Sum256(p)
	return b
}
