// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/receipts"
)

type FactoidSubmitResponse struct {
	Message string `json:"message"`
	TxID    string `json:"txid"`
}

type CommitChainResponse struct {
	Message string `json:"message"`
	TxID    string `json:"txid"`
}

type RevealChainResponse struct {
}

type CommitEntryResponse struct {
	Message string `json:"message"`
	TxID    string `json:"txid"`
}

type RevealEntryResponse struct {
	Message string `json:"message"`
	TxID    string `json:"txid"`
}

type DirectoryBlockResponse struct {
	Header struct {
		PrevBlockKeyMR string `json:"prevblockkeymr"`
		SequenceNumber uint32 `json:"sequencenumber"`
		Timestamp      uint32 `json:"timestamp"`
	} `json:"header"`
	EntryBlockList []EBlockAddr `json:"entryblocklist"`
}

type DirectoryBlockHeadResponse struct {
	KeyMR string `json:"keymr"`
	//Add height, etc?
}

type DirectoryBlockHeightResponse struct {
	Height int64 `json:"height"`
}

type RawDataResponse struct {
	Data string `json:"data"`
	//TODO: add
}

type ReceiptResponse struct {
	Receipt *receipts.Receipt `json:"receipt"`
}

type EntryBlockResponse struct {
	Header struct {
		BlockSequenceNumber uint32 `json:"blocksequencenumber"`
		ChainID             string `json:"chainid"`
		PrevKeyMR           string `json:"prevkeymr"`
		Timestamp           uint32 `json:"timestamp"`
		DBHeight            uint32 `json:"dbheight"`
	} `json:"header"`
	EntryList []EntryAddr `json:"timestamp"`
}

type EntryResponse struct {
	ChainID string   `json:"chainid"`
	Content string   `json:"content"`
	ExtIDs  []string `json:"extids"`
}

type ChainHeadResponse struct {
	ChainHead string `json:"chainhead"`
}

type EntryCreditBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type FactoidBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type FactoidFeeResponse struct {
	Fee uint64 `json:"fee"`
}

type PropertiesResponse struct {
	FactomdVersion string `json:"factomdversion"`
	ApiVersion     string `json:"apiversion"`
}

/*********************************************************************/

type DBHead struct {
	KeyMR string `json:"keymr"`
}

type EBlockAddr struct {
	ChainID string `json:"chainid"`
	KeyMR   string `json:"keymr"`
}

type DBlock struct {
	Header struct {
		PrevBlockKeyMR string `json:"prevblockkeymr"`
		SequenceNumber uint32 `json:"sequencenumber"`
		Timestamp      uint32 `json:"timestamp"`
	} `json:"header"`
	EntryBlockList []EBlockAddr `json:"entryblocklist"`
}

func (e *DBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type EntryAddr struct {
	EntryHash string `json:"entryhash"`
	Timestamp uint32 `json:"timestamp"`
}

type EBlock struct {
	Header struct {
		BlockSequenceNumber uint32 `json:"blocksequencenumber"`
		ChainID             string `json:"chainid"`
		PrevKeyMR           string `json:"prevkeymr"`
		Timestamp           uint32 `json:"timestamp"`
		DBHeight            uint32 `json:"dbheight"`
	}
	EntryList []EntryAddr `json:"entrylist"`
}

func (e *EBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type EntryStruct struct {
	ChainID string   `json:"chainid"`
	Content string   `json:"content"`
	ExtIDs  []string `json:"extids"`
}

func (e *EntryStruct) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type CHead struct {
	ChainHead string `json:"chainhead"`
}

func (e *CHead) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type FactoidBalance struct {
	Response string `json:"response"`
	Success  bool   `json:"success"`
}

//Requests

type AddressRequest struct {
	Address string `json:"address"`
}

type ChainIDRequest struct {
	ChainID string `json:"chainid"`
}

type EntryRequest struct {
	Entry string `json:"entry"`
}

type HashRequest struct {
	Hash string `json:"hash"`
}

type KeyMRRequest struct {
	KeyMR string `json:"keymr"`
}

type KeyRequest struct {
	Key string `json:"key"`
}

type MessageRequest struct {
	Message string `json:"message"`
}

type TransactionRequest struct {
	Transaction string `json:"transaction"`
}
