// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package databaseOverlay

import (
	"errors"
	"github.com/FactomProject/factomd/anchor"
	"github.com/FactomProject/factomd/common/directoryBlock/dbInfo"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var BitcoinAnchorChainID = "df3ade9eec4b08d5379cc64270c30ea7315d8a8a1a69efe2b98a60ecdd69e604"
var EthereumAnchorChainID = "6e4540d08d5ac6a1a394e982fb6a2ab8b516ee751c37420055141b94fe070bfe"
var ValidAnchorChains = map[string]bool{
	BitcoinAnchorChainID:  true,
	EthereumAnchorChainID: true,
}

func (dbo *Overlay) SetBitcoinAnchorRecordPublicKeysFromHex(publicKeys []string) error {
	dbo.BitcoinAnchorRecordPublicKeys = nil
	for _, v := range publicKeys {
		publicKey := new(primitives.PublicKey)
		err := publicKey.UnmarshalText([]byte(v))
		if err != nil {
			return err
		}
		dbo.BitcoinAnchorRecordPublicKeys = append(dbo.BitcoinAnchorRecordPublicKeys, publicKey)
	}
	return nil
}

func (dbo *Overlay) SetEthereumAnchorRecordPublicKeysFromHex(publicKeys []string) error {
	dbo.BitcoinAnchorRecordPublicKeys = nil
	for _, v := range publicKeys {
		publicKey := new(primitives.PublicKey)
		err := publicKey.UnmarshalText([]byte(v))
		if err != nil {
			return err
		}
		dbo.EthereumAnchorRecordPublicKeys = append(dbo.EthereumAnchorRecordPublicKeys, publicKey)
	}
	return nil
}

func (dbo *Overlay) ReparseAnchorChains() error {
	// Delete all DirBlockInfo buckets
	err := dbo.Clear(DIRBLOCKINFO)
	if err != nil {
		return err
	}
	err = dbo.Clear(DIRBLOCKINFO_UNCONFIRMED)
	if err != nil {
		return err
	}
	err = dbo.Clear(DIRBLOCKINFO_NUMBER)
	if err != nil {
		return err
	}
	err = dbo.Clear(DIRBLOCKINFO_SECONDARYINDEX)
	if err != nil {
		return err
	}

	// Fetch all potential anchor records
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

	// Validate structure, verify signatures, and store in database
	entries := append(btcAnchorEntries, ethAnchorEntries...)
	for _, entry := range entries {
		_ = dbo.SaveAnchorInfoFromEntry(entry, false)
	}
	return nil
}

func (dbo *Overlay) SaveAnchorInfoFromEntry(entry interfaces.IEBEntry, multiBatch bool) error {
	var anchorRecord *anchor.AnchorRecord
	var ok bool
	var err error

	switch entry.GetChainID().String() {
	case BitcoinAnchorChainID:
		// Bitcoin has mixed v1 and v2 AnchorRecords
		anchorRecord, ok, err = anchor.UnmarshalAndValidateAnchorEntryAnyVersion(entry, dbo.BitcoinAnchorRecordPublicKeys)
	case EthereumAnchorChainID:
		// Ethereum has v2 AnchorRecords only
		anchorRecord, ok, err = anchor.UnmarshalAndValidateAnchorRecordV2(entry.GetContent(), entry.ExternalIDs(), dbo.EthereumAnchorRecordPublicKeys)
	default:
		// Given where this function is called from, we shouldn't hit this. But just in case...
		return errors.New("unsupported anchor chain")
	}

	if err != nil {
		return err
	} else if ok == false || anchorRecord == nil {
		return nil
	}

	// We have a valid, signed anchor record entry
	// Now either create the DirBlockInfo for this block or update the existing DirBlockInfo with new found data
	dbi, err := dbo.CreateUpdatedDirBlockInfoFromAnchorRecord(anchorRecord)
	if err != nil {
		return err
	}
	if dbi.EthereumConfirmed && dbi.EthereumAnchorRecordEntryHash.IsSameAs(primitives.ZeroHash) {
		dbi.EthereumAnchorRecordEntryHash = entry.GetHash()
	}

	if multiBatch {
		return dbo.ProcessDirBlockInfoMultiBatch(dbi)
	}
	return dbo.ProcessDirBlockInfoBatch(dbi)
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
