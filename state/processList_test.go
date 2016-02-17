// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	. "github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestUpdateState(t *testing.T) {
	//UpdateState()
}

func TestGet(t *testing.T) {
	p := createProcessLists()
	index := 5
	list := p.Get(uint32(index))

	if list == nil {
		t.Errorf("Wrong Get")
	}
	if list.DBHeight != uint32(index) {
		t.Errorf("Wrong Get")
	}

	if len(p.Lists) != index+1 {
		t.Errorf("Wrong len of Lists - %v vs %v", len(p.Lists), index+1)
	}

	for i := 0; i < index; i++ {
		if p.Lists[i] != nil {
			t.Errorf("List isn't nil when it should be")
		}
	}

	if p.Lists[index] == nil {
		t.Errorf("List is nil when it shouldn't be")
	}

	list = p.Get(uint32(index * 2))
	if list == nil {
		t.Errorf("Wrong Get")
	}
	if list.DBHeight != uint32(index*2) {
		t.Errorf("Wrong Get")
	}
	if len(p.Lists) != index*2+1 {
		t.Errorf("Wrong len of Lists - %v vs %v", len(p.Lists), index*2+1)
	}
}

func TestGetLen(t *testing.T) {
	p := createProcessList()

	for i := range p.Servers {
		if p.GetLen(i) != len(p.Servers)-i {
			t.Errorf("Wrong GetLen - %v vs %v", p.GetLen(i), len(p.Servers)-i)
		}
	}
}

func TestSetSigComplete(t *testing.T) {
	p := createProcessList()

	p.ServerIndex = 1

	p.SetSigComplete(true)

	if p.Servers[p.ServerIndex].SigComplete != true {
		t.Errorf("Wrong server.SigComplete")
	}

	p.SetSigComplete(false)

	if p.Servers[p.ServerIndex].SigComplete != false {
		t.Errorf("Wrong server.SigComplete")
	}

	p.ServerIndex = 3

	p.SetSigComplete(true)

	if p.Servers[p.ServerIndex].SigComplete != true {
		t.Errorf("Wrong server.SigComplete")
	}

	p.SetSigComplete(false)

	if p.Servers[p.ServerIndex].SigComplete != false {
		t.Errorf("Wrong server.SigComplete")
	}
}

func TestSetEomComplete(t *testing.T) {
	p := createProcessList()

	p.ServerIndex = 1

	p.SetEomComplete(true)

	if p.Servers[p.ServerIndex].EomComplete != true {
		t.Errorf("Wrong server.EomComplete")
	}

	p.SetEomComplete(false)

	if p.Servers[p.ServerIndex].EomComplete != false {
		t.Errorf("Wrong server.EomComplete")
	}

	p.ServerIndex = 3

	p.SetEomComplete(true)

	if p.Servers[p.ServerIndex].EomComplete != true {
		t.Errorf("Wrong server.EomComplete")
	}

	p.SetEomComplete(false)

	if p.Servers[p.ServerIndex].EomComplete != false {
		t.Errorf("Wrong server.EomComplete")
	}
}

func TestPutAndGetNewEBlocks(t *testing.T) {
	p := createProcessList()
	var prev interfaces.IEntryBlock = nil
	for i := 0; i < 10; i++ {
		prev, _ = testHelper.CreateTestEntryBlock(prev)
		h, err := prev.Hash()
		if err != nil {
			t.Errorf("v", err)
		}
		p.PutNewEBlocks(100, h, prev)
	}
	prev = nil
	for i := 0; i < 10; i++ {
		prev, _ = testHelper.CreateTestEntryBlock(prev)
		h, err := prev.Hash()
		if err != nil {
			t.Errorf("v", err)
		}
		block := p.GetNewEBlocks(h)
		h1, err := block.Hash()
		if err != nil {
			t.Errorf("v", err)
		}
		h2, err := prev.Hash()
		if err != nil {
			t.Errorf("v", err)
		}
		if h1.IsSameAs(h2) == false {
			t.Errorf("Blocks are not equal")
		}
		h1, err = block.HeaderHash()
		if err != nil {
			t.Errorf("v", err)
		}
		h2, err = prev.HeaderHash()
		if err != nil {
			t.Errorf("v", err)
		}
		if h1.IsSameAs(h2) == false {
			t.Errorf("Blocks are not equal")
		}
	}
}

