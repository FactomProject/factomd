// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (s *State) IsStateFullySynced() bool {
	ll := s.ProcessLists.LastList()
	return s.ProcessLists.DBHeightBase < ll.DBHeight
}

//returns status, proper transaction ID, transaction timestamp, block timestamp, and an error
func (s *State) GetACKStatus(hash interfaces.IHash) (int, interfaces.IHash, interfaces.Timestamp, interfaces.Timestamp, error) {
	pl := s.ProcessLists.LastList()
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

	fBlock, err = s.DB.FetchFBlock(in)
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

	ecBlock, err = s.DB.FetchECBlock(in)
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

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchFactoidTransaction(hash)
}

func (s *State) FetchPaidFor(hash interfaces.IHash) (interfaces.IHash, error) {
	//TODO: expand to search data from outside database
	if hash == nil {
		return nil, nil
	}

	ecBlock := s.ProcessLists.LastList().EntryCreditBlock
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

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchPaidFor(hash)
}

func (s *State) FetchEntryByHash(hash interfaces.IHash) (interfaces.IEBEntry, error) {
	//TODO: expand to search data from outside database
	if hash == nil {
		return nil, nil
	}

	pl := s.ProcessLists.LastList()
	keys := pl.GetKeysNewEntries()

	for _, key := range keys {
		tx := pl.GetNewEntry(key)
		if hash.IsSameAs(tx.GetHash()) {
			return tx, nil
		}
	}

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchEntry(hash)
}
