package state_test

import (
	"testing"

	"time"

	"math/rand"

	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/state"
)

// The lists that are sorted in an increasing order
type GenericIncreasingList interface {
	//Len() int
	Add(uint32)
	Del(uint32)
}

type RecievedOverrideList struct {
	*state.StatesReceived
}

func (l *RecievedOverrideList) Add(height uint32) {
	msg := new(messages.DBStateMsg)
	newdb := new(directoryBlock.DirectoryBlock)
	header := new(directoryBlock.DBlockHeader)
	header.DBHeight = height
	newdb.Header = header
	msg.DirectoryBlock = newdb

	l.StatesReceived.Add(height, msg)
}

func TestWaitingListThreadSafety(t *testing.T) {
	list := state.NewStatesWaiting()
	testListThreadSafety(list, t)
}

func TestRecievedListThreadSafety(t *testing.T) {
	list := state.NewStatesReceived()
	override := new(RecievedOverrideList)
	override.StatesReceived = list
	testListThreadSafety(override, t)
}

func TestMissingListThreadSafety(t *testing.T) {
	list := state.NewStatesMissing()
	testListThreadSafety(list, t)
}

func testListThreadSafety(list GenericIncreasingList, t *testing.T) {
	done := false

	added := make(chan int, 1000)

	// Many random adds
	adds := func() {
		for {
			if done {
				return
			}

			list.Add(uint32(rand.Intn(1000)))
			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		}
	}

	// Random removes
	dels := func() {
		for {
			if done {
				return
			}

			n := <-added
			list.Del(uint32(n))
			time.Sleep(time.Duration(rand.Intn(100)) * time.Microsecond)
		}
	}

	for i := 0; i < 5; i++ {
		go adds()
		go dels()
	}

	timer := make(chan bool)
	go func() {
		time.Sleep(5 * time.Second)
		timer <- true
	}()

	<-timer
	done = true
	// Unit test will panic if there is race conditions
}

func TestDBStateListAdditionsMissing(t *testing.T) {
	list := state.NewStatesMissing()

	// Check overlapping adds and out of order
	for i := 50; i < 100; i++ {
		list.Add(uint32(i))
	}
	for i := 70; i >= 0; i-- {
		list.Add(uint32(i))
	}

	if list.Len() != 100 {
		t.Errorf("Exp len of 100, found %d", list.Len())
	}

	// Check out of order retrievals
	for i := 0; i < 100; i++ {
		r := uint32(rand.Intn(100)) // Random spot check
		h := list.Get(r)
		if h.Height() != r {
			t.Errorf("Random retrival failed. Exp %d, found %d", r, h.Height())
		}
	}

	// Check sorted list
	for i := 0; i < 100; i++ {
		if h := list.GetNext(); h.Height() != uint32(i) {
			t.Errorf("Exp %d, found %d", i, h.Height())
		}
	}

	if list.Len() != 0 {
		t.Errorf("Exp len of 0, found %d", list.Len())
	}
}
