package p2p_test

import (
	"testing"

	. "github.com/FactomProject/factomd/p2p"
)

func TestBlockFreeChannelSend(t *testing.T) {
	//BlockFreeChannelSend

	c := make(chan interface{}, StandardChannelSize)

	for i := 0; i < StandardChannelSize; i++ {
		BlockFreeChannelSend(c, i)
	}

	maxLen := 8990
	if len(c) != maxLen {
		t.Errorf("Invalid channel length - %v vs %v", len(c), maxLen)
	}

	offset := 1010
	for i := 0; i < maxLen; i++ {
		interf := <-c
		j := interf.(int)
		if i+offset != j {
			t.Errorf("Invalid message returned - %v vs %v", i+offset, j)
		}
	}
}
