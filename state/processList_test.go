// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	"github.com/FactomProject/factomd/common/interfaces"
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

	//Get(dbheight uint32) *ProcessList
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

func TestGetNewEBlocks(t *testing.T) {
	//GetNewEBlocks(key [32]byte) interfaces.IEntryBlock
}

func TestGetCommits(t *testing.T) {
	//GetCommits(key [32]byte) interfaces.IMsg
}

func TestPutNewEBlocks(t *testing.T) {
	//PutNewEBlocks(dbheight uint32, key interfaces.IHash, value interfaces.IEntryBlock)
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
	//(p *ProcessList) AddToProcessList(ack *messages.Ack, m interfaces.IMsg)
}

func TestPutCommits(t *testing.T) {
	//(p *ProcessList) PutCommits(key interfaces.IHash, value interfaces.IMsg)
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
