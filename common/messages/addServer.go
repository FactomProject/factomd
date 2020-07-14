// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	llog "github.com/FactomProject/factomd/log"
	log "github.com/sirupsen/logrus"
)

// Communicate a Directory Block State

type AddServerMsg struct {
	msgbase.MessageBase
	Timestamp     interfaces.Timestamp // Message Timestamp
	ServerChainID interfaces.IHash     // ChainID of new server
	ServerType    int                  // 0 = Federated, 1 = Audit

	// Unsorted list of unvalidated pubkey/sig pairs
	Signatures interfaces.IFullSignatureBlock

	sigMtx   sync.Mutex
	sigCache bool // true if this set of Signatures has been verified
	// List of validated pubkey/sig pairs, sorted by byte order
	validSignatures []interfaces.IFullSignature
}

var _ interfaces.IMsg = (*AddServerMsg)(nil)
var _ interfaces.MultiSignable = (*AddServerMsg)(nil)

func (m *AddServerMsg) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AddServerMsg.GetRepeatHash") }()

	return m.GetMsgHash()
}

func (m *AddServerMsg) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AddServerMsg.GetHash") }()

	return m.GetMsgHash()
}

func (m *AddServerMsg) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AddServerMsg.GetMsgHash") }()

	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *AddServerMsg) Type() byte {
	return constants.ADDSERVER_MSG
}

func (m *AddServerMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp.Clone()
}

// Validate takes the set of signatures in the message and compares them to the state's current authority set.
// If more than 50% ( >= n/2+1) of identities in the authority set (fed+audit) have signed the message, it is valid.
func (m *AddServerMsg) Validate(state interfaces.IState) int {
	auth := state.GetAuthorities()
	valid, err := m.VerifySignatures()
	if err != nil { // unable to marshal
		return -1
	}

	// not enough signatures
	if len(valid) < (len(auth)/2 + 1) {
		return -1
	}

	realKeys := 0
	for _, v := range valid {
		for _, a := range auth {
			if bytes.Equal(v.GetKey(), a.GetSigningKey()) {
				realKeys++
				break
			}
		}
	}

	if realKeys < len(auth)/2+1 {
		return -1
	}

	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *AddServerMsg) ComputeVMIndex(state interfaces.IState) {
	m.VMIndex = state.ComputeVMIndex(constants.ADMIN_CHAINID)
}

// Execute the leader functions of the given message
func (m *AddServerMsg) LeaderExecute(state interfaces.IState) {
	state.LeaderExecute(m)
}

func (m *AddServerMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

func (e *AddServerMsg) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessAddServer(dbheight, e)
}

func (e *AddServerMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AddServerMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *AddServerMsg) AddSignature(key interfaces.Signer) error {
	data, err := m.MarshalForSignature()
	if err != nil {
		return err
	}
	signature := key.Sign(data)

	m.sigMtx.Lock()
	defer m.sigMtx.Unlock()
	m.sigCache = false
	m.validSignatures = nil

	m.Signatures.AddSignature(signature)
	return nil
}

func (m *AddServerMsg) GetSignatures() []interfaces.IFullSignature {
	return m.Signatures.GetSignatures()
}

func (m *AddServerMsg) VerifySignatures() ([]interfaces.IFullSignature, error) {
	m.sigMtx.Lock()
	defer m.sigMtx.Unlock()

	if m.sigCache {
		return m.validSignatures, nil
	}

	data, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}

	sigs := m.Signatures.GetSignatures()

	duplicate := make(map[string]bool)
	valid := make([]interfaces.IFullSignature, 0, len(sigs)) // might be fewer if duplicate
	for _, sig := range sigs {
		key := fmt.Sprintf("%x", sig.GetKey())
		if !duplicate[key] && sig.Verify(data) {
			duplicate[key] = true
			valid = append(valid, sig)
		}
	}

	sort.Slice(valid, func(i, j int) bool {
		return bytes.Compare(valid[i].GetKey(), valid[j].GetKey()) < 0
	})

	m.validSignatures = valid
	return valid, nil
}

func (m *AddServerMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Add Server Message: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling Add Server Message: %v", r)
		}
	}()
	newData = data
	if newData[0] != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}
	newData = newData[1:]

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.ServerChainID = new(primitives.Hash)
	newData, err = m.ServerChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.ServerType = int(newData[0])
	newData = newData[1:]

	m.sigMtx.Lock()
	defer m.sigMtx.Unlock()
	m.sigCache = false
	m.validSignatures = nil
	m.Signatures = new(factoid.FullSignatureBlock)

	for len(newData) > 32+64 {
		sig := new(primitives.Signature)
		newData, err = sig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Signatures.AddSignature(sig)
	}
	return
}

func (m *AddServerMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *AddServerMsg) MarshalForSignature() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddServerMsg.MarshalForSignature err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	d, err := m.ServerChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(d)

	binary.Write(&buf, binary.BigEndian, uint8(m.ServerType))

	return buf.DeepCopyBytes(), nil
}

func (m *AddServerMsg) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddServerMsg.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	data, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.Signatures.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.DeepCopyBytes(), nil
}

func (m *AddServerMsg) String() string {
	var stype string
	if m.ServerType == 0 {
		stype = "Federated"
	} else {
		stype = "Audit"
	}
	return fmt.Sprintf("AddServer (%s): ChainID: %x Time: %x Msg Hash %x ",
		stype,
		m.ServerChainID.Bytes()[3:6],
		&m.Timestamp,
		m.GetMsgHash().Bytes()[:3])

}

func (m *AddServerMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "addserver", "server": m.ServerChainID.String(),
		"hash": m.GetHash().String()}
}

func (m *AddServerMsg) IsSameAs(b *AddServerMsg) bool {
	if b == nil {
		return false
	}
	if m.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}
	if !m.ServerChainID.IsSameAs(b.ServerChainID) {
		return false
	}
	if m.ServerType != b.ServerType {
		return false
	}

	if !m.Signatures.IsSameAs(b.Signatures) {
		return false
	}
	return true
}

func NewAddServerMsg(state interfaces.IState, serverType int) interfaces.IMsg {
	msg := new(AddServerMsg)
	msg.ServerChainID = state.GetIdentityChainID()
	msg.ServerType = serverType
	msg.Timestamp = state.GetTimestamp()
	msg.Signatures = factoid.NewFullSignatureBlock()

	return msg

}

func NewAddServerByHashMsg(state interfaces.IState, serverType int, newServerHash interfaces.IHash) interfaces.IMsg {
	msg := new(AddServerMsg)
	msg.ServerChainID = newServerHash
	msg.ServerType = serverType
	msg.Timestamp = state.GetTimestamp()
	msg.Signatures = factoid.NewFullSignatureBlock()
	return msg
}
