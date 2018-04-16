// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	//"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"github.com/FactomProject/factomd/util"
	log "github.com/sirupsen/logrus"
)

var _ = log.Print
var _ = util.ReadConfig

func TestInit(t *testing.T) {
	s := testHelper.CreateEmptyTestState()
	PrintState(s)
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

func TestStateKeys(t *testing.T) {
	s := testHelper.CreateEmptyTestState()
	sec := primitives.RandomPrivateKey()
	s.SimSetNewKeys(sec)
	act := s.GetServerPrivateKey()
	if act.PublicKeyString() != sec.PublicKeyString() {
		t.Error("Public key is not correct")
	}

	if act.PrivateKeyString() != sec.PrivateKeyString() {
		t.Error("Public key is not correct")
	}

	var _ = s.GetStatus()

}

/*
func TestDirBlockHead(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()
	height := state.GetHighestSavedBlk()
	if height != 1 {
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

	if d.GetKeyMR().String() != "3bafce89724fab70d40e1c4bd534b15250987198715ce360bff38e73424e13f0" {
		t.Errorf("Invalid DBLock KeyMR - got %v, expected 3bafce89724fab70d40e1c4bd534b15250987198715ce360bff38e73424e13f0", d.GetKeyMR().String())
	}
}
*/

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

func TestLoadHoldingMap(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()

	hque := state.LoadHoldingMap()
	if len(hque) != len(state.HoldingMap) {
		t.Errorf("Error with Holding Map Length")
	}
}

func TestLoadAcksMap(t *testing.T) {
	state := testHelper.CreateAndPopulateTestState()

	hque := state.LoadAcksMap()
	if len(hque) != len(state.HoldingMap) {
		t.Errorf("Error with Acks Map Length")
	}

}

func TestCalculateTransactionRate(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	to, _ := s.CalculateTransactionRate()
	time.Sleep(3 * time.Second)

	s.FactoidTrans = 333
	to2, i := s.CalculateTransactionRate()
	if to >= to2 {
		t.Errorf("Rate should be higher than %f, found %f", to, to2)
	}
	if i < 30 {
		t.Errorf("Instant transaction rate should be > 30 (roughly), found %f", i)
	}

}

func TestClone(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	s2i := s.Clone(1)
	s2, ok := s2i.(*State)
	if !ok {
		t.Error("Clone failed")
	}
	if s2.GetFactomNodeName() != "FNode01" {
		t.Error("Factom Node Name incorrect")
	}
	s.AddPrefix("x")
	s3i := s.Clone(2)
	s3, ok := s3i.(*State)
	if !ok {
		t.Error("Clone failed")
	}
	if s3.GetFactomNodeName() != "xFNode02" {
		t.Error("Factom Node Name incorrect")
	}
}

func TestLog(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	buf := new(bytes.Buffer)
	//s.Logger = log.New(buf, "debug", "unit_test")
	log.SetOutput(buf)
	log.SetLevel(log.DebugLevel)

	var levels []string = []string{"debug", "info", "warning", "error"}
	for _, l := range levels {
		msg := "A test message"
		s.Logf(l, "%s", msg)

		data := buf.Next(buf.Len())
		if !strings.Contains(string(data), msg) {
			t.Error("Logf did not log the msg")
		}

		msg2 := "Another test message"
		s.Log(l, msg2)
		data = buf.Next(buf.Len())
		if !strings.Contains(string(data), msg2) {
			t.Error("Log did not log the msg for level", l)
		}
	}

}

func TestSetKeys(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	p, _ := primitives.NewPrivateKeyFromHex("0000000000000000000000000000000000000000000000000000000000000000")
	s.SimSetNewKeys(p)

	if s.SimGetSigKey() != p.PublicKeyString() {
		t.Error("Public keys do not match")
	}
}

func TestPrintState(t *testing.T) {
	s := testHelper.CreateAndPopulateTestState()
	PrintState(s)
}

/*
func (s *State) SimSetNewKeys(p *primitives.PrivateKey) {
	s.serverPrivKey = p
	s.serverPubKey = p.Pub
}

func (s *State) SimGetSigKey() string {
	return s.serverPrivKey.Pub.String()
}

*/

/*
func TestBootStrappingIdentity(t *testing.T) {
	state := testHelper.CreateEmptyTestState()

	state.NetworkNumber = constants.NETWORK_MAIN
	if !state.GetNetworkBootStrapIdentity().IsSameAs(primitives.NewZeroHash()) {
		t.Errorf("Bootstrap Identity Mismatch on MAIN")
	}
	key, _ := primitives.HexToHash("0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a")
	if !state.GetNetworkBootStrapKey().IsSameAs(key) {IsInPendingEntryList
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
*/

func TestIsStalled(t *testing.T) {
	s := testHelper.CreateEmptyTestState()
	s.Syncing = false
	s.ProcessLists.DBHeightBase = 20
	s.CurrentMinuteStartTime = time.Now().UnixNano()
	if !s.IsStalled() {
		t.Error("Should be stalled as we are behind: Stalled:", s.IsStalled())
	}

	s.CurrentMinuteStartTime = 0
	s.ProcessLists.DBHeightBase = 0

	if s.IsStalled() {
		t.Error("When current minute start is 0, should not say stalled")
	}

	n := time.Now()
	then := n.Add(-1600 * time.Millisecond)
	s.CurrentMinuteStartTime = then.UnixNano()
	s.DirectoryBlockInSeconds = 10

	if !s.IsStalled() {
		t.Error("Should be stalled as 1.6x blktime behind")
	}

	then = time.Now().Add(-1200 * time.Millisecond)
	s.CurrentMinuteStartTime = then.UnixNano()
	if s.IsStalled() {
		t.Error("Should not be stalled as 1.2x blktime behind")
	}

}
