// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package receipts

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
)

type Receipt struct {
	Entry              *JSON
	EntryBlock         *JSON
	DirectoryBlock     *JSON
	BitcoinTransaction *JSON
	BitcoinBlock       *JSON `json:",omitempty"`
}

func (e *Receipt) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Receipt) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Receipt) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *Receipt) String() string {
	str, _ := e.JSONString()
	return str
}

type JSON struct {
	Raw  string `json:",omitempty"`
	Key  string `json:",omitempty"`
	Json string `json:",omitempty"`
}

func (e *JSON) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *JSON) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *JSON) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *JSON) String() string {
	str, _ := e.JSONString()
	return str
}

func CreateFullReceipt(dbo *databaseOverlay.Overlay, entryID interfaces.IHash) (*Receipt, error) {
	receipt := new(Receipt)
	receipt.Entry = new(JSON)
	receipt.Entry.Key = entryID.String()

	hash, err := dbo.LoadIncludedIn(entryID)
	if err != nil {
		return nil, err
	}

	if hash == nil {
		return nil, fmt.Errorf("Block containing entry not found")
	}

	receipt.EntryBlock = new(JSON)
	receipt.EntryBlock.Key = hash.String()

	eBlock, err := dbo.FetchEBlockByKeyMR(hash)
	if err != nil {
		return nil, err
	}

	if eBlock == nil {
		return nil, fmt.Errorf("EBlock not found")
	}
	hex, err := eBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	receipt.EntryBlock.Raw = fmt.Sprintf("%x", hex)

	hash = eBlock.DatabasePrimaryIndex()

	hash, err = dbo.LoadIncludedIn(hash)
	if err != nil {
		return nil, err
	}

	if hash == nil {
		return nil, fmt.Errorf("Block containing EBlock not found")
	}

	receipt.DirectoryBlock = new(JSON)
	receipt.DirectoryBlock.Key = hash.String()

	dBlock, err := dbo.FetchDBlockByKeyMR(hash)
	if err != nil {
		return nil, err
	}

	if dBlock == nil {
		return nil, fmt.Errorf("DBlock not found")
	}
	hex, err = dBlock.MarshalBinary()
	if err != nil {
		return nil, err
	}
	receipt.DirectoryBlock.Raw = fmt.Sprintf("%x", hex)

	hash = dBlock.DatabasePrimaryIndex()

	dirBlockInfo, err := dbo.FetchDirBlockInfoByKeyMR(hash)
	if err != nil {
		return nil, err
	}

	if dirBlockInfo == nil {
		return nil, fmt.Errorf("dirBlockInfo not found")
	}
	dbi := dirBlockInfo.(*dbInfo.DirBlockInfo)

	receipt.BitcoinTransaction = new(JSON)
	receipt.BitcoinTransaction.Key = dbi.BTCTxHash.String()
	receipt.BitcoinBlock = new(JSON)
	receipt.BitcoinBlock.Key = dbi.BTCBlockHash.String()

	return receipt, nil
}

func CreateMinimalReceipt(dbo *databaseOverlay.Overlay, entryID interfaces.IHash) {

}
