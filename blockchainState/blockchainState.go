// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package blockchainState

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

const EBLOCKEXPIRATION uint32 = 20 //TODO: set properly

var Expired int = 0
var LatestReveal int = 0
var TotalEntries int = 0

type BlockchainState struct {
	DBlockHeadKeyMR interfaces.IHash
	DBlockHeadHash  interfaces.IHash
	DBlockHeight    uint32

	ECBlockHeadKeyMR interfaces.IHash
	ECBlockHeadHash  interfaces.IHash

	FBlockHeadKeyMR interfaces.IHash
	FBlockHeadHash  interfaces.IHash

	ABlockHeadRefHash interfaces.IHash

	BlockHeads map[string]interfaces.IHash

	ECBalances   map[string]uint64
	FBalances    map[string]uint64
	ExchangeRate uint64

	PendingCommits map[string]*PendingCommit //entry hash: current DBlock height
}

func (bs *BlockchainState) HasFreeCommit(h interfaces.IHash) bool {
	pc, ok := bs.PendingCommits[h.String()]
	if ok == false {
		return false
	}
	switch h.String() {
	case "":
		fmt.Printf("Missing commit - %v\n", pc.String())
		break
	}
	return pc.HasFreeCommit()
}

func (bs *BlockchainState) PopCommit(h interfaces.IHash) error {
	pc, ok := bs.PendingCommits[h.String()]
	if ok == false {
		return fmt.Errorf("No commits found")
		//return nil
	}
	return pc.PopCommit(bs.DBlockHeight)
}

func (bs *BlockchainState) PushCommit(entryHash interfaces.IHash, commitTxID interfaces.IHash) {
	if bs.PendingCommits[entryHash.String()] == nil {
		bs.PendingCommits[entryHash.String()] = new(PendingCommit)
	}
	bs.PendingCommits[entryHash.String()].PushCommit(commitTxID, bs.DBlockHeight)
}

func (bs *BlockchainState) ClearExpiredCommits() error {
	for k, v := range bs.PendingCommits {
		v.ClearExpiredCommits(bs.DBlockHeight)
		if v.HasFreeCommit() == false {
			delete(bs.PendingCommits, k)
		}
	}
	return nil
}

type PendingCommit struct {
	Commits []SingleCommit
}

func (pc *PendingCommit) String() string {
	str, _ := primitives.EncodeJSONString(pc)
	return str
}

func (pc *PendingCommit) HasFreeCommit() bool {
	if len(pc.Commits) > 0 {
		return true
	}
	return false
}

func (pc *PendingCommit) PopCommit(dblockHeight uint32) error {
	if len(pc.Commits) == 0 {
		return fmt.Errorf("No commits found")
	}
	if int(dblockHeight-pc.Commits[0].DBlockHeight) > LatestReveal {
		LatestReveal = int(dblockHeight - pc.Commits[0].DBlockHeight)
	}
	pc.Commits = pc.Commits[1:]
	return nil
}

func (pc *PendingCommit) PushCommit(commitTxID interfaces.IHash, dblockHeight uint32) {
	pc.Commits = append(pc.Commits, SingleCommit{DBlockHeight: dblockHeight, CommitTxID: commitTxID.String()})
}

func (pc *PendingCommit) ClearExpiredCommits(dblockHeight uint32) {
	for {
		if len(pc.Commits) == 0 {
			return
		}
		if pc.Commits[0].DBlockHeight+EBLOCKEXPIRATION < dblockHeight {
			pc.Commits = pc.Commits[1:]
			Expired++
		} else {
			return
		}
	}
}

type SingleCommit struct {
	DBlockHeight uint32
	CommitTxID   string
}

