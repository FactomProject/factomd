// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package messages

import (
	"bytes"
	"fmt"
	"io"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// MaxBlockLocatorsPerMsg is the maximum number of Directory block locator hashes allowed
// per message.
const MaxBlockLocatorsPerMsg = 500

// MsgGetDirBlocks implements the MsgGetDirBlocks interface and represents a factom
// getdirblocks message.  It is used to request a list of blocks starting after the
// last known hash in the slice of block locator hashes.  The list is returned
// via an inv message (MsgInv) and is limited by a specific hash to stop at or
// the maximum number of blocks per message, which is currently 500.
//
// Set the HashStop field to the hash at which to stop and use
// AddBlockLocatorHash to build up the list of block locator hashes.
//
// The algorithm for building the block locator hashes should be to add the
// hashes in reverse order until you reach the genesis block.  In order to keep
// the list of locator hashes to a reasonable number of entries, first add the
// most recent 10 block hashes, then double the step each loop iteration to
// exponentially decrease the number of hashes the further away from head and
// closer to the genesis block you get.
type MsgGetDirBlocks struct {
	ProtocolVersion    uint32
	BlockLocatorHashes []interfaces.IHash
	HashStop           interfaces.IHash
}

// AddBlockLocatorHash adds a new block locator hash to the message.
func (msg *MsgGetDirBlocks) AddBlockLocatorHash(hash interfaces.IHash) error {
	//util.Trace()
	if len(msg.BlockLocatorHashes)+1 > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message [max %v]",
			MaxBlockLocatorsPerMsg)
		return messageError("MsgGetDirBlocks.AddBlockLocatorHash", str)
	}

	msg.BlockLocatorHashes = append(msg.BlockLocatorHashes, hash)
	return nil
}

// BtcDecode decodes r using the bitcoin protocol encoding into the receiver.
// This is part of the MsgGetDirBlocks interface implementation.
func (msg *MsgGetDirBlocks) BtcDecode(r io.Reader, pver uint32) error {
	//util.Trace()
	err := readElement(r, &msg.ProtocolVersion)
	if err != nil {
		return err
	}

	// Read num block locator hashes and limit to max.
	count, err := readVarInt(r, pver)
	if err != nil {
		return err
	}
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %v, max %v]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgGetDirBlocks.BtcDecode", str)
	}

	msg.BlockLocatorHashes = make([]interfaces.IHash, 0, count)
	for i := uint64(0); i < count; i++ {
		sha := new(primitives.Hash)
		err := readElement(r, sha)
		if err != nil {
			return err
		}
		msg.AddBlockLocatorHash(sha)
	}

	err = readElement(r, msg.HashStop)
	if err != nil {
		return err
	}

	return nil
}

// BtcEncode encodes the receiver to w using the bitcoin protocol encoding.
// This is part of the MsgGetDirBlocks interface implementation.
func (msg *MsgGetDirBlocks) BtcEncode(w io.Writer, pver uint32) error {
	//util.Trace()
	count := len(msg.BlockLocatorHashes)
	if count > MaxBlockLocatorsPerMsg {
		str := fmt.Sprintf("too many block locator hashes for message "+
			"[count %v, max %v]", count, MaxBlockLocatorsPerMsg)
		return messageError("MsgGetDirBlocks.BtcEncode", str)
	}

	err := writeElement(w, msg.ProtocolVersion)
	if err != nil {
		return err
	}

	err = writeVarInt(w, pver, uint64(count))
	if err != nil {
		return err
	}

	for _, hash := range msg.BlockLocatorHashes {
		err = writeElement(w, hash)
		if err != nil {
			return err
		}
	}

	err = writeElement(w, msg.HashStop)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the protocol command string for the message.  This is part
// of the MsgGetDirBlocks interface implementation.
func (msg *MsgGetDirBlocks) Command() string {
	//util.Trace()
	return CmdGetDirBlocks
}

// MaxPayloadLength returns the maximum length the payload can be for the
// receiver.  This is part of the MsgGetDirBlocks interface implementation.
func (msg *MsgGetDirBlocks) MaxPayloadLength(pver uint32) uint32 {
	// Protocol version 4 bytes + num hashes (varInt) + max block locator
	// hashes + hash stop.
	//util.Trace()
	return uint32(4 + MaxVarIntPayload + (MaxBlockLocatorsPerMsg * constants.HASH_LENGTH) + constants.HASH_LENGTH)
}

// NewMsgGetDirBlocks returns a new bitcoin getdirblocks message that conforms to the
// MsgGetDirBlocks interface using the passed parameters and defaults for the remaining
// fields.
func NewMsgGetDirBlocks(hashStop interfaces.IHash) *MsgGetDirBlocks {
	//util.Trace()
	return &MsgGetDirBlocks{
		ProtocolVersion:    ProtocolVersion,
		BlockLocatorHashes: make([]interfaces.IHash, 0, MaxBlockLocatorsPerMsg),
		HashStop:           hashStop,
	}
}

var _ interfaces.IMsg = (*MsgGetDirBlocks)(nil)

func (m *MsgGetDirBlocks) Process(uint32, interfaces.IState) {}

func (m *MsgGetDirBlocks) GetHash() interfaces.IHash {
	return nil
}

func (m *MsgGetDirBlocks) GetTimestamp() interfaces.Timestamp {
	return 0
}

func (m *MsgGetDirBlocks) Type() int {
	return -1
}

func (m *MsgGetDirBlocks) Int() int {
	return -1
}

func (m *MsgGetDirBlocks) Bytes() []byte {
	return nil
}

func (m *MsgGetDirBlocks) UnmarshalBinaryData(data []byte) (newdata []byte, err error) {
	return nil, nil
}

func (m *MsgGetDirBlocks) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *MsgGetDirBlocks) MarshalBinary() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetDirBlocks) MarshalForSignature() (data []byte, err error) {
	return nil, nil
}

func (m *MsgGetDirBlocks) String() string {
	return ""
}

// Validate the message, given the state.  Three possible results:
//  < 0 -- MsgGetDirBlocks is invalid.  Discard
//  0   -- Cannot tell if message is Valid
//  1   -- MsgGetDirBlocks is valid
func (m *MsgGetDirBlocks) Validate(dbheight uint32, state interfaces.IState) int {
	return 0
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *MsgGetDirBlocks) Leader(state interfaces.IState) bool {
	switch state.GetNetworkNumber() {
	case 0: // Main Network
		panic("Not implemented yet")
	case 1: // Test Network
		panic("Not implemented yet")
	case 2: // Local Network
		panic("Not implemented yet")
	default:
		panic("Not implemented yet")
	}

}

// Execute the leader functions of the given message
func (m *MsgGetDirBlocks) LeaderExecute(state interfaces.IState) error {
	return nil
}

// Returns true if this is a message for this server to execute as a follower
func (m *MsgGetDirBlocks) Follower(interfaces.IState) bool {
	return true
}

func (m *MsgGetDirBlocks) FollowerExecute(interfaces.IState) error {
	return nil
}

func (e *MsgGetDirBlocks) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MsgGetDirBlocks) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MsgGetDirBlocks) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}
