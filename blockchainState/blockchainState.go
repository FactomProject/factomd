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
	DBlockHead   interfaces.IHash
	DBlockHeight uint32

	ECBlockHead interfaces.IHash
	FBlockHead  interfaces.IHash
	ABlockHead  interfaces.IHash

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

	if bs.DBlockHead == nil {
		bs.DBlockHead = primitives.NewZeroHash()
	}
	if bs.FBlockHead == nil {
		bs.FBlockHead = primitives.NewZeroHash()
	}
	if bs.ECBlockHead == nil {
		bs.ECBlockHead = primitives.NewZeroHash()
	}
	if bs.ABlockHead == nil {
		bs.ABlockHead = primitives.NewZeroHash()
	}
}

func (bs *BlockchainState) ProcessBlockSet(dBlock interfaces.IDirectoryBlock, aBlock interfaces.IAdminBlock, fBlock interfaces.IFBlock, ecBlock interfaces.IEntryCreditBlock,
	eBlocks []interfaces.IEntryBlock) error {
	bs.Init()
	err := bs.ProcessDBlock(dBlock)
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

	if bs.DBlockHead.String() != dBlock.GetHeader().GetPrevKeyMR().String() {
		fmt.Printf("Invalid DBlock %v previous hash - expected %v, got %v\n", dBlock.DatabasePrimaryIndex().String(), bs.DBlockHead.String(), dBlock.GetHeader().GetPrevKeyMR().String())
	}

	bs.DBlockHead = dBlock.DatabasePrimaryIndex()
	bs.DBlockHeight = dBlock.GetDatabaseHeight()

	dbEntries := dBlock.GetDBEntries()
	for _, v := range dbEntries {
		bs.BlockHeads[v.GetChainID().String()] = v.GetKeyMR()
	}

	return nil
}

func (bs *BlockchainState) ProcessABlock(aBlock interfaces.IAdminBlock) error {
	return nil
}

func (bs *BlockchainState) ProcessFBlock(fBlock interfaces.IFBlock) error {
	bs.Init()

	if bs.FBlockHead.String() != fBlock.GetPrevKeyMR().String() {
		fmt.Printf("Invalid FBlock %v previous hash - expected %v, got %v\n", fBlock.DatabasePrimaryIndex().String(), bs.FBlockHead.String(), fBlock.GetPrevKeyMR().String())
	}
	bs.FBlockHead = fBlock.DatabasePrimaryIndex()

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

	if bs.ECBlockHead.String() != ecBlock.GetHeader().GetPrevHeaderHash().String() {
		fmt.Printf("Invalid ECBlock %v previous hash - expected %v, got %v\n", ecBlock.DatabasePrimaryIndex().String(), bs.ECBlockHead.String(), ecBlock.GetHeader().GetPrevHeaderHash().String())
	}
	bs.ECBlockHead = ecBlock.DatabasePrimaryIndex()

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
		fmt.Printf("Non-committed entry found in an eBlock - %v, %v, %v, %v\n", bs.DBlockHead.String(), bs.DBlockHeight, block.String(), v.String())
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
