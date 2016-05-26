// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"encoding/hex"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func HandleV2FactoidACK(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	ackReq, ok := params.(AckRequest)
	if !ok {
		return nil, NewInvalidParamsError()
	}

	if ackReq.TxID == "" && ackReq.FullTransaction == "" {
		return nil, NewInvalidParamsError()
	}

	txid := ackReq.TxID

	if txid == "" {
		b, err := hex.DecodeString(ackReq.FullTransaction)
		if err != nil {
			return nil, NewUnableToDecodeTransactionError()
		}
		tx := new(factoid.Transaction)
		err = tx.UnmarshalBinary(b)
		if err != nil {
			return nil, NewUnableToDecodeTransactionError()
		}
		txid = tx.GetHash().String()
	}

	txhash, err := primitives.NewShaHashFromStr(txid)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	status, err := state.GetACKStatus(txhash)
	if err != nil {
		return nil, NewInternalError()
	}

	answer := new(FactoidTxStatus)
	answer.TxID = txid

	switch status {
	case constants.AckStatusInvalid:
		answer.Status = AckStatusInvalid
		break
	case constants.AckStatusUnknown:
		answer.Status = AckStatusUnknown
		break
	case constants.AckStatusNotConfirmed:
		answer.Status = AckStatusNotConfirmed
		break
	case constants.AckStatusACK:
		answer.Status = AckStatusACK
		break
	case constants.AckStatus1Minute:
		answer.Status = AckStatus1Minute
		break
	case constants.AckStatusDBlockConfirmed:
		answer.Status = AckStatusDBlockConfirmed
		break
	default:
		return nil, NewInternalError()
		break
	}

	return answer, nil
}

