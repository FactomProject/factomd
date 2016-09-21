package main

import (
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
)

func TestChain(t *testing.T) {
	head, err := GetDBlockHead()
	if err != nil {
		t.Errorf("%v", err)
	}
	dHead, err := GetDBlock(head)
	if err != nil {
		t.Errorf("%v", err)
	}
	var h interfaces.IHash

	for _, e := range dHead.GetDBEntries() {
		if e.GetChainID().String() == "000000000000000000000000000000000000000000000000000000000000000a" {
			h = e.GetKeyMR()
			break
		}
	}

	for {
		if h.String() == "0000000000000000000000000000000000000000000000000000000000000000" {
			break
		}
		aBlock, err := GetABlock(h.String())
		if err != nil {
			t.Errorf("Error fetching %v - %v", h.String(), err)
			t.FailNow()
		}
		h = aBlock.GetHeader().GetPrevBackRefHash()
		t.Logf("Fetched ABlock #%v", aBlock.GetDBHeight())
	}

}
