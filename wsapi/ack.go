// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"encoding/hex"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

func HandleV2FactoidACK(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	n := time.Now()
	defer HandleV2APICallFctAck.Observe(float64(time.Since(n).Nanoseconds()))

	ackReq := new(AckRequest)
	err := MapToObject(params, ackReq)
	if err != nil {
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
		txid = tx.GetSigHash().String()
	}

	txhash, err := primitives.NewShaHashFromStr(txid)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	status, h, txTime, blockTime, err := state.GetACKStatus(txhash)
	if err != nil {
		return nil, NewInternalError()
	}
	answer := new(FactoidTxStatus)
	answer.TxID = h.String()

	if txTime != nil {
		answer.TransactionDate = txTime.GetTimeMilli()
		if txTime.GetTimeMilli() > 0 {
			answer.TransactionDateString = txTime.String()
		}
	}
	if blockTime != nil {
		answer.BlockDate = blockTime.GetTimeMilli()
		if blockTime.GetTimeMilli() > 0 {
			answer.BlockDateString = blockTime.String()
		}
	}

	answer.Status = constants.AckStatusString(status)
	if answer.Status == "na" {
		return nil, NewInternalError()
	}

	return answer, nil
}

// HandleV2ACKWithChain is the ack call with a given chainID. The chainID serves as a directive on what type
// of hash we are given, and we can act appropriately.
func HandleV2ACKWithChain(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	n := time.Now()
	defer HandleV2APICallEntryAck.Observe(float64(time.Since(n).Nanoseconds()))

	ackReq := new(EntryAckWithChainRequest)
	err := MapToObject(params, ackReq)
	if err != nil {
		return nil, NewInvalidParamsError()
	}

	if len(ackReq.ChainID) == 1 && (ackReq.ChainID == "c" || ackReq.ChainID == "f") {
		ackReq.ChainID = "000000000000000000000000000000000000000000000000000000000000000" + ackReq.ChainID
	}

	if len(ackReq.ChainID) != 64 {
		return nil, NewCustomInvalidParamsError("ChainID must be 64 hex encoded characters")
	}

	chainid, err := hex.DecodeString(ackReq.ChainID)
	if err != nil {
		return nil, NewCustomInvalidParamsError("ChainID must be 64 hex encoded characters")
	}
	var _ = chainid

	if ackReq.Hash == "" {
		// If it is a factoid transaction, it will be handled by the factoidack handler
		if ackReq.FullTransaction != "" && ackReq.FullTransaction != "000000000000000000000000000000000000000000000000000000000000000f" {
			ehash, commithash := DecodeTransactionToHashes(ackReq.FullTransaction)
			if ackReq.ChainID == "000000000000000000000000000000000000000000000000000000000000000c" {
				// Take the commit hash
				ackReq.Hash = commithash
			} else {
				// Take the entry hash
				ackReq.Hash = ehash
			}
		}
	}

	hash, err := primitives.HexToHash(ackReq.Hash)
	if err != nil {
		return nil, NewCustomInvalidParamsError("Hash must be 64 hex encoded characters")
	}

	answer := new(EntryStatus)
	switch ackReq.ChainID {
	case hex.EncodeToString(constants.EC_CHAINID):
		// This is an entry commit
		answer.CommitTxID = hash.String()
		status, blktime, com, entryhash := state.GetEntryCommitAckByTXID(hash)
		if blktime != nil {
			// If we have a block time, we can set that
			answer.CommitData.BlockDate = blktime.GetTime().Unix()
			answer.CommitData.BlockDateString = blktime.String()
			if entryhash != nil {
				// Looks like we found it's entryhash partner
				answer.EntryHash = entryhash.String()
				answer.EntryData.Status = constants.AckStatusString(status)
				answer.EntryData.BlockDate = blktime.GetTime().Unix()
				answer.EntryData.BlockDateString = blktime.String()
			}
		}

		if com != nil {
			answer.CommitData.TransactionDate = com.GetTimestamp().GetTime().Unix()
			answer.CommitData.TransactionDateString = com.GetTimestamp().String()
		}

		answer.CommitData.Status = constants.AckStatusString(status)
		return answer, nil
	case hex.EncodeToString(constants.FACTOID_CHAINID):
		// This is a factoid transaction, just use the old implementation for now
		otherAckReq := new(AckRequest)
		otherAckReq.TxID = ackReq.Hash
		otherAckReq.FullTransaction = ackReq.FullTransaction
		return HandleV2FactoidACK(state, otherAckReq)
	case hex.EncodeToString(constants.ADMIN_CHAINID):
		return nil, NewCustomInvalidParamsError("ChainID cannot be admin chain")
	}

	// If the chainid is not one of the special ones, that means the hash is an entry-hash
	// Will make a second function because of it's length
	return handleAckByEntryHash(hash, state)
}

