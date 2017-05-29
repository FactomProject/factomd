// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func (s *State) IsStateFullySynced() bool {
	ll := s.ProcessLists.LastList()

	return s.ProcessLists.DBHeightBase < ll.DBHeight
}

// GetACKStatus also checks the oldmsgs map
func (s *State) GetACKStatus(hash interfaces.IHash) (int, interfaces.IHash, interfaces.Timestamp, interfaces.Timestamp, error) {
	return s.getACKStatus(hash, true)
}

// GetSpecificACKStatus does NOT check the oldmsgs map. This is because the processlists map for entries and entry blocks is
// updated after the oldmsgs. This means an EntryACK will returns TransactionACK, but GetChain will return not found
// To fix this, for some calls (entries) we don't want to check the oldmsgs.
func (s *State) GetSpecificACKStatus(hash interfaces.IHash) (int, interfaces.IHash, interfaces.Timestamp, interfaces.Timestamp, error) {
	return s.getACKStatus(hash, false)
}

//returns status, proper transaction ID, transaction timestamp, block timestamp, and an error
func (s *State) getACKStatus(hash interfaces.IHash, useOldMsgs bool) (int, interfaces.IHash, interfaces.Timestamp, interfaces.Timestamp, error) {
	msg := s.GetInvalidMsg(hash)
	if msg != nil {
		return constants.AckStatusInvalid, hash, nil, nil, nil
	}

	in, err := s.DB.FetchIncludedIn(hash)
	if err != nil {
		return 0, hash, nil, nil, err
	}

	if in == nil {
		// Not in database.  Check Process Lists

		for _, pl := range s.ProcessLists.Lists {
			//pl := s.ProcessLists.LastList()
			if useOldMsgs {
				m := pl.GetOldMsgs(hash)
				if m != nil || pl.DirectoryBlock == nil {
					return constants.AckStatusACK, hash, m.GetTimestamp(), nil, nil
				}
			}

			ts := pl.DirectoryBlock.GetHeader().GetTimestamp()

			keys := pl.GetKeysNewEntries()
			for _, k := range keys {
				tx := pl.GetNewEntry(k)
				if hash.IsSameAs(tx.GetHash()) {
					return constants.AckStatusACK, hash, nil, ts, nil
				}
			}
			ecBlock := pl.EntryCreditBlock
			if ecBlock != nil {
				tx := ecBlock.GetEntryByHash(hash)
				if tx != nil {
					return constants.AckStatusACK, tx.GetSigHash(), tx.GetTimestamp(), ts, nil
				}
			}

			fBlock := s.FactoidState.GetCurrentBlock()
			if fBlock != nil {
				tx := fBlock.GetTransactionByHash(hash)
				if tx != nil {
					return constants.AckStatusACK, tx.GetSigHash(), tx.GetTimestamp(), ts, nil
				}
			}
		}

		//	 We are now looking into the holding queue.  it should have been found by now if it is going to be
		//	  if included has not been found, but we have no information, it should be unknown not unconfirmed.

		if s.IsStateFullySynced() {
			status, _, _, _ := s.FetchHoldingMessageByHash(hash)
			return status, hash, nil, nil, nil

		} else {
			return constants.AckStatusUnknown, hash, nil, nil, nil
		}
	}

	in2, err := s.DB.FetchIncludedIn(in)
	if err != nil {
		return 0, hash, nil, nil, err
	}
	dBlock, err := s.DB.FetchDBlock(in2)
	if err != nil {
		return 0, hash, nil, nil, err
	}
	fBlock, err := s.DB.FetchFBlock(in)
	if err != nil {
		return 0, hash, nil, nil, err
	}
	if fBlock != nil {
		tx := fBlock.GetTransactionByHash(hash)
		if tx == nil {
			return 0, hash, nil, nil, fmt.Errorf("Transaction not found in a block we were expecting")
		}
		return constants.AckStatusDBlockConfirmed, tx.GetSigHash(), tx.GetTimestamp(), dBlock.GetHeader().GetTimestamp(), nil
	}
	ecBlock, err := s.DB.FetchECBlock(in)
	if err != nil {
		return 0, hash, nil, nil, err
	}
	if ecBlock != nil {
		tx := ecBlock.GetEntryByHash(hash)
		if tx == nil {
			return 0, hash, nil, nil, fmt.Errorf("Transaction not found in a block we were expecting")
		}
		return constants.AckStatusDBlockConfirmed, tx.GetSigHash(), tx.GetTimestamp(), dBlock.GetHeader().GetTimestamp(), nil
	}

	//entries have no timestamp of their own, so return nil

	return constants.AckStatusDBlockConfirmed, hash, nil, dBlock.GetHeader().GetTimestamp(), nil

}

