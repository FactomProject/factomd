// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package btcd

import (
	"bytes"
	"encoding/hex"
	//"errors"
	cp "github.com/FactomProject/factomd/controlpanel"
	"github.com/davecgh/go-spew/spew"
	"strconv"
	"time"

	//"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/directoryBlock"
	//"github.com/FactomProject/factomd/common/entryBlock"
	//"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	//"github.com/FactomProject/factomd/common/primitives"
)

// processDirBlock validates dir block and save it to factom b.server.State.GetDB().
// similar to blockChain.BC_ProcessBlock
func (b *blockManager) processDirBlock(msg *messages.MsgDirBlock) error {
	blk, _ := b.server.State.GetDB().FetchDBlockByHeight(msg.DBlk.GetHeader().GetDBHeight())
	if blk != nil {
		bmgrLog.Info("DBlock already exists for height:" + string(msg.DBlk.GetHeader().GetDBHeight()))
		cp.CP.AddUpdate(
			"DBOverlap",                                                                    // tag
			"warning",                                                                      // Category
			"Directory Block Overlap",                                                      // Title
			"DBlock already exists for height:"+string(msg.DBlk.GetHeader().GetDBHeight()), // Message
			0) // Expire
		return nil
	}

	//Add it to mem pool before saving it in db
	b.fMemPool.addBlockMsg(msg, strconv.Itoa(int(msg.DBlk.GetHeader().GetDBHeight()))) // store in mempool with the height as the key

	bmgrLog.Debug("SyncUp: MsgDirBlock DBHeight=", msg.DBlk.GetHeader().GetDBHeight())
	cp.CP.AddUpdate(
		"DBSyncUp", // tag
		"Status",   // Category
		"SyncUp:",  // Title
		"MsgDirBlock DBHeigth=:"+string(msg.DBlk.GetHeader().GetDBHeight()), // Message
		0) // Expire

	return nil
}

// processFBlock validates admin block and save it to factom b.server.State.GetDB().
// similar to blockChain.BC_ProcessBlock
func (b *blockManager) processFBlock(msg *messages.MsgFBlock) error {
	key := hex.EncodeToString(msg.FBlck.GetHash().Bytes())
	//Add it to mem pool before saving it in db
	b.fMemPool.addBlockMsg(msg, string(key)) // stored in mem pool with the MR as the key
	bmgrLog.Debug("SyncUp: MsgFBlock DBHeight=", msg.FBlck.GetDBHeight())
	return nil
}

// processABlock validates admin block and save it to factom b.server.State.GetDB().
// similar to blockChain.BC_ProcessBlock
func (b *blockManager) processABlock(msg *messages.MsgABlock) error {
	abHash, err := msg.ABlk.PartialHash()
	if err != nil {
		return err
	}
	b.fMemPool.addBlockMsg(msg, abHash.String()) // store in mem pool with ABHash as key
	bmgrLog.Debug("SyncUp: MsgABlock DBHeight=", msg.ABlk.GetHeader().GetDBHeight())
	return nil
}

// procesFBlock validates entry credit block and save it to factom b.server.State.GetDB().
// similar to blockChain.BC_ProcessBlock
func (b *blockManager) processECBlock(msg *messages.MsgECBlock) error {
	hash, err := msg.ECBlock.HeaderHash()
	if err != nil {
		return err
	}
	b.fMemPool.addBlockMsg(msg, hash.String())
	bmgrLog.Debug("SyncUp: MsgCBlock DBHeight=", msg.ECBlock.GetHeader().GetDBHeight())
	return nil
}

// processEBlock validates entry block and save it to factom b.server.State.GetDB().
// similar to blockChain.BC_ProcessBlock
func (b *blockManager) processEBlock(msg *messages.MsgEBlock) error {
	keyMR, err := msg.EBlk.KeyMR()
	if err != nil {
		return err
	}
	b.fMemPool.addBlockMsg(msg, keyMR.String()) // store it in mem pool with MR as the key
	bmgrLog.Debug("SyncUp: MsgEBlock DBHeight=", msg.EBlk.GetHeader().GetDBHeight())
	return nil
}

// processEntry validates entry and save it to factom b.server.State.GetDB().
// similar to blockChain.BC_ProcessBlock
func (b *blockManager) processEntry(msg *messages.MsgEntry) error {
	h := msg.Entry.GetHash()
	b.fMemPool.addBlockMsg(msg, h.String()) // store it in mem pool with hash as the key
	bmgrLog.Debug("SyncUp: MsgEntry hash=", msg.Entry.GetHash())
	return nil
}

