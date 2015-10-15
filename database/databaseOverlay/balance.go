package databaseOverlay

import (
	"bytes"
	"encoding/binary"
	"fmt"
	. "github.com/FactomProject/factomd/common/constants"
	. "github.com/FactomProject/factomd/common/interfaces"
)

type FSbalance struct {
	IBlock
	number uint64
}

func (FSbalance) GetNewInstance() IBlock {
	return new(FSbalance)
}

func (f *FSbalance) UnmarshalBinaryData(data []byte) ([]byte, error) {
	num, data := binary.BigEndian.Uint64(data), data[8:]
	f.number = num
	return data, nil
}

func (f *FSbalance) UnmarshalBinary(data []byte) error {
	data, err := f.UnmarshalBinaryData(data)
	return err
}

func (f FSbalance) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer
	binary.Write(&out, binary.BigEndian, uint64(f.number))
	return out.Bytes(), nil
}

// Any address that is not defined has a zero balance.
func (db *Overlay) GetBalance(address IAddress) uint64 {
	balance := uint64(0)
	b, _ := db.DB.Get([]byte(DB_F_BALANCES), address.Bytes(), new(FSbalance))
	if b != nil {
		balance = b.(*FSbalance).number
	}
	return balance
}

// Any address that is not defined has a zero balance.
func (db *Overlay) GetECBalance(address IAddress) uint64 {
	balance := uint64(0)
	b, _ := db.DB.Get([]byte(DB_EC_BALANCES), address.Bytes(), new(FSbalance))
	if b != nil {
		balance = b.(*FSbalance).number
	}
	return balance
}

// Update balance throws an error if your update will drive the balance negative.
func (db *Overlay) UpdateBalance(address IAddress, amount int64) error {
	nbalance := int64(db.GetBalance(address)) + amount
	if nbalance < 0 {
		return fmt.Errorf("The update to this address would drive the balance negative.")
	}
	balance := uint64(nbalance)
	err := db.DB.Put([]byte(DB_F_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return err
}

// Update ec balance throws an error if your update will drive the balance negative.
func (db *Overlay) UpdateECBalance(address IAddress, amount int64) error {
	nbalance := int64(db.GetECBalance(address)) + amount
	if nbalance < 0 {
		return fmt.Errorf("The update to this Entry Credit address would drive the balance negative.")
	}
	balance := uint64(nbalance)
	err := db.DB.Put([]byte(DB_EC_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return err
}

// Add to Entry Credit Balance.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits. Using lowers the Entry Credit Balance.
func (db *Overlay) AddToECBalance(address IAddress, amount uint64, factoshisPerEC uint64) error {
	ecs := amount / factoshisPerEC
	balance := db.GetECBalance(address) + ecs
	err := db.DB.Put([]byte(DB_EC_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return err
}

// Use Entry Credits.  Note Entry Credit balances are maintained
// as entry credits, not Factoids.  But adding is done in Factoids, using
// done in Entry Credits.  Using lowers the Entry Credit Balance.
func (db *Overlay) UseECs(address IAddress, amount uint64) error {
	balance := db.GetECBalance(address) - amount
	if balance < 0 {
		return fmt.Errorf("Overdraft of Entry Credits attempted.")
	}
	err := db.DB.Put([]byte(DB_EC_BALANCES), address.Bytes(), &FSbalance{number: balance})
	return err
}

func (db *Overlay) PutTransactionBlock(hash IHash, trans IFBlock) error {
	return db.DB.Put([]byte(DB_FACTOID_BLOCKS), hash.Bytes(), trans)
}

func (db *Overlay) GetTransactionBlock(hash IHash, dst IFBlock) (IFBlock, error) {
	transblk, err := db.DB.Get([]byte(DB_FACTOID_BLOCKS), hash.Bytes(), dst)
	if err != nil {
		return nil, err
	}
	if transblk == nil {
		return nil, nil
	}
	return transblk.(IFBlock), nil
}
