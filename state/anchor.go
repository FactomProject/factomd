// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	//"fmt"
	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
	"sort"
)

var AnchorBlockID string = "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604"

func RebuildDirBlockInfo(dbo *databaseOverlay.Overlay) error {
	ars, err := FetchAllAnchorInfo(dbo)
	if err != nil {
		return err
	}
	err = SaveAnchorInfoAsDirBlockInfo(dbo, ars)
	if err != nil {
		return err
	}

	return nil
}

func FetchAllAnchorInfo(dbo *databaseOverlay.Overlay) ([]*anchor.AnchorRecord, error) {
	chainID, err := primitives.NewShaHashFromStr(AnchorBlockID)
	if err != nil {
		panic(err)
		return nil, err
	}
	entries, err := dbo.FetchAllEntriesByChainID(chainID)
	if err != nil {
		panic(err)
		return nil, err
	}
	answer := []*anchor.AnchorRecord{}
	for _, entry := range entries {
		if entry.DatabasePrimaryIndex().String() == "24674e6bc3094eb773297de955ee095a05830e431da13a37382dcdc89d73c7d7" {
			continue
		}
		content := entry.GetContent()
		ar, err := anchor.UnmarshalAnchorRecord(content)
		if err != nil {
			panic(err)
			return nil, err
		}
		answer = append(answer, ar)
	}
	sort.Sort(ByAnchorDBHeightAccending(answer))
	return answer, nil
}

func SaveAnchorInfoAsDirBlockInfo(dbo *databaseOverlay.Overlay, ars []*anchor.AnchorRecord) error {
	sort.Sort(ByAnchorDBHeightAccending(ars))

	for _, v := range ars {
		dbi, err := AnchorRecordToDirBlockInfo(v, dbo)
		if err != nil {
			return err
		}
		err = dbo.SaveDirBlockInfo(dbi)
		if err != nil {
			return err
		}
	}

	return nil
}

func AnchorRecordToDirBlockInfo(ar *anchor.AnchorRecord, dbo *databaseOverlay.Overlay) (*dbInfo.DirBlockInfo, error) {
	dbi := new(dbInfo.DirBlockInfo)
	var err error

	//TODO: fetch proper data
	//dbi.DBHash =
	dbi.DBHash, err = primitives.NewShaHashFromStr(ar.KeyMR)
	if err != nil {
		return nil, err
	}
	dbi.DBHeight = ar.DBHeight
	//dbi.Timestamp =
	dbi.BTCTxHash, err = primitives.NewShaHashFromStr(ar.Bitcoin.TXID)
	if err != nil {
		return nil, err
	}
	dbi.BTCTxOffset = ar.Bitcoin.Offset
	dbi.BTCBlockHeight = ar.Bitcoin.BlockHeight
	dbi.BTCBlockHash, err = primitives.NewShaHashFromStr(ar.Bitcoin.BlockHash)
	if err != nil {
		return nil, err
	}
	dbi.DBMerkleRoot, err = primitives.NewShaHashFromStr(ar.KeyMR)
	if err != nil {
		return nil, err
	}
	dbi.BTCConfirmed = true

	return dbi, nil
}

// AnchorRecord array sorting implementation - accending
type ByAnchorDBHeightAccending []*anchor.AnchorRecord

func (f ByAnchorDBHeightAccending) Len() int {
	return len(f)
}
func (f ByAnchorDBHeightAccending) Less(i, j int) bool {
	return f[i].DBHeight < f[j].DBHeight
}
func (f ByAnchorDBHeightAccending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
