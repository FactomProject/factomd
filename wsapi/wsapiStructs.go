// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/common/primitives"
)

type FactoidSubmitResponse {

}

type CommitChainResponse {

}

type RevealChainResponse {

}

type CommitEntryResponse {

}

type RevealEntryResponse {

}

type DirectoryBlockHeadResponse {

}

type GetRawDataResponse {

}

type GetReceiptResponse {

}

type DirectoryBlockResponse {

}

type EntryBlockResponse {

}

type EntryResponse {

}

type ChainHeadResponse {

}

type EntryCreditBalanceResponse {

}

type FactoidBalanceResponse {

}

type FactoidGetFeeResponse {

}

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
