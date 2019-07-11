// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// logger is based on github.com/alexcesaro/log and
// github.com/alexcesaro/log/golog (MIT License)

package anchor

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AnchorRecord is used to construct the anchor chain. The Factom Protocol writes an anchor into
// a parent blockchain (Bitcoin or Ethereum) approximately every 10 minutes. The accumulated entries
// from the previous 10 minutes are organized into a Directory Block (DBlock). The information
// from the directory block is stored in the AnchorRecord.
type AnchorRecord struct {
	AnchorRecordVer int    // version 1 places signature content in the field, version 2 uses external IDs for signature
	DBHeight        uint32 // Factom Directory Block Height - the unique number associated with this DBlock
	KeyMR           string // key merkle root of the directory block

	DBHeightMax uint32 `json:",omitempty"` // The highest directory block height included in this anchor window
	DBHeightMin uint32 `json:",omitempty"` // The lowest directory block height included in this anchor window
	WindowMR    string `json:",omitempty"` // Merkle root of all directory block KeyMRs from DBHeightMin to DBHeightMax

	RecordHeight uint32 // Likely to be deprecated soon. This is the future DBlock height a confirmation of anchoring event X is
	// **intended** to be written to (usually X+1). The field is fairly useless since you can't know for sure
	// what block an entry will be included in until after the entry is confirmed, and by that point, you can
	// get the block height from it's parent Entry Block.

	Bitcoin  *BitcoinStruct  `json:",omitempty"`
	Ethereum *EthereumStruct `json:",omitempty"`
}

// BitcoinStruct contains relevant data for a Bitcoin transaction
type BitcoinStruct struct {
	Address     string //"1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1",
	TXID        string //"9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5", BTC Hash - in reverse byte order
	BlockHeight int32  //345678,
	BlockHash   string //"00000000000000000cc14eacfc7057300aea87bed6fee904fd8e1c1f3dc008d4", BTC Hash - in reverse byte order
	Offset      int32  //87
}

// EthereumStruct contains relevant data for an Ethereum transaction
type EthereumStruct struct {
	ContractAddress string // Address of the Ethereum anchor contract
	TxID            string // Transaction ID of this particular anchor
	BlockHeight     int64  // Ethereum block height that this anchor was included in
	BlockHash       string // Hash of the Ethereum block that this anchor was included in
	TxIndex         int64  // Where the anchor tx is located within that block
}

var _ interfaces.Printable = (*AnchorRecord)(nil)
var _ interfaces.IAnchorRecord = (*AnchorRecord)(nil)

// JSONByte returns a []byte of the AnchorRecord encoded in Json: nil, error returned upon error
func (ar *AnchorRecord) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(ar)
}

// JSONString returns a string of the AnchorRecord encoded in Json: "", error returned upon error
func (ar *AnchorRecord) JSONString() (string, error) {
	return primitives.EncodeJSONString(ar)
}

// String returns a string of AnchorRecord encoded in Json: "" returned upon error
func (ar *AnchorRecord) String() string {
	str, _ := ar.JSONString()
	return str
}

// Marshal marshals the AnchorRecord into json format
func (ar *AnchorRecord) Marshal() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AnchorRecord.Marshal err:%v", *pe)
		}
	}(&err)
	data, err := json.Marshal(ar)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MarshalAndSign marshals the AnchorRecord into json and signs it with the input
// Signer, returning concatenated data of (data,signature)
func (ar *AnchorRecord) MarshalAndSign(priv interfaces.Signer) ([]byte, error) {
	data, sigbytes, err := ar.MarshalAndSignV2(priv)
	if err != nil {
		return nil, err
	}
	return append(data, []byte(fmt.Sprintf("%x", sigbytes))...), nil
}

// MarshalAndSignV2 marshals the AnchorRecord into json and signs, returning separate anchor and signature data
func (ar *AnchorRecord) MarshalAndSignV2(priv interfaces.Signer) ([]byte, []byte, error) {
	data, err := ar.Marshal()
	if err != nil {
		return nil, nil, err
	}
	sig := priv.Sign(data)
	return data, sig.Bytes(), nil
}

// Non-exported, refactored function that splits the AnchorRecord and its signature
func splitAnchorAndSignature(data []byte) (string, string, error) {
	if len(data) == 0 {
		return "", "", fmt.Errorf("Invalid data passed")
	}
	str := string(data)
	end := strings.LastIndex(str, "}}")
	if end < 0 {
		return "", "", fmt.Errorf("Found no closing bracket in `%v`", str)
	}
	// Signature comes after anchor
	anchorStr := str[:end+2]
	sigStr := str[end+2:]

	return anchorStr, sigStr, nil
}

// Unmarshal unmarshals json format input into this AnchorRecord
func (ar *AnchorRecord) Unmarshal(data []byte) error {
	str, _, err := splitAnchorAndSignature(data)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(str), ar)
	if err != nil {
		return err
	}

	return nil
}

