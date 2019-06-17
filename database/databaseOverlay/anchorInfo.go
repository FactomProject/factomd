// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var BitcoinAnchorChainID = "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604"
var EthereumAnchorChainID = "6e4540d08d5ac6a1a394e982fb6a2ab8b516ee751c37420055141b94fe070bfe"
var AnchorSigKeys = []string{
	"0426a802617848d4d16d87830fc521f4d136bb2d0c352850919c2679f189613a", //m1 key
	"d569419348ed7056ec2ba54f0ecd9eea02648b260b26e0474f8c07fe9ac6bf83", //m2 key
	// "547d837160766b9ca47e689e52ed55fdc05cb3430ad2328dcc431083db083ee6", // TODO: swap out this anchormaker test key
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

func (dbo *Overlay) ReparseAnchorChains() error {
	btcChainID, err := primitives.NewShaHashFromStr(BitcoinAnchorChainID)
	if err != nil {
		panic(err)
		return err
	}
	btcAnchorEntries, err := dbo.FetchAllEntriesByChainID(btcChainID)
	if err != nil {
		panic(err)
		return err
	}

	ethChainID, err := primitives.NewShaHashFromStr(EthereumAnchorChainID)
	if err != nil {
		panic(err)
		return err
	}
	ethAnchorEntries, err := dbo.FetchAllEntriesByChainID(ethChainID)
	if err != nil {
		panic(err)
		return err
	}

	entries := append(btcAnchorEntries, ethAnchorEntries...)
	for _, entry := range entries {
		_ = dbo.SaveAnchorInfoFromEntry(entry)
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
	dbi, err := dbo.CreateUpdatedDirBlockInfoFromAnchorRecord(ar)
	if err != nil {
		return err
	}
	if dbi.EthereumConfirmed && dbi.EthereumAnchorRecordEntryHash.IsSameAs(primitives.ZeroHash) {
		dbi.EthereumAnchorRecordEntryHash = entry.GetHash()
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
	dbi, err := dbo.CreateUpdatedDirBlockInfoFromAnchorRecord(ar)
	if err != nil {
		return err
	}
	if dbi.EthereumConfirmed && dbi.EthereumAnchorRecordEntryHash.IsSameAs(primitives.ZeroHash) {
		dbi.EthereumAnchorRecordEntryHash = entry.GetHash()
	}
	return dbo.ProcessDirBlockInfoMultiBatch(dbi)
}

func (dbo *Overlay) CreateUpdatedDirBlockInfoFromAnchorRecord(ar *anchor.AnchorRecord) (*dbInfo.DirBlockInfo, error) {
	height := ar.DBHeight
	if ar.DBHeightMax != 0 && ar.DBHeightMax != ar.DBHeight {
		height = ar.DBHeightMax
	}
	dirBlockKeyMR, err := dbo.FetchDBKeyMRByHeight(height)
	if err != nil {
		return nil, err
	}

	dirBlockInfo, err := dbo.FetchDirBlockInfoByKeyMR(dirBlockKeyMR)
	if err != nil {
		return nil, err
	}

	var dbi *dbInfo.DirBlockInfo
	if dirBlockInfo == nil {
		dbi = dbInfo.NewDirBlockInfo()
		dbi.DBHash = dirBlockKeyMR
		dbi.DBMerkleRoot = dirBlockKeyMR
		dbi.DBHeight = height
	} else {
		dbi = dirBlockInfo.(*dbInfo.DirBlockInfo)
	}

	if ar.Bitcoin != nil {
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
		dbi.BTCConfirmed = true
	} else if ar.Ethereum != nil {
		dbi.EthereumConfirmed = true
	}

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