func (bs *BlockchainState) Init() {
	if bs.BlockHeads == nil {
		bs.BlockHeads = map[string]interfaces.IHash{}
	}
	if bs.ECBalances == nil {
		bs.ECBalances = map[string]uint64{}
	}
	if bs.FBalances == nil {
		bs.FBalances = map[string]uint64{}
	}
	if bs.PendingCommits == nil {
		bs.PendingCommits = map[string]*PendingCommit{}
	}

	if bs.DBlockHeadKeyMR == nil {
		bs.DBlockHeadKeyMR = primitives.NewZeroHash()
	}
	if bs.DBlockHeadHash == nil {
		bs.DBlockHeadHash = primitives.NewZeroHash()
	}
	if bs.ECBlockHeadKeyMR == nil {
		bs.ECBlockHeadKeyMR = primitives.NewZeroHash()
	}
	if bs.ECBlockHeadHash == nil {
		bs.ECBlockHeadHash = primitives.NewZeroHash()
	}
	if bs.FBlockHeadKeyMR == nil {
		bs.FBlockHeadKeyMR = primitives.NewZeroHash()
	}
	if bs.FBlockHeadHash == nil {
		bs.FBlockHeadHash = primitives.NewZeroHash()
	}
	if bs.ABlockHeadRefHash == nil {
		bs.ABlockHeadRefHash = primitives.NewZeroHash()
	}
}

func (bs *BlockchainState) ProcessBlockSet(dBlock interfaces.IDirectoryBlock, aBlock interfaces.IAdminBlock, fBlock interfaces.IFBlock, ecBlock interfaces.IEntryCreditBlock,
	eBlocks []interfaces.IEntryBlock) error {
	bs.Init()
	err := bs.HandleErrors(dBlock.DatabasePrimaryIndex())
	if err != nil {
		return err
	}
	err = bs.ProcessDBlock(dBlock)
	if err != nil {
		return err
	}
	err = bs.ProcessABlock(aBlock)
	if err != nil {
		return err
	}
	err = bs.ProcessFBlock(fBlock)
	if err != nil {
		return err
	}
	err = bs.ProcessECBlock(ecBlock)
	if err != nil {
		return err
	}
	err = bs.ProcessEBlocks(eBlocks)
	if err != nil {
		return err
	}
	return nil
}

func (bs *BlockchainState) ProcessDBlock(dBlock interfaces.IDirectoryBlock) error {
	bs.Init()

	if bs.DBlockHeadKeyMR.String() != dBlock.GetHeader().GetPrevKeyMR().String() {
		fmt.Printf("Invalid DBlock %v previous KeyMR - expected %v, got %v\n", dBlock.DatabasePrimaryIndex().String(), bs.DBlockHeadKeyMR.String(), dBlock.GetHeader().GetPrevKeyMR().String())
	}
	if bs.DBlockHeadHash.String() != dBlock.GetHeader().GetPrevFullHash().String() {
		fmt.Printf("Invalid DBlock %v previous hash - expected %v, got %v\n", dBlock.DatabasePrimaryIndex().String(), bs.DBlockHeadHash.String(), dBlock.GetHeader().GetPrevFullHash().String())
	}

	bs.DBlockHeadKeyMR = dBlock.DatabasePrimaryIndex()
	bs.DBlockHeadHash = dBlock.DatabaseSecondaryIndex()
	bs.DBlockHeight = dBlock.GetDatabaseHeight()

	dbEntries := dBlock.GetDBEntries()
	for _, v := range dbEntries {
		bs.BlockHeads[v.GetChainID().String()] = v.GetKeyMR()
	}

	return nil
}

func (bs *BlockchainState) ProcessABlock(aBlock interfaces.IAdminBlock) error {
	bs.Init()

	if bs.ABlockHeadRefHash.String() != aBlock.GetHeader().GetPrevBackRefHash().String() {
		fmt.Printf("Invalid ABlock %v previous KeyMR - expected %v, got %v\n", aBlock.GetHash(), bs.ABlockHeadRefHash.String(), aBlock.GetHeader().GetPrevBackRefHash().String())
	}
	bs.ABlockHeadRefHash = aBlock.DatabaseSecondaryIndex()

	return nil
}

