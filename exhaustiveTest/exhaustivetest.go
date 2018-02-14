package exhaustiveTest

import (
	//	. "github.com/FactomProject/electiontesting/primitives"
	"bytes"
	"encoding/gob"
	"fmt"
	. "github.com/FactomProject/electiontesting/errorhandling"
)

type DummyMessage struct {
	S string // this will be a message
}

var messageMasks map[DummyMessage]uint32

type DummyElection struct {
	Seen map[*DummyMessage]bool
	Best *DummyMessage
}

func (e *DummyElection) Execute(m *DummyMessage) *DummyMessage {

	// more efficient to do this at a higher level but ... for now do it Paul's way
	// from my simplistic election this is just extra work
	if _, ok := e.Seen[m]; ok {
		return nil
	} // ignore messages I have seen
	e.Seen[m] = true // remember message  have seen

	if m.S > e.Best.S {
		e.Best = m
		return m // return messages that changed our state so they keep circulating till everyone has has a shot at them
	} else {
		return nil
	}
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
func clone(src []*DummyElection) []*DummyElection {
	dst := make([]*DummyElection, len(src))
	for i := 0; i < len(src); i++ {
		dst[i] = new(DummyElection)
		enc.Encode(src[i])
		dec.Decode(dst[i])
	}
	return dst[:]
}

// need better reflect based deep copy
func clone2(src *DummyElection) *DummyElection {
	dst := new(DummyElection)
	buff := new(bytes.Buffer)
	enc := gob.NewEncoder(buff)
	dec := gob.NewDecoder(buff)
	enc.Encode(src)
	dec.Decode(dst)
	return dst
}

func factorial(n int) int {
	factVal := 1
	if n < 0 {
		HandleError("Factorial of negative number doesn't exist.")
	} else {
		for i := 1; i <= n; i++ {
			factVal *= i
		}
	}
	return factVal // Wonder how much slower the recursive version is ...
}

func log2(x uint) uint {
	var i uint
	for i = 0; x < (1 << i); i++ {
		/* do nothing */
	}
	return i
}

/* Function to swap values at two indexes */
func swap(messages []*DummyMessage, x int, y int) {
	var temp *DummyMessage = messages[x]
	messages[x] = messages[y]
	messages[y] = temp
}

// Create all permutation of a set of messages
func permute(messages []*DummyMessage, l int, r int, results chan ([]*DummyMessage)) {
	if l == r {
		results <- messages
		return
	}
	for i := l; i <= r; i++ {
		swap(messages, l, i)
		permute(messages, l+1, r, results)
		swap(messages, l, i) //backtrack
	}
}

// check if a message exists in a list of messages
func notIn(messages []*DummyMessage, om *DummyMessage) bool {
	for _, m := range messages {
		if m.S == om.S {
			return false
		}
	}
	return true
}

// Test all message masks
func exhaustiveTest3(messages []*DummyMessage, nodes []*DummyElection, masks []int) {
	var outputMessages []*DummyMessage = make([]*DummyMessage, 0)

	for i := 0; i < len(messages); i++ {
		m := messages[i] // get the next message
		nodes2 := clone(nodes)
		for n := 0; n < len(nodes); n++ {
			output := (masks[n] & (1 << uint(i))) != 0 // check if we are sending this message
			if output {
				om := nodes2[n].Execute(m)
				if om != nil {
					outputMessages = append(outputMessages, om)
				}
			}
		} // for all nodes

		// OK we send one message to very node. Now see if we have to recurse for another layer
		if len(outputMessages) > 0 { // If we got output messages we have to descend
			newMessages := make([]*DummyMessage, 0)
			newMessages = append(newMessages, messages[i:]...) // append the unprocessed input messages
			for _, om := range outputMessages {
				if notIn(newMessages, om) {
					newMessages = append(newMessages, om) // append the non duplicate output messages
				}
			} // for all output messages
			exhaustiveTest1(newMessages, nodes2)
		}
	} // for all messages
}

// Test all message masks
func exhaustiveTest2(messages []*DummyMessage, nodes []*DummyElection) {
	var mCount int = len(messages)
	var nCount int = len(nodes)
	var mMax int = 1 << uint(mCount) // need one mask bit per message

	nodes2 := clone(nodes)
	nodes = nodes2
	for i := 0; i < len(nodes)*mMax; i++ {
		masks := make([]int, nCount)
		for j := 0; j < nCount; j++ {
			masks[j] = (i >> uint(j*nCount)) % mMax
		} // for each node
		exhaustiveTest3(messages, nodes, masks)

	} // for all nodes * all message masks
}

var level int
var nodeNames map[*DummyElection]string

func mString(m *DummyMessage) string { return m.S }
func mmstring(a []*DummyMessage) (rval string) {
	for _, m := range a {
		rval = rval + mString(m)
	}
	return rval
}

func nString(n *DummyElection) string {
	if n.Best == nil {
		return fmt.Sprintf("%s< > ", nodeNames[n])

	}
	return fmt.Sprintf("%s<%s> ", nodeNames[n], n.Best.S)
}

func mnString(a []*DummyElection) (rval string) {
	for _, n := range a {
		rval = rval + nString(n)
	}
	return rval
}

// test all permutation of message order
func exhaustiveTest1(messages []*DummyMessage, nodes []*DummyElection) {
	var results chan []*DummyMessage
	results = make(chan []*DummyMessage)

	//ok, so it's unused at this point
	if nodeNames == nil {
		nodeNames = make(map[*DummyElection]string, 0)
		for i := 0; i < len(nodes); i++ {
			nodes[i] = new(DummyElection)
			nodeNames[nodes[i]] = fmt.Sprintf("Node%d", i)
		}
	}

	fmt.Printf("Testing level %d messages[%+v] for nodes [%+v]\n", level, mmstring(messages), mnString(nodes))

	go permute(messages, 0, len(messages)-1, results)
	level++
	//
	for permutedMessages := range results {
		exhaustiveTest2(permutedMessages, nodes)
	} // for all message orders
	level--
}
