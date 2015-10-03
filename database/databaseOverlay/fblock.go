package databaseOverlay

import (
	//. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid/block"
	. "github.com/FactomProject/factomd/common/interfaces"
)

/*
// ProcessFBlockBatch inserts the factoid block
func (db *Overlay) ProcessFBlockBatch(block IFBlock) error {

	if block != nil {
		if db.lbatch == nil {
			db.lbatch = new(leveldb.Batch)
		}

		defer db.lbatch.Reset()

		binaryBlock, err := block.MarshalBinary()
		if err != nil {
			return err
		}

		scHash := block.GetHash()

		// Insert the binary factom block
		var key []byte = []byte{byte(TBL_SC)}
		key = append(key, scHash.Bytes()...)
		db.lbatch.Put(key, binaryBlock)

		// Insert the sc block number cross reference
		key = []byte{byte(TBL_SC_NUM)}
		key = append(key, block.GetChainID().Bytes()...)
		bytes := make([]byte, 4)
		binary.BigEndian.PutUint32(bytes, block.GetDBHeight())
		key = append(key, bytes...)
		db.lbatch.Put(key, scHash.Bytes())

		// Update the chain head reference
		key = []byte{byte(TBL_CHAIN_HEAD)}
		key = append(key, FACTOID_CHAINID...)
		db.lbatch.Put(key, scHash.Bytes())

		err = db.lDb.Write(db.lbatch, db.wo)
		if err != nil {
			log.Println("batch failed %v\n", err)
			return err
		}

	}
	return nil
}

// FetchFBlockByHash gets an factoid block by hash from the database.
func (db *Overlay) FetchFBlockByHash(hash IHash) (FBlock IFBlock, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_SC)}
	key = append(key, hash.Bytes()...)
	data, err := db.lDb.Get(key, db.ro)

	if data != nil {
		FBlock = new(block.FBlock)
		_, err := FBlock.UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}
	return FBlock, nil
}*/

// FetchAllFBlocks gets all of the factoid blocks
func (db *Overlay) FetchAllFBlocks() (FBlocks []IFBlock, err error) {
	bucket := []byte{byte(TBL_SC)}

	list, err:=db.DB.GetAll(bucket, new(FBlock))
	if err!=nil {
		return nil, err
	}
	answer:=make([]IFBlock, len(list))
	for i, v:=range(list) {
		answer[i] = v.(*FBlock)
	}
	return answer, nil
}
