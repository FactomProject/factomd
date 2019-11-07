package p2p_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/FactomProject/factomd/p2p"
)

func TestNetworkID(t *testing.T) {
	var n NetworkID = MainNet
	if n.String() != "MainNet" {
		t.Errorf("Exp %s, got %s", "MainNet", n.String())
	}

	n = TestNet
	if n.String() != "TestNet" {
		t.Errorf("Exp %s, got %s", "TestNet", n.String())
	}

	n = LocalNet
	if n.String() != "LocalNet" {
		t.Errorf("Exp %s, got %s", "LocalNet", n.String())
	}

	n = 10
	if n.String() != fmt.Sprintf("CustomNet ID: %x\n", 10) {
		t.Errorf("Exp %s, got %s", fmt.Sprintf("CustomNet ID: %x\n", 10), n.String())
	}
}

func TestBlockFreeChannelSend(t *testing.T) {
	//BlockFreeChannelSend
	c := make(chan interface{}, 100)
	d := make(chan struct{}, 10)

	go addToBlockFree(c, d)

	// We will check to make sure the thread has finished, indicated the BlockFreeChannelSend
	// is non-blocking
	didnotblock := false
CheckForBlockingLoop:
	for i := 0; i < 10; i++ {
		select {
		case <-d:
			didnotblock = true
			break CheckForBlockingLoop
		default:
		}
		// Channel should never fill above 95%
		if len(c) > int(float64(cap(c))*0.95) {
			didnotblock = false
			break CheckForBlockingLoop
		}
		time.Sleep(20 * time.Millisecond)
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
				t.Errorf("Read in channel is incorrect. Exp %d, found %d", i, v)
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
