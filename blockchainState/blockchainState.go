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

type BlockchainState struct {
	DBlockHead   interfaces.IHash
	DBlockHeight uint32
	BlockHeads   map[string]interfaces.IHash

	ECBalances map[string]uint64
	FBalances  map[string]uint64
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
}

func (bs *BlockchainState) ProcessBlockSet(dBlock interfaces.IDirectoryBlock, fBlock interfaces.IFBlock, ecBlock interfaces.IEntryCreditBlock) error {
	bs.Init()
	err := bs.ProcessDBlock(dBlock)
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
	return nil
}

func (bs *BlockchainState) ProcessDBlock(dBlock interfaces.IDirectoryBlock) error {
	bs.Init()
	bs.DBlockHead = dBlock.DatabasePrimaryIndex()
	bs.DBlockHeight = dBlock.GetDatabaseHeight()

	dbEntries := dBlock.GetDBEntries()
	for _, v := range dbEntries {
		bs.BlockHeads[v.GetChainID().String()] = v.GetKeyMR()
	}

	return nil
}

func (bs *BlockchainState) ProcessFBlock(fBlock interfaces.IFBlock) error {
	bs.Init()
	entries := fBlock.GetTransactions()
	for _, v := range entries {
		ins := v.GetInputs()
		for _, w := range ins {
			if bs.FBalances[w.GetAddress().String()] < w.GetAmount() {
				return fmt.Errorf("Not enough factoids")
			}
			bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] - w.GetAmount()
		}
		outs := v.GetOutputs()
		for _, w := range outs {
			bs.FBalances[w.GetAddress().String()] = bs.FBalances[w.GetAddress().String()] + w.GetAmount()
		}
		ecOut := v.GetECOutputs()
		for _, w := range ecOut {
			bs.ECBalances[w.GetAddress().String()] = bs.ECBalances[w.GetAddress().String()] + w.GetAmount()
		}
	}
	return nil
}

func (bs *BlockchainState) ProcessECBlock(ecBlock interfaces.IEntryCreditBlock) error {
	bs.Init()
	entries := ecBlock.GetEntries()
	for _, v := range entries {
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
			break
		case entryCreditBlock.ECIDChainCommit:
			e := v.(*entryCreditBlock.CommitChain)
			if bs.ECBalances[e.ECPubKey.String()] < uint64(e.Credits) {
				return fmt.Errorf("Not enough ECs - %v:%v<%v", e.ECPubKey.String(), bs.ECBalances[e.ECPubKey.String()], uint64(e.Credits))
			}
			bs.ECBalances[e.ECPubKey.String()] = bs.ECBalances[e.ECPubKey.String()] - uint64(e.Credits)
			break
		default:
			break
		}
	}
	return nil
}

func (bs *BlockchainState) Clone() (*BlockchainState, error) {
	data, err := bs.MarshalBinary()
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

func (bs *BlockchainState) MarshalBinary() ([]byte, error) {
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
