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

// GetEntryCommitAckByTXID will fetch the status of a commit by TxID
//	Searches this order:
//		Database	--> Check if it made it to blockchain
//		PL			-->	See if it is still in the processlist
//		PL - 1		-->	Only if min 0, because then it's not in DB yet, but still in this PL
//		Holding 	--> See if it is in holding
func (s *State) GetEntryCommitAckByTXID(hash interfaces.IHash) (status int, blktime interfaces.Timestamp, commit interfaces.IMsg, entryhash interfaces.IHash) {
	status = constants.AckStatusUnknown
	// Check Database for commit
	ecblkHash, err := s.DB.FetchIncludedIn(hash)
	if err == nil && ecblkHash != nil {
		// See if it is a commit txid. If not found, then it is not
		ecblk, err := s.DB.FetchECBlock(ecblkHash)
		status = constants.AckStatusDBlockConfirmed
		if err == nil && ecblk != nil {
			// Found the ECBlock. We can find the transaction
			for _, e := range ecblk.GetEntries() {
				if e.GetHash().IsSameAs(hash) {
					// Found the entryhash
					entryhash = e.GetEntryHash()
					break
				}
			}

			// Get the time
			dblkHash, err := s.DB.FetchIncludedIn(ecblkHash)
			if err == nil && dblkHash != nil {
				dblk, err := s.DB.FetchDBlock(dblkHash)
				if err == nil && dblk != nil {
					blktime = dblk.GetTimestamp()
				}
			}

			// Found in DB, so return
			return
		}
	}

	for i := uint32(0); i < 2; i++ {
		// Also have to check the prior PL if we are in minute 0
		if i == 1 && s.GetCurrentMinute() != 0 {
			continue
		}
		if s.PLProcessHeight < i {
			// If i == 1 and PLHeight is 0, we don't want uint32 underflow
			continue
		}
		// Not found in the DB, we can check the highest PL
		pl := s.ProcessLists.GetSafe(s.PLProcessHeight - i)
		if pl != nil {
			// We will search the processlist for the commit
			// TODO: Is this thread safe?
			for _, v := range pl.VMs {
				for i, m := range v.List {
					switch m.Type() {
					case constants.COMMIT_CHAIN_MSG:
						cc, ok := m.(*messages.CommitChainMsg)
						if ok {
							if cc.CommitChain.GetSigHash().IsSameAs(hash) {
								// Msg found in the latest processlist
								if i > v.Height { // if it has not yet been processed ...
									status = constants.AckStatusNotConfirmed
								} else {
									status = constants.AckStatusACK
								}
								commit = cc
								entryhash = cc.CommitChain.EntryHash
								return
							}
						}
					case constants.COMMIT_ENTRY_MSG:
						ce, ok := m.(*messages.CommitEntryMsg)
						if ok {
							if ce.CommitEntry.GetSigHash().IsSameAs(hash) {
								// Msg found in the latest processlist
								if i > v.Height { // if it has not yet been processed ...
									status = constants.AckStatusNotConfirmed
								} else {
									status = constants.AckStatusACK
								}
								commit = ce
								entryhash = ce.CommitEntry.EntryHash
								return
							}
						}
					}
				}
			}
		}
	}

	// If it was found in the PL or DBlock, it would return. All that is left is the holding map
	_, commit = s.FetchEntryRevealAndCommitFromHolding(hash)
	if commit != nil { // Found in holding
		switch commit.Type() {
		case constants.COMMIT_CHAIN_MSG:
			cc, ok := commit.(*messages.CommitChainMsg)
			if ok {
				entryhash = cc.CommitChain.EntryHash
				status = constants.AckStatusNotConfirmed
			}
		case constants.COMMIT_ENTRY_MSG:
			ce, ok := commit.(*messages.CommitEntryMsg)
			if ok {
				entryhash = ce.CommitEntry.EntryHash
				status = constants.AckStatusNotConfirmed
			}
		}
	}
	return
}

