package p2p_test

import (
	"testing"
	"time"

	. "github.com/FactomProject/factomd/p2p"
)

func TestBlockFreeChannelSend(t *testing.T) {
	//BlockFreeChannelSend
	c := make(chan interface{}, 100)
	d := make(chan struct{}, 10)

	go addToBlockFree(c, d)

	// We will check to make sure the thread has finished, indicated the BlockFreeChannelSend
	// is non-blocking
	didnotblock := false
CheckForBlockingLoop:
	for c := 0; c < 10; c++ {
		select {
		case <-d:
			didnotblock = true
			break CheckForBlockingLoop
		default:
		}
		time.Sleep(5 * time.Millisecond)

	}

	if !didnotblock {
		t.Error("BlockFreeChannelSend has blocked")
		t.FailNow()
	}

	first := true
	for i := 0; i < cap(c)*2; i++ {
		select {
		case v := <-c:
			// Set the first value. It's an offset
			if first {
				first = false
				i = v.(int)
				break
			}
			if v.(int) != i {
				t.Errorf("Value in channel is incorrect. Exp %d, found %d", i, v)
			}
		default:
			t.Error("Channel is empty, but still expect elements")
		}
	}
}

// addToBlockFree adds 2 times as many elements as the channel can handle
func addToBlockFree(c chan interface{}, done chan struct{}) {
	for i := 0; i < cap(c)*2; i++ {
		BlockFreeChannelSend(c, i)
	}
	done <- struct{}{}
}
