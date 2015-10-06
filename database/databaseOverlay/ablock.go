package databaseOverlay

import (
	//	"errors"
	"encoding/binary"
	. "github.com/FactomProject/factomd/common/AdminBlock"
	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/interfaces"
)

// ProcessABlockBatch inserts the AdminBlock
func (db *Overlay) ProcessABlockBatch(block *AdminBlock) error {
	if block == nil {
		return nil
	}

	batch := []Record{}

	abHash, err := block.PartialHash()
	if err != nil {
		return err
	}
	batch = append(batch, Record{[]byte{byte(TBL_AB)}, abHash.Bytes(), block})

	bytes := make([]byte, 4)
	binary.BigEndian.PutUint32(bytes, block.Header.DBHeight)
	batch = append(batch, Record{[]byte{byte(TBL_AB_NUM)}, bytes, abHash})

	batch = append(batch, Record{[]byte{byte(TBL_CHAIN_HEAD)}, ADMIN_CHAINID, abHash})

	err = db.DB.PutInBatch(batch)
	if err != nil {
		return err
	}

	return nil
}

// FetchABlockByHash gets an admin block by hash from the database.
func (db *Overlay) FetchABlockByHash(aBlockHash IHash) (*AdminBlock, error) {
	bucket := []byte{byte(TBL_AB)}
	key := aBlockHash.Bytes()

	block, err := db.DB.Get(bucket, key, new(AdminBlock))
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, nil
	}
	return block.(*AdminBlock), nil
}

// FetchAllABlocks gets all of the admin blocks
func (db *Overlay) FetchAllABlocks() (aBlocks []*AdminBlock, err error) {
	bucket := []byte{byte(TBL_AB)}

	list, err := db.DB.GetAll(bucket, new(AdminBlock))
	if err != nil {
		return nil, err
	}
	answer := make([]*AdminBlock, len(list))
	for i, v := range list {
		answer[i] = v.(*AdminBlock)
	}
	return answer, nil
}
