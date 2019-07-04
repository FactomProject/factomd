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

//AnchorRecord is used to construct anchor chain
type AnchorRecord struct {
	AnchorRecordVer int
	DBHeight        uint32 `json:",omitempty"` // The only directory block height included in this anchor
	KeyMR           string `json:",omitempty"` // Merkle root of the only directory block included in this anchor

	DBHeightMax uint32 `json:",omitempty"` // The highest directory block height included in this anchor window
	DBHeightMin uint32 `json:",omitempty"` // The lowest directory block height included in this anchor window
	WindowMR    string `json:",omitempty"` // Merkle root of all directory block KeyMRs from DBHeightMin to DBHeightMax

	RecordHeight uint32 // Directory block height we intended to put the AnchorRecord into

	Bitcoin  *BitcoinStruct  `json:",omitempty"`
	Ethereum *EthereumStruct `json:",omitempty"`
}

type BitcoinStruct struct {
	Address     string //"1HLoD9E4SDFFPDiYfNYnkBLQ85Y51J3Zb1",
	TXID        string //"9b0fc92260312ce44e74ef369f5c66bbb85848f2eddd5a7a1cde251e54ccfdd5", BTC Hash - in reverse byte order
	BlockHeight int32  //345678,
	BlockHash   string //"00000000000000000cc14eacfc7057300aea87bed6fee904fd8e1c1f3dc008d4", BTC Hash - in reverse byte order
	Offset      int32  //87
}

type EthereumStruct struct {
	ContractAddress string // Address of the Ethereum anchor contract
	TxID            string // Transaction ID of this particular anchor
	BlockHeight     int64  // Ethereum block height that this anchor was included in
	BlockHash       string // Hash of the Ethereum block that this anchor was included in
	TxIndex         int64  // Where the anchor tx is located within that block
}

var _ interfaces.Printable = (*AnchorRecord)(nil)
var _ interfaces.IAnchorRecord = (*AnchorRecord)(nil)

func (e *AnchorRecord) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AnchorRecord) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AnchorRecord) String() string {
	str, _ := e.JSONString()
	return str
}

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

func (ar *AnchorRecord) MarshalAndSign(priv interfaces.Signer) ([]byte, error) {
	data, err := ar.Marshal()
	if err != nil {
		return nil, err
	}
	sig := priv.Sign(data)
	return append(data, []byte(fmt.Sprintf("%x", sig.Bytes()))...), nil
}

func (ar *AnchorRecord) MarshalAndSignV2(priv interfaces.Signer) ([]byte, []byte, error) {
	data, err := ar.Marshal()
	if err != nil {
		return nil, nil, err
	}
	sig := priv.Sign(data)
	return data, sig.Bytes(), nil
}

func (ar *AnchorRecord) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("Invalid data passed")
	}
	str := string(data)
	end := strings.LastIndex(str, "}}")
	if end < 0 {
		return fmt.Errorf("Found no closing bracket in `%v`", str)
	}
	str = str[:end+2]
	err := json.Unmarshal([]byte(str), ar)
	if err != nil {
		return err
	}

	return nil
}

func UnmarshalAnchorRecord(data []byte) (*AnchorRecord, error) {
	ar := new(AnchorRecord)
	err := ar.Unmarshal(data)
	if err != nil {
		return nil, err
	}
	return ar, nil
}

func UnmarshalAndValidateAnchorRecord(data []byte, publicKeys []interfaces.Verifier) (*AnchorRecord, bool, error) {
	if len(data) == 0 {
		return nil, false, fmt.Errorf("Invalid data passed")
	}
	str := string(data)
	end := strings.LastIndex(str, "}}")
	if end < 0 {
		return nil, false, fmt.Errorf("Found no closing bracket in `%v`", str)
	}
	anchorStr := str[:end+2]
	signatureStr := str[end+2:]

	sig := new(primitives.ByteSliceSig)
	err := sig.UnmarshalText([]byte(signatureStr))
	if err != nil {
		return nil, false, err
	}
	fixed, err := sig.GetFixed()
	if err != nil {
		return nil, false, err
	}

	valid := false
	for _, publicKey := range publicKeys {
		valid = publicKey.Verify([]byte(anchorStr), &fixed)
		if valid == true {
			break
		}
	}
	if valid == false {
		return nil, false, nil
	}

	ar := new(AnchorRecord)
	err = ar.Unmarshal(data)
	if err != nil {
		return nil, false, err
	}
	return ar, true, nil
}

func UnmarshalAndValidateAnchorRecordV2(data []byte, extIDs [][]byte, publicKeys []interfaces.Verifier) (*AnchorRecord, bool, error) {
	if len(data) == 0 {
		return nil, false, fmt.Errorf("Invalid data passed")
	}
	if len(extIDs) != 1 {
		return nil, false, fmt.Errorf("Invalid External IDs passed")
	}

	sig := new(primitives.ByteSliceSig)
	sig.UnmarshalBinary(extIDs[0])
	fixed, err := sig.GetFixed()
	if err != nil {
		return nil, false, err
	}

	valid := false
	for _, publicKey := range publicKeys {
		valid = publicKey.Verify(data, &fixed)
		if valid == true {
			break
		}
	}
	if valid == false {
		return nil, false, nil
	}

	ar := new(AnchorRecord)
	err = ar.Unmarshal(data)
	if err != nil {
		return nil, false, err
	}
	return ar, true, nil
}

func UnmarshalAndValidateAnchorEntryAnyVersion(entry interfaces.IEBEntry, publicKeys []interfaces.Verifier) (*AnchorRecord, bool, error) {
	ar, valid, err := UnmarshalAndValidateAnchorRecord(entry.GetContent(), publicKeys)
	if ar == nil {
		ar, valid, err = UnmarshalAndValidateAnchorRecordV2(entry.GetContent(), entry.ExternalIDs(), publicKeys)
		return ar, valid, err
	}
	return ar, valid, err
}

func CreateAnchorRecordFromDBlock(dBlock interfaces.IDirectoryBlock) *AnchorRecord {
	ar := new(AnchorRecord)
	ar.AnchorRecordVer = 1
	ar.DBHeight = dBlock.GetHeader().GetDBHeight()
	ar.KeyMR = dBlock.DatabasePrimaryIndex().String()
	ar.RecordHeight = ar.DBHeight
	return ar
}
