package databaseOverlay

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
)

func (db *Overlay) FetchFactoidTransactionByHash(hash interfaces.IHash) (interfaces.ITransaction, error) {
	in, err := db.FetchIncludedIn(hash)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, nil
	}
	block, err := db.FetchFBlockByKeyMR(in)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("Block not found, should not happen")
	}
	txs := block.GetTransactions()
	for _, tx := range txs {
		if tx.GetHash().IsSameAs(hash) {
			return tx, nil
		}
	}
	return nil, fmt.Errorf("Transaction not found in block, should not happen")
}

func (db *Overlay) FetchECTransactionByHash(hash interfaces.IHash) (interfaces.IECBlockEntry, error) {
	in, err := db.FetchIncludedIn(hash)
	if err != nil {
		return nil, err
	}
	if in == nil {
		return nil, nil
	}
	block, err := db.FetchECBlockByHeaderHash(in)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("Block not found, should not happen")
	}
	txs := block.GetBody().GetEntries()
	for _, tx := range txs {
		if tx.Hash().IsSameAs(hash) {
			return tx, nil
		}
	}
	return nil, fmt.Errorf("Transaction not found in block, should not happen")
}