// Validate the new blocks in mem pool and store them in db
func (b *blockManager) validateAndStoreBlocks() {
	var myDBHeight uint32
	var sleeptime = 1
	var dblk *directoryBlock.DirectoryBlock

	for true {
		dblk = nil
		myDBHeight = b.server.State.GetDBHeight()
		msg := b.fMemPool.getBlockMsg(string(myDBHeight + 1))
		if msg != nil {
			switch msg.(type) {
			case *messages.MsgDirBlock:
				dblk = msg.(*messages.MsgDirBlock).DBlk
				bmgrLog.Debug("SyncUp: validate height=%d, dirblock=%s\n ", myDBHeight+1, spew.Sdump(dblk))
				if dblk != nil {
					if b.validateBlocksFromMemPool(dblk) {
						err := b.storeBlocksFromMemPool(dblk)
						if err == nil {
							b.deleteBlocksFromMemPool(dblk)
						} else {
							panic("error in deleteBlocksFromMemPool.")
						}
					} else {
						time.Sleep(time.Duration(sleeptime * 1000000)) // Nanoseconds for duration
					}
				} else {
					//TODO: send an internal msg to sync up with peers
				}
			}
		} else {
			time.Sleep(time.Duration(sleeptime * 1000000)) // Nanoseconds for duration
		}
	}
}

// Validate the new blocks in mem pool and store them in db
func (b *blockManager) validateBlocksFromMemPool(dblk *directoryBlock.DirectoryBlock) bool {

	// Validate the genesis block
	if dblk.GetHeader().GetDBHeight() == 0 {
		h := dblk.GetHash() //CreateHash(b)
		if h.String() != constants.GENESIS_DIR_BLOCK_HASH {
			// panic for milestone 1
			panic("\nGenesis block hash expected: " + constants.GENESIS_DIR_BLOCK_HASH +
				"\nGenesis block hash found:    " + h.String() + "\n")
			//bmgrLog.Errorf("Genesis dir block is not as expected: " + h.String())
		}
	}

	for _, dbEntry := range dblk.GetDBEntries() {
		switch dbEntry.GetChainID().String() {
		case hex.EncodeToString(constants.EC_CHAINID[:]):
			if _, ok := b.fMemPool.blockpool[dbEntry.GetKeyMR().String()]; !ok {
				return false
			}
		case hex.EncodeToString(constants.ADMIN_CHAINID[:]):
			if _, ok := b.fMemPool.blockpool[dbEntry.GetKeyMR().String()]; !ok {
				return false
				//} else {
				// validate signature of the previous dir block
				//aBlkMsg, _ := msg.(*messages.MsgABlock)
				//if !b.validateDBSignature(aBlkMsg.ABlk) {
				//return false
				//}
			}
		case hex.EncodeToString(constants.FACTOID_CHAINID[:]):
			if _, ok := b.fMemPool.blockpool[dbEntry.GetKeyMR().String()]; !ok {
				return false
			}
		default:
			if msg, ok := b.fMemPool.blockpool[dbEntry.GetKeyMR().String()]; !ok {
				return false
			} else {
				eBlkMsg, _ := msg.(*messages.MsgEBlock)
				// validate every entry in EBlock
				for _, ebEntry := range eBlkMsg.EBlk.Body.EBEntries {
					if _, foundInMemPool := b.fMemPool.blockpool[ebEntry.String()]; !foundInMemPool {
						if !bytes.Equal(ebEntry.Bytes()[:31], constants.ZERO_HASH[:31]) {
							// continue if the entry arleady exists in db
							entry, _ := b.server.State.GetDB().FetchEntryByHash(ebEntry)
							if entry == nil {
								return false
							}
						}
					}
				}
			}
		}
	}
	return true
}

