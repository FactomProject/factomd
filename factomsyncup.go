package btcd

import (
	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/FactomCode/database"
	"github.com/FactomProject/btcd/wire"
	"strconv"
	"time"
)

// Validate the new blocks in mem pool and store them in db
func validateAndStoreBlocks(fMemPool *ftmMemPool, db database.Db, dchain *common.DChain, outCtlMsgQ chan wire.FtmInternalMsg) {
	var myDBHeight int64
	var sleeptime int
	var dblk *common.DirectoryBlock

	for true {

		_, myDBHeight, _ = db.FetchBlockHeightCache()

		sleeptime = 1 // need a formula??

		if len(dchain.Blocks) > int(myDBHeight+1) {
			dblk = dchain.Blocks[myDBHeight+1]
		}
		if dblk != nil {
			if validateBlocksFromMemPool(dblk, fMemPool, db) {
				err := storeBlocksFromMemPool(dblk, fMemPool, db)
				if err == nil {
					deleteBlocksFromMemPool(dblk, fMemPool)
				} else {
					panic("error in deleteBlocksFromMemPool.")
				}
			}
		} else {
			//send an internal msg to sync up with peers
		}

		time.Sleep(time.Duration(sleeptime * 1000000000))
	}

}

// Validate the new blocks in mem pool and store them in db
func validateBlocksFromMemPool(b *common.DirectoryBlock, fMemPool *ftmMemPool, db database.Db) bool {

	for _, dbEntry := range b.DBEntries {
		switch dbEntry.ChainID.String() {
		case ecchain.ChainID.String():
			if _, ok := fMemPool.blockpool[dbEntry.MerkleRoot.String()]; !ok {
				return false
			}
		case achain.ChainID.String():
			if _, ok := fMemPool.blockpool[dbEntry.MerkleRoot.String()]; !ok {
				return false
			}
		case scchain.ChainID.String():
			if _, ok := fMemPool.blockpool[dbEntry.MerkleRoot.String()]; !ok {
				return false
			}
		default:
			if msg, ok := fMemPool.blockpool[dbEntry.MerkleRoot.String()]; !ok {
				return false
			} else {
				eBlkMsg, _ := msg.(*wire.MsgEBlock)
				// validate every entry in EBlock
				for _, ebEntry := range eBlkMsg.EBlk.EBEntries {
					if _, foundInMemPool := fMemPool.blockpool[ebEntry.EntryHash.String()]; !foundInMemPool {
						// continue if the entry arleady exists in db
						entry, _ := db.FetchEntryByHash(ebEntry.EntryHash)
						if entry == nil {
							return false
						}
					}
				}
			}
		}
	}

	return true
}

// Validate the new blocks in mem pool and store them in db
func storeBlocksFromMemPool(b *common.DirectoryBlock, fMemPool *ftmMemPool, db database.Db) error {

	for _, dbEntry := range b.DBEntries {
		switch dbEntry.ChainID.String() {
		case ecchain.ChainID.String():
			ecBlkMsg := fMemPool.blockpool[dbEntry.MerkleRoot.String()].(*wire.MsgECBlock)
			err := db.ProcessECBlockBatch(ecBlkMsg.ECBlock)
			if err != nil {
				return err
			}
		case achain.ChainID.String():
			aBlkMsg := fMemPool.blockpool[dbEntry.MerkleRoot.String()].(*wire.MsgABlock)
			err := db.ProcessABlockBatch(aBlkMsg.ABlk)
			if err != nil {
				return err
			}
		case scchain.ChainID.String():
			fBlkMsg := fMemPool.blockpool[dbEntry.MerkleRoot.String()].(*wire.MsgFBlock)
			err := db.ProcessFBlockBatch(fBlkMsg.SC)
			if err != nil {
				return err
			}
		default:
			eBlkMsg, _ := fMemPool.blockpool[dbEntry.MerkleRoot.String()].(*wire.MsgEBlock)
			for _, ebEntry := range eBlkMsg.EBlk.EBEntries {
				if msg, foundInMemPool := fMemPool.blockpool[ebEntry.EntryHash.String()]; foundInMemPool {
					err := db.InsertEntry(ebEntry.EntryHash, msg.(*wire.MsgEntry).Entry)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// Validate the new blocks in mem pool and store them in db
func deleteBlocksFromMemPool(b *common.DirectoryBlock, fMemPool *ftmMemPool) error {

	for _, dbEntry := range b.DBEntries {
		switch dbEntry.ChainID.String() {
		case ecchain.ChainID.String():
			delete(fMemPool.blockpool, dbEntry.MerkleRoot.String())
		case achain.ChainID.String():
			delete(fMemPool.blockpool, dbEntry.MerkleRoot.String())
		case scchain.ChainID.String():
			delete(fMemPool.blockpool, dbEntry.MerkleRoot.String())
		default:
			eBlkMsg, _ := fMemPool.blockpool[dbEntry.MerkleRoot.String()].(*wire.MsgEBlock)
			for _, ebEntry := range eBlkMsg.EBlk.EBEntries {
				delete(fMemPool.blockpool, ebEntry.EntryHash.String())
			}
			delete(fMemPool.blockpool, dbEntry.MerkleRoot.String())
		}
	}
	delete(fMemPool.blockpool, strconv.Itoa(int(b.Header.BlockHeight)))

	return nil
}