// handleAckByEntryHash assumes the hash given is an entryhash
func handleAckByEntryHash(hash interfaces.IHash, state interfaces.IState) (interface{}, *primitives.JSONError) {
	answer := new(EntryStatus)
	// This is an entry
	revStatus, revBlktime, commit := state.GetEntryRevealAckByEntryHash(hash)
	answer.EntryHash = hash.String()
	answer.EntryData.Status = constants.AckStatusString(revStatus)
	if revBlktime != nil {
		answer.EntryData.BlockDate = revBlktime.GetTime().Unix()
		answer.EntryData.BlockDateString = revBlktime.String()
	}

	// If the reveal entry is found in the blockchain, that mean the reveal is in there too. We don't want to bother looking
	// anywhere else
	if revStatus == constants.AckStatusDBlockConfirmed {
		txid, err := state.FetchPaidFor(hash)
		if err == nil && txid != nil {
			answer.CommitTxID = txid.String()
			answer.CommitData.Status = constants.AckStatusString(constants.AckStatusDBlockConfirmed)
		}
		_, _, txTime, _, err := state.GetSpecificACKStatus(txid)
		if err != nil {
			return nil, NewInternalError()
		}
		if txTime != nil {
			answer.CommitData.TransactionDate = txTime.GetTimeMilli()
			if txTime.GetTimeMilli() > 0 {
				answer.CommitData.TransactionDateString = txTime.String()
			}
		}

		// Now we will exit, as any commit found below will be less than dblock confirmed.
		return answer, nil
	}

	// Continue here is not DBlockConfirmed

	if commit != nil {
		// This means it was found in the holding queue
		answer.CommitData.Status = constants.AckStatusString(constants.AckStatusNotConfirmed)
	} else {
		var status int
		// This means we have not found the commit yet
		status, commit = state.GetEntryCommitAckByEntryHash(hash)
		answer.CommitData.Status = constants.AckStatusString(status)
	}

	// If we found the commit, either by holding or by other means, set the variables on the response.
	if commit != nil {
		answer.CommitData.TransactionDateString = commit.GetTimestamp().String()
		answer.CommitData.TransactionDate = commit.GetTimestamp().GetTime().Unix()

		ce, ok := commit.(*messages.CommitEntryMsg)
		if ok {
			answer.CommitTxID = ce.CommitEntry.GetSigHash().String()
		}
		cc, ok := commit.(*messages.CommitChainMsg)
		if ok {
			answer.CommitTxID = cc.CommitChain.GetSigHash().String()
		}
	}

	return answer, nil
}

// HandleV2EntryACK will return the status of an entryhash or commit hash. We will assume the input is an entry hash, as with that
// the reveal and/or commit can be found. This is also beneficial as if the commit was rejected because a commit already exists,
// calling the entry-ack with the entryhash can find the exisiting commit.
// 		Order of search
//			Searching the Reveal
//				- ReplayFilter (_____________)
//					- Checking the replay filter tells us if an entry was recently submitted into the blockchain
//				- Check ProcessList (TransAck)
//					- Check NewEntry map
//					- Linear Search (Also check commits while going through, could save us a pass later)
//						- A non-processed reveal/commit will not be in the newentry map
//				- Check the Database (DblockConfirmed)
//					- We check the database last, despite it being the highest level of confirmation
//			Searching for the Commit
//				- Holding
//				- ProcessList (TransAck)
//					- Commit Map
//					- Linear search
//
// func handleV2EntryACK(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
// 	n := time.Now()
// 	defer HandleV2APICallEntryAck.Observe(float64(time.Since(n).Nanoseconds()))

// 	ackReq := new(AckRequest)

// 	err := MapToObject(params, ackReq)

// 	if err != nil {
// 		return nil, NewInvalidParamsError()
// 	}

// 	if ackReq.TxID == "" && ackReq.FullTransaction == "" {
// 		return nil, NewInvalidParamsError()
// 	}

// 	return nil, nil
// }