func HandleV2EntryACK(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	ackReq, ok := params.(AckRequest)
	if !ok {
		return nil, NewInvalidParamsError()
	}

	if ackReq.TxID == "" && ackReq.FullTransaction == "" {
		return nil, NewInvalidParamsError()
	}

	eTxID := ""
	ecTxID := ""

	if ackReq.TxID == "" {
		b, err := hex.DecodeString(ackReq.FullTransaction)
		if err != nil {
			return nil, NewUnableToDecodeTransactionError()
		}
		e := new(entryBlock.Entry)
		err = e.UnmarshalBinary(b)
		if err != nil {
			ec := new(entryCreditBlock.CommitEntry)
			err = ec.UnmarshalBinary(b)
			if err != nil {
				cc := new(entryCreditBlock.CommitChain)
				err = cc.UnmarshalBinary(b)
				if err != nil {
					return nil, NewUnableToDecodeTransactionError()
				} else {
					eTxID = cc.EntryHash.String()
					ecTxID = ackReq.TxID
				}
			} else {
				eTxID = ec.EntryHash.String()
				ecTxID = ackReq.TxID
			}
		} else {
			eTxID = ackReq.TxID
		}
	}

	//TODO: fetch entries, ec TXs from state as well

	//We didn't receive a full transaction, but a transaction hash
	//We have to figure out which transaction hash we got
	if ecTxID == "" && eTxID == "" {
		h, err := primitives.NewShaHashFromStr(ackReq.TxID)
		if err != nil {
			return nil, NewInvalidParamsError()
		}
		entry, err := state.FetchEntryByHash(h)
		if err != nil {
			return nil, NewInternalError()
		}
		if entry != nil {
			eTxID = ackReq.TxID
		} else {
			ec, err := state.FetchECTransactionByHash(h)
			if err != nil {
				return nil, NewInternalError()
			}
			if ec != nil {
				ecTxID = ackReq.TxID
				eTxID = ec.GetEntryHash().String()
			}
		}
	}

	answer := new(EntryStatus)
	answer.CommitTxID = ecTxID
	answer.EntryHash = eTxID

	if answer.CommitTxID == "" && answer.EntryHash == "" {
		//We know nothing about the transaction, so we return unknown status
		answer.CommitData.Status = AckStatusUnknown
		answer.EntryData.Status = AckStatusUnknown
		return answer, nil
	}

	//Fetching the second part of the transaction pair
	if answer.EntryHash == "" {
		h, err := primitives.NewShaHashFromStr(answer.EntryHash)
		if err != nil {
			return nil, NewInvalidParamsError()
		}
		ec, err := state.FetchECTransactionByHash(h)
		if err != nil {
			return nil, NewInternalError()
		}
		if ec != nil {
			answer.EntryHash = ec.GetEntryHash().String()
		}
	}

	if answer.CommitTxID == "" {
		h, err := primitives.NewShaHashFromStr(answer.EntryHash)
		if err != nil {
			return nil, NewInvalidParamsError()
		}
		ec, err := state.FetchPaidFor(h)
		if err != nil {
			return nil, NewInternalError()
		}
		if ec != nil {
			answer.CommitTxID = ec.String()
		}
	}

	//Fetching statuses
	if answer.CommitTxID == "" {
		answer.CommitData.Status = AckStatusUnknown
	} else {
		h, err := primitives.NewShaHashFromStr(answer.EntryHash)
		if err != nil {
			return nil, NewInvalidParamsError()
		}

		status, err := state.GetACKStatus(h)
		if err != nil {
			return nil, NewInternalError()
		}

		switch status {
		case constants.AckStatusInvalid:
			answer.CommitData.Status = AckStatusInvalid
			break
		case constants.AckStatusUnknown:
			answer.CommitData.Status = AckStatusUnknown
			break
		case constants.AckStatusNotConfirmed:
			answer.CommitData.Status = AckStatusNotConfirmed
			break
		case constants.AckStatusACK:
			answer.CommitData.Status = AckStatusACK
			break
		case constants.AckStatus1Minute:
			answer.CommitData.Status = AckStatus1Minute
			break
		case constants.AckStatusDBlockConfirmed:
			answer.CommitData.Status = AckStatusDBlockConfirmed
			break
		default:
			return nil, NewInternalError()
			break
		}
	}

	if answer.EntryHash == "" {
		answer.EntryData.Status = AckStatusUnknown
	} else {
		h, err := primitives.NewShaHashFromStr(answer.EntryHash)
		if err != nil {
			return nil, NewInvalidParamsError()
		}

		status, err := state.GetACKStatus(h)
		if err != nil {
			return nil, NewInternalError()
		}

		switch status {
		case constants.AckStatusInvalid:
			answer.EntryData.Status = AckStatusInvalid
			break
		case constants.AckStatusUnknown:
			answer.EntryData.Status = AckStatusUnknown
			break
		case constants.AckStatusNotConfirmed:
			answer.EntryData.Status = AckStatusNotConfirmed
			break
		case constants.AckStatusACK:
			answer.EntryData.Status = AckStatusACK
			break
		case constants.AckStatus1Minute:
			answer.EntryData.Status = AckStatus1Minute
			break
		case constants.AckStatusDBlockConfirmed:
			answer.EntryData.Status = AckStatusDBlockConfirmed
			break
		default:
			return nil, NewInternalError()
			break
		}
	}

	return answer, nil
}

type AckRequest struct {
	TxID            string `json:"txid,omitempty"`
	FullTransaction string `json:"fulltransaction,omitempty"`
}

type FactoidTxStatus struct {
	TxID string `json:"txid"`
	GeneralTransactionData
}

type EntryStatus struct {
	CommitTxID string `json:"committxid"`
	EntryHash  string `json:"entryhash"`

	CommitData GeneralTransactionData `json:"commitdata"`
	EntryData  GeneralTransactionData `json:"entrydata"`

	ReserveTransactions          []ReserveInfo `json:"reserveinfo,omitempty"`
	ConflictingRevealEntryHashes []string      `json:"conflictingrevealentryhashes,omitempty"`
}

type ReserveInfo struct {
	TxID    string `json:"txid"`
	Timeout int64  `json:"timeout"` //Unix time
}

type GeneralTransactionData struct {
	TransactionDate int64      `json:"transactiondate"` //Unix time
	Malleated       *Malleated `json:"malleated,omitempty"`
	Status          string     `json:"status"`
}

type Malleated struct {
	MalleatedTxIDs []string `json:"malleatedtxids"`
}

const (
	AckStatusInvalid         = "Invalid"
	AckStatusUnknown         = "Unknown"
	AckStatusNotConfirmed    = "NotConfirmed"
	AckStatusACK             = "TransactionACK"
	AckStatus1Minute         = "1Minute"
	AckStatusDBlockConfirmed = "DBlockConfirmed"
)