// GetEntryCommitAck will fetch the status of a entrycommit by ENTRYHASH. The places it checks are:
//		CommitMap 	--> This indicates if the entry made it into the processlist within the last 4 hrs
//		Last PL		--> Check if still in a processList
//		Holding 	--> See if it is in holding
//
//	Returns:
//		status 		= Status of reveal from possible ack responses
//		commit 		= The commit message
func (s *State) GetEntryCommitAckByEntryHash(hash interfaces.IHash) (status int, commit interfaces.IMsg) {
	// We begin as unknown
	status = constants.AckStatusUnknown

	// Check if the commit is in the the latest processlist
	for i := uint32(0); i < 2; i++ {
		// Also have to check the prior PL if we are in minute 0
		if i == 1 && s.GetCurrentMinute() != 0 {
			continue
		}
		if s.PLProcessHeight < i {
			// If i == 1 and PLHeight is 0, we don't want uint32 underflow
			continue
		}
		pl := s.ProcessLists.GetSafe(s.PLProcessHeight - i)
		if pl != nil {
			// We will search the processlist for the commit
			// TODO: Is this thread safe?
			for _, v := range pl.VMs {
				for _, m := range v.List {
					switch m.Type() {
					case constants.COMMIT_CHAIN_MSG:
						cc, ok := m.(*messages.CommitChainMsg)
						if ok {
							if cc.CommitChain.EntryHash.IsSameAs(hash) {
								// Msg found in the latest processlist
								status = constants.AckStatusACK
								commit = cc
								return
							}
						}
					case constants.COMMIT_ENTRY_MSG:
						ce, ok := m.(*messages.CommitEntryMsg)
						if ok {
							if ce.CommitEntry.EntryHash.IsSameAs(hash) {
								// Msg found in the latest processlist
								status = constants.AckStatusACK
								commit = ce
								return
							}
						}
					}
				}
			}
		}
	}

	// If found in commit map, and not in PL, it is in the DB
	c := s.Commits.Get(hash.Fixed())
	if c != nil {
		// The commit was found and valid. We will change this to AckStatus if found
		// in the latest processlist
		status = constants.AckStatusDBlockConfirmed
		commit = c
		return
	}

	// At this point, we have the status of unknown. Any DBlock or Ack level has been covered above.
	// If 'c' is not nil, then commit was found in the holding map.
	_, c = s.FetchEntryRevealAndCommitFromHolding(hash)
	if c != nil {
		status = constants.AckStatusNotConfirmed
		commit = c
	}

	return
}

