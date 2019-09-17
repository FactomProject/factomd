package dbInfo

import (
	"encoding/gob"
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DirBlockInfo holds the information for a factom directory block
type DirBlockInfo struct {
	// Serial hash for the directory block
	DBHash    interfaces.IHash `json:"dbhash"`
	DBHeight  uint32           `json:"dbheight"`  //directory block height
	Timestamp int64            `json:"timestamp"` // time of this dir block info being created in milliseconds
	// BTCTxHash is the Tx hash returned from rpcclient.SendRawTransaction
	BTCTxHash interfaces.IHash `json:"btctxhash"` // use string or *btcwire.ShaHash ???
	// BTCTxOffset is the index of the TX in this BTC block
	BTCTxOffset int32 `json:"btctxoffset"`
	// BTCBlockHeight is the height of the block where this TX is stored in BTC
	BTCBlockHeight int32 `json:"btcblockheight"`
	//BTCBlockHash is the hash of the block where this TX is stored in BTC
	BTCBlockHash interfaces.IHash `json:"btcblockhash"` // use string or *btcwire.ShaHash ???
	// DBMerkleRoot is the merkle root of the Directory Block
	// and is written into BTC as OP_RETURN data
	DBMerkleRoot interfaces.IHash `json:"dbmerkleroot"`
	// A flag to to show BTC anchor confirmation
	BTCConfirmed bool `json:"btcconfirmed"`

	EthereumAnchorRecordEntryHash interfaces.IHash `json:"ethereumanchorrecordentryhash"`
	EthereumConfirmed             bool             `json:"ethereumconfirmed"`
}

var _ interfaces.Printable = (*DirBlockInfo)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*DirBlockInfo)(nil)
var _ interfaces.DatabaseBatchable = (*DirBlockInfo)(nil)
var _ interfaces.IDirBlockInfo = (*DirBlockInfo)(nil)

// Init initializes any nil hash in the DirBlockInfo to the zero hash
func (d *DirBlockInfo) Init() {
	if d.DBHash == nil {
		d.DBHash = primitives.NewZeroHash()
	}
	if d.BTCTxHash == nil {
		d.BTCTxHash = primitives.NewZeroHash()
	}
	if d.BTCBlockHash == nil {
		d.BTCBlockHash = primitives.NewZeroHash()
	}
	if d.DBMerkleRoot == nil {
		d.DBMerkleRoot = primitives.NewZeroHash()
	}
	if d.EthereumAnchorRecordEntryHash == nil {
		d.EthereumAnchorRecordEntryHash = primitives.NewZeroHash()
	}
}

// NewDirBlockInfo creates a newly initialized DirBlockInfo
func NewDirBlockInfo() *DirBlockInfo {
	dbi := new(DirBlockInfo)
	dbi.Init()
	return dbi
}

// String  encodes the DirBlockInfo into a JSON string
func (d *DirBlockInfo) String() string {
	str, _ := d.JSONString()
	return str
}

// JSONByte encodes the DirBlockInfo into a JSON byte array
func (d *DirBlockInfo) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(d)
}

// JSONString encodes the DirBlockInfo into a JSON string
func (d *DirBlockInfo) JSONString() (string, error) {
	return primitives.EncodeJSONString(d)
}

// New returns a new DirBlockInfo
func (d *DirBlockInfo) New() interfaces.BinaryMarshallableAndCopyable {
	return NewDirBlockInfo()
}

// GetDatabaseHeight returns this directory blocks height
func (d *DirBlockInfo) GetDatabaseHeight() uint32 {
	return d.DBHeight
}

// GetDBHeight returns this directory blocks height
func (d *DirBlockInfo) GetDBHeight() uint32 {
	return d.DBHeight
}

// GetBTCConfirmed returns the whether the BTC anchor has been confirmed
func (d *DirBlockInfo) GetBTCConfirmed() bool {
	return d.BTCConfirmed
}

