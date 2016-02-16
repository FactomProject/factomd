// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state_test

import (
	. "github.com/FactomProject/factomd/state"
	"testing"
)

func TestUpdateState(t *testing.T) {
	//UpdateState()
}

func TestGet(t *testing.T) {
	//Get(dbheight uint32) *ProcessList
}

func TestGetLen(t *testing.T) {
	//GetLen(list int) int
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

	//SetSigComplete(value bool)
}

func TestSetEomComplete(t *testing.T) {
	//SetEomComplete(value bool)
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
	//Complete() bool
}

func TestSetComplete(t *testing.T) {
	//SetComplete(v bool)
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

func createProcessList() *ProcessList {
	p := NewProcessList(5, 10)
	p.ServerIndex = 1
	return p
}