func HandleV2EntryACK(state interfaces.IState, params interface{}) (interface{}, *primitives.JSONError) {
	n := time.Now()
	defer HandleV2APICallEntryAck.Observe(float64(time.Since(n).Nanoseconds()))

	ackReq := new(AckRequest)

	err := MapToObject(params, ackReq)

	if err != nil {
		return nil, NewInvalidParamsError()
	}

	if ackReq.TxID == "" && ackReq.FullTransaction == "" {
		return nil, NewInvalidParamsError()
	}

	eTxID := ""
	ecTxID := ""

	if ackReq.TxID == "" {
		eTxID, ecTxID = DecodeTransactionToHashes(ackReq.FullTransaction)
		if ecTxID == "" && eTxID == "" {
			return nil, NewUnableToDecodeTransactionError()
		}
	}

	//We didn't receive a full transaction, but a transaction hash
	//We have to figure out which transaction hash we got
	if eTxID == "" {
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
				//	ecTxID = ackReq.TxID
				eTxID = ec.GetEntryHash().String()
			}

			// havent found entry or chain transaction.  check all of the Process Lists
			if eTxID == "" {
				eHash, err := state.FetchEntryHashFromProcessListsByTxID(ackReq.TxID)
				if err != nil {
					wsLog.Println("FetchEntryHashFromProcessListsByTxID:", err)
				} else {
					eTxID = eHash.String()
				}
			}

			//pend := state.GetPendingEntries(params)  // covered elsewhere
			/// still havent found them.  Check the Acks queue
			aQue := state.LoadAcksMap()

			for _, a := range aQue {
				if a.Type() == constants.REVEAL_ENTRY_MSG {
					var rm messages.RevealEntryMsg
					enb, err := a.MarshalBinary()
					if err != nil {
						return nil, NewInternalError()
					}
					err = rm.UnmarshalBinary(enb)
					if err != nil {
						return nil, NewInternalError()
					}
					if rm.Entry.GetHash().String() == ackReq.TxID {
						eTxID = rm.Entry.GetHash().String()
					}
					//	ecTxID = rm.Entry. GetChainIDHash().String()
				} else if a.Type() == constants.COMMIT_ENTRY_MSG {
					var rm messages.CommitEntryMsg
					enb, err := a.MarshalBinary()
					if err != nil {
						return nil, NewInternalError()
					}
					err = rm.UnmarshalBinary(enb)
					if err != nil {
						return nil, NewInternalError()
					}

					if rm.CommitEntry.GetSigHash().String() == ackReq.TxID {
						eTxID = rm.CommitEntry.GetEntryHash().String()

					}
					//	ecTxID = rm.CommitEntry.GetEntryHash().String()
				} else if a.Type() == constants.COMMIT_CHAIN_MSG {
					var rm messages.CommitChainMsg
					enb, err := a.MarshalBinary()
					if err != nil {
						return nil, NewInternalError()
					}
					err = rm.UnmarshalBinary(enb)
					if err != nil {
						return nil, NewInternalError()
					}
					//	ecTxID = rm.CommitChain.ChainIDHash.String()
					if rm.CommitChain.GetSigHash().String() == ackReq.TxID {
						eTxID = rm.CommitChain.GetSigHash().String()
					}
				}
			}

			// still havent found them.  Check the holding queue
			if ecTxID == "" && eTxID == "" {
				hQue := state.LoadHoldingMap()

				for _, h := range hQue {
					if h.Type() == constants.REVEAL_ENTRY_MSG {
						var rm messages.RevealEntryMsg
						enb, err := h.MarshalBinary()
						if err != nil {
							return nil, NewInternalError()
						}
						err = rm.UnmarshalBinary(enb)
						if err != nil {
							return nil, NewInternalError()
						}
						if rm.Entry.GetHash().String() == ackReq.TxID {
							eTxID = rm.Entry.GetHash().String()
							//		ecTxID = ackReq.TxID
						}
					} else if h.Type() == constants.COMMIT_ENTRY_MSG {
						var rm messages.CommitEntryMsg
						enb, err := h.MarshalBinary()
						if err != nil {
							return nil, NewInternalError()
						}
						err = rm.UnmarshalBinary(enb)
						if err != nil {
							return nil, NewInternalError()
						}

						if rm.CommitEntry.GetSigHash().String() == ackReq.TxID {
							eTxID = rm.CommitEntry.GetEntryHash().String()
							//		ecTxID = ackReq.TxID

						}

					} else if h.Type() == constants.COMMIT_CHAIN_MSG {
						var rm messages.CommitChainMsg
						enb, err := h.MarshalBinary()
						if err != nil {
							return nil, NewInternalError()
						}
						err = rm.UnmarshalBinary(enb)
						if err != nil {
							return nil, NewInternalError()
						}

						if rm.CommitChain.GetSigHash().String() == ackReq.TxID {
							eTxID = rm.CommitChain.GetSigHash().String()
							//		ecTxID = ackReq.TxID
						}

					} else {
						//	fmt.Println("I DONT KNOW THIS Holding Message TYPE:", h.Type())
					}
				}
			}
		}
	}

	answer := new(EntryStatus)
	answer.CommitTxID = ecTxID
	answer.EntryHash = eTxID
	//	answer.CommitData.Status = AckStatusACK
	//	answer.EntryData.Status = AckStatusACK
	//	return answer, nil
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
		h, err := primitives.NewShaHashFromStr(answer.CommitTxID)
		if err != nil {
			return nil, NewInvalidParamsError()
		}

		status, txid, txTime, blockTime, err := state.GetSpecificACKStatus(h)
		if err != nil {
			return nil, NewInternalError()
		}

		answer.CommitTxID = txid.String()

		if txTime != nil {
			answer.CommitData.TransactionDate = txTime.GetTimeMilli()
			if txTime.GetTimeMilli() > 0 {
				answer.CommitData.TransactionDateString = txTime.String()
			}
		}
		if blockTime != nil {
			answer.CommitData.BlockDate = blockTime.GetTimeMilli()
			if blockTime.GetTimeMilli() > 0 {
				answer.CommitData.BlockDateString = blockTime.String()
			}
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
		}
	}

	if answer.EntryHash == "" {
		answer.EntryData.Status = AckStatusUnknown
	} else {
		h, err := primitives.NewShaHashFromStr(answer.EntryHash)
		if err != nil {
			return nil, NewInvalidParamsError()
		}

		status, txid, txTime, blockTime, err := state.GetSpecificACKStatus(h)
		if err != nil {
			return nil, NewInternalError()
		}

		answer.EntryHash = txid.String()

		if txTime != nil {
			answer.EntryData.TransactionDate = txTime.GetTimeMilli()
			if txTime.GetTimeMilli() > 0 {
				answer.EntryData.TransactionDateString = txTime.String()
			}
		}
		if blockTime != nil {
			answer.EntryData.BlockDate = blockTime.GetTimeMilli()
			if blockTime.GetTimeMilli() > 0 {
				answer.EntryData.BlockDateString = blockTime.String()
			}
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
		}
	}

	return answer, nil
}

