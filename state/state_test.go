// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"testing"

	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/primitives"
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
	num1 := s.GetSalt(ts1)
	num2 := s.GetSalt(ts1)
	if num1 != num2 {
		t.Error("Secret Number failure")
	}
	ts1.SetTime(uint64(ts1.GetTimeMilli() + 1000))
	num3 := s.GetSalt(ts1)
	if num1 == num3 {
		t.Error("Secret Number bad match")
	}
	fmt.Printf("Secret Numbers %x %x %x\n", num1, num2, num3)
}

func TestDirBlockHead(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	height := state.GetHighestSavedBlk()
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

func TestBootStrappingIdentity(t *testing.T) {
	state := testHelper.CreateEmptyTestState()

	state.NetworkNumber = constants.NETWORK_MAIN
	if !state.GetNetworkBootStrapIdentity().IsSameAs(primitives.NewZeroHash()) {
		t.Errorf("Bootstrap Identity Mismatch on MAIN")
	}
	key, _ := primitives.HexToHash("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a")
	if !state.GetNetworkBootStrapKey().IsSameAs(key) {
		t.Errorf("Bootstrap Identity Key Mismatch on MAIN")
	}

	state.NetworkNumber = constants.NETWORK_TEST
	if !state.GetNetworkBootStrapIdentity().IsSameAs(primitives.NewZeroHash()) {
		t.Errorf("Bootstrap Identity Mismatch on TEST")
	}

	key, _ = primitives.HexToHash("49b6edd274e7d07c94d4831eca2f073c207248bde1bf989d2183a8cebca227b7")
	if !state.GetNetworkBootStrapKey().IsSameAs(key) {
		t.Errorf("Bootstrap Identity Key Mismatch on TEST")
	}

	state.NetworkNumber = constants.NETWORK_LOCAL
	id, _ := primitives.HexToHash("38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9")
	if !state.GetNetworkBootStrapIdentity().IsSameAs(id) {
		t.Errorf("Bootstrap Identity Mismatch on LOCAL")
	}
	key, _ = primitives.HexToHash("cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a")
	if !state.GetNetworkBootStrapKey().IsSameAs(key) {
		t.Errorf("Bootstrap Identity Key Mismatch on LOCAL")
	}

	state.NetworkNumber = constants.NETWORK_CUSTOM
	id, _ = primitives.HexToHash("38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9")
	if !state.GetNetworkBootStrapIdentity().IsSameAs(id) {
		t.Errorf("Bootstrap Identity Mismatch on CUSTOM")
	}
	key, _ = primitives.HexToHash("cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a")
	if !state.GetNetworkBootStrapKey().IsSameAs(key) {
		t.Errorf("Bootstrap Identity Key Mismatch on CUSTOM")
	}

}
