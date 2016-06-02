// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
)

func (s *State) IsStateFullySynced() bool {
	return s.ProcessLists.DBHeightBase == uint32(len(s.ProcessLists.Lists))
}

func (s *State) GetACKStatus(hash interfaces.IHash) (int, error) {
	_, found := s.ProcessLists.LastList().OldMsgs[hash.Fixed()]
	if found {
		return constants.AckStatusACK, nil
	}

	//TODO: check if message is invalid

	in, err := s.DB.FetchIncludedIn(hash)
	if err != nil {
		return 0, err
	}

	if in == nil {
		if s.IsStateFullySynced() {
			return constants.AckStatusNotConfirmed, nil
		} else {
			return constants.AckStatusUnknown, nil
		}
	}

	return constants.AckStatusDBlockConfirmed, nil
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

	return dbase.FetchECTransactionByHash(hash)
}

func (s *State) FetchFactoidTransactionByHash(hash interfaces.IHash) (interfaces.ITransaction, error) {
	//TODO: expand to search data from outside database
	if hash == nil {
		return nil, nil
	}
	/*
		fBlock := s.ProcessLists.LastList().FactoidCreditBlock
		if fBlock != nil {
			tx := fBlock.GetTransactionByHash(hash)
			if tx != nil {
				return tx, nil
			}
		}
	*/

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchFactoidTransactionByHash(hash)
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

	for _, tx := range s.ProcessLists.LastList().NewEntries {
		if hash.IsSameAs(tx.GetHash()) {
			return tx, nil
		}
	}

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchEntryByHash(hash)
}
