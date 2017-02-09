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

const EBLOCKEXPIRATION uint32 = 1000 //TODO: set properly

var Expired int = 0
var LatestReveal int = 0
var TotalEntries int = 0

var MES *MissingEntries

func init() {
	MES = new(MissingEntries)
	MES.missing = map[string]*Entry{}
	MES.found = []*Entry{}
	/*
		MES.NewMissing("7ee6c3021bb3e95cc1495c47b4e505fb1e90792879067062ff1286b3595202e1", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("57befcfc4e236ab450285b4bb33b0f069ae15c994a4409dd705b9a50146b2117", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("0fc878f3f8011034a66bea1dd6bb6e8e88f15ab36e00c736e2de48a74436798a", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("b735d05adbcff7057673d38f2b59e1c6c747e7095c6f51199d4e8bc80f520985", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("9d3877130398cac5a460fa5989e5251dcc76ab0bd39bd51e5840e42da5143ca6", "a2139054b6c7f4f48440c22fd5fe189300bc7c442ffe825ff4e071e9b47d56d9", 35640)
		MES.NewMissing("bd2068e84d703c31a25880ab62263e161936c7b16d03e8724a8ac31100e6132b", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("db4cbf8a9148d7cf91e5369dd0c8d1b7f74e4190cdc25775c95f8b3868080629", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("5837bb414cf003a2ac8911bc9f98490cc07d516335161faf505104518d5fdcde", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("c6f8b0884e6494289ad9fe46f0623a174caf40dfa21471bb2e746db0eea0ee73", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("04e4f61c02ba862d7ad537ec6b9102f8197462ec62e3d4e04636e5899e0858d9", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("607951a8fc6dfe384a770b56fc92c4d6384e629d67907ef44ea97f64065970b9", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("a43b83d5fd59f444edec5475a9e446bcc8c2565e8f7e3ffe86981d2a42b0694c", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("9b293d17991a95253ecc92c0e8089be1cde0cd1a21344e0c03b84abb68f29a8e", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("55ced301094e0ef4c0545768b8a30ec4137a47883e51614f681687c842141d18", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("e6ff494b2468897e98d333c7685f000c6be6d4e1023c0b2d2b4080a33e911245", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("959eb6a885fb4df07f410856ffdc971e1021ce777c918c7705252bb47253f663", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("dcd3e45560b421650c5b8314229a4771eccea15179b04df8318ca4c7dfcc92dc", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("7aea0ad5db315af7ef25127570c774b75add27108d4f7723cbd0f0e8039b9784", "a2139054b6c7f4f48440c22fd5fe189300bc7c442ffe825ff4e071e9b47d56d9", 35640)
		MES.NewMissing("f29aa952ee1de9a64f702a3b6615f8de41ee31e7581b0ab2e2283d98b62f7876", "a2139054b6c7f4f48440c22fd5fe189300bc7c442ffe825ff4e071e9b47d56d9", 35640)
		MES.NewMissing("165d626e1f41fab623d88b2816db3b6ca3e5ac3a774402125e36a3fb4ce4d116", "a2139054b6c7f4f48440c22fd5fe189300bc7c442ffe825ff4e071e9b47d56d9", 35640)
		MES.NewMissing("8f044304745e3df7594420f05f2e6330473eeda4c0559a242f227ca5b25002a1", "a2139054b6c7f4f48440c22fd5fe189300bc7c442ffe825ff4e071e9b47d56d9", 35640)
		MES.NewMissing("08e3031aac2334411c4bd26176ca14daadcf04158c9e222a21b676cd8c38c98d", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("27f390a40190d18c49ad400b6f1e98b5d88c346e897dcdfe80e3379467e52255", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("c4ea62b5e28fb116018537592a5bc8ae1b21aba51232a9cb2d231e52c3f7fd2e", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("51f666f22cdd506cffacd4e81451d649dcafd80ac50154ffcdd04efb24d68778", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("cf0b41b057bb2b1f088bcccd552304b21aede671aaa631c86e586e77f81276e9", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("5a3d58dc5366b15e85b76b83805c29d73bf1044d03eaf0a5f0515ae6eaab56d1", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("43f4fc2a107b27e62d84b988872413654bce904498be8152dfce81b0a5f612d2", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("dcbc532ce32761781e209f7eb2bc6a56d82469d5997c03745127ce000b653f93", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("f8142476974e66de99422b4a987ae0250d0c1c7d372b3d5150e80cc2ec726b03", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("20a8f165bb58a9b1e93e88049cce1d7508e9e5292e62638f6df8152b82fd698c", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("28b486c4d748ee699a847eb0e284d5f8227bcb7d4a948a39df51fd8725579952", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("5d42f0f8fc6857ed207286aa68b32734a639d769fb5349d1f2b1a3d16515411f", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("32ffef710041bdbf91a74a41287d0bf735f06703cab06f3f82379c68b993037c", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("3381c6f558c0cd72bb047a08bc5379d952e3319de10d006ce0bd465761190b18", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("940e372536e96d9742b23665f976c2fb610bc52a1edbe4bcc9f2fa6c1efa8f8a", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("9c4cb42a3250788bcfcca0b35e5bbf42bfe9fdd436e9394dec6d3a15e54ecc0b", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("693a135f6d747b0cd46d9a9c83653e8b3d5598e2592250a5d50074390dcbcedb", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("29aee0c092a3ad7abdac275a2de0b644510c77889eb4a8ff1db3f60d8a652375", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("c2cb669c7c7f3beabf0b0e757291c672b0f0395ad93c0f827e8c857272922b1a", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
		MES.NewMissing("f3d7683205a190b5577b1863190c7eac2152e3d7071b147d4935171cc287995a", "53e0a7d4c8bc355f15716a845d846169190e4b6dd4b0f7c18ed2a208cc653b68", 35639)
	*/
}

