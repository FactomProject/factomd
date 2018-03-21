// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	//"fmt"
	"sort"

	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var AnchorBlockID string = "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604"
var AnchorSigKeys []string = []string{
	"0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a", //m1 key
	"d569419348ed7056ec2ba54f0ecd9eea02648b260b26e0474f8c07fe9ac6bf83", //m2 key
}
var AnchorSigPublicKeys []interfaces.Verifier

func init() {
	for _, v := range AnchorSigKeys {
		pubKey := new(primitives.PublicKey)
		err := pubKey.UnmarshalText([]byte(v))
		if err != nil {
			panic(err)
		}
		AnchorSigPublicKeys = append(AnchorSigPublicKeys, pubKey)
	}
}

func (dbo *Overlay) RebuildDirBlockInfo() error {
	ars, err := dbo.FetchAllAnchorInfo()
	if err != nil {
		return err
	}
	err = dbo.SaveAnchorInfoAsDirBlockInfo(ars)
	if err != nil {
		return err
	}

	return nil
}

func (dbo *Overlay) SaveAnchorInfoFromEntry(entry interfaces.IEBEntry) error {
	if entry.DatabasePrimaryIndex().String() == "24674e6bc3094eb773297de955ee095a05830e431da13a37382dcdc89d73c7d7" {
		return nil
	}
	ar, ok, err := anchor.UnmarshalAndValidateAnchorEntryAnyVersion(entry, AnchorSigPublicKeys)
	if err != nil {
		return err
	}
	if ok == false {
		return nil
	}
	if ar == nil {
		return nil
	}
	dbi, err := AnchorRecordToDirBlockInfo(ar)
	if err != nil {
		return err
	}
	return dbo.ProcessDirBlockInfoBatch(dbi)
}

func (dbo *Overlay) SaveAnchorInfoFromEntryMultiBatch(entry interfaces.IEBEntry) error {
	if entry.DatabasePrimaryIndex().String() == "24674e6bc3094eb773297de955ee095a05830e431da13a37382dcdc89d73c7d7" {
		return nil
	}
	ar, ok, err := anchor.UnmarshalAndValidateAnchorEntryAnyVersion(entry, AnchorSigPublicKeys)
	if err != nil {
		return err
	}
	if ok == false {
		return nil
	}
	if ar == nil {
		return nil
	}
	dbi, err := AnchorRecordToDirBlockInfo(ar)
	if err != nil {
		return err
	}
	return dbo.ProcessDirBlockInfoMultiBatch(dbi)
}

func (dbo *Overlay) FetchAllAnchorInfo() ([]*anchor.AnchorRecord, error) {
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
	sort.Sort(ByAnchorDBHeightAscending(answer))
	return answer, nil
}

func (dbo *Overlay) SaveAnchorInfoAsDirBlockInfo(ars []*anchor.AnchorRecord) error {
	sort.Sort(ByAnchorDBHeightAscending(ars))

	for _, v := range ars {
		dbi, err := AnchorRecordToDirBlockInfo(v)
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

func AnchorRecordToDirBlockInfo(ar *anchor.AnchorRecord) (*dbInfo.DirBlockInfo, error) {
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

// AnchorRecord array sorting implementation - ascending
type ByAnchorDBHeightAscending []*anchor.AnchorRecord

func (f ByAnchorDBHeightAscending) Len() int {
	return len(f)
}
func (f ByAnchorDBHeightAscending) Less(i, j int) bool {
	return f[i].DBHeight < f[j].DBHeight
}
func (f ByAnchorDBHeightAscending) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
