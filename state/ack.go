// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

func (s *State) IsStateFullySynced() bool {
	return s.ProcessLists.DBHeightBase == uint32(len(s.ProcessLists.Lists))
}

func (s *State) GetACKStatus(hash interfaces.IHash) (int, interfaces.Timestamp, error) {
	m, found := s.ProcessLists.LastList().OldMsgs[hash.Fixed()]
	if found {
		return constants.AckStatusACK, m.GetTimestamp(), nil
	}

	for _, tx := range s.ProcessLists.LastList().NewEntries {
		if hash.IsSameAs(tx.GetHash()) {
			return constants.AckStatusACK, s.ProcessLists.LastList().DirectoryBlock.GetHeader().GetTimestamp(), nil
		}
	}
	ecBlock := s.ProcessLists.LastList().EntryCreditBlock
	if ecBlock != nil {
		tx := ecBlock.GetEntryByHash(hash)
		if tx != nil {
			return constants.AckStatusACK, s.ProcessLists.LastList().DirectoryBlock.GetHeader().GetTimestamp(), nil
		}
	}
	fBlock := s.FactoidState.GetCurrentBlock()
	if fBlock != nil {
		tx := fBlock.GetTransactionByHash(hash)
		if tx != nil {
			return constants.AckStatusACK, tx.GetTimestamp(), nil
		}
	}

	msg := s.GetInvalidMsg(hash)
	if msg != nil {
		return constants.AckStatusInvalid, primitives.NewTimestampFromSeconds(0), nil
	}

	in, err := s.DB.FetchIncludedIn(hash)
	if err != nil {
		return 0, primitives.NewTimestampFromSeconds(0), err
	}

	if in == nil {
		if s.IsStateFullySynced() {
			return constants.AckStatusNotConfirmed, primitives.NewTimestampFromSeconds(0), nil
		} else {
			return constants.AckStatusUnknown, primitives.NewTimestampFromSeconds(0), nil
		}
	}

	in2, err := s.DB.FetchIncludedIn(in)
	if err != nil {
		return 0, primitives.NewTimestampFromSeconds(0), err
	}

	dBlock, err := s.DB.FetchDBlock(in2)
	if err != nil {
		return 0, primitives.NewTimestampFromSeconds(0), err
	}

	return constants.AckStatusDBlockConfirmed, dBlock.GetHeader().GetTimestamp(), nil
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

	for _, tx := range s.ProcessLists.LastList().NewEntries {
		if hash.IsSameAs(tx.GetHash()) {
			return tx, nil
		}
	}

	dbase := s.GetAndLockDB()
	defer s.UnlockDB()

	return dbase.FetchEntry(hash)
}
