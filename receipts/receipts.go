// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package receipts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"strings"
)

type Receipt struct {
	Entry              *JSON
	EntryBlock         *JSON
	DirectoryBlock     *JSON
	BitcoinTransaction *JSON
	BitcoinBlock       *JSON `json:",omitempty"`
}

func (e *Receipt) IsSameAs(r *Receipt) bool {
	if e.Entry == nil {
		if r.Entry != nil {
			return false
		}
	}
	if e.Entry.IsSameAs(r.Entry) == false {
		return false
	}

	if e.EntryBlock == nil {
		if r.EntryBlock != nil {
			return false
		}
	}
	if e.EntryBlock.IsSameAs(r.EntryBlock) == false {
		return false
	}

	if e.DirectoryBlock == nil {
		if r.DirectoryBlock != nil {
			return false
		}
	}
	if e.DirectoryBlock.IsSameAs(r.DirectoryBlock) == false {
		return false
	}

	if e.BitcoinTransaction == nil {
		if r.BitcoinTransaction != nil {
			return false
		}
	}
	if e.BitcoinTransaction.IsSameAs(r.BitcoinTransaction) == false {
		return false
	}

	if e.BitcoinBlock == nil {
		if r.BitcoinBlock != nil {
			return false
		}
	}
	if e.BitcoinBlock.IsSameAs(r.BitcoinBlock) == false {
		return false
	}

	return true
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

func (e *Receipt) DecodeString(str string) error {
	jsonByte := []byte(str)
	err := json.Unmarshal(jsonByte, e)
	if err != nil {
		return err
	}
	return nil
}

func DecodeReceiptString(str string) (*Receipt, error) {
	receipt := new(Receipt)
	err := receipt.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return receipt, err
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

func (e *JSON) IsSameAs(r *JSON) bool {
	if r == nil {
		return false
	}
	if e.Raw != r.Raw {
		return false
	}
	if e.Key != r.Key {
		return false
	}
	if e.Json != r.Json {
		return false
	}
	return true
}

func CreateFullReceipt(dbo interfaces.DBOverlay, entryID interfaces.IHash) (*Receipt, error) {
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

func CreateMinimalReceipt(dbo interfaces.DBOverlay, entryID interfaces.IHash) (*Receipt, error) {
	receipt, err := CreateFullReceipt(dbo, entryID)
	if err != nil {
		return nil, err
	}
	if receipt.Entry != nil {
		receipt.Entry.Raw = ""
	}
	if receipt.EntryBlock != nil {
		receipt.EntryBlock.Raw = ""
	}
	if receipt.DirectoryBlock != nil {
		receipt.DirectoryBlock.Raw = ""
	}
	if receipt.BitcoinTransaction != nil {
		receipt.BitcoinTransaction.Raw = ""
	}
	if receipt.BitcoinBlock != nil {
		receipt.BitcoinBlock.Raw = ""
	}
	return receipt, nil
}

func VerifyFullReceipt(dbo interfaces.DBOverlay, receiptStr string) error {
	receipt, err := DecodeReceiptString(receiptStr)
	if err != nil {
		return err
	}
	if receipt.Entry == nil {
		return fmt.Errorf("receipt.Entry not found")
	}
	if receipt.EntryBlock == nil {
		return fmt.Errorf("receipt.EntryBlock not found")
	}
	if receipt.DirectoryBlock == nil {
		return fmt.Errorf("receipt.DirectoryBlock not found")
	}
	if receipt.BitcoinTransaction == nil {
		return fmt.Errorf("receipt.BitcoinTransaction not found")
	}
	if receipt.BitcoinBlock == nil {
		return fmt.Errorf("receipt.BitcoinBlock not found")
	}

	if receipt.Entry.Key == "" {
		return fmt.Errorf("receipt.Entry.Key is empty")
	}
	if receipt.EntryBlock.Key == "" {
		return fmt.Errorf("receipt.EntryBlock.Key is empty")
	}
	if receipt.DirectoryBlock.Key == "" {
		return fmt.Errorf("receipt.DirectoryBlock.Key is empty")
	}
	if receipt.BitcoinTransaction.Key == "" {
		return fmt.Errorf("receipt.BitcoinTransaction.Key is empty")
	}
	if receipt.BitcoinBlock.Key == "" {
		return fmt.Errorf("receipt.BitcoinBlock.Key is empty")
	}

	if strings.Contains(receipt.EntryBlock.Raw, receipt.Entry.Key) == false {
		return fmt.Errorf("Entry.Key not found in EntryBlock.Raw")
	}

	if strings.Contains(receipt.DirectoryBlock.Raw, receipt.EntryBlock.Key) == false {
		return fmt.Errorf("EntryBlock.Key not found in DirectoryBlock.Raw")
	}

	hash, err := primitives.NewShaHashFromStr(receipt.Entry.Key)
	if err != nil {
		return err
	}

	//TODO: verify better?
	receipt2, err := CreateFullReceipt(dbo, hash)
	if err != nil {
		return err
	}

	if receipt.IsSameAs(receipt2) == false {
		return fmt.Errorf("Receipt appears invalid - it doesn't match a recreated receipt")
	}

	return nil
}

func VerifyMinimalReceipt(dbo interfaces.DBOverlay, receiptStr string) error {
	return nil
}