func DecodeTransactionToHashes(fullTransaction string) (eTxID string, ecTxID string) {
	//fmt.Printf("DecodeTransactionToHashes - %v\n", fullTransaction)
	b, err := hex.DecodeString(fullTransaction)
	if err != nil {
		return
	}

	cc := new(entryCreditBlock.CommitChain)
	rest, err := cc.UnmarshalBinaryData(b)
	if err != nil || len(rest) > 0 {
		ec := new(entryCreditBlock.CommitEntry)
		rest, err = ec.UnmarshalBinaryData(b)
		if err != nil || len(rest) > 0 {
			e := new(entryBlock.Entry)
			rest, err = e.UnmarshalBinaryData(b)
			if err != nil || len(rest) > 0 {
				return
			} else {
				eTxID = e.GetHash().String()
			}
		} else {
			eTxID = ec.GetEntryHash().String()
			ecTxID = ec.GetHash().String()
		}
	} else {
		eTxID = cc.GetEntryHash().String()
		ecTxID = cc.GetHash().String()
	}

	//fmt.Printf("eTxID - %v\n", eTxID)
	//fmt.Printf("ecTxID - %v\n", ecTxID)
	return
}

type AckRequest struct {
	TxID            string `json:"txid,omitempty"`
	FullTransaction string `json:"fulltransaction,omitempty"`
}

type EntryAckWithChainRequest struct {
	Hash            string `json:"hash,omitempty"`
	ChainID         string `json:"chainid,omitempty"`
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
	TransactionDate       int64  `json:"transactiondate,omitempty"`       //Unix time
	TransactionDateString string `json:"transactiondatestring,omitempty"` //ISO8601 time
	BlockDate             int64  `json:"blockdate,omitempty"`             //Unix time
	BlockDateString       string `json:"blockdatestring,omitempty"`       //ISO8601 time

	Malleated *Malleated `json:"malleated,omitempty"`
	Status    string     `json:"status"`
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
