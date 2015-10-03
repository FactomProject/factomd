package databaseOverlay

import (
	"encoding/binary"
	"errors"
	. "github.com/FactomProject/factomd/common/EntryBlock"
	. "github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/database/bytestore"


)

// ProcessEBlockBatche inserts the EBlock and update all it's ebentries in DB
func (db *Overlay) ProcessEBlockBatch(eblock *EBlock) error {
	if eblock == nil {
		return nil
	}

	if len(eblock.Body.EBEntries) < 1 {
		return errors.New("Empty eblock!")
	}
	
	batch:=[]Record{}

	// Insert the binary entry block
	bucket := []byte{byte(TBL_EB)}
	hash, err := eblock.Hash()
	if err != nil {
		return err
	}
	key := hash.Bytes()

	batch = append(batch, Record{bucket, key, eblock})

	// Insert the entry block merkle root cross reference
	bucket = []byte{byte(TBL_EB_MR)}
	keyMR, err := eblock.KeyMR()
	if err != nil {
		return err
	}
	key = keyMR.Bytes()

	eBlockHash, err := eblock.Hash()
	if err != nil {
		return err
	}
	batch = append(batch, Record{bucket, key, eBlockHash})


	// Insert the entry block number cross reference
	bucket = []byte{byte(TBL_EB_CHAIN_NUM)}
	key = eblock.Header.ChainID.Bytes()

	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, eblock.Header.EBSequence)
	
	bs:=bytestore.NewByteStore(bytes)

	batch = append(batch, Record{bucket, key, bs})

	// Update the chain head reference
	bucket = []byte{byte(TBL_CHAIN_HEAD)}
	key = eblock.Header.ChainID.Bytes()
	keyMR, err = eblock.KeyMR()
	if err != nil {
		return err
	}
	batch = append(batch, Record{bucket, key, keyMR})

	err = db.DB.PutInBatch(batch)
	if err!=nil {
		return err
	}
	
	return nil
}
/*
// FetchEBlockByMR gets an entry block by merkle root from the database.
func (db *Overlay) FetchEBlockByMR(eBMR IHash) (eBlock *EBlock, err error) {
	eBlockHash, err := db.FetchEBHashByMR(eBMR)
	if err != nil {
		return nil, err
	}

	if eBlockHash != nil {
		eBlock, err = db.FetchEBlockByHash(eBlockHash)
		if err != nil {
			return nil, err
		}
	}

	return eBlock, nil
}

// FetchEntryBlock gets an entry by hash from the database.
func (db *Overlay) FetchEBlockByHash(eBlockHash IHash) (*EBlock, error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_EB)}
	key = append(key, eBlockHash.Bytes()...)
	data, err := db.lDb.Get(key, db.ro)
	if err != nil {
		return nil, err
	}

	eBlock := NewEBlock()
	if data != nil {
		_, err := eBlock.UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}
	return eBlock, nil
}

// FetchEBlockByHeight gets an entry block by height from the database.
// Need to rewrite since only the cross ref is stored in db ??
/*func (db *Overlay) FetchEBlockByHeight(chainID IHash, eBlockHeight uint32) (*EBlock, error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_EB_CHAIN_NUM)}
	key = append(key, chainID.Bytes...)
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, eBlockHeight)
	key = append(key, bytes...)
	data, err := db.lDb.Get(key, db.ro)
	if err != nil {
		return nil, err
	}

	eBlock := NewEBlock()
	if data != nil {
		_, err:=eBlock.UnmarshalBinaryData(data)
		if err!=nil {
			return nil, err
		}
	}
	return eBlock, nil
}
*//*

// FetchEBHashByMR gets an entry by hash from the database.
func (db *Overlay) FetchEBHashByMR(eBMR IHash) (IHash, error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_EB_MR)}
	key = append(key, eBMR.Bytes()...)
	data, err := db.lDb.Get(key, db.ro)
	if err != nil {
		return nil, err
	}

	eBlockHash := NewZeroHash()
	_, err = eBlockHash.UnmarshalBinaryData(data)
	if err != nil {
		return nil, err
	}

	return eBlockHash, nil
}*/

// InsertChain inserts the newly created chain into db
func (db *Overlay) InsertChain(chain *EChain) (error) {
	bucket := []byte{byte(TBL_CHAIN_HASH)}
	key :=chain.ChainID.Bytes()
	err := db.DB.Put(bucket, key, chain)
	if err != nil {
		return err
	}
	return nil
}
/*
// FetchChainByHash gets a chain by chainID
func (db *Overlay) FetchChainByHash(chainID IHash) (*EChain, error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var key []byte = []byte{byte(TBL_CHAIN_HASH)}
	key = append(key, chainID.Bytes()...)
	data, err := db.lDb.Get(key, db.ro)
	if err != nil {
		return nil, err
	}

	chain := NewEChain()
	if data != nil {
		_, err := chain.UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}
	return chain, nil
}

// FetchAllChains get all of the cahins
func (db *Overlay) FetchAllChains() (chains []*EChain, err error) {
	db.dbLock.Lock()
	defer db.dbLock.Unlock()

	var fromkey []byte = []byte{byte(TBL_CHAIN_HASH)}   // Table Name (1 bytes)
	var tokey []byte = []byte{byte(TBL_CHAIN_HASH + 1)} // Table Name (1 bytes)

	chainSlice := make([]*EChain, 0, 10)

	iter := db.lDb.NewIterator(&util.Range{Start: fromkey, Limit: tokey}, db.ro)
	for iter.Next() {
		chain := NewEChain()
		_, err := chain.UnmarshalBinaryData(iter.Value())
		if err != nil {
			return nil, err
		}
		chainSlice = append(chainSlice, chain)
	}
	iter.Release()
	err = iter.Error()

	return chainSlice, err
}
*/

// FetchAllEBlocksByChain gets all of the blocks by chain id
func (db *Overlay) FetchAllEBlocksByChain(chainID IHash) ([]*EBlock, error) {
	bucket := append([]byte{byte(TBL_EB_CHAIN_NUM)}, chainID.Bytes()...)

	list, err:=db.DB.GetAll(bucket, new(Hash))
	if err!=nil {
		return nil, err
	}
	bucket = []byte{byte(TBL_EB)}
	answer:=make([]*EBlock, len(list))
	for i, v:=range(list) {
		key := v.(*Hash).Bytes()

		data, err:=db.DB.Get(bucket, key, new(EBlock))
		if err!= nil {
			return nil, err
		}
		answer[i] = data.(*EBlock)
	}
	return answer, nil
}