type MissingEntries struct {
	missing map[string]*Entry
	found   []*Entry
}

func (me *MissingEntries) IsEntryMissing(h string) bool {
	if me.missing[h] != nil {
		return true
	}
	return false
}

func (me *MissingEntries) NewMissing(entryID, dBlock string, height uint32) {
	fmt.Printf("New missing - %v\n", entryID)
	if me.missing[entryID] != nil {
		panic("Duplicate missing entry " + entryID)
	}
	e := new(Entry)
	e.EntryID = entryID
	e.EntryDBlock = dBlock
	e.EntryDBlockHeight = height

	me.missing[entryID] = e
}

func (me *MissingEntries) FoundMissing(entryID, commitID, dBlock string, height uint32) {
	fmt.Printf("Found missing - %v\n", entryID)
	e := me.missing[entryID]
	if e == nil {
		panic("Found non-missing entry! " + entryID)
	}
	e.CommitID = commitID
	e.CommitDBlock = dBlock
	e.CommitHeight = height

	me.found = append(me.found, e)
	//delete(me.missing, entryID)
}

func (me *MissingEntries) Print() {
	fmt.Printf("Missing:\n")
	for _, v := range me.missing {
		fmt.Printf("%v\t%v\t%v\n", v.EntryID, v.EntryDBlock, v.EntryDBlockHeight)
	}
	fmt.Printf("Found:\n")
	for _, v := range me.found {
		fmt.Printf("%v\t%v\t%v\t%v\t%v\t%v\n", v.EntryID, v.EntryDBlock, v.EntryDBlockHeight, v.CommitID, v.CommitDBlock, v.CommitHeight)
	}
}

type Entry struct {
	EntryID           string
	EntryDBlock       string
	EntryDBlockHeight uint32

	CommitID     string
	CommitDBlock string
	CommitHeight uint32
}

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
	if MES.IsEntryMissing(entryHash.String()) {
		MES.FoundMissing(entryHash.String(), commitTxID.String(), bs.DBlockHeadKeyMR.String(), bs.DBlockHeight)
		return
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
	err := bs.HandlePreBlockErrors(dBlock.DatabasePrimaryIndex())
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
	err = bs.HandlePostBlockErrors(dBlock.DatabasePrimaryIndex())
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
		MES.NewMissing(v.String(), bs.DBlockHeadKeyMR.String(), bs.DBlockHeight)
	}
	err := bs.PopCommit(v)
	if err != nil {
		//fmt.Printf("Error - %v\n", err)
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
