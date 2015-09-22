package DirectoryBlock

import (
	. "github.com/FactomProject/factomd/common/interfaces"
)

type DBEntry struct {
	ChainID IHash
	KeyMR   IHash // Different MR in EBlockHeader
}

var _ Printable = (*DBEntry)(nil)
var _ BinaryMarshallable = (*DBEntry)(nil)
var _ IDBEntry = (*DBEntry)(nil)

func (c *DBEntry) MarshalledSize() uint64 {
	panic("Function not implemented")
	return 0
}

func (c *DBEntry) GetChainID() IHash {
	return c.ChainID
}
func (c *DBEntry) GetKeyMR() (IHash, error) {
	return c.KeyMR, nil
}

func NewDBEntry(entry IDBEntry) (*DBEntry, error) {
	e := new(DBEntry)

	e.ChainID = entry.GetChainID()
	var err error
	e.KeyMR, err = entry.GetKeyMR()
	if err != nil {
		return nil, err
	}

	return e, nil
}
