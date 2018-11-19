// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/sirupsen/logrus"
)

var faultLogger = packageLogger.WithFields(log.Fields{"subpack": "fault"})

type FaultCore struct {
	// The following 5 fields represent the "Core" of the message
	// This should match the Core of FullServerFault messages
	ServerID      interfaces.IHash
	AuditServerID interfaces.IHash
	VMIndex       byte
	DBHeight      uint32
	Height        uint32
	SystemHeight  uint32
	Timestamp     interfaces.Timestamp
}

func (fc *FaultCore) GetHash() interfaces.IHash {
	data, err := fc.MarshalCore()
	if err != nil {
		return nil
	}
	return primitives.Sha(data)
}

func (fc *FaultCore) MarshalCore() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Server Fault Core: %v", r)
		}
	}()

	var buf primitives.Buffer

	if d, err := fc.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := fc.AuditServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(fc.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(fc.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(fc.Height))
	binary.Write(&buf, binary.BigEndian, uint32(fc.SystemHeight))

	if d, err := fc.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func markFault(pl *ProcessList, vmIndex int, faultReason int) {
	// We can use the "IgnoreMissing" boolean to track if enough time has elapsed
	// since bootup to start faulting servers on the network
	if pl.State.IgnoreMissing {
		return
	}

	if pl.State.Leader && pl.State.LeaderVMIndex == vmIndex {
		return
	}

	now := time.Now().Unix()
	vm := pl.VMs[vmIndex]

	if vm.WhenFaulted == 0 {
		// if we did not previously consider this VM faulted
		// we simply mark it as faulted (by assigning it a nonzero WhenFaulted time)
		// and keep track of the ProcessList height it has faulted at
		vm.WhenFaulted = now
		vm.FaultFlag = faultReason
	}

	c := pl.State.CurrentMinute
	if c > 9 {
		c = 9
	}
	index := pl.ServerMap[c][vmIndex]
	if index < len(pl.FedServers) {
		pl.FedServers[index].SetOnline(false)
	}
}

func markNoFault(pl *ProcessList, vmIndex int) {
	vm := pl.VMs[vmIndex]

	vm.WhenFaulted = 0
	vm.FaultFlag = -1

	nextIndex := (vmIndex + 1) % len(pl.FedServers)
	if pl.VMs[nextIndex].FaultFlag > 0 {
		markNoFault(pl, nextIndex)
	}

	c := pl.State.CurrentMinute
	if c > 9 {
		c = 9
	}
	index := pl.ServerMap[c][vmIndex]
	if index < len(pl.FedServers) {
		pl.FedServers[index].SetOnline(true)
	}
}

func FaultCheck(pl *ProcessList) {
	now := time.Now().Unix()

	for i := 0; i < len(pl.FedServers); i++ {
		if i == pl.State.LeaderVMIndex {
			continue
		}
		vm := pl.VMs[i]
		if vm.WhenFaulted > 0 && int(now-vm.WhenFaulted) > pl.State.FaultTimeout*2 {
			newVMI := (i + 1) % len(pl.FedServers)
			markFault(pl, newVMI, 1)
		}

	}
	return
}

func (s *State) FollowerExecuteSFault(m interfaces.IMsg) {
}

// When we execute a FullFault message, it could be complete (includes all
// necessary signatures + pledge) or incomplete, in which case it is just
// a negotiation ping
func (s *State) FollowerExecuteFullFault(m interfaces.IMsg) {

}

func (s *State) Reset() {
	// We are no longer using Reset
	// s.ResetRequest = true
}

// Set to reprocess all messages and states
func (s *State) DoReset() {
	return
}
