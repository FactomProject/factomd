package databaseOverlay

import (
	//	"errors"
	. "github.com/FactomProject/factomd/common/EntryCreditBlock"
	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/interfaces"
)

// ProcessECBlockBatche inserts the ECBlock and update all it's cbentries in DB
func (db *Overlay) ProcessECBlockBatch(block *ECBlock) error {
	if block == nil {
		return nil
	}

	batch := []Record{}

	hash, err := block.HeaderHash()
	if err != nil {
		return err
	}
	batch = append(batch, Record{[]byte{byte(TBL_CB)}, hash.Bytes(), block})

	hash, err = block.HeaderHash()
	if err != nil {
		return err
	}
	batch = append(batch, Record{[]byte{byte(TBL_CHAIN_HEAD)}, EC_CHAINID, hash})

	err = db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

// FetchECBlockByHash gets an Entry Credit block by hash from the database.
func (db *Overlay) FetchECBlockByHash(ecBlockHash IHash) (*ECBlock, error) {
	bucket := []byte{byte(TBL_CB)}
	key := ecBlockHash.Bytes()

	block, err := db.DB.Get(bucket, key, new(ECBlock))
	if err != nil {
		return nil, err
	}
	return block.(*ECBlock), nil
}

// FetchAllECBlocks gets all of the entry credit blocks
func (db *Overlay) FetchAllECBlocks() ([]*ECBlock, error) {
	bucket := []byte{byte(TBL_CB)}

	list, err := db.DB.GetAll(bucket, new(ECBlock))
	if err != nil {
		return nil, err
	}
	answer := make([]*ECBlock, len(list))
	for i, v := range list {
		answer[i] = v.(*ECBlock)
	}
	return answer, nil
}
