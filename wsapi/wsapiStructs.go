// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/receipts"
)

type FactoidSubmitResponse struct {
	Message string `json:"message"`
	TxID    string `json:"txid"`
}

type CommitChainResponse struct {
	Message     string `json:"message"`
	TxID        string `json:"txid"`
	EntryHash   string `json:"entryhash,omitempty"`
	ChainIDHash string `json:"chainidhash,omitempty"`
}

type RevealChainResponse struct {
}

type CommitEntryResponse struct {
	Message   string `json:"message"`
	TxID      string `json:"txid"`
	EntryHash string `json:"entryhash,omitempty"`
}

type RevealEntryResponse struct {
	Message   string `json:"message"`
	EntryHash string `json:"entryhash"`
	ChainID   string `json:"chainid,omitempty"`
}

type DirectoryBlockResponse struct {
	Header struct {
		PrevBlockKeyMR string `json:"prevblockkeymr"`
		SequenceNumber int64  `json:"sequencenumber"`
		Timestamp      int64  `json:"timestamp"`
	} `json:"header"`
	EntryBlockList []EBlockAddr `json:"entryblocklist"`
}

type DirectoryBlockHeadResponse struct {
	KeyMR string `json:"keymr"`
	//Add height, etc?
}

// type DirectoryBlockHeightResponse struct {
// 	Height int64 `json:"height"`
// }

type HeightsResponse struct {
	DirectoryBlockHeight         int64 `json:"directoryblockheight"`
	LeaderHeight                 int64 `json:"leaderheight"`
	EntryBlockHeight             int64 `json:"entryblockheight"`
	EntryHeight                  int64 `json:"entryheight"`
	MissingEntryCount            int64 `json:"-"`
	EntryBlockDBHeightProcessing int64 `json:"-"`
	EntryBlockDBHeightComplete   int64 `json:"-"`
}

type CurrentMinuteResponse struct {
	LeaderHeight            int64 `json:"leaderheight"`
	DirectoryBlockHeight    int64 `json:"directoryblockheight"`
	Minute                  int64 `json:"minute"`
	CurrentBlockStartTime   int64 `json:"currentblockstarttime"`
	CurrentMinuteStartTime  int64 `json:"currentminutestarttime"`
	CurrentTime             int64 `json:"currenttime"`
	DirectoryBlockInSeconds int64 `json:"directoryblockinseconds"`
	StallDetected           bool  `json:"stalldetected"`
	FaultTimeOut            int64 `json:"faulttimeout"`
	RoundTimeOut            int64 `json:"roundtimeout"`
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
		BlockSequenceNumber int64  `json:"blocksequencenumber"`
		ChainID             string `json:"chainid"`
		PrevKeyMR           string `json:"prevkeymr"`
		Timestamp           int64  `json:"timestamp"`
		DBHeight            int64  `json:"dbheight"`
	} `json:"header"`
	EntryList []EntryAddr `json:"entrylist"`
}

type EntryCreditBlockResponse struct {
	ECBlock struct {
		Header     interfaces.IECBlockHeader `json:"header"`
		Body       interfaces.IECBlockBody   `json:"body"`
		HeaderHash interfaces.IHash          `json:"headerhash"`
		FullHash   interfaces.IHash          `json:"fullhash"`
	} `json:"ecblock"`
	RawData string `json:"rawdata"`
}

type EntryResponse struct {
	ChainID string   `json:"chainid"`
	Content string   `json:"content"`
	ExtIDs  []string `json:"extids"`
}

type ChainHeadResponse struct {
	ChainHead          string `json:"chainhead"`
	ChainInProcessList bool   `json:"chaininprocesslist"`
}

type EntryCreditBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type FactoidBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type EntryCreditRateResponse struct {
	Rate int64 `json:"rate"`
}

type PropertiesResponse struct {
	FactomdVersion string `json:"factomdversion"`
	ApiVersion     string `json:"factomdapiversion"`
}

type SendRawMessageResponse struct {
	Message string `json:"message"`
}

type TransactionRateResponse struct {
	TotalTransactionRate   float64 `json:"totaltxrate"`
	InstantTransactionRate float64 `json:"instanttxrate"`
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
		SequenceNumber int64  `json:"sequencenumber"`
		Timestamp      int64  `json:"timestamp"`
	} `json:"header"`
	EntryBlockList []EBlockAddr `json:"entryblocklist"`
}

