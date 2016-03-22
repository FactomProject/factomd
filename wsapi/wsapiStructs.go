// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/receipts"
)

type FactoidSubmitResponse struct {
	Message string
	TxID    string
}

type CommitChainResponse struct {
	Message string
	TxID    string
}

type RevealChainResponse struct {
}

type CommitEntryResponse struct {
	Message string
	TxID    string
}

type RevealEntryResponse struct {
	Message string
	TxID    string
}

type DirectoryBlockHeadResponse struct {
	KeyMR string
	//Add height, etc?
}

type GetRawDataResponse struct {
	Data string
	//TODO: add
}

type GetReceiptResponse struct {
	Receipt *receipts.Receipt
}

type DirectoryBlockResponse struct {
	Header struct {
		PrevBlockKeyMR string
		SequenceNumber uint32
		Timestamp      uint32
	}
	EntryBlockList []EBlockAddr
}

type EntryBlockResponse struct {
	Header struct {
		BlockSequenceNumber uint32
		ChainID             string
		PrevKeyMR           string
		Timestamp           uint32
		DBHeight            uint32
	}
	EntryList []EntryAddr
}

type EntryResponse struct {
	ChainID string
	Content string
	ExtIDs  []string
}

type ChainHeadResponse struct {
	ChainHead string
}

type EntryCreditBalanceResponse struct {
	Balance int64
}

type FactoidBalanceResponse struct {
	Balance int64
}

type FactoidGetFeeResponse struct {
	Fee uint64
}

type DirectoryBlockHeightResponse struct {
	Height int64
}

type PropertiesResponse struct {
	Factomd_Version string
	Protocol_Version string
}

/*********************************************************************/

type DBHead struct {
	KeyMR string
}

type RawData struct {
	Data string
}

type EBlockAddr struct {
	ChainID string
	KeyMR   string
}

type DBlock struct {
	Header struct {
		PrevBlockKeyMR string
		SequenceNumber uint32
		Timestamp      uint32
	}
	EntryBlockList []EBlockAddr
}

func (e *DBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type EntryAddr struct {
	EntryHash string
	Timestamp uint32
}

type EBlock struct {
	Header struct {
		BlockSequenceNumber uint32
		ChainID             string
		PrevKeyMR           string
		Timestamp           uint32
		DBHeight            uint32
	}
	EntryList []EntryAddr
}

func (e *EBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type EntryStruct struct {
	ChainID string
	Content string
	ExtIDs  []string
}

func (e *EntryStruct) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type CHead struct {
	ChainHead string
}

func (e *CHead) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type FactoidBalance struct {
	Response string
	Success  bool
}