// GetEntryRevealAck will fetch the status of a entryreveal. The places it checks are:
//		ReplayMap	--> This indicates if the entry made it into the processlist within the last 4 hrs
//		Database	-->	Check if it made it to blockchain
//		Holding		-->	See if it is in holding. Will also look for commit if it finds that
//
//	Returns:
//		status 		= Status of reveal from possible ack responses
// 		blktime		= The time of the block if found in the database, nil if not found in blockchain
//		commit		 = Only returned if found from holding. This will be empty if found in dbase or in processlist
func (s *State) GetEntryRevealAckByEntryHash(hash interfaces.IHash) (status int, blktime interfaces.Timestamp, commit interfaces.IMsg) {
	// We begin as unknown
	status = constants.AckStatusUnknown

	// Fetch the EBlock
	eblk, err := s.DB.FetchIncludedIn(hash)
	if err == nil && eblk != nil {
		// Ensure that it was found in an eblock, not an ecblock.
		if eblock, err := s.DB.FetchEBlock(eblk); eblock != nil && err == nil {
			// This means the entry was found in the database
			status = constants.AckStatusDBlockConfirmed
			dblk, err := s.DB.FetchIncludedIn(eblk)
			if err != nil {
				return
			}

			dBlock, err := s.DB.FetchDBlock(dblk)
			if err != nil {
				return
			}
			blktime = dBlock.GetTimestamp()

			// Exit as it was found in the blockchain, and we have the time
			return
		}
	}

	// Not found in the database. Check if it was found in the processlist
	if !s.Replay.IsHashUnique(constants.REVEAL_REPLAY, hash.Fixed()) {
		// It has been in the PL, so return
		status = constants.AckStatusACK
		return
	}

	// Not found in the database or the processlist. We can still check holding.
	// If 'r' is not nil, then reveal was found in the holding map. Also return the
	// commit msg if we had to look this far, it could save someone else calling us a lookup
	r, c := s.FetchEntryRevealAndCommitFromHolding(hash)
	if r != nil {
		status = constants.AckStatusNotConfirmed
		if c != nil {
			commit = c
		}
	}

	return
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
				aMsg := pl.GetOldAck(hash)
				if aMsg != nil { // No ack then it's not "known"
					a, ok := aMsg.(*messages.Ack)
					if !ok {
						// probably deserves a panic here if we got an old ack and it wasn't an ack
						return constants.AckStatusUnknown, hash, nil, nil, nil
					}
					if pl.VMs[a.GetVMIndex()].Height < int(a.Height) {
						// if it is in the process list but has not yet been process then claim it's unknown
						// Otherwise it might get an ack status but still be un-spendable
						return constants.AckStatusNotConfirmed, hash, nil, nil, nil
					} else {
						return constants.AckStatusACK, hash, a.GetTimestamp(), nil, nil
					}
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

// FetchEntryRevealAndCommitFromHolding will look for the commit and reveal for a given hash.
// It will check the hash as an entryhash and a txid, and return any reveals that match the entryhash
// and any commits that match the entryhash or txid
//
//		Returns
//			reveal = The reveal message if found
//			commit = The commit message if found
func (s *State) FetchEntryRevealAndCommitFromHolding(hash interfaces.IHash) (reveal interfaces.IMsg, commit interfaces.IMsg) {
	q := s.LoadHoldingMap()
	for _, h := range q {
		switch {
		case h.Type() == constants.COMMIT_CHAIN_MSG:
			cm, ok := h.(*messages.CommitChainMsg)
			if ok {
				if cm.CommitChain.EntryHash.IsSameAs(hash) {
					commit = cm
				}

				if hash.IsSameAs(cm.CommitChain.GetSigHash()) {
					commit = cm
				}
			}
		case h.Type() == constants.COMMIT_ENTRY_MSG:
			cm, ok := h.(*messages.CommitEntryMsg)
			if ok {
				if cm.CommitEntry.EntryHash.IsSameAs(hash) {
					commit = cm
				}

				if hash.IsSameAs(cm.CommitEntry.GetSigHash()) {
					commit = cm
				}
			}
		case h.Type() == constants.REVEAL_ENTRY_MSG:
			rm, ok := h.(*messages.RevealEntryMsg)
			if ok {
				if rm.Entry.GetHash().IsSameAs(hash) {
					reveal = rm
				}
			}
		}
	}
	return
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

	dbase := s.GetDB()

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

	dbase := s.GetDB()

	return dbase.FetchFactoidTransaction(hash)
}

func (s *State) FetchPaidFor(hash interfaces.IHash) (interfaces.IHash, error) {
	//TODO: expand to search data from outside database
	if hash == nil {
		return nil, nil
	}

	for _, pls := range s.ProcessLists.Lists {
		if pls == nil {
			continue
		}
		ecBlock := pls.EntryCreditBlock
		for _, tx := range ecBlock.GetEntries() {
			switch tx.ECID() {
			case constants.ECIDEntryCommit:
				if hash.IsSameAs(tx.(*entryCreditBlock.CommitEntry).EntryHash) {
					return tx.GetSigHash(), nil
				}
				break
			case constants.ECIDChainCommit:
				if hash.IsSameAs(tx.(*entryCreditBlock.CommitChain).EntryHash) {
					return tx.GetSigHash(), nil
				}
				break
			}
		}
	}
	dbase := s.GetDB()

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
			if hash.IsSameAs(tx.GetHash()) {
				return tx, nil
			}
		}
	}

	dbase := s.GetDB()

	return dbase.FetchEntry(hash)
}
