// Copyright 2015 Factom Foundation
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

//returns status, proper transaction ID, transaction timestamp, block timestamp, and an error
func (s *State) GetACKStatus(hash interfaces.IHash) (int, interfaces.IHash, interfaces.Timestamp, interfaces.Timestamp, error) {
	for _, pl := range s.ProcessLists.Lists {
		//pl := s.ProcessLists.LastList()
		m := pl.GetOldMsgs(hash)
		if m != nil || pl.DirectoryBlock == nil {
			return constants.AckStatusACK, hash, m.GetTimestamp(), nil, nil
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
	msg := s.GetInvalidMsg(hash)
	if msg != nil {
		return constants.AckStatusInvalid, hash, nil, nil, nil
	}
	in, err := s.DB.FetchIncludedIn(hash)
	if err != nil {
		return 0, hash, nil, nil, err
	}
	if in == nil {
		if s.IsStateFullySynced() {
			return constants.AckStatusNotConfirmed, hash, nil, nil, nil
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
func (s *State) FetchECTransactionByHash(hash interfaces.IHash) (interfaces.IECBlockEntry, error) {
	//TODO: expand to search data from outside database
	if hash == nil {
		return nil, nil
	}

	ecBlock := s.ProcessLists.LastList().EntryCreditBlock
	if ecBlock != nil {
		tx := ecBlock.GetEntryByHash(hash)
		if tx != nil {
			return tx, nil
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
		if pl.DBHeight > currentHeightComplete {
			cb := pl.State.FactoidState.GetCurrentBlock()
			ct := cb.GetTransactions()
			for _, tx := range ct {
				if tx.GetHash() == hash {
					return tx, nil
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
			if tx.GetHash() == hash {
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
			return tx, nil
		}
	}

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchEntry(hash)
}
