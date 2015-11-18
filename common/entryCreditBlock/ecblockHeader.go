// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package entryCreditBlock

import (
	"bytes"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type ECBlockHeader struct {
	ECChainID           interfaces.IHash
	BodyHash            interfaces.IHash
	PrevHeaderHash      interfaces.IHash
	PrevLedgerKeyMR     interfaces.IHash
	DBHeight            uint32
	HeaderExpansionArea []byte
	ObjectCount         uint64
	BodySize            uint64
}

var _ = fmt.Print
var _ interfaces.Printable = (*ECBlockHeader)(nil)

func (e *ECBlockHeader) SetBodySize(cnt uint64){
	e.BodySize = cnt
}

func (e *ECBlockHeader) GetBodySize() (uint64 ){
	return e.BodySize
}

func (e *ECBlockHeader) SetObjectCount(cnt uint64){
	e.ObjectCount = cnt
}

func (e *ECBlockHeader) GetObjectCount() (uint64 ){
	return e.ObjectCount
}

func (e *ECBlockHeader) SetHeaderExpansionArea(area []byte){
	e.HeaderExpansionArea = area
}

func (e *ECBlockHeader) GetHeaderExpansionArea() (area []byte ){
	return e.HeaderExpansionArea
}

func (e *ECBlockHeader) SetBodyHash(prev interfaces.IHash){
	e.BodyHash = prev
}

func (e *ECBlockHeader) GetBodyHash() (prev interfaces.IHash){
	return e.BodyHash 
}


func (e *ECBlockHeader) SetECChainID(prev interfaces.IHash){
	e.ECChainID = prev
}

func (e *ECBlockHeader) GetECChainID() (prev interfaces.IHash){
	return e.ECChainID 
}

func (e *ECBlockHeader) SetPrevHeaderHash(prev interfaces.IHash){
	e.PrevHeaderHash = prev
}

func (e *ECBlockHeader) GetPrevHeaderHash() (prev interfaces.IHash){
	return e.PrevHeaderHash 
}

func (e *ECBlockHeader) SetPrevLedgerKeyMR(prev interfaces.IHash) {
	e.PrevLedgerKeyMR = prev
}

func (e *ECBlockHeader) GetPrevLedgerKeyMR() (prev interfaces.IHash) {
	return e.PrevLedgerKeyMR
}

func (e *ECBlockHeader) SetDBHeight(height uint32) {
	e.DBHeight = height
}

func (e *ECBlockHeader) GetDBHeight() (height uint32) {
	return e.DBHeight
}

func NewECBlockHeader() *ECBlockHeader {
	h := new(ECBlockHeader)
	h.ECChainID = primitives.NewZeroHash()
	h.ECChainID.SetBytes(constants.EC_CHAINID)
	h.BodyHash = primitives.NewZeroHash()
	h.PrevHeaderHash = primitives.NewZeroHash()
	h.PrevLedgerKeyMR = primitives.NewZeroHash()
	h.HeaderExpansionArea = make([]byte, 0)
	return h
}

func (e *ECBlockHeader) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ECBlockHeader) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *ECBlockHeader) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *ECBlockHeader) String() string {
	str, _ := e.JSONString()
	return str
}