// IsSame returns true iff all the fields of ar==ar2
func (ar *AnchorRecord) IsSame(ar2 *AnchorRecord) bool {
	if ar.AnchorRecordVer != ar2.AnchorRecordVer || ar.DBHeight != ar2.DBHeight || ar.KeyMR != ar2.KeyMR ||
		ar.DBHeightMax != ar2.DBHeightMax || ar.DBHeightMin != ar2.DBHeightMin || ar.WindowMR != ar2.WindowMR ||
		ar.RecordHeight != ar2.RecordHeight || (ar.Bitcoin != nil && ar.Bitcoin.IsSame(ar2.Bitcoin) == false) ||
		(ar.Ethereum != nil && ar.Ethereum.IsSame(ar2.Ethereum) == false) {
		return false
	}
	return true
}

// UnmarshalAnchorRecord unmarshals a json format input into a new AnchorRecord
func UnmarshalAnchorRecord(data []byte) (*AnchorRecord, error) {
	ar := new(AnchorRecord)
	err := ar.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return ar, nil
}

// verifyAnchorAndSignature verifies the data and signature from the public keys - unexported
func verifyAnchorAndSignature(data []byte, sig *primitives.ByteSliceSig, publicKeys []interfaces.Verifier) (bool, error) {
	fixed, err := sig.GetFixed()
	if err != nil {
		return false, err
	}

	valid := false
	for _, publicKey := range publicKeys {
		valid = publicKey.Verify(data, &fixed)
		if valid == true {
			break
		}
	}
	return valid, nil
}

// UnmarshalAndValidateAnchorRecord unmarshals signed json data and verifies signature with public keys for AnchorRecord v1
func UnmarshalAndValidateAnchorRecord(data []byte, publicKeys []interfaces.Verifier) (*AnchorRecord, bool, error) {
	anchorStr, signatureStr, err := splitAnchorAndSignature(data)
	if err != nil {
		return nil, false, err
	}

	sig := new(primitives.ByteSliceSig)
	err = sig.UnmarshalText([]byte(signatureStr))
	if err != nil {
		return nil, false, err
	}

	valid := false
	valid, err = verifyAnchorAndSignature([]byte(anchorStr), sig, publicKeys)
	if valid == false {
		return nil, false, err
	}

	ar, err := UnmarshalAnchorRecord(data)
	if err != nil {
		return nil, false, err
	}
	return ar, true, nil
}

// UnmarshalAndValidateAnchorRecordV2 unmarshals json data and verifies external signature with public keys using AnchorRecord v2
func UnmarshalAndValidateAnchorRecordV2(data []byte, extIDs [][]byte, publicKeys []interfaces.Verifier) (*AnchorRecord, bool, error) {
	if len(data) == 0 {
		return nil, false, fmt.Errorf("Invalid data passed")
	}
	if len(extIDs) != 1 {
		return nil, false, fmt.Errorf("Invalid External IDs passed")
	}

	sig := new(primitives.ByteSliceSig)
	sig.UnmarshalBinary(extIDs[0])
	valid, err := verifyAnchorAndSignature(data, sig, publicKeys)
	if valid == false {
		return nil, false, err
	}

	ar := new(AnchorRecord)
	err = ar.Unmarshal(data)
	if err != nil {
		return nil, false, err
	}
	return ar, true, nil
}

// UnmarshalAndValidateAnchorEntryAnyVersion unmarshals json data and verifies with either internal or external signatures against public keys
func UnmarshalAndValidateAnchorEntryAnyVersion(entry interfaces.IEBEntry, publicKeys []interfaces.Verifier) (*AnchorRecord, bool, error) {
	ar, valid, err := UnmarshalAndValidateAnchorRecord(entry.GetContent(), publicKeys)
	if ar == nil {
		ar, valid, err = UnmarshalAndValidateAnchorRecordV2(entry.GetContent(), entry.ExternalIDs(), publicKeys)
		return ar, valid, err
	}
	return ar, valid, err
}

// IsSame returns true iff all fields of BitcoinStructs bc==bc2
func (bc *BitcoinStruct) IsSame(bc2 *BitcoinStruct) bool {
	if bc2 == nil || bc.Address != bc2.Address || bc.TXID != bc2.TXID || bc.BlockHeight != bc2.BlockHeight || bc.BlockHash != bc2.BlockHash ||
		bc.Offset != bc2.Offset {
		return false
	}
	return true
}

// IsSame returns true iff all fields of EthereumStructs es==es2
func (es *EthereumStruct) IsSame(es2 *EthereumStruct) bool {
	if es2 == nil || es.ContractAddress != es2.ContractAddress || es.TxID != es2.TxID || es.BlockHeight != es2.BlockHeight || es.BlockHash != es2.BlockHash ||
		es.TxIndex != es.TxIndex {
		return false
	}
	return true
}