// Validate the new blocks in mem pool and store them in db
// Need to make a batch insert in db in milestone 2
func (b *blockManager) storeBlocksFromMemPool(dblk *directoryBlock.DirectoryBlock) error {

	for _, dbEntry := range dblk.DBEntries {
		switch dbEntry.GetChainID().String() {
		case hex.EncodeToString(constants.EC_CHAINID[:]):
			ecBlkMsg := b.fMemPool.blockpool[dbEntry.GetKeyMR().String()].(*messages.MsgECBlock)
			err := b.server.State.GetDB().ProcessECBlockBatch(ecBlkMsg.ECBlock)
			if err != nil {
				return err
			}
			b.server.State.SetEntryCreditBlock(b.server.State.GetDBHeight(), ecBlkMsg.ECBlock)
			//initializeECreditMap(ecBlkMsg.ECBlock)
		case hex.EncodeToString(constants.ADMIN_CHAINID[:]):
			aBlkMsg := b.fMemPool.blockpool[dbEntry.GetKeyMR().String()].(*messages.MsgABlock)
			err := b.server.State.GetDB().ProcessABlockBatch(aBlkMsg.ABlk)
			if err != nil {
				return err
			}
			b.server.State.SetAdminBlock(b.server.State.GetDBHeight(), aBlkMsg.ABlk)
		case hex.EncodeToString(constants.FACTOID_CHAINID[:]):
			fBlkMsg := b.fMemPool.blockpool[dbEntry.GetKeyMR().String()].(*messages.MsgFBlock)
			err := b.server.State.GetDB().ProcessFBlockBatch(fBlkMsg.FBlck)
			if err != nil {
				return err
			}
			// Initialize the Factoid State
			err = b.server.State.GetFactoidState(b.server.State.GetDBHeight()).AddTransactionBlock(fBlkMsg.FBlck)
			//FactoshisPerCredit = fBlkMsg.FBlck.GetExchRate()
			if err != nil {
				return err
			}
//			b.server.State.SetFactoidKeyMR(b.server.State.GetDBHeight(), fBlkMsg.FBlck.GetKeyMR()) // ???
		default:
			// handle Entry Block
			eBlkMsg, _ := b.fMemPool.blockpool[dbEntry.GetKeyMR().String()].(*messages.MsgEBlock)
			// store entry in db first
			for _, ebEntry := range eBlkMsg.EBlk.Body.EBEntries {
				if msg, foundInMemPool := b.fMemPool.blockpool[ebEntry.String()]; foundInMemPool {
					err := b.server.State.GetDB().InsertEntry(msg.(*messages.MsgEntry).Entry)
					if err != nil {
						return err
					}
				}
			}
			// Store Entry Block in db
			err := b.server.State.GetDB().ProcessEBlockBatch(eBlkMsg.EBlk)
			if err != nil {
				return err
			}
			/*
				// create a chain in db if it's not existing
				chain := chainIDMap[eBlkMsg.EBlk.Header.ChainID.String()]
				if chain == nil {
					chain = new(EChain)
					chain.ChainID = eBlkMsg.EBlk.Header.ChainID
					if eBlkMsg.EBlk.Header.EBSequence == 0 {
						chain.FirstEntry, _ = b.server.State.GetDB().FetchEntryByHash(eBlkMsg.EBlk.Body.EBEntries[0])
					}
					b.server.State.GetDB().InsertChain(chain)
					chainIDMap[chain.ChainID.String()] = chain
				} else if chain.FirstEntry == nil && eBlkMsg.EBlk.Header.EBSequence == 0 {
					chain.FirstEntry, _ = b.server.State.GetDB().FetchEntryByHash(eBlkMsg.EBlk.Body.EBEntries[0])
					b.server.State.GetDB().InsertChain(chain)
				}*/
		}
	}

	// Store the dir block
	err := b.server.State.GetDB().ProcessDBlockBatch(dblk)
	if err != nil {
		return err
	}

	// Update State with block height & current/previous blocks
	b.server.State.SetDirectoryBlock(b.server.State.GetDBHeight(), dblk)
	return nil
}

// Validate the new blocks in mem pool and store them in db
func (b *blockManager) deleteBlocksFromMemPool(dblk *directoryBlock.DirectoryBlock) error {

	for _, dbEntry := range dblk.GetDBEntries() {
		switch dbEntry.GetChainID().String() {
		case hex.EncodeToString(constants.EC_CHAINID[:]):
			b.fMemPool.deleteBlockMsg(dbEntry.GetKeyMR().String())
		case hex.EncodeToString(constants.ADMIN_CHAINID[:]):
			b.fMemPool.deleteBlockMsg(dbEntry.GetKeyMR().String())
		case hex.EncodeToString(constants.FACTOID_CHAINID[:]):
			b.fMemPool.deleteBlockMsg(dbEntry.GetKeyMR().String())
		default:
			eBlkMsg, _ := b.fMemPool.blockpool[dbEntry.GetKeyMR().String()].(*messages.MsgEBlock)
			for _, ebEntry := range eBlkMsg.EBlk.Body.EBEntries {
				b.fMemPool.deleteBlockMsg(ebEntry.String())
			}
			b.fMemPool.deleteBlockMsg(dbEntry.GetKeyMR().String())
		}
	}
	b.fMemPool.deleteBlockMsg(strconv.Itoa(int(dblk.Header.GetDBHeight())))
	return nil
}

/*
func (b *blockManager) validateDBSignature(aBlock *adminBlock.AdminBlock) bool {

	dbSigEntry := aBlock.GetDBSignature()
	if dbSigEntry == nil {
		if aBlock.Header.GetDBHeight() == 0 {
			return true
		} else {
			return false
		}
	} else {
		dbSig := dbSigEntry.(*DBSignatureEntry)
		if serverPubKey.String() != dbSig.PubKey.String() {
			return false
		} else {
			// obtain the previous directory block
			dblk := dchain.Blocks[aBlock.Header.DBHeight-1]
			if dblk == nil {
				return false
			} else {
				// validatet the signature
				bHeader, _ := dblk.Header.MarshalBinary()
				if !serverPubKey.Verify(bHeader, (*[64]byte)(dbSig.PrevDBSig)) {
					bmgrLog.Infof("No valid signature found in Admin Block = %s\n", spew.Sdump(aBlock))
					return false
				}
			}
		}
	}
	return true
}
*/
