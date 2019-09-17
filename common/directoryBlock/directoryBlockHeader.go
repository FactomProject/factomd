// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package directoryBlock

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

// DBlockHeader contains information related to a specific directory block.
type DBlockHeader struct {
	Version      byte             `json:"version"`      // The nework version, only supported version seems to be constants.VERSION_0
	NetworkID    uint32           `json:"networkid"`    // Three supported networks in constants: MAIN_NETWORK_ID, TEST_NETWORK_ID, LOCAL_NETWORK_ID
	BodyMR       interfaces.IHash `json:"bodymr"`       // The Merkle root of the 'body' (entries only) of this directory block
	PrevKeyMR    interfaces.IHash `json:"prevkeymr"`    // The key Merkle root of the previous directory block
	PrevFullHash interfaces.IHash `json:"prevfullhash"` // The FullHash of the previous directory block
	Timestamp    uint32           `json:"timestamp"`    // In theory, the timestamp the directory block was created, in minutes
	DBHeight     uint32           `json:"dbheight"`     // The directory block height this header information is relevant to
	BlockCount   uint32           `json:"blockcount"`   // The number of entry blocks in this directory block
}

var _ interfaces.Printable = (*DBlockHeader)(nil)
var _ interfaces.BinaryMarshallable = (*DBlockHeader)(nil)
var _ interfaces.IDirectoryBlockHeader = (*DBlockHeader)(nil)

// Init initializes any nil interfaces to the Zero hash in the DBlockHeader
func (h *DBlockHeader) Init() {
	if h.BodyMR == nil {
		h.BodyMR = primitives.NewZeroHash()
	}
	if h.PrevKeyMR == nil {
		h.PrevKeyMR = primitives.NewZeroHash()
	}
	if h.PrevFullHash == nil {
		h.PrevFullHash = primitives.NewZeroHash()
	}
}

// GetHeaderHash marshals the header and returns the hash of the marshaled data
func (h *DBlockHeader) GetHeaderHash() (interfaces.IHash, error) {

	binaryEBHeader, err := h.MarshalBinary()
	if err != nil {
		return nil, err
	}

	h2 := primitives.Sha(binaryEBHeader)

	return h2, nil
}

// IsSameAs returns true iff the input header is identical to this DBlockHeader
func (h *DBlockHeader) IsSameAs(b interfaces.IDirectoryBlockHeader) bool {
	if h == nil || b == nil {
		if h == nil && b == nil {
			return true
		}
		return false
	}

	bb, ok := b.(*DBlockHeader)
	if ok == false {
		return false
	}

	if h.Version != bb.Version {
		return false
	}
	if h.NetworkID != bb.NetworkID {
		return false
	}

	if h.BodyMR.IsSameAs(bb.BodyMR) == false {
		return false
	}
	if h.PrevKeyMR.IsSameAs(bb.PrevKeyMR) == false {
		return false
	}
	if h.PrevFullHash.IsSameAs(bb.PrevFullHash) == false {
		return false
	}

	if h.Timestamp != bb.Timestamp {
		return false
	}
	if h.DBHeight != bb.DBHeight {
		return false
	}
	if h.BlockCount != bb.BlockCount {
		return false
	}

	return true
}

// GetVersion returns the network version
func (h *DBlockHeader) GetVersion() byte {
	return h.Version
}

// SetVersion sets the network version to the input
func (h *DBlockHeader) SetVersion(version byte) {
	h.Version = version
}

// GetNetworkID returns the network id
func (h *DBlockHeader) GetNetworkID() uint32 {
	return h.NetworkID
}

// SetNetworkID sets the nework id to the input
func (h *DBlockHeader) SetNetworkID(networkID uint32) {
	h.NetworkID = networkID
}

// GetBodyMR returns the stored bodyMR of the directory block this header is associated with
func (h *DBlockHeader) GetBodyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBlockHeader.GetBodyMR() saw an interface that was nil")
		}
	}()

	return h.BodyMR
}

// SetBodyMR sets the bodyMR to the input
func (h *DBlockHeader) SetBodyMR(bodyMR interfaces.IHash) {
	h.BodyMR = bodyMR
}

// GetPrevKeyMR returns the previous key Merkle root
func (h *DBlockHeader) GetPrevKeyMR() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBlockHeader.GetPrevKeyMR() saw an interface that was nil")
		}
	}()

	return h.PrevKeyMR
}

// SetPrevKeyMR sets the previous key Merkle root to the input
func (h *DBlockHeader) SetPrevKeyMR(prevKeyMR interfaces.IHash) {
	h.PrevKeyMR = prevKeyMR
}

// GetPrevFullHash returns the stored FullHash of the previous directory block to this directory block
func (h *DBlockHeader) GetPrevFullHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBlockHeader.GetPrevFullHash() saw an interface that was nil")
		}
	}()

	return h.PrevFullHash
}

