package state_test

import (
	"testing"

	"time"

	"math/rand"

	"github.com/FactomProject/factomd/state"
)

func TestListThreadSafety(t *testing.T) {
	list := state.NewStatesMissing()
	done := false

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

			n, _ := list.NextConsecutiveMissing(1)
			list.Del(n)
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