func (bs *BlockchainState) ProcessFBlock(fBlock interfaces.IFBlock) error {
	bs.Init()

	if bs.FBlockHeadKeyMR.String() != fBlock.GetPrevKeyMR().String() {
		fmt.Printf("Invalid FBlock %v previous KeyMR - expected %v, got %v\n", fBlock.DatabasePrimaryIndex().String(), bs.FBlockHeadKeyMR.String(), fBlock.GetPrevKeyMR().String())
	}
	if bs.FBlockHeadHash.String() != fBlock.GetPrevLedgerKeyMR().String() {
		fmt.Printf("Invalid FBlock %v previous hash - expected %v, got %v\n", fBlock.DatabasePrimaryIndex().String(), bs.FBlockHeadHash.String(), fBlock.GetPrevLedgerKeyMR().String())
	}
	bs.FBlockHeadKeyMR = fBlock.DatabasePrimaryIndex()
	bs.FBlockHeadHash = fBlock.DatabaseSecondaryIndex()

	transactions := fBlock.GetTransactions()
	for _, v := range transactions {
		err := bs.ProcessFactoidTransaction(v)
		if err != nil {
			return err
		}
	}
	bs.ExchangeRate = fBlock.GetExchRate()
	return nil
}

func (bs *BlockchainState) ProcessFactoidTransaction(tx interfaces.ITransaction) error {
	bs.Init()
	ins := tx.GetInputs()
	for _, w := range ins {
		if bs.FBalances[w.GetAddress().String()] < w.GetAmount() {
			return fmt.Errorf("Not enough factoids")
		}
		bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] - w.GetAmount()
	}
	outs := tx.GetOutputs()
	for _, w := range outs {
		bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] + w.GetAmount()
	}
	ecOut := tx.GetECOutputs()
	for _, w := range ecOut {
		bs.ECBalances[w.GetAddress().String()] = bs.ECBalances[w.GetAddress().String()] + w.GetAmount()
	}
	return nil
}