func (s *State) FetchHoldingMessageByHash(hash interfaces.IHash) (int, byte, interfaces.IMsg, error) {
	q := s.LoadHoldingMap()
	for _, h := range q {
		switch {
		//	case h.Type() == constants.EOM_MSG :
		//	case h.Type() == constants.ACK_MSG :
		//	case h.Type() == constants.FED_SERVER_FAULT_MSG :
		//	case h.Type() == constants.AUDIT_SERVER_FAULT_MSG :
		//	case h.Type() == constants.FULL_SERVER_FAULT_MSG :
		case h.Type() == constants.COMMIT_CHAIN_MSG:
			var rm messages.CommitChainMsg
			enb, err := h.MarshalBinary()
			err = rm.UnmarshalBinary(enb)
			if hash.IsSameAs(rm.CommitChain.GetSigHash()) {
				return constants.AckStatusNotConfirmed, constants.REVEAL_ENTRY_MSG, h, err
			}
		case h.Type() == constants.COMMIT_ENTRY_MSG:
			var rm messages.CommitEntryMsg
			enb, err := h.MarshalBinary()
			err = rm.UnmarshalBinary(enb)
			if hash.IsSameAs(rm.CommitEntry.GetSigHash()) {
				return constants.AckStatusNotConfirmed, constants.REVEAL_ENTRY_MSG, h, err
			}
			//	case h.Type() == constants.DIRECTORY_BLOCK_SIGNATURE_MSG :
			//	case h.Type() == constants.EOM_TIMEOUT_MSG :
		case h.Type() == constants.FACTOID_TRANSACTION_MSG:
			var rm messages.FactoidTransaction
			enb, err := h.MarshalBinary()
			err = rm.UnmarshalBinary(enb)
			if hash.IsSameAs(rm.Transaction.GetSigHash()) {
				return constants.AckStatusNotConfirmed, constants.FACTOID_TRANSACTION_MSG, h, err
			}
			//	case h.Type() == constants.HEARTBEAT_MSG :
			//	case h.Type() == constants.INVALID_ACK_MSG :
			//	case h.Type() == constants.INVALID_DIRECTORY_BLOCK_MSG :
		case h.Type() == constants.REVEAL_ENTRY_MSG:
			var rm messages.RevealEntryMsg
			enb, err := h.MarshalBinary()
			err = rm.UnmarshalBinary(enb)
			if hash.IsSameAs(rm.Entry.GetHash()) {
				return constants.AckStatusNotConfirmed, constants.REVEAL_ENTRY_MSG, h, err
			}
			//	case  h.Type() == constants.REQUEST_BLOCK_MSG :
			//	case h.Type() == constants.SIGNATURE_TIMEOUT_MSG:
			//	case h.Type() == constants.MISSING_MSG :
			//	case h.Type() == constants.MISSING_DATA :
			//	case h.Type() == constants.DATA_RESPONSE :
			//	case h.Type() == constants.MISSING_MSG_RESPONSE:
			//	case h.Type() == constants.DBSTATE_MSG :
			//	case h.Type() == constants.DBSTATE_MISSING_MSG:
			//	case h.Type() == constants.ADDSERVER_MSG:
			//	case h.Type() == constants.CHANGESERVER_KEY_MSG:
			//	case h.Type() == constants.REMOVESERVER_MSG:
			//	case h.Type() == constants.BOUNCE_MSG:
			//	case h.Type() == constants.BOUNCEREPLY_MSG:
			//	case h.Type() == constants.MISSING_ENTRY_BLOCKS:
			//	case h.Type() == constants.ENTRY_BLOCK_RESPONSE :

		}
	}
	return constants.AckStatusUnknown, byte(0), nil, fmt.Errorf("Not Found")
}

