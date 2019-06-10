package tools

import (
	"encoding/hex"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

const level string = "level"
const bolt string = "bolt"

type Fetcher interface {
	SetChainHeads(primaryIndexes, chainIDs []interfaces.IHash) error

	FetchDBlockHead() (interfaces.IDirectoryBlock, error)
	//FetchDBlock(hash interfaces.IHash) (interfaces.IDirectoryBlock, error)
	FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error)
	FetchEBlock(hash interfaces.IHash) (interfaces.IEntryBlock, error)

	FetchEntry(hash interfaces.IHash) (interfaces.IEBEntry, error)
	FetchDBlockByHeight(dBlockHeight uint32) (interfaces.IDirectoryBlock, error)
	FetchABlockByHeight(blockHeight uint32) (interfaces.IAdminBlock, error)
	FetchFBlockByHeight(blockHeight uint32) (interfaces.IFBlock, error)
	FetchECBlockByHeight(blockHeight uint32) (interfaces.IEntryCreditBlock, error)
	FetchECBlockByPrimary(keymr interfaces.IHash) (interfaces.IEntryCreditBlock, error)
}

var _ Fetcher = (*APIReader)(nil)
var _ Fetcher = (*databaseOverlay.Overlay)(nil)

func NewDBReader(levelBolt string, path string) *databaseOverlay.Overlay {
	var dbase *hybridDB.HybridDB
	var err error
	if levelBolt == bolt {
		dbase = hybridDB.NewBoltMapHybridDB(nil, path)
	} else {
		dbase, err = hybridDB.NewLevelMapHybridDB(path, false)
		if err != nil {
			panic(err)
		}
	}

	dbo := databaseOverlay.NewOverlay(dbase)
	return dbo
}

type APIReader struct {
	location string
}

func NewAPIReader(loc string) *APIReader {
	a := new(APIReader)
	a.location = loc
	factom.SetFactomdServer(loc)

	return a
}

func (a *APIReader) SetChainHeads(primaryIndexes, chainIDs []interfaces.IHash) error {
	return nil
}

func (a *APIReader) FetchEntry(hash interfaces.IHash) (interfaces.IEBEntry, error) {
	raw, err := factom.GetRaw(hash.String())
	if err != nil {
		return nil, err
	}

	entry := entryBlock.NewEntry()
	err = UnmarshalGeneric(entry, raw)
	return entry, err
}

func (a *APIReader) FetchEBlock(hash interfaces.IHash) (interfaces.IEntryBlock, error) {
	raw, err := factom.GetRaw(hash.String())
	if err != nil {
		return nil, err
	}

	block := entryBlock.NewEBlock()
	err = UnmarshalGeneric(block, raw)
	return block, err
}

func (a *APIReader) FetchDBlockHead() (interfaces.IDirectoryBlock, error) {
	head, err := factom.GetDBlockHead()
	if err != nil {
		return nil, err
	}
	raw, err := factom.GetRaw(head)
	if err != nil {
		return nil, err
	}

	block := directoryBlock.NewDirectoryBlock(nil)
	err = UnmarshalGeneric(block, raw)
	return block, err
}

func (a *APIReader) FetchDBlockByHeight(height uint32) (interfaces.IDirectoryBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("d", int64(height))
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(raw.RawData)
	if err != nil {
		return nil, err
	}

	block := directoryBlock.NewDirectoryBlock(nil)
	err = UnmarshalGeneric(block, data)
	return block, err
}

func (a *APIReader) FetchFBlockByHeight(height uint32) (interfaces.IFBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("f", int64(height))
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(raw.RawData)
	if err != nil {
		return nil, err
	}

	block := factoid.NewFBlock(nil)
	err = UnmarshalGeneric(block, data)
	return block, err
}

func (a *APIReader) FetchABlockByHeight(height uint32) (interfaces.IAdminBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("a", int64(height))
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(raw.RawData)
	if err != nil {
		return nil, err
	}

	ablock := adminBlock.NewAdminBlock(nil)
	err = UnmarshalGeneric(ablock, data)
	return ablock, err
}

func (a *APIReader) FetchECBlockByPrimary(keymr interfaces.IHash) (interfaces.IEntryCreditBlock, error) {
	data, err := factom.GetRaw(keymr.String())
	if err != nil {
		return nil, err
	}

	ecblock := entryCreditBlock.NewECBlock()
	err = UnmarshalGeneric(ecblock, data)
	return ecblock, err
}

func (a *APIReader) FetchECBlockByHeight(height uint32) (interfaces.IEntryCreditBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("ec", int64(height))
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(raw.RawData)
	if err != nil {
		return nil, err
	}

	ecblock := entryCreditBlock.NewECBlock()
	err = UnmarshalGeneric(ecblock, data)
	return ecblock, err
}

func (a *APIReader) FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error) {
	resp, err := factom.GetChainHead(chainID.String())
	if err != nil {
		return nil, err
	}
	return primitives.HexToHash(resp)
}

func UnmarshalGeneric(i interfaces.BinaryMarshallable, raw []byte) error {
	err := i.UnmarshalBinary(raw)
	if err != nil {
		return err
	}
	return nil
}