func (bs *BlockchainState) ProcessECBlock(ecBlock interfaces.IEntryCreditBlock) error {
	bs.Init()

	if bs.ECBlockHeadKeyMR.String() != ecBlock.GetHeader().GetPrevHeaderHash().String() {
		fmt.Printf("Invalid ECBlock %v previous KeyMR - expected %v, got %v\n", ecBlock.DatabasePrimaryIndex().String(), bs.ECBlockHeadKeyMR.String(), ecBlock.GetHeader().GetPrevHeaderHash().String())
	}
	if bs.ECBlockHeadHash.String() != ecBlock.GetHeader().GetPrevFullHash().String() {
		fmt.Printf("Invalid ECBlock %v previous hash - expected %v, got %v\n", ecBlock.DatabasePrimaryIndex().String(), bs.ECBlockHeadHash.String(), ecBlock.GetHeader().GetPrevFullHash().String())
	}
	bs.ECBlockHeadKeyMR = ecBlock.DatabasePrimaryIndex()
	bs.ECBlockHeadHash = ecBlock.DatabaseSecondaryIndex()

	entries := ecBlock.GetEntries()
	for _, v := range entries {
		err := bs.ProcessECEntries(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockchainState) HandleErrors(dBlockHash interfaces.IHash) error {
	bs.Init()
	hash := dBlockHash.String()

	//ECBlock discontinuity
	switch hash {
	case "7bd9704ec9e8e5238fe7669cc235b56c109052e7a8be4281b2a985922c24e038": //22881
		h, _ := primitives.HexToHash("f3fa09154eeba632d88542a20637a8b3eb86758f0a46b5cca364483b815e6b45")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("0d3b181f15177156e8381ffc21203b31114676902947987170ac73008d20a352")
		bs.ECBlockHeadHash = h
		break
	case "3957f3f30fa71d1586e0fe2c575a11775fc52bce08326084ab459b2cea28d7fb": //22884
		h, _ := primitives.HexToHash("c3e8fda5dfea0d7ea0dad5d59a6f6bd10ca18cae5f57545913915b635cb1667c")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("19c73168fed48738bfdc53c63dc10beb18cdf12d63339f3eefb9b2bdc35af1ee")
		bs.ECBlockHeadHash = h
		break
	case "59b2a648c5a32ec3dcf2254e6e50e492e7443e390edb36f3908fecd870e20ec5": //22941
		h, _ := primitives.HexToHash("7962567ff7b9cd4e9283e8120d20af03c221205035702ea517e4a9cf7874b828")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("916a5e9a653ae50597f6a98515687621925d4403b0582aa054f3f5363b784a46")
		bs.ECBlockHeadHash = h
		break
	case "4d19fde10fdaf3873b4dd827e5360a6ee9b848d73d469b274e6bd24a5ffa2096": //22950
		h, _ := primitives.HexToHash("04faa69b36313712c6d166a7f74af56cfe846128af20ee422126b900b10f0ddb")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("d369b72492ee9c91424be0b0866959c4c02bd3f888abc3fa71359a6b260f4c10")
		bs.ECBlockHeadHash = h
		break
	case "0c93039ccd3b9d685c354323c356ed56e12447b8663d69d539ed17104a1bbc4a": //22977
		h, _ := primitives.HexToHash("750bafb62bd5783c3459e65ec15a80d56c9007399ab8c78e3205085a99c13b4b")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("dc77eed730b88c0eca3e3d8157eb0376de2a9fe1bcf36acf9fa78f5994a83517")
		bs.ECBlockHeadHash = h
		break
	case "4e184bf1326e9cd5cea23d5a409bbbcf45450754d6e3c16938cb709d2cfebf1c": //22979
		h, _ := primitives.HexToHash("45146a4b32883031e1ffbfadc25688375a7d78964224a77e018c19224a8f9445")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("8ab1001168b5427c13064c849fa3e33e0ada01579a45a468aa40c419b7235e0d")
		bs.ECBlockHeadHash = h
		break
	case "8441b4e44f5b1696e17696ddfae81448b77f10347a72931a7b2f896a2deef063": //23269
		h, _ := primitives.HexToHash("88a7ee2eb0bbc3d66b64531ee877620d8b13b327628f47d914e19f6ee9075233")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("f878c85e0c22b4980423971ece2dd43636d67d46eeabbf8e394443d0d30cd885")
		bs.ECBlockHeadHash = h
		break
	case "043a2153b6905ffc9560c5aa36569010dbcba97fb86edb00315e1eb1f063cfa8": //31461
		h, _ := primitives.HexToHash("d5d59a11b01907fca10610f1ab909c1f2d549bff129bc25c399a848d0046eaf9")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("f3ef19fdac26528336e8edefd46ad1b56fe6ba824e855c7fead879c8e6c2cf65")
		bs.ECBlockHeadHash = h
		break
	case "ceb025d7b3e1e75adab24696d8c9c6f352f8b672e80ecca16631cce886b56001": //49237
		h, _ := primitives.HexToHash("a39f0ea8c6db864615afcd112450dbe2be01ae8d8e13f8be15a50dcef820b10d")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("651d784531c3d46e62911de5f0298a2e256b6308e9d91e10697b7228fe67a979")
		bs.ECBlockHeadHash = h
		break
	case "8a2fd42e9fa3bcaff68ecf3396991954688c766bba18d24cf4d8dbaaba016ca0": //50159
		h, _ := primitives.HexToHash("d3363256fc494c78b76218b936979b31ea83bc6c3684946ed990d0052151ea48")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("bc7e49eb3cb7e97d8fae8fc3542590687d5b477118f269a7d557b2e99902dce8")
		bs.ECBlockHeadHash = h
		break
	case "cb32117379e6cae4097b29ef7e9a40995fa21f38b0723c0c3ec2d1f489b137cd": //54355
		h, _ := primitives.HexToHash("4550edb9b41c2e92a4dd76d0648f3d45f852918113d3281fb522fb890b79ddcb")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("fe66be1fee8230c9efe3893de052f61f0575b23780e0fccdce676b2a087fd578")
		bs.ECBlockHeadHash = h
		break
	case "446bdeeb4cbb811ced0d784c1384279da9dd452cc9fbdfde58eeb11b7869d10a": //57216
		h, _ := primitives.HexToHash("82f7551efc2be9c12d58715cde6bf4dca7cc2974ecc7b59991590ca529ca3a39")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("2d51bda94923de4a838808b576d8a04be0adcf0a7253ef9187cc8c77709e38b5")
		bs.ECBlockHeadHash = h
		break
	case "960d90279b03b14d00a9bac86a69d5e8408d97bdd4da0120b79b1f9e1ab9bf18": //62783
		h, _ := primitives.HexToHash("c84ef4fdc379a62c39ba8b00602a8582eb8ec3d1340e6e49c7f6230774dfb055")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("fe63d6644cbf2ca581d9a0ba0b4d81d0e4b1a5d4c9e9b89ba15a356059fbac35")
		bs.ECBlockHeadHash = h
		break
	case "af55e8ab385b64f1ca2bc8202d4571c4832b1a58a50467e708571bf176cc9aa3": //67813
		h, _ := primitives.HexToHash("abc7ba807e13b92ff9e0fbae1f1f702ce4d2821506fc35b39fc016770f31714e")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("3c91b12d0a7a5372bd852df6f5cde7d0de1d3dce03cd1fbe58af9d79a1144dc7")
		bs.ECBlockHeadHash = h
		break
	case "6cb8e09d851c94425aea470d7fa771ccd56bf9408d52d6f76f8bdc022c81bfaf": //69088
		h, _ := primitives.HexToHash("8abc6032ad76242a6c3f1ed2adcf7cf4256c81f869f915b08a93a4601fe0f3b7")
		bs.ECBlockHeadKeyMR = h
		h, _ = primitives.HexToHash("89c1dcc0f7e4822e3fe6c9482e20b15c2b003fd37e3106bb25c27a1479b585ff")
		bs.ECBlockHeadHash = h
		break
	}

	//Expired Commits
	switch hash {
	case "90424572f5663cbe6d2fc6d97c0b97384c20d8c8b0182c9b0948ea1051f0dcf5": //5050
		eHash, _ := primitives.NewShaHashFromStr("75acd41e736406404bf6705a0ac3e5435442458954a413c32c472d7f46dc69af")
		ecHash, _ := primitives.NewShaHashFromStr("0fb30d88fb38ddaea3fdf99ba03f310e7b2a471a4390bfd4e19cd06f3ac49b62")
		bs.PushCommit(eHash, ecHash)
		eHash, _ = primitives.NewShaHashFromStr("280829e1be009fd5ff4b04028a4794dc027f2ba9667d7b3695b42c1039ee61b7")
		ecHash, _ = primitives.NewShaHashFromStr("21473e3bfd058ba941d2e2d4095237787c2de0e1cf32c9977366819f9385deec")
		bs.PushCommit(eHash, ecHash)
		break
	case "4b68855a7185bc100b9d7f106737e1277a9756a565d7b86e2cc282a08e7f4251": //5051
		eHash, _ := primitives.NewShaHashFromStr("7601c6a77560402999ba91b5be7439ad81e132c0b5b221bfb8afce3abb6de917")
		ecHash, _ := primitives.NewShaHashFromStr("3396cbd422bcbf1f45dca8e41e4814233b61f484460ca3ee5a67177f21745888")
		bs.PushCommit(eHash, ecHash)
		break
	case "d3b3663b218183ce89dbe177621f4d4214de2413b094713b7511a04aad638380": //25283
		eHash, _ := primitives.NewShaHashFromStr("b3742461ed073ff0f185b858d3ebe8ab1638f0c055c1950e0709964da77f71b2")
		ecHash, _ := primitives.NewShaHashFromStr("d11473f4a0cb0595e7c1dce8708dc01de0972f55e314e68bd656b0a0b96be5a8")
		bs.PushCommit(eHash, ecHash)
		break
	}

	return nil
}

func (bs *BlockchainState) ProcessECEntries(v interfaces.IECBlockEntry) error {
	bs.Init()
	switch v.ECID() {
	case entryCreditBlock.ECIDBalanceIncrease:
		e := v.(*entryCreditBlock.IncreaseBalance)
		bs.ECBalances[e.ECPubKey.String()] = bs.ECBalances[e.ECPubKey.String()] + e.NumEC
		break
	case entryCreditBlock.ECIDEntryCommit:
		e := v.(*entryCreditBlock.CommitEntry)
		if bs.ECBalances[e.ECPubKey.String()] < uint64(e.Credits) {
			return fmt.Errorf("Not enough ECs - %v:%v<%v", e.ECPubKey.String(), bs.ECBalances[e.ECPubKey.String()], uint64(e.Credits))
		}
		bs.ECBalances[e.ECPubKey.String()] = bs.ECBalances[e.ECPubKey.String()] - uint64(e.Credits)
		bs.PushCommit(e.GetEntryHash(), v.Hash())
		break
	case entryCreditBlock.ECIDChainCommit:
		e := v.(*entryCreditBlock.CommitChain)
		if bs.ECBalances[e.ECPubKey.String()] < uint64(e.Credits) {
			return fmt.Errorf("Not enough ECs - %v:%v<%v", e.ECPubKey.String(), bs.ECBalances[e.ECPubKey.String()], uint64(e.Credits))
		}
		bs.ECBalances[e.ECPubKey.String()] = bs.ECBalances[e.ECPubKey.String()] - uint64(e.Credits)
		bs.PushCommit(e.GetEntryHash(), v.Hash())
		break
	default:
		break
	}
	return nil
}

func (bs *BlockchainState) ProcessEBlocks(eBlocks []interfaces.IEntryBlock) error {
	bs.Init()
	for _, v := range eBlocks {
		err := bs.ProcessEBlock(v)
		if err != nil {
			return err
		}
	}
	return bs.ClearExpiredCommits()
}

func (bs *BlockchainState) ProcessEBlock(eBlock interfaces.IEntryBlock) error {
	bs.Init()
	eHashes := eBlock.GetEntryHashes()
	for _, v := range eHashes {
		err := bs.ProcessEntryHash(v, eBlock.GetHash())
		if err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockchainState) ProcessEntryHash(v, block interfaces.IHash) error {
	bs.Init()
	if v.IsMinuteMarker() {
		return nil
	}
	TotalEntries++
	if bs.HasFreeCommit(v) == true {

	} else {
		fmt.Printf("Non-committed entry found in an eBlock - %v, %v, %v, %v\n", bs.DBlockHeadKeyMR.String(), bs.DBlockHeight, block.String(), v.String())
	}
	err := bs.PopCommit(v)
	if err != nil {
		fmt.Printf("Error - %v\n", err)
		//panic("")
	}
	return nil
}

func (bs *BlockchainState) Clone() (*BlockchainState, error) {
	data, err := bs.MarshalBinaryData()
	if err != nil {
		return nil, err
	}
	b := new(BlockchainState)
	err = b.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (bs *BlockchainState) MarshalBinaryData() ([]byte, error) {
	b := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(b)
	err := enc.Encode(bs)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (bs *BlockchainState) UnmarshalBinary(data []byte) error {
	bs.Init()
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)
	return dec.Decode(bs)
}

func (e *BlockchainState) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *BlockchainState) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *BlockchainState) String() string {
	str, _ := e.JSONString()
	return str
}
