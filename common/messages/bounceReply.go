// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package messages

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/messages/msgbase"
	llog "github.com/FactomProject/factomd/log"
	log "github.com/sirupsen/logrus"
)

type BounceReply struct {
	msgbase.MessageBase
	Name      string
	Number    int32
	Timestamp interfaces.Timestamp
	Stamps    []interfaces.Timestamp
	size      int
}

var _ interfaces.IMsg = (*BounceReply)(nil)

func (m *BounceReply) GetRepeatHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("BounceReply.GetRepeatHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

// We have to return the hash of the underlying message.
func (m *BounceReply) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("BounceReply.GetHash() saw an interface that was nil")
		}
	}()

	return m.GetMsgHash()
}

func (m *BounceReply) SizeOf() int {
	m.GetMsgHash()
	return m.size
}

func (m *BounceReply) GetMsgHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("BounceReply.GetMsgHash() saw an interface that was nil")
		}
	}()

	data, err := m.MarshalForSignature()

	m.size = len(data)

	if err != nil {
		return nil
	}
	m.MsgHash = primitives.Sha(data)
	return m.MsgHash
}

func (m *BounceReply) Type() byte {
	return constants.BOUNCEREPLY_MSG
}

func (m *BounceReply) GetTimestamp() interfaces.Timestamp {
	return m.Timestamp.Clone()
}

func (m *BounceReply) VerifySignature() (bool, error) {
	return true, nil
}

func (m *BounceReply) WellFormed() bool {
	// TODO: Flush this out
	return true
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- Message is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- Message is valid
func (m *BounceReply) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *BounceReply) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *BounceReply) LeaderExecute(state interfaces.IState) {
}

func (m *BounceReply) FollowerExecute(state interfaces.IState) {
}

// Acknowledgements do not go into the process list.
func (e *BounceReply) Process(dbheight uint32, state interfaces.IState) bool {
	return true
}

func (e *BounceReply) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *BounceReply) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *BounceReply) Sign(key interfaces.Signer) error {
	return nil
}

func (m *BounceReply) GetSignature() interfaces.IFullSignature {
	return nil
}

func (m *BounceReply) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()

	m.SetPeer2Peer(true)

	newData = data

	if newData[0] != m.Type() {
		return nil, errors.New("Invalid Message type")
	}
	newData = newData[1:]

	m.Name = string(newData[:32])
	newData = newData[32:]

	m.Number, newData = int32(binary.BigEndian.Uint32(newData[0:4])), newData[4:]

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	numTS, newData := binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := uint32(0); i < numTS; i++ {
		ts := new(primitives.Timestamp)
		newData, err = ts.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		m.Stamps = append(m.Stamps, ts)
	}
	return
}

func (m *BounceReply) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *BounceReply) MarshalForSignature() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "BounceReply.MarshalForSignature err:%v", *pe)
		}
	}(&err)
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, m.Type())

	var buff [32]byte

	copy(buff[:32], []byte(fmt.Sprintf("%32s", m.Name)))
	buf.Write(buff[:])

	binary.Write(&buf, binary.BigEndian, m.Number)

	t := m.GetTimestamp()
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, int32(len(m.Stamps)))

	for _, ts := range m.Stamps {
		data, err := ts.MarshalBinary()
		if err != nil {
			return nil, err
		}
		buf.Write(data)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *BounceReply) MarshalBinary() (data []byte, err error) {
	return m.MarshalForSignature()
}

func (m *BounceReply) String() string {
	// bbbb Origin: 2016-09-05 12:26:20.426954586 -0500 CDT left BounceReply Start:             2016-09-05 12:26:05 Hops:     1 Size:    43 Last Hop Took 14.955 Average Hop: 14.955
	now := time.Now()
	t := fmt.Sprintf("%2d:%2d:%2d.%03d", now.Hour(), now.Minute(), now.Second(), now.Nanosecond()/1000000)
	mill := m.Timestamp.GetTimeMilli()
	mills := mill % 1000
	mill = mill / 1000
	secs := mill % 60
	mill = mill / 60
	mins := mill % 60
	mill = mill / 60
	hrs := mill % 24
	t2 := fmt.Sprintf("%2d:%2d:%2d.%03d", hrs, mins, secs, mills)
	b := m.SizeOf() % 1000
	kb := (m.SizeOf() / 1000) % 1000
	mb := (m.SizeOf() / 1000 / 1000)
	sz := fmt.Sprintf("%d,%03d", kb, b)
	if mb > 0 {
		sz = fmt.Sprintf("%d,%03d,%03d", mb, kb, b)
	}

	str := fmt.Sprintf("Origin: %12s  %10s-%03d-%03d BounceReply Start: %12s Hops: %5d Size: [%12s] ",
		t,
		strings.TrimSpace(m.Name),
		m.Number,
		len(m.Stamps),
		t2,
		len(m.Stamps),
		sz)

	elapse := primitives.NewTimestampNow().GetTimeMilli() - m.Stamps[len(m.Stamps)-1].GetTimeMilli()

	str = str + fmt.Sprintf("Last Hop Took %d.%03d", elapse/1000, elapse%1000)
	return str
}

func (m *BounceReply) LogFields() log.Fields {
	return log.Fields{}
}

func (a *BounceReply) IsSameAs(b *BounceReply) bool {
	return true
}

func (a *BounceReply) Label() string {
	return msgbase.GetLabel(a)
}