func (s *State) FetchECTransactionByHash(hash interfaces.IHash) (interfaces.IECBlockEntry, error) {
	//TODO: expand to search data from outside database
	if hash == nil {
		return nil, nil
	}

	var currentHeightComplete = s.GetDBHeightComplete()
	pls := s.ProcessLists.Lists
	for _, pl := range pls {
		if pl != nil {
			if pl.DBHeight > currentHeightComplete {
				ecBlock := pl.EntryCreditBlock
				if ecBlock != nil {
					tx := ecBlock.GetEntryByHash(hash)
					if tx != nil {
						return tx, nil
					}
				}
			}
		}
	}

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchECTransaction(hash)
}

func (s *State) FetchFactoidTransactionByHash(hash interfaces.IHash) (interfaces.ITransaction, error) {
	//TODO: expand to search data from outside database
	if hash == nil {
		return nil, nil
	}
	fBlock := s.FactoidState.GetCurrentBlock()
	if fBlock != nil {
		tx := fBlock.GetTransactionByHash(hash)
		if tx != nil {
			return tx, nil
		}
	}
	// not in FactoidSate lists.  try process listsholding queue
	// check holding queue

	var currentHeightComplete = s.GetDBHeightComplete()
	pls := s.ProcessLists.Lists
	for _, pl := range pls {
		// ignore old process lists
		// watch for nil while syncing blockchain
		if pl != nil {
			if pl.DBHeight > currentHeightComplete {
				cb := pl.State.FactoidState.GetCurrentBlock()
				ct := cb.GetTransactions()
				for _, tx := range ct {
					if tx.GetHash().IsSameAs(hash) {
						return tx, nil
					}
				}
			}
		}
	}

	q := s.LoadHoldingMap()
	for _, h := range q {
		if h.Type() == constants.FACTOID_TRANSACTION_MSG {
			var rm messages.FactoidTransaction
			enb, err := h.MarshalBinary()
			if err != nil {
				return nil, err
			}
			err = rm.UnmarshalBinary(enb)
			if err != nil {
				return nil, err
			}
			tx := rm.GetTransaction()
			if tx.GetHash().IsSameAs(hash) {
				return tx, nil
			}
		}
	}

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchFactoidTransaction(hash)
}

func (s *State) FetchPaidFor(hash interfaces.IHash) (interfaces.IHash, error) {
	//TODO: expand to search data from outside database
	if hash == nil {
		return nil, nil
	}

	for _, pls := range s.ProcessLists.Lists {
		ecBlock := pls.EntryCreditBlock
		for _, tx := range ecBlock.GetEntries() {
			switch tx.ECID() {
			case entryCreditBlock.ECIDEntryCommit:
				if hash.IsSameAs(tx.(*entryCreditBlock.CommitEntry).EntryHash) {
					return tx.GetSigHash(), nil
				}
				break
			case entryCreditBlock.ECIDChainCommit:
				if hash.IsSameAs(tx.(*entryCreditBlock.CommitChain).EntryHash) {
					return tx.GetSigHash(), nil
				}
				break
			}
		}
	}
	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchPaidFor(hash)
}

func (s *State) FetchEntryByHash(hash interfaces.IHash) (interfaces.IEBEntry, error) {
	//TODO: expand to search data from outside database
	if hash == nil {
		return nil, nil
	}

	//pl := s.ProcessLists.LastList()
	for _, pl := range s.ProcessLists.Lists {
		keys := pl.GetKeysNewEntries()

		for _, key := range keys {
			tx := pl.GetNewEntry(key)
			if hash.IsSameAs(tx.GetHash()) {
				fmt.Println("returningProcesslist hash")
				return tx, nil
			}
		}
	}

	// not in process lists.  try holding queue
	// check holding queue

	q := s.LoadHoldingMap()
	var re messages.RevealEntryMsg
	for _, h := range q {
		if h.Type() == constants.REVEAL_ENTRY_MSG {
			enb, err := h.MarshalBinary()
			if err != nil {
				return nil, err
			}
			err = re.UnmarshalBinary(enb)
			if err != nil {
				return nil, err
			}
			tx := re.Entry
			if hash.IsSameAs(tx.GetHash()) {
				return tx, nil
			}
		}
	}

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchEntry(hash)
}
