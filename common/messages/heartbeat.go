// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"bytes"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	log "github.com/sirupsen/logrus"
)

//A placeholder structure for messages
type Heartbeat struct {
	msgbase.MessageBase
	Timestamp       interfaces.Timestamp
	SecretNumber    uint32
	DBHeight        uint32
	DBlockHash      interfaces.IHash //Hash of last Directory Block
	IdentityChainID interfaces.IHash //Identity Chain ID

	Signature interfaces.IFullSignature

	//Not marshalled
	hash         interfaces.IHash
	sigvalid     bool
	marshalCache []byte
}

var _ interfaces.IMsg = (*Heartbeat)(nil)
var _ interfaces.Signable = (*Heartbeat)(nil)

func (a *Heartbeat) IsSameAs(b *Heartbeat) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if a.DBlockHash == nil && b.DBlockHash != nil {
		return false
	}
	if a.DBlockHash != nil {
		if a.DBlockHash.IsSameAs(b.DBlockHash) == false {
			return false
		}
	}

	if a.IdentityChainID == nil && b.IdentityChainID != nil {
		return false
	}
	if a.IdentityChainID != nil {
		if a.IdentityChainID.IsSameAs(b.IdentityChainID) == false {
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

func (m *Heartbeat) Process(uint32, interfaces.IState) bool {
	return true
}

func (m *Heartbeat) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *Heartbeat) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *Heartbeat) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *Heartbeat) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *Heartbeat) Type() byte {
	return constants.HEARTBEAT_MSG
}

func (m *Heartbeat) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling HeartBeat: %v", r)
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

	m.SecretNumber, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	hash := new(primitives.Hash)

	newData, err = hash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.DBlockHash = hash

	hash = new(primitives.Hash)
	newData, err = hash.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}
	m.IdentityChainID = hash

	if len(newData) > 0 {
		sig := new(primitives.Signature)
		newData, err = sig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Signature = sig
	}

	m.marshalCache = append(m.marshalCache, data[:len(data)-len(newData)]...)

	return nil, nil
}

func (m *Heartbeat) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *Heartbeat) MarshalForSignature() (data []byte, err error) {
	if m.DBlockHash == nil || m.IdentityChainID == nil {
		return nil, fmt.Errorf("Message is incomplete")
	}

	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	binary.Write(&buf, binary.BigEndian, m.SecretNumber)
	binary.Write(&buf, binary.BigEndian, m.DBHeight)

	if d, err := m.DBlockHash.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	if d, err := m.IdentityChainID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *Heartbeat) MarshalBinary() (data []byte, err error) {

	if m.marshalCache != nil {
		return m.marshalCache, nil
	}

	resp, err := m.MarshalForSignature()
	if err != nil {
		return nil, err
	}
	sig := m.GetSignature()
	if sig != nil {
		sigBytes, err := sig.MarshalBinary()
		if err != nil {
			return nil, err
		}
		return append(resp, sigBytes...), nil
	}
	return resp, nil
}

func (m *Heartbeat) String() string {
	return fmt.Sprintf("HeartBeat ID[%x] dbht %d ts %d", m.IdentityChainID.Bytes()[3:6], m.DBHeight, m.Timestamp.GetTimeSeconds())
}

func (m *Heartbeat) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "heartbeat",
		"vm":        m.VMIndex,
		"dbheight":  m.DBHeight,
		"server":    m.IdentityChainID.String(),
		"timestamp": m.Timestamp.GetTimeSeconds()}
}

func (m *Heartbeat) ChainID() []byte {
	return nil
}

func (m *Heartbeat) ListHeight() int {
	return 0
}

func (m *Heartbeat) SerialHash() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *Heartbeat) Validate(is interfaces.IState) int {
	now := is.GetTimestamp()

	if now.GetTimeSeconds()-m.Timestamp.GetTimeSeconds() > 60 {
		return -1
	}

	if m.GetSignature() == nil {
		// the message has no signature (and so is invalid)
		return -1
	}

	// Ignore old heartbeats
	if m.DBHeight <= is.GetHighestSavedBlk() {
		return -1
	}

	if !m.sigvalid {
		auth := is.GetAuthorityInterface(m.IdentityChainID)
		if auth == nil {
			return -1
		}

		if bytes.Compare(m.Signature.GetKey(), auth.GetSigningKey()) != 0 {
			return -1
		}

		isVer, err := m.VerifySignature()
		if err != nil || !isVer {
			// if there is an error during signature verification
			// or if the signature is invalid
			// the message is considered invalid
			return -1
		}
		m.sigvalid = true
	}

	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *Heartbeat) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
func (m *Heartbeat) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *Heartbeat) FollowerExecute(is interfaces.IState) {
	for _, auditServer := range is.GetAuditServers(is.GetLeaderHeight()) {
		if auditServer.GetChainID().IsSameAs(m.IdentityChainID) {
			if m.IdentityChainID.IsSameAs(is.GetIdentityChainID()) {
				if m.SecretNumber != is.GetSalt(m.Timestamp) {
					panic("We have seen a heartbeat using our Identity that isn't ours")
				}
			}
			auditServer.SetOnline(true)
		}
	}
}

func (e *Heartbeat) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Heartbeat) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *Heartbeat) Sign(key interfaces.Signer) error {
	signature, err := msgbase.SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *Heartbeat) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *Heartbeat) VerifySignature() (bool, error) {
	return msgbase.VerifyMessage(m)
}