func TestPutAndGetCommits(t *testing.T) {
	p := createProcessList()
	m1 := new(messages.Ack)
	k1, err := primitives.NewShaHashFromStr("0000000000000000000000000000000000000000000000000000000000000001")
	if err != nil {
		t.Errorf("v", err)
	}
	p.PutCommits(k1, m1)

	m2 := new(messages.CommitChainMsg)
	k2, err := primitives.NewShaHashFromStr("0000000000000000000000000000000000000000000000000000000000000002")
	if err != nil {
		t.Errorf("v", err)
	}
	p.PutCommits(k2, m2)

	m3 := new(messages.CommitEntryMsg)
	k3, err := primitives.NewShaHashFromStr("0000000000000000000000000000000000000000000000000000000000000003")
	if err != nil {
		t.Errorf("v", err)
	}
	p.PutCommits(k3, m3)
	p.PutCommits(k3, m3)
	p.PutCommits(k3, m3)

	_, ok := p.GetCommits(k1).(*messages.Ack)
	if ok == false {
		t.Errorf("Wrong message type")
	}
	mm2, ok := p.GetCommits(k2).(*messages.CommitChainMsg)
	if ok == false {
		t.Errorf("Wrong message type")
	}
	if mm2.GetCount() != 0 {
		t.Errorf("mm2.GetCount()!=0 - %v", mm2.GetCount())
	}
	mm3, ok := p.GetCommits(k3).(*messages.CommitEntryMsg)
	if ok == false {
		t.Errorf("Wrong message type")
	}
	if mm3.GetCount() != 2 {
		t.Errorf("mm3.GetCount()!=2 - %v", mm3.GetCount())
	}
}

func TestComplete(t *testing.T) {
	p := createProcessList()

	if p.Complete() != false {
		t.Errorf("Wrong Complete")
	}

	for i := 1; i < len(p.Servers); i++ {
		p.Servers[i].SigComplete = true
		if p.Complete() != false {
			t.Errorf("Wrong Complete")
		}
	}
	p.Servers[0].SigComplete = true
	if p.Complete() != true {
		t.Errorf("Wrong Complete")
	}
	p.SetComplete(false)
	if p.Complete() != false {
		t.Errorf("Wrong Complete")
	}
	p.SetComplete(true)
	if p.Complete() != true {
		t.Errorf("Wrong Complete")
	}
}

func TestSetComplete(t *testing.T) {
	p := createProcessList()

	p.SetComplete(true)

	for i := range p.Servers {
		if p.Servers[i].SigComplete != true {
			t.Errorf("Wrong server.SigComplete")
		}
	}

	p.SetComplete(false)

	for i := range p.Servers {
		if p.Servers[i].SigComplete != false {
			t.Errorf("Wrong server.SigComplete")
		}
	}
}

func TestProcess(t *testing.T) {
	//Process(state interfaces.IState)
}

func TestAddToProcessList(t *testing.T) {
	p := createProcessList()
	max := len(p.Servers)
	max2 := 10
	for i := 0; i < max; i++ {
		for j := 0; j < max2; j++ {
			ack := new(messages.Ack)
			ack.ServerIndex = byte(i)
			ack.Height = uint32(j)
			p.AddToProcessList(ack, ack)
		}
	}
	for i := range p.Servers {
		if len(p.Servers[i].List) != max2 {
			t.Errorf("Wrong len of server List")
		}
		for j, msg := range p.Servers[i].List {
			m, ok := msg.(*messages.Ack)
			if ok == false {
				t.Errorf("Wrong message type")
			}
			if m.ServerIndex != byte(i) || m.Height != uint32(j) {
				t.Errorf("Wrong data loaded from message")
			}
		}
	}
}

func TestNewProcessLists(t *testing.T) {
	//NewProcessLists(state interfaces.IState) *ProcessLists
}

func TestNewProcessList(t *testing.T) {
	//NewProcessList(totalServers int, dbheight uint32) *ProcessList
}

func createProcessLists() *ProcessLists {
	state := testHelper.CreateAndPopulateTestState()
	p := NewProcessLists(state)
	return p
}

func createProcessList() *ProcessList {
	serverCount := 5
	p := NewProcessList(serverCount, 10)
	p.ServerIndex = 1

	for i := range p.Servers {
		p.Servers[i].List = make([]interfaces.IMsg, serverCount-i)
	}

	return p
}