// GetChainID returns the hard coded chain id "DirBlockInfo"
func (d *DirBlockInfo) GetChainID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirBlockInfo.GetChainID() saw an interface that was nil")
		}
	}()

	id := make([]byte, 32)
	copy(id, []byte("DirBlockInfo"))
	return primitives.NewHash(id)
}

// DatabasePrimaryIndex returns the Merkle Root of this directory block (initializing hashes if needed)
func (d *DirBlockInfo) DatabasePrimaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirBlockInfo.DatabasePrimaryIndex() saw an interface that was nil")
		}
	}()

	d.Init()
	return d.DBMerkleRoot
}

// DatabaseSecondaryIndex returns the serial hash of the directory block
func (d *DirBlockInfo) DatabaseSecondaryIndex() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirBlockInfo.DatabaseSecondaryIndex() saw an interface that was nil")
		}
	}()

	d.Init()
	return d.DBHash
}

// GetDBMerkleRoot returns the Merkle Root of this directory block (initializing hashes if needed)
func (d *DirBlockInfo) GetDBMerkleRoot() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirBlockInfo.GetDBMerkleRoot() saw an interface that was nil")
		}
	}()

	d.Init()
	return d.DBMerkleRoot
}

// GetBTCTxHash returns the BTC Tx hash (initializing hashes if needed)
func (d *DirBlockInfo) GetBTCTxHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DirBlockInfo.GetBTCTxHash() saw an interface that was nil")
		}
	}()

	d.Init()
	return d.BTCTxHash
}

// GetTimestamp returns the timestamp in milliseconds of the directory block creation
func (d *DirBlockInfo) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampFromMilliseconds(uint64(d.Timestamp))
}

// GetBTCBlockHeight returns the BTC block height this block is stored
func (d *DirBlockInfo) GetBTCBlockHeight() int32 {
	return d.BTCBlockHeight
}

// MarshalBinary marshals this DirBlockInfo into a byte array
func (d *DirBlockInfo) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DirBlockInfo.MarshalBinary err:%v", *pe)
		}
	}(&err)
	d.Init()
	var data primitives.Buffer

	enc := gob.NewEncoder(&data)

	err = enc.Encode(newDirBlockInfoCopyFromDBI(d))
	if err != nil {
		return nil, err
	}
	return data.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this DirBlockInfo
func (d *DirBlockInfo) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	dec := gob.NewDecoder(primitives.NewBuffer(data))
	dbic := newDirBlockInfoCopy()
	err = dec.Decode(dbic)
	if err != nil {
		return nil, err
	}
	d.parseDirBlockInfoCopy(dbic)
	return nil, nil
}

// UnmarshalBinary unmarshals th einput data into this DirBlockInfo
func (d *DirBlockInfo) UnmarshalBinary(data []byte) (err error) {
	_, err = d.UnmarshalBinaryData(data)
	return
}

// SetTimestamp sets the directory block's timestamp to now
func (d *DirBlockInfo) SetTimestamp(timestamp interfaces.Timestamp) {
	d.Timestamp = timestamp.GetTimeMilli()
}

// dirBlockInfoCopy is a literal copy of the DirBlockInfo struct above. This copy is needed to prevent
// an infinite recursion which causes the stack to overflow upon the MarshalBinary call in the test
// TestMarshalUnmarshal. I do not understand what is causing the infinite recursion, but am merely documenting
// this fact here.
type dirBlockInfoCopy struct {
	// Serial hash for the directory block
	DBHash    interfaces.IHash
	DBHeight  uint32 //directory block height
	Timestamp int64  // time of this dir block info being created
	// BTCTxHash is the Tx hash returned from rpcclient.SendRawTransaction
	BTCTxHash interfaces.IHash // use string or *btcwire.ShaHash ???
	// BTCTxOffset is the index of the TX in this BTC block
	BTCTxOffset int32
	// BTCBlockHeight is the height of the block where this TX is stored in BTC
	BTCBlockHeight int32
	//BTCBlockHash is the hash of the block where this TX is stored in BTC
	BTCBlockHash interfaces.IHash // use string or *btcwire.ShaHash ???
	// DBMerkleRoot is the merkle root of the Directory Block
	// and is written into BTC as OP_RETURN data
	DBMerkleRoot interfaces.IHash
	// A flag to to show BTC anchor confirmation
	BTCConfirmed bool

	EthereumAnchorRecordEntryHash interfaces.IHash
	EthereumConfirmed             bool
}

