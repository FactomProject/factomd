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
	"github.com/FactomProject/factomd/common/messages/msgbase"
	"github.com/FactomProject/factomd/common/primitives"

	llog "github.com/FactomProject/factomd/log"
	log "github.com/sirupsen/logrus"
)

// Communicate a Directory Block State

type RemoveServerMsg struct {
	msgbase.MessageBase
	Timestamp     interfaces.Timestamp // Message Timestamp
	ServerChainID interfaces.IHash     // ChainID of new server
	ServerType    int                  // 0 = Federated, 1 = Audit

	Signatures interfaces.IFullSignatureBlock

	sigMtx   sync.Mutex
	sigCache bool // true if this set of Signatures has been verified
	// List of validated pubkey/sig pairs, sorted by byte order
	validSignatures []interfaces.IFullSignature
}

var _ interfaces.IMsg = (*RemoveServerMsg)(nil)
var _ interfaces.MultiSignable = (*RemoveServerMsg)(nil)

func (m *RemoveServerMsg) GetRepeatHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RemoveServerMsg.GetRepeatHash") }()

	return m.GetMsgHash()
}

func (m *RemoveServerMsg) GetHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RemoveServerMsg.GetHash") }()

	return m.GetMsgHash()
}

func (m *RemoveServerMsg) GetMsgHash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RemoveServerMsg.GetMsgHash") }()

	if m.MsgHash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *RemoveServerMsg) Type() byte {
	return constants.REMOVESERVER_MSG
}

func (m *RemoveServerMsg) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp.Clone()
}

func (m *RemoveServerMsg) Validate(state interfaces.IState) int {
	auth := state.GetAuthorities()
	found := false
	for _, a := range auth {
		if a.GetAuthorityChainID().IsSameAs(m.ServerChainID) {
			found = true
			break
		}
	}
	if !found {
		return -1
	}

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

func (m *RemoveServerMsg) GetSignatures() []interfaces.IFullSignature {
	return m.Signatures.GetSignatures()
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *RemoveServerMsg) ComputeVMIndex(state interfaces.IState) {
	m.VMIndex = state.ComputeVMIndex(constants.ADMIN_CHAINID)
}

// Execute the leader functions of the given message
func (m *RemoveServerMsg) LeaderExecute(state interfaces.IState) {
	state.LeaderExecute(m)
}

func (m *RemoveServerMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

// Acknowledgements do not go into the process list.
func (e *RemoveServerMsg) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessRemoveServer(dbheight, e)
}

func (e *RemoveServerMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RemoveServerMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *RemoveServerMsg) AddSignature(key interfaces.Signer) error {
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

func (m *RemoveServerMsg) VerifySignatures() ([]interfaces.IFullSignature, error) {
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

func (m *RemoveServerMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Add Server Message: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling Add Server Message: %v", r)
		}
		return
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
	newData, err = m.Signatures.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	return
}

func (m *RemoveServerMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *RemoveServerMsg) MarshalForSignature() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "RemoveServerMsg.MarshalForSignature err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	data, err = m.ServerChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, uint8(m.ServerType))

	return buf.DeepCopyBytes(), nil
}

func (m *RemoveServerMsg) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "RemoveServerMsg.MarshalBinary err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	data, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	if m.Signatures != nil {
		data, err = m.Signatures.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *RemoveServerMsg) String() string {
	var stype string
	if m.ServerType == 0 {
		stype = "Federated"
	} else {
		stype = "Audit"
	}
	return fmt.Sprintf("RemoveServer (%s): ChainID: %x Time: %x Msg Hash %x ",
		stype,
		m.ServerChainID.Bytes()[:3],
		&m.Timestamp,
		m.GetMsgHash().Bytes()[:3])

}

func (m *RemoveServerMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "removeserver",
		"server":     m.ServerChainID.String()[4:10],
		"servertype": m.ServerType,
		"hash":       m.GetMsgHash().String()[:6]}
}

func (m *RemoveServerMsg) IsSameAs(b *RemoveServerMsg) bool {
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

func NewRemoveServerMsg(state interfaces.IState, chainId interfaces.IHash, serverType int) interfaces.IMsg {
	msg := new(RemoveServerMsg)
	msg.ServerChainID = chainId
	msg.ServerType = serverType
	msg.Timestamp = state.GetTimestamp()
	msg.Signatures = factoid.NewFullSignatureBlock()

	return msg

}