// SetPrevFullHash sets the previous full hash value to the input
func (h *DBlockHeader) SetPrevFullHash(PrevFullHash interfaces.IHash) {
	h.PrevFullHash = PrevFullHash
}

// GetTimestamp returns a timestamp for when this directory block was created in minutes
func (h *DBlockHeader) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampFromMinutes(h.Timestamp)
}

// SetTimestamp sets the timestamp in minutes from the input
func (h *DBlockHeader) SetTimestamp(timestamp interfaces.Timestamp) {
	h.Timestamp = timestamp.GetTimeMinutesUInt32()
}

// GetDBHeight returns the height this directory block is located in the blockchain
func (h *DBlockHeader) GetDBHeight() uint32 {
	return h.DBHeight
}

// SetDBHeight sets the height this directory block is locsated in the blockchain
func (h *DBlockHeader) SetDBHeight(dbheight uint32) {
	h.DBHeight = dbheight
}

// GetBlockCount returns the number of entry blocks within this directory block
func (h *DBlockHeader) GetBlockCount() uint32 {
	return h.BlockCount
}

// SetBlockCount sets the block count to the input
func (h *DBlockHeader) SetBlockCount(blockcount uint32) {
	h.BlockCount = blockcount
}

// JSONByte returns the directory block header encoded into json format
func (h *DBlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(h)
}

// JSONString returns the directory block header encoded into json format
func (h *DBlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(h)
}

// String returns the directory block header into a string format
func (h *DBlockHeader) String() string {
	h.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("  version:         %v\n", h.Version))
	out.WriteString(fmt.Sprintf("  networkid:       %x\n", h.NetworkID))
	out.WriteString(fmt.Sprintf("  bodymr:          %s\n", h.BodyMR.String()[:6]))
	out.WriteString(fmt.Sprintf("  prevkeymr:       %s\n", h.PrevKeyMR.String()[:6]))
	out.WriteString(fmt.Sprintf("  prevfullhash:    %s\n", h.PrevFullHash.String()[:6]))
	out.WriteString(fmt.Sprintf("  timestamp:       %d\n", h.Timestamp))
	out.WriteString(fmt.Sprintf("  timestamp str:   %s\n", h.GetTimestamp().String()))
	out.WriteString(fmt.Sprintf("  dbheight:        %d\n", h.DBHeight))
	out.WriteString(fmt.Sprintf("  blockcount:      %d\n", h.BlockCount))

	return (string)(out.DeepCopyBytes())
}

// MarshalBinary marshals the directory block
func (h *DBlockHeader) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DBlockHeader.MarshalBinary err:%v", *pe)
		}
	}(&err)
	h.Init()
	buf := primitives.NewBuffer(nil)

	err = buf.PushByte(h.Version)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(h.NetworkID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(h.BodyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(h.PrevKeyMR)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(h.PrevFullHash)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(h.Timestamp)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(h.DBHeight)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(h.BlockCount)
	if err != nil {
		return nil, err
	}

	if h.BlockCount > constants.MaxDirectoryBlockEntryCount {
		panic("Send: Blockcount too great in directory block")
	}

	return buf.DeepCopyBytes(), err
}

// UnmarshalBinaryData unmarshals the input data into the directory block header
func (h *DBlockHeader) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	var err error

	h.Version, err = buf.PopByte()
	if err != nil {
		return nil, err
	}
	h.NetworkID, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	h.BodyMR = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(h.BodyMR)
	if err != nil {
		return nil, err
	}
	h.PrevKeyMR = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(h.PrevKeyMR)
	if err != nil {
		return nil, err
	}
	h.PrevFullHash = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(h.PrevFullHash)
	if err != nil {
		return nil, err
	}

	h.Timestamp, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	h.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	h.BlockCount, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	if h.BlockCount > constants.MaxDirectoryBlockEntryCount {
		panic("Receive: Blockcount too great in directory block" + fmt.Sprintf(":::: %d", h.BlockCount))
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into the directory block header
func (h *DBlockHeader) UnmarshalBinary(data []byte) (err error) {
	_, err = h.UnmarshalBinaryData(data)
	return
}

type ExpandedDBlockHeader DBlockHeader

func (h DBlockHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ExpandedDBlockHeader
		ChainID string `json:"chainid"`
	}{
		ExpandedDBlockHeader: ExpandedDBlockHeader(h),
		ChainID:              "000000000000000000000000000000000000000000000000000000000000000d",
	})
}

/************************************************
 * Support Functions
 ************************************************/

// NewDBlockHeader returns a new directory block header with zero hashes
func NewDBlockHeader() *DBlockHeader {
	d := new(DBlockHeader)

	d.BodyMR = primitives.NewZeroHash()
	d.PrevKeyMR = primitives.NewZeroHash()
	d.PrevFullHash = primitives.NewZeroHash()

	return d
}
