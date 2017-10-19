// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/FactomProject/logrus"
)

//A placeholder structure for messages
type CommitChainMsg struct {
	MessageBase
	CommitChain *entryCreditBlock.CommitChain

	Signature interfaces.IFullSignature

	// Not marshaled... Just used by the leader
	count    int
	validsig bool
}

var _ interfaces.IMsg = (*CommitChainMsg)(nil)
var _ Signable = (*CommitChainMsg)(nil)

func (a *CommitChainMsg) IsSameAs(b *CommitChainMsg) bool {
	if a == nil || b == nil {
		if a == nil && b == nil {
			return true
		}
		return false
	}

	if a.CommitChain == nil && b.CommitChain != nil {
		return false
	}
	if a.CommitChain != nil {
		if a.CommitChain.IsSameAs(b.CommitChain) == false {
			return false
		}
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}

	return true
}

func (m *CommitChainMsg) GetCount() int {
	return m.count
}

func (m *CommitChainMsg) IncCount() {
	m.count += 1
}

func (m *CommitChainMsg) SetCount(cnt int) {
	m.count = cnt
}

func (m *CommitChainMsg) Process(dbheight uint32, state interfaces.IState) bool {
	return state.ProcessCommitChain(dbheight, m)
}

func (m *CommitChainMsg) GetRepeatHash() interfaces.IHash {
	return m.CommitChain.GetSigHash()
}

func (m *CommitChainMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *CommitChainMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		m.MsgHash = m.CommitChain.GetSigHash()
	}
	return m.MsgHash
}

func (m *CommitChainMsg) GetTimestamp() interfaces.Timestamp {
	return m.CommitChain.GetTimestamp()
}

func (m *CommitChainMsg) Type() byte {
	return constants.COMMIT_CHAIN_MSG
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *CommitChainMsg) Validate(state interfaces.IState) int {
	if !m.validsig && !m.CommitChain.IsValid() {
		return -1
	}
	m.validsig = true

	ebal := state.GetFactoidState().GetECBalance(*m.CommitChain.ECPubKey)
	v := int(ebal) - int(m.CommitChain.Credits)
	if v < 0 {
		return 0
	}

	return 1
}

func (m *CommitChainMsg) ComputeVMIndex(state interfaces.IState) {
	m.VMIndex = state.ComputeVMIndex(constants.EC_CHAINID)
}

// Execute the leader functions of the given message
func (m *CommitChainMsg) LeaderExecute(state interfaces.IState) {
	// Check if we have yet to see an entry.  If we have seen one (NoEntryYet == false) then
	// we can record it.
	if state.NoEntryYet(m.CommitChain.EntryHash, m.CommitChain.GetTimestamp()) {
		state.LeaderExecuteCommitChain(m)
	} else {
		state.FollowerExecuteCommitChain(m)
	}
}

func (m *CommitChainMsg) FollowerExecute(state interfaces.IState) {
	state.FollowerExecuteMsg(m)
}

func (e *CommitChainMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *CommitChainMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *CommitChainMsg) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *CommitChainMsg) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *CommitChainMsg) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *CommitChainMsg) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	t, err := buf.PopByte()
	if t != m.Type() {
		return nil, fmt.Errorf("Invalid Message type")
	}

	m.CommitChain = entryCreditBlock.NewCommitChain()
	err = buf.PopBinaryMarshallable(m.CommitChain)
	if err != nil {
		return nil, err
	}

	if buf.Len() > 0 {
		m.Signature = new(primitives.Signature)
		err = buf.PopBinaryMarshallable(m.Signature)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *CommitChainMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *CommitChainMsg) MarshalForSignature() ([]byte, error) {
	buf := primitives.NewBuffer(nil)

	err := buf.PushByte(m.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(m.CommitChain)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (m *CommitChainMsg) MarshalBinary() (data []byte, err error) {
	h, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(h)

	sig := m.GetSignature()
	if sig != nil {
		err = buf.PushBinaryMarshallable(sig)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (m *CommitChainMsg) String() string {
	if m.LeaderChainID == nil {
		m.LeaderChainID = primitives.NewZeroHash()
	}
	str := fmt.Sprintf("%6s-VM%3d: entryhash[%x] hash[%x]",
		"CChain",
		m.VMIndex,
		m.CommitChain.EntryHash.Bytes()[:3],
		m.GetHash().Bytes()[:3])
	return str
}

func (m *CommitChainMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "commitchain", "vmindex": m.VMIndex,
		"server":      m.LeaderChainID.String()[4:12],
		"commitchain": m.CommitChain.EntryHash.String()[:6],
		"hash":        m.GetHash().String()[:6]}
}
