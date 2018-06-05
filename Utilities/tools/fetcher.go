package tools

import (
	"encoding/hex"

	"fmt"

	"github.com/AdamSLevy/factom"
	"github.com/FactomProject/factomd/common/directoryBlock"
	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"github.com/FactomProject/factomd/database/hybridDB"
)

const level string = "level"
const bolt string = "bolt"

// Able to be either a datbase or api
type Fetcher interface {
	FetchDBlockHead() (interfaces.IDirectoryBlock, error)
	FetchDBlockByHeight(dBlockHeight uint32) (interfaces.IDirectoryBlock, error)
	//FetchDBlock(hash interfaces.IHash) (interfaces.IDirectoryBlock, error)
	FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error)
	FetchEBlock(hash interfaces.IHash) (interfaces.IEntryBlock, error)
	SetChainHeads(primaryIndexes, chainIDs []interfaces.IHash) error
}

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

func (a *APIReader) FetchEBlock(hash interfaces.IHash) (interfaces.IEntryBlock, error) {
	return nil, fmt.Errorf("Not implmented for api")
	//raw, err := factom.GetRaw(hash.String())
	//if err != nil {
	//	return nil, err
	//}
	//return rawBytesToEblock(raw)
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
	return rawBytesToDblock(raw)
}

func (a *APIReader) FetchDBlockByHeight(dBlockHeight uint32) (interfaces.IDirectoryBlock, error) {
	raw, err := factom.GetBlockByHeightRaw("d", int64(dBlockHeight))
	if err != nil {
		return nil, err
	}

	return rawRespToBlock(raw.RawData)
}

func (a *APIReader) FetchHeadIndexByChainID(chainID interfaces.IHash) (interfaces.IHash, error) {
	resp, err := factom.GetChainHead(chainID.String())
	if err != nil {
		return nil, err
	}
	return primitives.HexToHash(resp)
}

func rawBytesToEblock(raw []byte) (interfaces.IEntryBlock, error) {
	eblock := entryBlock.NewEBlock()
	err := eblock.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return eblock, nil
}

func rawBytesToDblock(raw []byte) (interfaces.IDirectoryBlock, error) {
	dblock := directoryBlock.NewDirectoryBlock(nil)
	err := dblock.UnmarshalBinary(raw)
	if err != nil {
		return nil, err
	}
	return dblock, nil
}

func rawRespToBlock(raw string) (interfaces.IDirectoryBlock, error) {
	by, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	return rawBytesToDblock(by)
}
