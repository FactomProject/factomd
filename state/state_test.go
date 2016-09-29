// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"fmt"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/util"
)

var _ = log.Print
var _ = util.ReadConfig

func TestInit(t *testing.T) {
	testHelper.CreateEmptyTestState()
}

func TestSecretCode(t *testing.T) {
	s := new(state.State)
	ts1 := s.GetTimestamp()
	num1 := s.GetSecretNumber(ts1)
	num2 := s.GetSecretNumber(ts1)
	if num1 != num2 {
		t.Error("Secret Number failure")
	}
	ts1.SetTime(uint64(ts1.GetTimeMilli() + 1000))
	num3 := s.GetSecretNumber(ts1)
	if num1 == num3 {
		t.Error("Secret Number bad match")
	}
	fmt.Printf("Secret Numbers %x %x %x\n", num1, num2, num3)
}

func TestDirBlockHead(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	height := state.GetHighestCompletedBlock()
	if height != 9 {
		t.Errorf("Invalid DBLock Height - got %v, expected 10", height+1)
	}
	d := state.GetDirectoryBlockByHeight(height)

	//fmt.Println(d)
	//fmt.Println("------------")
	//fmt.Println(d.String())
	//data, _ := d.MarshalBinary()
	//fmt.Printf("%x\n", data)
	//fmt.Printf("nwtwork number %d\n", state.NetworkNumber)
	//fmt.Printf("network id %x\n", d.GetHeader().GetNetworkID())
	//fmt.Printf("network id %x\n", d.GetHeader.GetBodyMR)

	if d.GetKeyMR().String() != "12d6c012e3598ca1c10dbf60ac12af9fa8904b8fd98968e86f4c66c14884c225" {
		t.Errorf("Invalid DBLock KeyMR - got %v, expected 12d6c012e3598ca1c10dbf60ac12af9fa8904b8fd98968e86f4c66c14884c225", d.GetKeyMR().String())
	}
}

func TestGetDirectoryBlockByHeight(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	blocks := testHelper.CreateFullTestBlockSet()
	for i, block := range blocks {
		dBlock := state.GetDirectoryBlockByHeight(uint32(i))
		if dBlock.GetKeyMR().IsSameAs(block.DBlock.GetKeyMR()) == false {
			t.Errorf("DBlocks are not the same at height %v", i+1)
			continue
		}
		if dBlock.GetFullHash().IsSameAs(block.DBlock.GetFullHash()) == false {
			t.Errorf("DBlocks are not the same at height %v", i+1)
			continue
		}
	}
}
