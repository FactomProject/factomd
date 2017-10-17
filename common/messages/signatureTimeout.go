// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	log "github.com/FactomProject/logrus"
)

//A placeholder structure for messages
type SignatureTimeout struct {
	MessageBase
	Timestamp interfaces.Timestamp

	Signature interfaces.IFullSignature

	//Not marshalled
	hash interfaces.IHash
}

var _ interfaces.IMsg = (*SignatureTimeout)(nil)
var _ Signable = (*SignatureTimeout)(nil)

func (a *SignatureTimeout) IsSameAs(b *SignatureTimeout) bool {
	if b == nil {
		return false
	}
	if a.Timestamp.GetTimeMilli() != b.Timestamp.GetTimeMilli() {
		return false
	}

	if a.Signature == nil && b.Signature != nil {
		return false
	}
	if a.Signature != nil {
		if a.Signature.IsSameAs(b.Signature) == false {
			return false
		}
	}
	//TODO: expand

	return true
}

func (m *SignatureTimeout) Process(uint32, interfaces.IState) bool { return true }

func (m *SignatureTimeout) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *SignatureTimeout) GetHash() interfaces.IHash {
	if m.hash == nil {
		data, err := m.MarshalForSignature()
		if err != nil {
			panic(fmt.Sprintf("Error in CommitChain.GetHash(): %s", err.Error()))
		}
		m.hash = primitives.Sha(data)
	}
	return m.hash
}

func (m *SignatureTimeout) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *SignatureTimeout) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp
}

func (m *SignatureTimeout) Type() byte {
	return constants.SIGNATURE_TIMEOUT_MSG
}

func (m *SignatureTimeout) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
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

	if len(newData) > 0 {
		m.Signature = new(primitives.Signature)
		newData, err = m.Signature.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
	}

	return newData, nil
}

func (m *SignatureTimeout) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *SignatureTimeout) MarshalForSignature() (data []byte, err error) {
	var buf primitives.Buffer
	buf.Write([]byte{m.Type()})
	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	//TODO: expand

	return buf.DeepCopyBytes(), nil
}
func (m *SignatureTimeout) MarshalBinary() (data []byte, err error) {
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

func (m *SignatureTimeout) GetSignature() interfaces.IFullSignature {
	return m.Signature
}

func (m *SignatureTimeout) VerifySignature() (bool, error) {
	return VerifyMessage(m)
}

func (m *SignatureTimeout) Sign(key interfaces.Signer) error {
	signature, err := SignSignable(m, key)
	if err != nil {
		return err
	}
	m.Signature = signature
	return nil
}

func (m *SignatureTimeout) String() string {
	return "Signature Timeout"
}

func (m *SignatureTimeout) LogFields() log.Fields {
	return log.Fields{}
}

func (m *SignatureTimeout) DBHeight() int {
	return 0
}

func (m *SignatureTimeout) ChainID() []byte {
	return nil
}

func (m *SignatureTimeout) ListHeight() int {
	return 0
}

func (m *SignatureTimeout) SerialHash() []byte {
	return nil
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *SignatureTimeout) Validate(state interfaces.IState) int {
	return 0
}

func (m *SignatureTimeout) ComputeVMIndex(state interfaces.IState) {
}

func (m *SignatureTimeout) LeaderExecute(state interfaces.IState) {
}

func (m *SignatureTimeout) FollowerExecute(interfaces.IState) {
}

func (e *SignatureTimeout) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *SignatureTimeout) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}
