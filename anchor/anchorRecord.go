// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

// logger is based on github.com/alexcesaro/log and
// github.com/alexcesaro/log/golog (MIT License)

package anchor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

//AnchorRecord is used to construct anchor chain
type AnchorRecord struct {
	AnchorRecordVer int
	DBHeight        uint32
	KeyMR           string
	RecordHeight    uint32

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
	Address     string //0x30aa981f6d2fce81083e584c8ee2f822b548752f
	TXID        string //0x50ea0effc383542811a58704a6d6842ed6d76439a2d942d941896ad097c06a78
	BlockHeight int64  //293003
	BlockHash   string //0x3b504616495fc9cf7be9b5b776692a9abbfb95491fa62abf62dcdf4d53ff5979
	Offset      int64  //0
}

var _ interfaces.Printable = (*AnchorRecord)(nil)
var _ interfaces.IAnchorRecord = (*AnchorRecord)(nil)

func (e *AnchorRecord) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AnchorRecord) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AnchorRecord) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *AnchorRecord) String() string {
	str, _ := e.JSONString()
	return str
}

func (ar *AnchorRecord) Marshal() ([]byte, error) {
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
	return append(data, sig.Bytes()...), nil
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

func UnmarshalAndvalidateAnchorRecord(data []byte, publicKey interfaces.Verifier) (*AnchorRecord, bool, error) {
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
	sig.UnmarshalText([]byte(signatureStr))
	fixed, err := sig.GetFixed()
	if err != nil {
		return nil, false, err
	}

	valid := publicKey.Verify([]byte(anchorStr), &fixed)
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

func CreateAnchorRecordFromDBlock(dBlock interfaces.IDirectoryBlock) *AnchorRecord {
	ar := new(AnchorRecord)
	ar.AnchorRecordVer = 2
	ar.DBHeight = dBlock.GetHeader().GetDBHeight()
	ar.KeyMR = dBlock.DatabasePrimaryIndex().String()
	ar.RecordHeight = ar.DBHeight
	return ar
}
