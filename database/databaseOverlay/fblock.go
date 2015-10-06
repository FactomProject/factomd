package databaseOverlay

import (
	"encoding/binary"
	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/factoid/block"
	. "github.com/FactomProject/factomd/common/interfaces"
)

// ProcessFBlockBatch inserts the factoid block
func (db *Overlay) ProcessFBlockBatch(block IFBlock) error {
	if block == nil {
		return nil
	}

	batch := []Record{}

	scHash := block.GetHash()
	batch = append(batch, Record{[]byte{byte(TBL_SC)}, scHash.Bytes(), block})

	// Insert the sc block number cross reference
	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, block.GetDBHeight())
	batch = append(batch, Record{[]byte{byte(TBL_SC_NUM)}, bytes, scHash})

	batch = append(batch, Record{[]byte{byte(TBL_CHAIN_HEAD)}, FACTOID_CHAINID, scHash})

	err := db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

// FetchFBlockByHash gets an factoid block by hash from the database.
func (db *Overlay) FetchFBlockByHash(hash IHash) (IFBlock, error) {
	bucket := []byte{byte(TBL_SC)}
	key := hash.Bytes()

	block, err := db.DB.Get(bucket, key, new(FBlock))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*FBlock), nil
}

// FetchAllFBlocks gets all of the factoid blocks
func (db *Overlay) FetchAllFBlocks() (FBlocks []IFBlock, err error) {
	bucket := []byte{byte(TBL_SC)}

	list, err := db.DB.GetAll(bucket, new(FBlock))
	if err != nil {
		return nil, err
	}
	answer := make([]IFBlock, len(list))
	for i, v := range list {
		answer[i] = v.(*FBlock)
	}
	return answer, nil
}