// newDirBlockInfoFromDBI creates a new DirBlockInfo that is a copy of the input
func newDirBlockInfoCopyFromDBI(dbi *DirBlockInfo) *dirBlockInfoCopy {
	dbic := new(dirBlockInfoCopy)
	dbic.DBHash = dbi.DBHash
	dbic.DBHeight = dbi.DBHeight
	dbic.Timestamp = dbi.Timestamp
	dbic.BTCTxHash = dbi.BTCTxHash
	dbic.BTCTxOffset = dbi.BTCTxOffset
	dbic.BTCBlockHeight = dbi.BTCBlockHeight
	dbic.BTCBlockHash = dbi.BTCBlockHash
	dbic.DBMerkleRoot = dbi.DBMerkleRoot
	dbic.BTCConfirmed = dbi.BTCConfirmed
	dbic.EthereumAnchorRecordEntryHash = dbi.EthereumAnchorRecordEntryHash
	dbic.EthereumConfirmed = dbi.EthereumConfirmed
	return dbic
}

// newDirBlockInfoCopy creates a new dirBlockInfoCopy with zero hashes filled
func newDirBlockInfoCopy() *dirBlockInfoCopy {
	dbi := new(dirBlockInfoCopy)
	dbi.DBHash = primitives.NewZeroHash()
	dbi.BTCTxHash = primitives.NewZeroHash()
	dbi.BTCBlockHash = primitives.NewZeroHash()
	dbi.DBMerkleRoot = primitives.NewZeroHash()
	dbi.EthereumAnchorRecordEntryHash = primitives.NewZeroHash()
	return dbi
}

// parseDirBlockInfoCopy sets this DirBlockInfo object to the values from the input
func (d *DirBlockInfo) parseDirBlockInfoCopy(dbi *dirBlockInfoCopy) {
	d.DBHash = dbi.DBHash
	d.DBHeight = dbi.DBHeight
	d.Timestamp = dbi.Timestamp
	d.BTCTxHash = dbi.BTCTxHash
	d.BTCTxOffset = dbi.BTCTxOffset
	d.BTCBlockHeight = dbi.BTCBlockHeight
	d.BTCBlockHash = dbi.BTCBlockHash
	d.DBMerkleRoot = dbi.DBMerkleRoot
	d.BTCConfirmed = dbi.BTCConfirmed
	d.EthereumAnchorRecordEntryHash = dbi.EthereumAnchorRecordEntryHash
	d.EthereumConfirmed = dbi.EthereumConfirmed
}

// NewDirBlockInfoFromDirBlock creates a new DirBlockInfo from the input
func NewDirBlockInfoFromDirBlock(dirBlock interfaces.IDirectoryBlock) *DirBlockInfo {
	dbi := new(DirBlockInfo)
	dbi.DBHash = dirBlock.GetHash()
	dbi.DBHeight = dirBlock.GetDatabaseHeight()
	dbi.DBMerkleRoot = dirBlock.GetKeyMR()
	dbi.SetTimestamp(dirBlock.GetHeader().GetTimestamp())
	dbi.BTCTxHash = primitives.NewZeroHash()
	dbi.BTCBlockHash = primitives.NewZeroHash()
	dbi.BTCConfirmed = false
	dbi.EthereumAnchorRecordEntryHash = primitives.NewZeroHash()
	dbi.EthereumConfirmed = false
	return dbi
}