func (e *DBlock) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

type EntryAddr struct {
	EntryHash string `json:"entryhash"`
	Timestamp int64  `json:"timestamp"`
}

type EBlock struct {
	Header struct {
		BlockSequenceNumber int64  `json:"blocksequencenumber"`
		ChainID             string `json:"chainid"`
		PrevKeyMR           string `json:"prevkeymr"`
		Timestamp           int64  `json:"timestamp"`
		DBHeight            int64  `json:"dbheight"`
	} `json:"header"`
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

type TransactionResponse struct {
	ECTranasction      interfaces.IECBlockEntry `json:"ectransaction,omitempty"`
	FactoidTransaction interfaces.ITransaction  `json:"factoidtransaction,omitempty"`
	Entry              interfaces.IEBEntry      `json:"entry,omitempty"`

	//F/EC/E block the transaction is included in
	IncludedInTransactionBlock string `json:"includedintransactionblock"`
	//DirectoryBlock the tranasction is included in
	IncludedInDirectoryBlock string `json:"includedindirectoryblock"`
	//The DBlock height
	IncludedInDirectoryBlockHeight int64 `json:"includedindirectoryblockheight"`
}

type BlockHeightResponse struct {
	DBlock  *JStruct `json:"dblock,omitempty"`
	ABlock  *JStruct `json:"ablock,omitempty"`
	FBlock  *JStruct `json:"fblock,omitempty"`
	ECBlock *JStruct `json:"ecblock,omitempty"`
	RawData string   `json:"rawdata,omitempty"`
}

//Requests

type AddressRequest struct {
	Address string `json:"address"`
}

type HeightRequest struct {
	Height int64 `json:"height"`
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

type PendingEntry struct {
	EntryHash interfaces.IHash `json:"entryhash"`
	ChainID   interfaces.IHash `json:"chainid"`
	Status    string           `json:"status"`
}

type PendingTransaction struct {
	TransactionID interfaces.IHash `json:"transactionid"`
	Status        string           `json:"status"`
}

type TransactionRequest struct {
	Transaction string `json:"transaction"`
}

type SendRawMessageRequest struct {
	Message string `json:"message"`
}

type FactiodAccounts struct {
	NumbOfAccounts string   `json:numberofacc`
	Height         uint32   `json:"height"`
	Accounts       []string `json:accounts`
}

type MultipleFTBalances struct {
	CurrentHeight   uint32        `json:"currentheight"`
	LastSavedHeight uint32        `json:"lastsavedheight"`
	Balances        []interface{} `json:"balances"`
}

type MultipleECBalances struct {
	CurrentHeight   uint32        `json:"currentheight"`
	LastSavedHeight uint32        `json:"lastsavedheight"`
	Balances        []interface{} `json:"balances"`
}

type DiagnosticsResponse struct {
	Name              string `json:"name"`
	ID                string `json:"id,omitempty"`
	PublicKey         string `json:"publickey,omitempty"`
	Role              string `json:"role"`
	SyncingState      string `json:"syncingstate"`
	SyncingReceived   int    `json:"syncingreceived"`
	SyncingExpected   int
	SyncingMissingIDs []string `json:"missingnodesids"`
	VMs               []VM     `json:"vms"`

	ElectionInfo *ElectionInfo `json:"elections"`
}

type ElectionInfo struct {
	InProgress      bool     `json:"inprogress"`
	StateAuthSet    AuthSet  `json:"stateauthset"`
	ElectionAuthSet AuthSet  `json:"electionauthset"`
	AuditHeartBeats []string `json:"auditheartbeats"`

	VmIndex  *int   `json:"vmindex,omitempty"`
	FedIndex *int   `json:"fedindex,omitempty"`
	FedID    string `json:"fedid,omitempty"`
	Round    *int   `json:"round,omitempty"`
}

type AuthSet struct {
	Leaders []string `json:"leaders"`
	Audits  []string `json:"audits"`
}

type VM struct {
	CurrentProcessList struct {
		Height  int    `json:"height"`
		Length  int    `json:"length"`
		NextNil string `json:"nextnil"`
	} `json:"currentprocesslist"`
	PrevBlockCreatedFrom string `json:"prevblockcreatedfrom"`
}
